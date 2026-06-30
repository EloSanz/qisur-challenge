#!/bin/bash
set -e

echo "=== Health Check ==="
curl -s http://localhost:8086/health | jq || echo "Health check failed"
echo ""

echo "=== Login ==="
LOGIN_RES=$(curl -s -X POST http://localhost:8086/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "123"}')
echo $LOGIN_RES | jq
TOKEN=$(echo $LOGIN_RES | jq -r .token)
echo ""

echo "=== Create Category ==="
CAT_RES=$(curl -s -X POST http://localhost:8086/api/categories \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name": "Electronics", "description": "Gadgets"}')
echo $CAT_RES | jq
CAT_ID=$(echo $CAT_RES | jq -r .data.id)
echo ""

echo "=== Create Product ==="
PROD_RES=$(curl -s -X POST http://localhost:8086/api/products \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d "{\"name\": \"Smartphone X\", \"price\": 999.99, \"stock\": 50, \"category_ids\": [\"$CAT_ID\"]}")
echo $PROD_RES | jq
PROD_ID=$(echo $PROD_RES | jq -r .data.id)
echo ""

echo "=== Get Product ==="
curl -s http://localhost:8086/api/products/$PROD_ID | jq
echo ""

echo "=== Update Product ==="
curl -s -X PUT http://localhost:8086/api/products/$PROD_ID \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d "{\"price\": 899.99}" | jq
echo ""

echo "=== Get Product History ==="
curl -s http://localhost:8086/api/products/$PROD_ID/history | jq
echo ""

echo "=== Search ==="
curl -s "http://localhost:8086/api/search?type=product&q=Smartphone" | jq
echo ""
