package cache

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestCacheEventBusPublishAndClose(t *testing.T) {
	bus := NewCacheEventBus()
	ch := bus.Subscribe()

	event := CacheEvent{EventID: "evt-1", Operation: "CREATE"}
	bus.Publish(event)

	select {
	case received := <-ch:
		if received.EventID != "evt-1" {
			t.Fatalf("expected event evt-1, got %s", received.EventID)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for event")
	}

	bus.Close()
	if _, ok := <-ch; ok {
		t.Fatalf("expected channel to be closed after bus.Close")
	}

	closed := bus.Subscribe()
	if _, ok := <-closed; ok {
		t.Fatalf("expected closed subscription channel")
	}
}

func TestCacheEventToOrganization(t *testing.T) {
	createdAt := time.Now().UTC().Format(time.RFC3339)
	event := CacheEvent{
		TenantID: "tenant",
		Data: map[string]interface{}{
			"code":       "ORG-1",
			"name":       "Org One",
			"unitType":   "TEAM",
			"status":     "ACTIVE",
			"level":      float64(2),
			"sortOrder":  float64(10),
			"parentCode": "ROOT",
			"createdAt":  createdAt,
			"updatedAt":  createdAt,
		},
	}

	org := event.ToOrganization()
	if org.Code != "ORG-1" || org.TenantID != "tenant" {
		t.Fatalf("unexpected organization conversion: %#v", org)
	}
	if org.SortOrder != 10 {
		t.Fatalf("expected sort order 10")
	}

	direct := CacheEvent{Data: Organization{Code: "DIRECT", TenantID: "tenant"}}
	if got := direct.ToOrganization(); got.Code != "DIRECT" {
		t.Fatalf("expected existing organization to be returned as-is")
	}
}

func TestSmartCacheUpdaterOperations(t *testing.T) {
	updater := NewSmartCacheUpdater(log.New(io.Discard, "", 0))
	existing := []Organization{
		{Code: "A", Name: "Alpha", SortOrder: 1},
		{Code: "B", Name: "Beta", SortOrder: 2},
	}
	params := QueryParams{}

	// create new matching organization
	created, ok := updater.UpdateListCache(existing, &Organization{Code: "C", Name: "Gamma", SortOrder: 0}, "CREATE", params)
	if !ok || len(created) != 3 || created[0].Code != "C" {
		t.Fatalf("expected new organization to be inserted and sorted")
	}

	// update existing organization and keep order
	updated, ok := updater.UpdateListCache(existing, &Organization{Code: "B", Name: "Beta v2", SortOrder: 1}, "UPDATE", params)
	if !ok || updated[1].Name != "Beta v2" {
		t.Fatalf("expected organization B to be updated in-place")
	}

	// delete removes organization
	deleted, ok := updater.UpdateListCache(existing, &Organization{Code: "A"}, "DELETE", params)
	if !ok || len(deleted) != 1 || deleted[0].Code != "B" {
		t.Fatalf("expected organization A to be removed")
	}
}

func TestSmartCacheUpdaterSearchFiltering(t *testing.T) {
	updater := NewSmartCacheUpdater(log.New(io.Discard, "", 0))
	existing := []Organization{
		{Code: "A", Name: "Alpha"},
	}
	params := QueryParams{SearchText: "Alpha"}

	// matches search text
	_, ok := updater.UpdateListCache(existing, &Organization{Code: "B", Name: "Alphabet"}, "CREATE", params)
	if !ok {
		t.Fatalf("expected create to match search text")
	}

	// update that no longer matches should remove
	list, ok := updater.UpdateListCache(existing, &Organization{Code: "A", Name: "Omega"}, "UPDATE", params)
	if !ok || len(list) != 0 {
		t.Fatalf("expected updated organization removed when not matching search")
	}
}

func TestL1CacheBasicOperations(t *testing.T) {
	cache := NewL1Cache(2, 20*time.Millisecond)
	entry := CacheEntry{Key: "k1", Data: json.RawMessage(`"value"`)}

	cache.Set("k1", entry)
	if got, ok := cache.Get("k1"); !ok || string(got.Data) != `"value"` {
		t.Fatalf("expected to retrieve cached entry")
	}

	cache.Delete("k1")
	if _, ok := cache.Get("k1"); ok {
		t.Fatalf("expected entry to be deleted")
	}

	cache.Set("k2", entry)
	time.Sleep(30 * time.Millisecond)
	if _, ok := cache.Get("k2"); ok {
		t.Fatalf("expected entry to expire via TTL")
	}
}

func TestL1CacheStatsAndEviction(t *testing.T) {
	cache := NewL1Cache(1, time.Minute)
	entry1 := CacheEntry{Key: "a", Data: json.RawMessage(`"a"`)}
	entry2 := CacheEntry{Key: "b", Data: json.RawMessage(`"b"`)}

	cache.Set("a", entry1)
	cache.Set("b", entry2) // should evict "a"

	if _, ok := cache.Get("a"); ok {
		t.Fatalf("expected oldest entry to be evicted")
	}
	if _, ok := cache.Get("b"); !ok {
		t.Fatalf("expected newest entry to remain")
	}

	stats := cache.GetStats()
	if stats.Size != 1 || stats.HitCount == 0 || stats.HitRate <= 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}

	cache.Clear()
	if stats := cache.GetStats(); stats.Size != 0 {
		t.Fatalf("expected cache to be cleared")
	}
}

