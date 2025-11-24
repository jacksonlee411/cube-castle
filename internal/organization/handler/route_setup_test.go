package handler

import (
	"testing"

	standardobject "cube-castle/internal/standardobject"
	pkglogger "cube-castle/pkg/logger"
	clockpkg "cube-castle/pkg/temporal/clock"
	"github.com/go-chi/chi/v5"
)

func TestSetupRoutes_NoPanic(_ *testing.T) {
	r := chi.NewRouter()
	// DevTools (dev mode)
	dh := NewDevToolsHandler(nil, pkglogger.NewNoopLogger(), true, nil)
	dh.SetupRoutes(r)

	// Operational
	oh := NewOperationalHandler(nil, nil, nil, pkglogger.NewNoopLogger())
	oh.SetupRoutes(r)

	// JobCatalog
	jh := NewJobCatalogHandler(nil, pkglogger.NewNoopLogger())
	jh.SetupRoutes(r)

	// Organization
	ohOrg := NewOrganizationHandler(nil, nil, nil, pkglogger.NewNoopLogger(), nil, nil, nil, standardobject.NewNoopService(), clockpkg.NewSystemClock(), nil)
	ohOrg.SetupRoutes(r)
}
