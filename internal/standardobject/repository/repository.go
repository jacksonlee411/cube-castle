package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/standardobject"
	sqlc "cube-castle/internal/standardobject/repository/sqlc"
	"cube-castle/pkg/temporal/constraints"
	"github.com/google/uuid"
)

var objectConstraints = map[standardobject.ObjectType]constraints.ConstraintType{
	standardobject.ObjectTypeOrganizationUnit: constraints.ConstraintTypeTC1,
}

// Repository wraps sqlc generated queries and exposes higher level helpers for the standard object port.
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository builds a Repository backed by the provided DB/transaction handle.
func NewRepository(db sqlc.DBTX) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

// Upsert persists the kernel, versions and links of an aggregate.
func (r *Repository) Upsert(ctx context.Context, aggregate standardobject.ObjectAggregate) error {
	kernel := aggregate.Kernel
	kernelRec, err := r.queries.UpsertStandardObject(ctx, sqlc.UpsertStandardObjectParams{
		ObjectType:         string(kernel.ObjectType),
		Code:               kernel.Code,
		TenantCode:         kernel.TenantCode,
		DisplayName:        kernel.DisplayName,
		Status:             string(kernel.Status),
		Labels:             mustJSON(kernel.Labels),
		SchemaVersion:      kernel.SchemaVersion,
		DataClassification: kernel.DataClassification,
		RetentionPolicy:    kernel.RetentionPolicy,
		CreatedBy:          parseUUID(kernel.CreatedBy),
	})
	if err != nil {
		return fmt.Errorf("upsert kernel: %w", err)
	}

	if aggregate.Version.VersionCode != "" {
		version := aggregate.Version
		validTo := version.EndDate
		txTo := version.TransactionTo
		existing, err := r.queries.ListStandardObjectVersionsForObject(ctx, kernelRec.ID)
		if err != nil {
			return fmt.Errorf("load existing versions: %w", err)
		}
		constraintType := constraintForObjectType(kernel.ObjectType)
		validation, err := constraints.Validate(constraintType, toRanges(existing), constraints.RangeWindow{
			From: version.EffectiveDate,
			To:   version.EndDate,
		})
		if err != nil {
			return err
		}
		if err := r.closeOpenWindow(ctx, existing, defaultTime(version.TransactionFrom), constraintType, validation, version.EffectiveDate); err != nil {
			return err
		}
		if _, err := r.queries.InsertStandardObjectVersion(ctx, sqlc.InsertStandardObjectVersionParams{
			ObjectID:         kernelRec.ID,
			VersionCode:      version.VersionCode,
			EffectiveDate:    version.EffectiveDate,
			EndDate:          sql.NullTime{Time: derefTime(validTo), Valid: validTo != nil},
			ValidityRange:    buildRangeLiteral(version.EffectiveDate, version.EndDate),
			TransactionRange: buildRangeLiteral(defaultTime(version.TransactionFrom), txTo),
			IsCurrent:        version.IsCurrent,
			Payload:          mustJSON(version.Payload),
			Audit:            mustJSON(version.AuditTrail),
			Checksum:         sql.NullString{String: version.Checksum, Valid: version.Checksum != ""},
		}); err != nil {
			return fmt.Errorf("insert version: %w", err)
		}
	}

	for _, link := range aggregate.Links {
		targetType := link.TargetType
		if targetType == "" {
			targetType = kernel.ObjectType
		}
		targetID, err := r.lookupObjectID(ctx, kernel.TenantCode, targetType, link.TargetCode)
		if err != nil {
			return fmt.Errorf("lookup link target %s: %w", link.TargetCode, err)
		}
		createdBy := link.CreatedBy
		if createdBy == "" {
			createdBy = kernel.CreatedBy
		}
		if _, err := r.queries.UpsertStandardObjectLink(ctx, sqlc.UpsertStandardObjectLinkParams{
			LinkType:         link.LinkType,
			SourceObjectID:   kernelRec.ID,
			TargetObjectID:   targetID,
			TenantCode:       kernel.TenantCode,
			ValidityRange:    buildRangeLiteral(defaultTime(link.ValidFrom), link.ValidTo),
			TransactionRange: buildRangeLiteral(defaultTime(link.TransactionFrom), link.TransactionTo),
			Attributes:       mustJSON(link.Attributes),
			CreatedBy:        parseUUID(createdBy),
		}); err != nil {
			return fmt.Errorf("upsert link (%s): %w", link.LinkType, err)
		}
	}

	return nil
}

