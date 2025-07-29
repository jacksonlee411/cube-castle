// tests/setup/env.setup.js
// Environment setup for tests
process.env.NODE_ENV = 'test';
process.env.NEXT_PUBLIC_API_URL = 'http://localhost:3000/api';
process.env.GRAPHQL_ENDPOINT = 'http://localhost:8080/graphql';
process.env.TEST_TOKEN = 'test-token-12345';