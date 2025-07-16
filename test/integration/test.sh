#!/bin/bash
set -e

echo "=== Integration Test for kubectl-eex ==="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to check if value matches expected
check_value() {
    local env_name=$1
    local expected=$2
    local actual=$3
    
    if [ "$actual" = "$expected" ]; then
        echo -e "${GREEN}✓ $env_name: $actual${NC}"
        return 0
    else
        echo -e "${RED}✗ $env_name: expected '$expected', got '$actual'${NC}"
        return 1
    fi
}

# Test 1: Extract environment variables from deployment
echo ""
echo "Test 1: Extracting environment variables from deployment..."
OUTPUT=$(./kubectl-eex deployment/test-app -n test-keex)

# Create a temporary file to source the environment variables
TMPFILE=$(mktemp)
echo "$OUTPUT" > "$TMPFILE"

# Source the file to get all environment variables
set -a
source "$TMPFILE"
set +a

# Clean up
rm -f "$TMPFILE"

# Test direct environment variable
echo ""
echo "Testing direct environment variables:"
check_value "DIRECT_ENV" "direct-value" "$DIRECT_ENV"

# Test individual secret references
echo ""
echo "Testing individual secret references:"
check_value "DB_HOST" "localhost" "$DB_HOST"
check_value "DB_PORT" "5432" "$DB_PORT"

# Test individual configmap references
echo ""
echo "Testing individual configmap references:"
check_value "APPLICATION_ENV" "production" "$APPLICATION_ENV"
check_value "APPLICATION_PORT" "8080" "$APPLICATION_PORT"

# Test envFrom secret (recursive functionality)
echo ""
echo "Testing envFrom secret (entire secret imported):"
check_value "DATABASE_HOST" "localhost" "$DATABASE_HOST"
check_value "DATABASE_PORT" "5432" "$DATABASE_PORT"
check_value "DATABASE_USER" "admin" "$DATABASE_USER"
check_value "DATABASE_PASS" "secret123" "$DATABASE_PASS"
check_value "API_KEY" "abcdefghijk" "$API_KEY"

# Test envFrom configmap (recursive functionality)
echo ""
echo "Testing envFrom configmap (entire configmap imported):"
check_value "APP_ENV" "production" "$APP_ENV"
check_value "APP_DEBUG" "false" "$APP_DEBUG"
check_value "APP_LOG_LEVEL" "info" "$APP_LOG_LEVEL"
check_value "APP_PORT" "8080" "$APP_PORT"
check_value "APP_NAME" "test-application" "$APP_NAME"

# Test envFrom with prefix
echo ""
echo "Testing envFrom with prefix (secret):"
check_value "PARTIAL_PARTIAL_KEY_1" "value1" "$PARTIAL_PARTIAL_KEY_1"
check_value "PARTIAL_PARTIAL_KEY_2" "value2" "$PARTIAL_PARTIAL_KEY_2"
check_value "PARTIAL_PARTIAL_KEY_3" "value3" "$PARTIAL_PARTIAL_KEY_3"

echo ""
echo "Testing envFrom with prefix (configmap):"
check_value "CONFIG_PREFIX_CONFIG_A" "value-a" "$CONFIG_PREFIX_CONFIG_A"
check_value "CONFIG_PREFIX_CONFIG_B" "value-b" "$CONFIG_PREFIX_CONFIG_B"
check_value "CONFIG_PREFIX_CONFIG_C" "value-c" "$CONFIG_PREFIX_CONFIG_C"

# Test 2: Test with different output formats
echo ""
echo "Test 2: Testing different output formats..."

echo "Testing docker format:"
./kubectl-eex deployment/test-app -n test-keex -o docker | grep -q "ENV DATABASE_HOST=\"localhost\"" || (echo "Docker format test failed" && exit 1)
echo -e "${GREEN}✓ Docker format test passed${NC}"

echo "Testing dotenv format:"
./kubectl-eex deployment/test-app -n test-keex -o dotenv | grep -q "DATABASE_HOST=localhost" || (echo "Dotenv format test failed" && exit 1)
echo -e "${GREEN}✓ Dotenv format test passed${NC}"

echo ""
echo -e "${GREEN}=== All integration tests passed! ===${NC}"
