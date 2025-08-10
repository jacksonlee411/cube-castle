/**
 * 时态管理类型定义单元测试
 */
import type { 
  TemporalMode,
  TemporalQueryParams,
  DateRange,
  TemporalOrganizationUnit,
  TimelineEvent,
  EventType,
  EventStatus,
  TemporalContext
} from '../types/temporal';

describe('Temporal Types', () => {
  describe('TemporalMode', () => {
    it('should accept valid temporal modes', () => {
      const modes: TemporalMode[] = ['current', 'historical', 'planning'];
      expect(modes).toHaveLength(3);
      expect(modes).toContain('current');
      expect(modes).toContain('historical');
      expect(modes).toContain('planning');
    });
  });

  describe('DateRange', () => {
    it('should create valid date range', () => {
      const range: DateRange = {
        start: '2024-01-01T00:00:00Z',
        end: '2024-12-31T23:59:59Z'
      };

      expect(range.start).toBe('2024-01-01T00:00:00Z');
      expect(range.end).toBe('2024-12-31T23:59:59Z');
    });

    it('should handle optional end date', () => {
      const range: DateRange = {
        start: '2024-01-01T00:00:00Z'
      };

      expect(range.start).toBe('2024-01-01T00:00:00Z');
      expect(range.end).toBeUndefined();
    });
  });

  describe('TemporalQueryParams', () => {
    it('should create valid temporal query params', () => {
      const params: TemporalQueryParams = {
        mode: 'historical',
        asOfDate: '2024-06-01T00:00:00Z',
        dateRange: {
          start: '2024-01-01T00:00:00Z',
          end: '2024-12-31T23:59:59Z'
        },
        limit: 100,
        includeInactive: true,
        eventTypes: ['create', 'update', 'delete']
      };

      expect(params.mode).toBe('historical');
      expect(params.asOfDate).toBe('2024-06-01T00:00:00Z');
      expect(params.limit).toBe(100);
      expect(params.includeInactive).toBe(true);
      expect(params.eventTypes).toEqual(['create', 'update', 'delete']);
    });

    it('should work with minimal params', () => {
      const params: TemporalQueryParams = {
        mode: 'current'
      };

      expect(params.mode).toBe('current');
      expect(params.asOfDate).toBeUndefined();
      expect(params.dateRange).toBeUndefined();
    });
  });

  describe('TemporalOrganizationUnit', () => {
    it('should create valid temporal organization unit', () => {
      const org: TemporalOrganizationUnit = {
        code: '1000001',
        name: 'Test Department',
        unit_type: 'DEPARTMENT',
        status: 'ACTIVE',
        level: 1,
        path: '/1000001',
        sort_order: 1,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-06-01T00:00:00Z',
        effective_from: '2024-01-01T00:00:00Z',
        effective_to: '2024-12-31T23:59:59Z',
        is_temporal: true,
        version: 1,
        change_reason: 'Initial creation'
      };

      expect(org.code).toBe('1000001');
      expect(org.name).toBe('Test Department');
      expect(org.is_temporal).toBe(true);
      expect(org.version).toBe(1);
      expect(org.change_reason).toBe('Initial creation');
    });
  });

  describe('TimelineEvent', () => {
    it('should create valid timeline event', () => {
      const event: TimelineEvent = {
        id: 'evt-001',
        organizationCode: '1000001',
        eventType: 'create',
        eventDate: '2024-01-01T00:00:00Z',
        effectiveDate: '2024-01-01T00:00:00Z',
        status: 'completed',
        title: 'Organization Created',
        description: 'New department created',
        triggeredBy: 'user-001',
        createdAt: '2024-01-01T00:00:00Z'
      };

      expect(event.id).toBe('evt-001');
      expect(event.organizationCode).toBe('1000001');
      expect(event.eventType).toBe('create');
      expect(event.status).toBe('completed');
      expect(event.title).toBe('Organization Created');
    });

    it('should handle optional fields', () => {
      const event: TimelineEvent = {
        id: 'evt-002',
        organizationCode: '1000001',
        eventType: 'update',
        eventDate: '2024-06-01T00:00:00Z',
        status: 'pending',
        title: 'Organization Updated',
        createdAt: '2024-06-01T00:00:00Z'
      };

      expect(event.effectiveDate).toBeUndefined();
      expect(event.description).toBeUndefined();
      expect(event.metadata).toBeUndefined();
    });
  });

  describe('EventType', () => {
    it('should accept valid event types', () => {
      const types: EventType[] = [
        'create', 'update', 'delete', 'activate', 'deactivate',
        'restructure', 'merge', 'split', 'transfer', 'rename'
      ];

      expect(types).toHaveLength(10);
      expect(types).toContain('create');
      expect(types).toContain('restructure');
      expect(types).toContain('merge');
    });
  });

  describe('EventStatus', () => {
    it('should accept valid event statuses', () => {
      const statuses: EventStatus[] = [
        'pending', 'approved', 'rejected', 'completed', 'cancelled'
      ];

      expect(statuses).toHaveLength(5);
      expect(statuses).toContain('pending');
      expect(statuses).toContain('approved');
      expect(statuses).toContain('completed');
    });
  });

  describe('TemporalContext', () => {
    it('should create valid temporal context', () => {
      const context: TemporalContext = {
        mode: 'historical',
        asOfDate: '2024-06-01T00:00:00Z',
        effectiveDate: '2024-06-01T00:00:00Z',
        timezone: 'UTC',
        version: 1
      };

      expect(context.mode).toBe('historical');
      expect(context.asOfDate).toBe('2024-06-01T00:00:00Z');
      expect(context.timezone).toBe('UTC');
      expect(context.version).toBe(1);
    });

    it('should work with optional fields', () => {
      const context: TemporalContext = {
        mode: 'current',
        asOfDate: new Date().toISOString(),
        effectiveDate: new Date().toISOString(),
        timezone: 'UTC',
        version: 1
      };

      expect(context.mode).toBe('current');
      expect(context.timezone).toBe('UTC');
      expect(context.version).toBe(1);
    });
  });
});