// Get retrieves an aggregate using the provided key.
func (r *Repository) Get(ctx context.Context, key standardobject.ObjectKey) (standardobject.ObjectAggregate, error) {
	asOf := time.Now().UTC()
	obj, err := r.queries.GetStandardObjectKernel(ctx, sqlc.GetStandardObjectKernelParams{
		TenantCode: key.TenantCode,
		ObjectType: string(key.ObjectType),
		Code:       key.Code,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return standardobject.ObjectAggregate{}, standardobject.ErrNotFound
		}
		return standardobject.ObjectAggregate{}, fmt.Errorf("load kernel: %w", err)
	}

	version, err := r.queries.GetVersionAsOf(ctx, sqlc.GetVersionAsOfParams{
		TenantCode: key.TenantCode,
		ObjectType: string(key.ObjectType),
		Code:       key.Code,
		Column4:    asOf,
		Column5:    asOf,
	})
	if err != nil {
		return standardobject.ObjectAggregate{}, fmt.Errorf("load version: %w", err)
	}

	rawLinks, err := r.queries.ListLinksForSource(ctx, obj.ID)
	if err != nil {
		return standardobject.ObjectAggregate{}, fmt.Errorf("load links: %w", err)
	}

	return standardobject.ObjectAggregate{
		Kernel: standardobject.ObjectKernel{
			ID:                 obj.ID.String(),
			ObjectType:         key.ObjectType,
			Code:               obj.Code,
			DisplayName:        obj.DisplayName,
			TenantCode:         obj.TenantCode,
			Status:             standardobject.LifecycleStatus(obj.Status),
			Labels:             mustLabelMap(obj.Labels),
			SchemaVersion:      obj.SchemaVersion,
			DataClassification: obj.DataClassification,
			RetentionPolicy:    obj.RetentionPolicy,
			CreatedBy:          obj.CreatedBy.UUID.String(),
			CreatedAt:          obj.CreatedAt,
			UpdatedAt:          obj.UpdatedAt,
		},
		Version: convertVersion(version),
		Links:   convertLinks(rawLinks),
	}, nil
}

func convertVersion(row sqlc.StandardObjectVersion) standardobject.TemporalVersion {
	_, validTo := parseRangeLiteral(row.ValidityRange)
	txFrom, txTo := parseRangeLiteral(row.TransactionRange)
	return standardobject.TemporalVersion{
		VersionID:       row.ID.String(),
		VersionCode:     row.VersionCode,
		EffectiveDate:   row.EffectiveDate,
		EndDate:         validTo,
		IsCurrent:       row.IsCurrent,
		Payload:         mustMap(row.Payload),
		AuditTrail:      mustMap(row.Audit),
		Checksum:        derefString(row.Checksum),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		TransactionFrom: txFrom,
		TransactionTo:   txTo,
	}
}

func toRanges(rows []sqlc.StandardObjectVersion) []constraints.RangeWindow {
	out := make([]constraints.RangeWindow, 0, len(rows))
	for _, row := range rows {
		_, end := parseRangeLiteral(row.ValidityRange)
		out = append(out, constraints.RangeWindow{
			From: row.EffectiveDate,
			To:   end,
		})
	}
	return out
}

func constraintForObjectType(t standardobject.ObjectType) constraints.ConstraintType {
	if value, ok := objectConstraints[t]; ok {
		return value
	}
	return constraints.ConstraintTypeTC2
}

