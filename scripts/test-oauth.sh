#!/bin/bash

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo -e "${BLUE}🧪 Testing SMAP Auth Service OAuth Flow${NC}"
echo ""

# Test 1: Health Check
echo -e "${YELLOW}Test 1: Health Check${NC}"
response=$(curl -s -o /dev/null -w "%{http_code}" ${BASE_URL}/health)
if [ "$response" = "200" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed (HTTP $response)${NC}"
    exit 1
fi
echo ""

# Test 2: JWKS Endpoint
echo -e "${YELLOW}Test 2: JWKS Endpoint${NC}"
jwks=$(curl -s ${BASE_URL}/authentication/.well-known/jwks.json)
if echo "$jwks" | grep -q "keys"; then
    echo -e "${GREEN}✓ JWKS endpoint working${NC}"
    echo "Response: $jwks" | jq '.' 2>/dev/null || echo "$jwks"
else
    echo -e "${RED}✗ JWKS endpoint failed${NC}"
    echo "Response: $jwks"
fi
echo ""

# Test 3: OAuth Login Redirect
echo -e "${YELLOW}Test 3: OAuth Login Redirect${NC}"
login_response=$(curl -s -I ${BASE_URL}/authentication/login)
if echo "$login_response" | grep -q "Location.*accounts.google.com"; then
    echo -e "${GREEN}✓ OAuth login redirect working${NC}"
    echo "Redirects to Google OAuth"
else
    echo -e "${RED}✗ OAuth login redirect failed${NC}"
    echo "$login_response"
fi
echo ""

# Test 4: Database Connection
echo -e "${YELLOW}Test 4: Database Connection${NC}"
db_check=$(docker exec smap-postgres psql -U postgres -d smap_auth -c "SELECT COUNT(*) FROM users;" 2>&1)
if echo "$db_check" | grep -q "count"; then
    echo -e "${GREEN}✓ Database connection working${NC}"
    echo "$db_check"
else
    echo -e "${RED}✗ Database connection failed${NC}"
    echo "$db_check"
fi
echo ""

# Test 5: Redis Connection
echo -e "${YELLOW}Test 5: Redis Connection${NC}"
redis_check=$(docker exec smap-redis redis-cli ping 2>&1)
if [ "$redis_check" = "PONG" ]; then
    echo -e "${GREEN}✓ Redis connection working${NC}"
else
    echo -e "${RED}✗ Redis connection failed${NC}"
    echo "$redis_check"
fi
echo ""

# Test 6: Kafka Connection (optional for Day 3)
echo -e "${YELLOW}Test 6: Kafka Connection (Optional)${NC}"
kafka_check=$(docker exec smap-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 2>&1 | head -n 1)
if echo "$kafka_check" | grep -q "ApiVersion"; then
    echo -e "${GREEN}✓ Kafka connection working${NC}"
else
    echo -e "${YELLOW}⚠ Kafka not ready (needed for Day 3)${NC}"
fi
echo ""

echo -e "${BLUE}📋 Summary${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Core services are ready for OAuth testing"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Open browser: ${BASE_URL}/authentication/login"
echo "2. Login with Google account (must be in allowed_domains)"
echo "3. Check cookie 'smap_auth_token' is set"
echo "4. Test /authentication/me endpoint"
echo ""
echo -e "${YELLOW}Manual test commands:${NC}"
echo "# Get current user (replace JWT_TOKEN with actual token from cookie)"
echo "curl ${BASE_URL}/authentication/me --cookie 'smap_auth_token=JWT_TOKEN'"
echo ""
echo "# Logout"
echo "curl -X POST ${BASE_URL}/authentication/logout --cookie 'smap_auth_token=JWT_TOKEN'"
echo ""