func TestConsistencyChecker(t *testing.T) {
	l1 := NewL1Cache(4, time.Minute)
	entry := CacheEntry{Key: "org", Data: json.RawMessage(`{"code":"A"}`)}
	l1.Set("org", entry)

	checker := NewConsistencyChecker(l1, nil, log.New(io.Discard, "", 0))

	report := checker.CheckConsistency(context.Background(), []string{"org"})
	if len(report.Inconsistencies) != 1 {
		t.Fatalf("expected inconsistency due to missing L2 data")
	}

	if _, err := checker.getFromL2(context.Background(), "org"); err == nil {
		t.Fatalf("expected placeholder getFromL2 to error")
	}

	if checker.hashContent(make(chan int)) != "error" {
		t.Fatalf("expected hashContent to return error for unsupported data")
	}
}

func TestCacheKeyManager(t *testing.T) {
	km := &CacheKeyManager{namespace: "org_v1"}
	key := km.GenerateKey("organizations", "tenant", "params")
	if len(key) == 0 || key[:6] != "cache:" {
		t.Fatalf("unexpected cache key: %s", key)
	}

	patterns := km.GetPatternForTags([]string{"tenant:one", "type:list"})
	if len(patterns) != 2 || patterns[0] != "org_v1:*tenant:one*" {
		t.Fatalf("unexpected patterns: %#v", patterns)
	}
}

type mockL3Query struct {
	listResult   []Organization
	singleResult map[string]*Organization
	statsResult  *OrganizationStats
	listCalls    int
	singleCalls  int
	statsCalls   int
}

func (m *mockL3Query) GetOrganizations(ctx context.Context, tenantID uuid.UUID, params QueryParams) ([]Organization, error) {
	m.listCalls++
	result := make([]Organization, len(m.listResult))
	copy(result, m.listResult)
	return result, nil
}

func (m *mockL3Query) GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*Organization, error) {
	m.singleCalls++
	if org, ok := m.singleResult[code]; ok && org != nil {
		clone := *org
		return &clone, nil
	}
	return nil, nil
}

func (m *mockL3Query) GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*OrganizationStats, error) {
	m.statsCalls++
	if m.statsResult == nil {
		return nil, nil
	}
	clone := *m.statsResult
	return &clone, nil
}

