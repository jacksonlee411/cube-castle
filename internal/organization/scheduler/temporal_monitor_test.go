package scheduler

import (
	"context"
	"io"
	"regexp"
	"testing"

	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
)

type monitorCounts struct {
	total        int
	current      int
	future       int
	historical   int
	duplicate    int
	missing      int
	timeline     int
	inconsistent int
	orphan       int
}

func expectMetricsQueries(mock sqlmock.Sqlmock, counts monitorCounts) {
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT code\) FROM organization_units WHERE status <> 'DELETED'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.total))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM organization_units WHERE is_current = true AND status <> 'DELETED'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.current))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM organization_units WHERE effective_date > CURRENT_DATE AND status <> 'DELETED'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.future))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM organization_units WHERE end_date IS NOT NULL AND end_date <= CURRENT_DATE AND status <> 'DELETED'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.historical))

	mock.ExpectQuery(`(?s)SELECT COUNT\(\*\) FROM \(\s*SELECT tenant_id, code\s*FROM organization_units\s*WHERE is_current = true AND status <> 'DELETED'\s*GROUP BY tenant_id, code\s*HAVING COUNT\(\*\) > 1\s*\) duplicates`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.duplicate))

	mock.ExpectQuery(`(?s)SELECT COUNT\(\*\) FROM \(\s*SELECT DISTINCT tenant_id, code\s*FROM organization_units\s*WHERE \(tenant_id, code\) NOT IN\s*\(\s*SELECT tenant_id, code\s*FROM organization_units\s*WHERE is_current = true AND status <> 'DELETED'\s*\)\s*AND \(tenant_id, code\) NOT IN\s*\(\s*SELECT tenant_id, code\s*FROM organization_units\s*GROUP BY tenant_id, code\s*HAVING MIN\(CASE WHEN status <> 'DELETED' THEN effective_date ELSE NULL END\) > CURRENT_DATE\s*\)\s*AND EXISTS\s*\(\s*SELECT 1 FROM organization_units u\s*WHERE u.tenant_id = organization_units.tenant_id\s*AND u.code = organization_units.code\s*AND u.status <> 'DELETED'\s*\)\s*\) missing`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.missing))

	mock.ExpectQuery(`(?s)SELECT COUNT\(\*\) FROM \(\s*SELECT DISTINCT o1.tenant_id, o1.code\s*FROM organization_units o1\s*JOIN organization_units o2\s*ON\s*\(\s*o1.tenant_id = o2.tenant_id\s*AND o1.code = o2.code\s*AND o1.record_id != o2.record_id\s*\)\s*WHERE\s*o1.status <> 'DELETED'\s*AND o2.status <> 'DELETED'\s*AND o1.effective_date < COALESCE\(o2.end_date, '9999-12-31'::date\)\s*AND o2.effective_date < COALESCE\(o1.end_date, '9999-12-31'::date\)\s*\) AS timeline_overlaps`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.timeline))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM organization_units\s*WHERE is_current != \(\s*effective_date <= CURRENT_DATE\s*AND \(end_date IS NULL OR end_date > CURRENT_DATE\)\s*\)\s*AND status <> 'DELETED'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.inconsistent))

	mock.ExpectQuery(`(?s)SELECT COUNT\(\*\) FROM organization_units o1\s*WHERE\s*parent_code IS NOT NULL\s*AND o1.status <> 'DELETED'\s*AND NOT EXISTS\s*\(\s*SELECT 1 FROM organization_units o2\s*WHERE o2.tenant_id = o1.tenant_id\s*AND o2.code = o1.parent_code\s*AND o2.is_current = true\s*AND o2.status <> 'DELETED'\s*\)`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(counts.orphan))
}

func TestTemporalMonitorCollectMetricsHealthy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	expectMetricsQueries(mock, monitorCounts{
		total:        10,
		current:      10,
		future:       2,
		historical:   3,
		duplicate:    0,
		missing:      0,
		timeline:     0,
		inconsistent: 0,
		orphan:       0,
	})

	logger := pkglogger.NewLogger(pkglogger.WithWriter(io.Discard))
	monitor := NewTemporalMonitor(db, logger)

	metrics, err := monitor.CollectMetrics(context.Background())
	if err != nil {
		t.Fatalf("collect metrics: %v", err)
	}

	if metrics.TotalOrganizations != 10 {
		t.Fatalf("unexpected total organizations: %d", metrics.TotalOrganizations)
	}
	if metrics.HealthScore != 100 {
		t.Fatalf("expected health score 100, got %.2f", metrics.HealthScore)
	}
	if metrics.AlertLevel != "HEALTHY" {
		t.Fatalf("expected alert level HEALTHY, got %s", metrics.AlertLevel)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}

func TestTemporalMonitorCheckAlertsCritical(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	expectMetricsQueries(mock, monitorCounts{
		total:        10,
		current:      8,
		future:       2,
		historical:   3,
		duplicate:    1,
		missing:      0,
		timeline:     0,
		inconsistent: 0,
		orphan:       0,
	})

	logger := pkglogger.NewLogger(pkglogger.WithWriter(io.Discard))
	monitor := NewTemporalMonitor(db, logger)

	alerts, err := monitor.CheckAlerts(context.Background())
	if err != nil {
		t.Fatalf("check alerts: %v", err)
	}

	if len(alerts) == 0 {
		t.Fatalf("expected alerts to be triggered")
	}

	criticalRegexp := regexp.MustCompile(`CRITICAL`)
	if !criticalRegexp.MatchString(alerts[0]) {
		t.Fatalf("expected critical alert message, got %s", alerts[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}
