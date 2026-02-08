#!/bin/bash
set -e

BASE_URL="http://localhost:9092"

echo "=== Testing IAM Service ==="
echo ""

# 1. Health check
echo "1. Testing health endpoints..."
curl -s $BASE_URL/health | jq '.'
curl -s $BASE_URL/ready | jq '.'
echo "✅ Health checks passed"
echo ""

# 2. Create session
echo "2. Creating new session..."
RESPONSE=$(curl -s -c cookies.txt -X POST $BASE_URL/auth/session \
  -H "Content-Type: application/json" \
  -d '{"metadata":{"ip":"127.0.0.1","test":"true"}}')

echo $RESPONSE | jq '.'

USER_ID=$(echo $RESPONSE | jq -r '.user_id')
SESSION_ID=$(echo $RESPONSE | jq -r '.session_id')
IS_NEW=$(echo $RESPONSE | jq -r '.is_new_user')

echo "User ID: $USER_ID"
echo "Session ID: $SESSION_ID"
echo "Is New User: $IS_NEW"
echo "✅ Session created"
echo ""

# 3. Validate session (ForwardAuth)
echo "3. Validating session..."
curl -s -b cookies.txt -i $BASE_URL/auth/validate | grep -E "HTTP|X-User-Id|X-Provider"
echo "✅ Session validated"
echo ""

# 4. Get user info
echo "4. Getting user info..."
curl -s $BASE_URL/users/$USER_ID | jq '.'
echo "✅ User info retrieved"
echo ""

# 5. Refresh session
echo "5. Refreshing session..."
curl -s -b cookies.txt -X PUT $BASE_URL/auth/session/refresh | jq '.'
echo "✅ Session refreshed"
echo ""

# 6. Logout
echo "6. Logging out..."
curl -s -b cookies.txt -X DELETE $BASE_URL/auth/logout -I | grep "HTTP"
echo "✅ Logged out"
echo ""

# 7. Verify session is invalid
echo "7. Verifying session is invalid..."
HTTP_CODE=$(curl -s -b cookies.txt -o /dev/null -w "%{http_code}" $BASE_URL/auth/validate)
if [ "$HTTP_CODE" == "401" ]; then
    echo "✅ Session correctly invalidated (401)"
else
    echo "❌ Expected 401, got $HTTP_CODE"
fi
echo ""

# Cleanup
rm -f cookies.txt

echo "=== All tests passed! ==="