func TestUnifiedCacheManagerQueryFlows(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger := log.New(io.Discard, "", 0)

	tenantID := uuid.New()
	l3 := &mockL3Query{
		listResult: []Organization{{Code: "A", TenantID: tenantID.String(), Name: "Alpha", SortOrder: 1}},
		singleResult: map[string]*Organization{
			"A": {Code: "A", TenantID: tenantID.String(), Name: "Alpha"},
		},
		statsResult: &OrganizationStats{TotalCount: 1},
	}

	cfg := &CacheConfig{
		L1TTL:           time.Minute,
		L2TTL:           time.Minute,
		L1MaxSize:       100,
		WriteThrough:    true,
		ConsistencyMode: "STRONG",
		Namespace:       "org_v1_test",
	}

	ucm := NewUnifiedCacheManager(client, l3, cfg, logger)
	defer ucm.Close()

	ctx := context.Background()
	params := QueryParams{First: 20, Offset: 0}

	orgs, err := ucm.GetOrganizations(ctx, tenantID, params)
	if err != nil || len(orgs) != 1 {
		t.Fatalf("expected organizations from L3: %v, err=%v", orgs, err)
	}
	if l3.listCalls != 1 {
		t.Fatalf("expected one L3 list call, got %d", l3.listCalls)
	}

	// L1 cache hit should not trigger additional L3 calls
	if _, err := ucm.GetOrganizations(ctx, tenantID, params); err != nil {
		t.Fatalf("unexpected error on L1 hit: %v", err)
	}
	if l3.listCalls != 1 {
		t.Fatalf("expected L1 hit to avoid L3 call")
	}

	// Clear L1 to force L2 hit
	ucm.l1Cache.Clear()
	if _, err := ucm.GetOrganizations(ctx, tenantID, params); err != nil {
		t.Fatalf("unexpected error on L2 hit: %v", err)
	}
	if l3.listCalls != 1 {
		t.Fatalf("expected L2 hit to avoid L3 call")
	}

	// Flush caches to force L3 again
	mr.FlushAll()
	ucm.l1Cache.Clear()
	if _, err := ucm.GetOrganizations(ctx, tenantID, params); err != nil {
		t.Fatalf("unexpected error on fallback L3 call: %v", err)
	}
	if l3.listCalls != 2 {
		t.Fatalf("expected L3 to be queried again, calls=%d", l3.listCalls)
	}

	// Single organization fetch
	org, err := ucm.GetOrganization(ctx, tenantID, "A")
	if err != nil || org == nil {
		t.Fatalf("expected organization fetch success: %v, err=%v", org, err)
	}
	if l3.singleCalls != 1 {
		t.Fatalf("expected single fetch to hit L3 once")
	}

	if _, err := ucm.GetOrganization(ctx, tenantID, "A"); err != nil {
		t.Fatalf("unexpected error on cached single fetch: %v", err)
	}
	if l3.singleCalls != 1 {
		t.Fatalf("expected cached single fetch to avoid L3")
	}

	// Missing organization should return nil
	if org, err := ucm.GetOrganization(ctx, tenantID, "missing"); err != nil || org != nil {
		t.Fatalf("expected nil result for missing organization, got %v, err=%v", org, err)
	}

	stats, err := ucm.GetOrganizationStats(ctx, tenantID)
	if err != nil || stats == nil || stats.TotalCount != 1 {
		t.Fatalf("expected stats from L3, got %v, err=%v", stats, err)
	}
	if l3.statsCalls != 1 {
		t.Fatalf("expected stats L3 call once")
	}

	// Cached stats path
	if _, err := ucm.GetOrganizationStats(ctx, tenantID); err != nil {
		t.Fatalf("unexpected error on cached stats: %v", err)
	}
	if l3.statsCalls != 1 {
		t.Fatalf("expected cached stats to avoid L3")
	}
}