func convertLinks(rows []sqlc.ListLinksForSourceRow) []standardobject.Link {
	result := make([]standardobject.Link, 0, len(rows))
	for _, row := range rows {
		validFrom, validTo := parseRangeLiteral(row.ValidityRange)
		txFrom, txTo := parseRangeLiteral(row.TransactionRange)
		result = append(result, standardobject.Link{
			LinkID:          row.ID.String(),
			LinkType:        row.LinkType,
			SourceCode:      row.SourceCode,
			TargetCode:      row.TargetCode,
			Attributes:      mustMap(row.Attributes),
			TenantCode:      row.TenantCode,
			ValidFrom:       validFrom,
			ValidTo:         validTo,
			TransactionFrom: txFrom,
			TransactionTo:   txTo,
			CreatedBy:       row.CreatedBy.UUID.String(),
		})
	}
	return result
}

func (r *Repository) closeOpenWindow(ctx context.Context, versions []sqlc.StandardObjectVersion, transactionClose time.Time, constraintType constraints.ConstraintType, validation constraints.ValidationResult, newEffectiveDate time.Time) error {
	if constraintType == constraints.ConstraintTypeTC3 {
		return nil
	}
	var openRow *sqlc.StandardObjectVersion
	for i := range versions {
		_, txEnd := parseRangeLiteral(versions[i].TransactionRange)
		if txEnd == nil {
			openRow = &versions[i]
			break
		}
	}
	if openRow == nil {
		return nil
	}
	args := sqlc.CloseVersionRangesParams{
		ID:        openRow.ID,
		Tstzrange: transactionClose,
		EndDate:   sql.NullTime{},
	}
	if validation.RequireContiguousValidity {
		args.Tstzrange_2 = newEffectiveDate
		args.EndDate = sql.NullTime{Time: truncateToDate(newEffectiveDate), Valid: true}
	}
	return r.queries.CloseVersionRanges(ctx, args)
}

func (r *Repository) lookupObjectID(ctx context.Context, tenant string, objectType standardobject.ObjectType, code string) (uuid.UUID, error) {
	rec, err := r.queries.GetStandardObjectKernel(ctx, sqlc.GetStandardObjectKernelParams{
		TenantCode: tenant,
		ObjectType: string(objectType),
		Code:       code,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return rec.ID, nil
}

func parseUUID(value string) uuid.NullUUID {
	if value == "" {
		return uuid.NullUUID{}
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

func mustJSON(v any) json.RawMessage {
	if v == nil {
		return json.RawMessage([]byte("{}"))
	}
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(b)
}

func mustMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func mustLabelMap(raw json.RawMessage) map[string]string {
	if len(raw) == 0 {
		return map[string]string{}
	}
	var out map[string]string
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]string{}
	}
	return out
}

func buildRangeLiteral(from time.Time, to *time.Time) string {
	lower := from.UTC().Format(time.RFC3339Nano)
	upper := "infinity"
	if to != nil {
		upper = to.UTC().Format(time.RFC3339Nano)
	}
	return fmt.Sprintf("[\"%s\",\"%s\")", lower, upper)
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func nullableTime(val sql.NullTime) *time.Time {
	if !val.Valid {
		return nil
	}
	return &val.Time
}

func derefString(val sql.NullString) string {
	if !val.Valid {
		return ""
	}
	return val.String
}

func parseRangeLiteral(lit string) (time.Time, *time.Time) {
	if lit == "" {
		return time.Time{}, nil
	}
	trim := strings.TrimSuffix(strings.TrimPrefix(lit, "["), ")")
	parts := strings.Split(trim, ",")
	start := parseRangeEdge(parts[0])
	var end *time.Time
	if len(parts) > 1 {
		candidate := parseRangeEdge(parts[1])
		if !candidate.IsZero() {
			c := candidate
			end = &c
		}
	}
	return start, end
}

func parseRangeEdge(value string) time.Time {
	value = strings.Trim(value, "\"")
	if value == "" || value == "infinity" {
		return time.Time{}
	}
	ts, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return ts
}

func defaultTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}

func truncateToDate(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
