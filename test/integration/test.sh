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
OUTPUT=$(./kubectl-eex deployment/test-app -n test-keex -f shell)

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

# Test 2: Test with TYPE NAME format (kubectl standard format)
echo ""
echo "Test 2: Testing TYPE NAME format..."
OUTPUT_TYPE_NAME=$(./kubectl-eex deployment test-app -n test-keex -f shell)

# Check if output is the same as TYPE/NAME format
if [ "$OUTPUT" = "$OUTPUT_TYPE_NAME" ]; then
    echo -e "${GREEN}✓ TYPE NAME format works correctly${NC}"
else
    echo -e "${RED}✗ TYPE NAME format output differs from TYPE/NAME format${NC}"
    echo "Length TYPE/NAME: $(echo -n "$OUTPUT" | wc -c)"
    echo "Length TYPE NAME: $(echo -n "$OUTPUT_TYPE_NAME" | wc -c)"
    echo "Diff output:"
    diff -u <(echo "$OUTPUT") <(echo "$OUTPUT_TYPE_NAME") || true
    exit 1
fi

# Test 3: Test with different output formats
echo ""
echo "Test 3: Testing different output formats..."

echo "Testing docker format:"
./kubectl-eex deployment/test-app -n test-keex -f docker | grep -q -- "-e DATABASE_HOST=\"localhost\"" || (echo "Docker format test failed" && exit 1)
echo -e "${GREEN}✓ Docker format test passed${NC}"

echo "Testing dotenv format:"
./kubectl-eex deployment/test-app -n test-keex -f dotenv | grep -q "DATABASE_HOST=localhost" || (echo "Dotenv format test failed" && exit 1)
echo -e "${GREEN}✓ Dotenv format test passed${NC}"

# Test TYPE NAME format with output format
echo "Testing TYPE NAME format with docker output:"
./kubectl-eex deployment test-app -n test-keex -f docker | grep -q -- "-e DATABASE_HOST=\"localhost\"" || (echo "TYPE NAME docker format test failed" && exit 1)
echo -e "${GREEN}✓ TYPE NAME with docker format test passed${NC}"

# Test 4: Test with different resource types (TYPE NAME format)
echo ""
echo "Test 4: Testing different resource types with TYPE NAME format..."

# Test statefulset with TYPE/NAME format
echo "Testing statefulset with TYPE/NAME format:"
OUTPUT_STS_SLASH=$(./kubectl-eex statefulset/test-statefulset -n test-keex -f shell)
echo "$OUTPUT_STS_SLASH" | grep -q "STATEFULSET_ENV=statefulset-value" || (echo "StatefulSet TYPE/NAME format failed" && exit 1)
echo "$OUTPUT_STS_SLASH" | grep -q "DB_PASSWORD=secret123" || (echo "StatefulSet secret ref failed" && exit 1)
echo "$OUTPUT_STS_SLASH" | grep -q "STS_APP_ENV=production" || (echo "StatefulSet envFrom prefix failed" && exit 1)
echo -e "${GREEN}✓ StatefulSet TYPE/NAME format test passed${NC}"

# Test statefulset with TYPE NAME format
echo "Testing statefulset with TYPE NAME format:"
OUTPUT_STS_SPACE=$(./kubectl-eex statefulset test-statefulset -n test-keex -f shell)
if [ "$OUTPUT_STS_SLASH" = "$OUTPUT_STS_SPACE" ]; then
    echo -e "${GREEN}✓ StatefulSet TYPE NAME format matches TYPE/NAME format${NC}"
else
    echo -e "${RED}✗ StatefulSet TYPE NAME format output differs${NC}"
    exit 1
fi

# Test short resource names (sts instead of statefulset)
echo "Testing short resource name 'sts' with TYPE NAME format:"
OUTPUT_STS_SHORT=$(./kubectl-eex sts test-statefulset -n test-keex -f shell)
if [ "$OUTPUT_STS_SLASH" = "$OUTPUT_STS_SHORT" ]; then
    echo -e "${GREEN}✓ Short resource name 'sts' works correctly${NC}"
else
    echo -e "${RED}✗ Short resource name 'sts' output differs${NC}"
    exit 1
fi

# Test deployment short name
echo "Testing short resource name 'deploy' with TYPE NAME format:"
OUTPUT_DEPLOY_SHORT=$(./kubectl-eex deploy test-app -n test-keex -f shell)
if [ "$OUTPUT" = "$OUTPUT_DEPLOY_SHORT" ]; then
    echo -e "${GREEN}✓ Short resource name 'deploy' works correctly${NC}"
else
    echo -e "${RED}✗ Short resource name 'deploy' output differs${NC}"
    exit 1
fi

# Test 5: Test container selection with TYPE NAME format
echo ""
echo "Test 5: Testing container selection with TYPE NAME format..."

# First, let's create a multi-container pod for testing
cat <<EOF | kubectl apply -f - -n test-keex
apiVersion: v1
kind: Pod
metadata:
  name: test-pod-multi
spec:
  containers:
  - name: container1
    image: busybox:latest
    command: ["sleep", "3600"]
    env:
    - name: CONTAINER1_ENV
      value: "container1-value"
  - name: container2
    image: busybox:latest
    command: ["sleep", "3600"]
    env:
    - name: CONTAINER2_ENV
      value: "container2-value"
EOF

# Wait for pod to be ready
kubectl wait --for=condition=ready pod/test-pod-multi -n test-keex --timeout=60s

# Test container selection with TYPE/NAME format
echo "Testing container selection with TYPE/NAME format:"
OUTPUT_POD_C1_SLASH=$(./kubectl-eex pod/test-pod-multi -n test-keex -c container1 -f shell)
echo "$OUTPUT_POD_C1_SLASH" | grep -q "CONTAINER1_ENV=container1-value" || (echo "Container1 selection failed" && exit 1)
echo "$OUTPUT_POD_C1_SLASH" | grep -v -q "CONTAINER2_ENV" || (echo "Container1 selection included container2 env" && exit 1)
echo -e "${GREEN}✓ Container selection with TYPE/NAME format works${NC}"

# Test container selection with TYPE NAME format
echo "Testing container selection with TYPE NAME format:"
OUTPUT_POD_C1_SPACE=$(./kubectl-eex pod test-pod-multi -n test-keex -c container1 -f shell)
if [ "$OUTPUT_POD_C1_SLASH" = "$OUTPUT_POD_C1_SPACE" ]; then
    echo -e "${GREEN}✓ Container selection with TYPE NAME format works${NC}"
else
    echo -e "${RED}✗ Container selection TYPE NAME format output differs${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=== All integration tests passed! ===${NC}"