func TestUnifiedCacheManagerCDCAndInvalidation(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger := log.New(io.Discard, "", 0)
	tenantID := uuid.New()

	cfg := &CacheConfig{
		L1TTL:           time.Minute,
		L2TTL:           time.Minute,
		L1MaxSize:       100,
		WriteThrough:    true,
		ConsistencyMode: "STRONG",
		Namespace:       "org_v1_test",
	}

	ucm := NewUnifiedCacheManager(client, &mockL3Query{}, cfg, logger)
	defer ucm.Close()

	ctx := context.Background()
	org := Organization{Code: "ORG-1", TenantID: tenantID.String(), Name: "Alpha"}
	event := CacheEvent{Operation: "CREATE", TenantID: tenantID.String(), Data: org, Timestamp: time.Now().Unix()}
	keyMgr := &CacheKeyManager{namespace: cfg.Namespace}
	orgKey := keyMgr.GenerateKey("organization", org.TenantID, org.Code)

	// Seed stale entries to verify invalidation occurs
	ucm.l1Cache.Set(orgKey, CacheEntry{Key: orgKey})
	client.Set(ctx, orgKey, "stale", 0)

	listKey := keyMgr.GenerateKey("organizations", org.TenantID, "50", "0", "")
	ucm.l1Cache.Set(listKey, CacheEntry{Key: listKey})
	client.Set(ctx, listKey, "stale", 0)

	if err := ucm.HandleCDCEvent(ctx, event); err != nil {
		t.Fatalf("unexpected error handling create event: %v", err)
	}

	if _, ok := ucm.l1Cache.Get(orgKey); ok {
		t.Fatalf("expected organization cache invalidated after event")
	}
	if mr.Exists(orgKey) {
		t.Fatalf("expected organization removed from redis after event")
	}

	if _, ok := ucm.l1Cache.Get(listKey); ok {
		t.Fatalf("expected list cache invalidated after event")
	}
	if mr.Exists(listKey) {
		t.Fatalf("expected list cache removed from redis after event")
	}

	event.Operation = "UPDATE"
	if err := ucm.HandleCDCEvent(ctx, event); err != nil {
		t.Fatalf("unexpected error handling update event: %v", err)
	}

	event.Operation = "DELETE"
	if err := ucm.HandleCDCEvent(ctx, event); err != nil {
		t.Fatalf("unexpected error handling delete event: %v", err)
	}
	if _, ok := ucm.l1Cache.Get(orgKey); ok {
		t.Fatalf("expected organization removed from L1 after delete")
	}

	// Traditional invalidation path (write-through disabled)
	cfgNoWriteThrough := &CacheConfig{
		L1TTL:           time.Minute,
		L2TTL:           time.Minute,
		L1MaxSize:       100,
		WriteThrough:    false,
		ConsistencyMode: "EVENTUAL",
		Namespace:       "org_v1_test2",
	}

	ucm2 := NewUnifiedCacheManager(client, &mockL3Query{}, cfgNoWriteThrough, logger)
	defer ucm2.Close()

	keyMgr2 := &CacheKeyManager{namespace: cfgNoWriteThrough.Namespace}
	orgKey2 := keyMgr2.GenerateKey("organization", org.TenantID, org.Code)
	listKey2 := cfgNoWriteThrough.Namespace + ":organizations:" + org.TenantID + ":page1"
	statsKey2 := keyMgr2.GenerateKey("stats", org.TenantID)

	client.Set(ctx, orgKey2, "value", 0)
	client.Set(ctx, listKey2, "value", 0)
	client.Set(ctx, statsKey2, "value", 0)
	ucm2.l1Cache.Set(orgKey2, CacheEntry{Key: orgKey2})

	traditionalEvent := CacheEvent{Operation: "UPDATE", TenantID: org.TenantID, Data: org, Timestamp: time.Now().Unix()}
	if err := ucm2.HandleCDCEvent(ctx, traditionalEvent); err != nil {
		t.Fatalf("unexpected error handling traditional invalidation: %v", err)
	}

	if _, ok := ucm2.l1Cache.Get(orgKey2); ok {
		t.Fatalf("expected L1 cache cleared for organization")
	}
}

func TestUnifiedCacheManagerCacheManagement(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger := log.New(io.Discard, "", 0)
	tenantID := uuid.New()

	cfg := &CacheConfig{
		L1TTL:           time.Minute,
		L2TTL:           time.Minute,
		L1MaxSize:       100,
		WriteThrough:    true,
		ConsistencyMode: "STRONG",
		Namespace:       "org_v1_test",
	}

	ucm := NewUnifiedCacheManager(client, &mockL3Query{}, cfg, logger)
	defer ucm.Close()

	ctx := context.Background()

	stats := ucm.GetCacheStats(ctx)
	if !stats.L2Connected {
		t.Fatalf("expected L2 to be reported as connected")
	}

	keyMgr := &CacheKeyManager{namespace: cfg.Namespace}
	orgKey := keyMgr.GenerateKey("organization", tenantID.String(), "A")
	entry := CacheEntry{Key: orgKey, Data: json.RawMessage(`{"code":"A"}`)}
	ucm.l1Cache.Set(orgKey, entry)
	client.Set(ctx, orgKey, `{"data":"value"}`, 0)

	if err := ucm.RefreshCache(ctx, tenantID, "organization", "A"); err != nil {
		t.Fatalf("unexpected error refreshing organization cache: %v", err)
	}
	if _, ok := ucm.l1Cache.Get(orgKey); ok {
		t.Fatalf("expected organization entry removed from L1 after refresh")
	}
	if mr.Exists(orgKey) {
		t.Fatalf("expected organization entry removed from redis after refresh")
	}

	patternKey := keyMgr.GenerateKey("organizations", tenantID.String(), "*")
	client.Set(ctx, patternKey, "value", 0)
	if err := ucm.RefreshCache(ctx, tenantID, "organizations", ""); err != nil {
		t.Fatalf("unexpected error refreshing list cache: %v", err)
	}

	if err := ucm.Close(); err != nil {
		t.Fatalf("expected close to succeed: %v", err)
	}
}
