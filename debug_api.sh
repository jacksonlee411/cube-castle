#!/bin/bash

echo "=== Testing API with different request formats ==="

echo "1. Testing with minimal request:"
curl -s -X POST http://localhost:8080/api/v1/interpret \
  -H "Content-Type: application/json" \
  -d '{"query": "hello"}' \
  -w "\nStatus: %{http_code}\n"

echo -e "\n2. Testing with full request:"
curl -s -X POST http://localhost:8080/api/v1/interpret \
  -H "Content-Type: application/json" \
  -d '{"query": "hello", "user_id": "00000000-0000-0000-0000-000000000000"}' \
  -w "\nStatus: %{http_code}\n"

echo -e "\n3. Testing with malformed JSON:"
curl -s -X POST http://localhost:8080/api/v1/interpret \
  -H "Content-Type: application/json" \
  -d '{"query": "hello", "user_id": "invalid-uuid"}' \
  -w "\nStatus: %{http_code}\n"

echo -e "\n4. Testing health endpoint:"
curl -s http://localhost:8080/health \
  -w "\nStatus: %{http_code}\n" 