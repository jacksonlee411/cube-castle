# GraphQL Directory - DEPRECATED

⚠️ **This directory contains legacy GraphQL code that is not currently used in production.**

## Current Architecture Status

- **Production API**: REST API only (`/api/v1/*`)
- **GraphQL Endpoint**: ❌ Not implemented
- **Frontend Data**: Uses SWR + REST API proxy

## Directory Contents

- `schema/`: GraphQL schema definitions (unused)
- `resolvers/`: GraphQL resolvers (unused)
- Purpose: Originally planned for future GraphQL implementation

## Architecture Decision

The current REST + SWR architecture provides:
- ✅ Simple CRUD operations
- ✅ Excellent caching with SWR
- ✅ Low complexity and high performance
- ✅ Standard HTTP patterns

GraphQL was evaluated but deemed unnecessary for current HR management use cases.

## Future Considerations

Consider GraphQL only if these requirements emerge:
- Complex nested relationship queries
- Mobile app with different field requirements
- Real-time subscriptions
- Multi-client field optimization needs

## Recommendation

Keep REST architecture, focus on business value delivery.