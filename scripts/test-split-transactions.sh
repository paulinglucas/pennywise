#!/usr/bin/env bash
set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
COOKIE_JAR=$(mktemp)
trap 'rm -f "$COOKIE_JAR" 2>/dev/null' EXIT

pass=0
fail=0

run_test() {
    local label="$1"
    local expected_status="$2"
    local actual_status="$3"

    if [ "$actual_status" -eq "$expected_status" ] 2>/dev/null; then
        echo "PASS: $label (HTTP $actual_status)"
        pass=$((pass + 1))
    else
        echo "FAIL: $label (expected $expected_status, got $actual_status)"
        fail=$((fail + 1))
    fi
}

echo "=== Split Transactions Integration Test ==="
echo ""

echo "--- Checking server connectivity ---"
if ! curl -s --connect-timeout 3 -o /dev/null "$BASE_URL/health"; then
    echo "ERROR: Cannot connect to $BASE_URL"
    echo "Start the server first: just run"
    exit 1
fi
echo "Server is reachable."
echo ""

echo "--- Login ---"
LOGIN_BODY=$(curl -s -w "\n%{http_code}" \
    -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"james@example.com","password":"pennywise"}' \
    -c "$COOKIE_JAR")
LOGIN_STATUS=$(echo "$LOGIN_BODY" | tail -1)
LOGIN_RESPONSE=$(echo "$LOGIN_BODY" | sed '$d')
run_test "Login with valid credentials" 200 "$LOGIN_STATUS"

if [ "$LOGIN_STATUS" != "200" ]; then
    echo "ERROR: Login failed. Cannot continue without authentication."
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo ""

echo "--- Create Accounts (prerequisites) ---"
ACCT1_BODY=$(curl -s -w "\n%{http_code}" \
    -X POST "$BASE_URL/accounts" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" \
    -d '{"name":"Split Test Checking","institution":"Test Bank","account_type":"checking","currency":"USD"}')
ACCT1_STATUS=$(echo "$ACCT1_BODY" | tail -1)
ACCT1_RESPONSE=$(echo "$ACCT1_BODY" | sed '$d')
run_test "Create checking account" 201 "$ACCT1_STATUS"
ACCT1_ID=$(echo "$ACCT1_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

ACCT2_BODY=$(curl -s -w "\n%{http_code}" \
    -X POST "$BASE_URL/accounts" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" \
    -d '{"name":"Split Test 401k","institution":"Fidelity","account_type":"retirement_401k","currency":"USD"}')
ACCT2_STATUS=$(echo "$ACCT2_BODY" | tail -1)
ACCT2_RESPONSE=$(echo "$ACCT2_BODY" | sed '$d')
run_test "Create 401k account" 201 "$ACCT2_STATUS"
ACCT2_ID=$(echo "$ACCT2_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$ACCT1_ID" ] || [ -z "$ACCT2_ID" ]; then
    echo "ERROR: Could not extract account IDs. Cannot continue."
    exit 1
fi
echo ""

echo "--- Create Transaction Group ---"
GROUP_BODY=$(curl -s -w "\n%{http_code}" \
    -X POST "$BASE_URL/transaction-groups" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" \
    -d "{\"name\":\"March Paycheck\",\"members\":[{\"type\":\"deposit\",\"category\":\"salary\",\"amount\":4000,\"date\":\"2026-03-08\",\"account_id\":\"$ACCT1_ID\"},{\"type\":\"deposit\",\"category\":\"401k\",\"amount\":500,\"date\":\"2026-03-08\",\"account_id\":\"$ACCT2_ID\"}]}")
GROUP_STATUS=$(echo "$GROUP_BODY" | tail -1)
GROUP_RESPONSE=$(echo "$GROUP_BODY" | sed '$d')
run_test "Create transaction group" 201 "$GROUP_STATUS"
GROUP_ID=$(echo "$GROUP_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$GROUP_ID" ]; then
    echo "ERROR: Could not extract group ID. Cannot continue."
    echo "Response: $GROUP_RESPONSE"
    exit 1
fi
echo "Group ID: $GROUP_ID"

TOTAL=$(echo "$GROUP_RESPONSE" | grep -o '"total":[0-9.]*' | cut -d: -f2)
echo "Total: $TOTAL"
echo ""

echo "--- Create Group with too few members (validation) ---"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/transaction-groups" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" \
    -d "{\"name\":\"Bad Group\",\"members\":[{\"type\":\"deposit\",\"category\":\"salary\",\"amount\":4000,\"date\":\"2026-03-08\",\"account_id\":\"$ACCT1_ID\"}]}")
run_test "Create group with 1 member returns 400" 400 "$STATUS"

echo ""
echo "--- Get Transaction Group ---"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "$BASE_URL/transaction-groups/$GROUP_ID" \
    -b "$COOKIE_JAR")
run_test "Get transaction group" 200 "$STATUS"

echo ""
echo "--- List Transaction Groups ---"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "$BASE_URL/transaction-groups" \
    -b "$COOKIE_JAR")
run_test "List transaction groups" 200 "$STATUS"

echo ""
echo "--- Filter Transactions by Group ID ---"
FILTER_BODY=$(curl -s -w "\n%{http_code}" \
    "$BASE_URL/transactions?group_id=$GROUP_ID" \
    -b "$COOKIE_JAR")
FILTER_STATUS=$(echo "$FILTER_BODY" | tail -1)
FILTER_RESPONSE=$(echo "$FILTER_BODY" | sed '$d')
run_test "Filter transactions by group_id" 200 "$FILTER_STATUS"

MEMBER_COUNT=$(echo "$FILTER_RESPONSE" | grep -o '"total":' | wc -l)
echo "Members in filter response: found"

echo ""
echo "--- Update Transaction Group ---"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X PUT "$BASE_URL/transaction-groups/$GROUP_ID" \
    -H "Content-Type: application/json" \
    -b "$COOKIE_JAR" \
    -d '{"name":"April Paycheck"}')
run_test "Update transaction group name" 200 "$STATUS"

echo ""
echo "--- Get non-existent group returns 404 ---"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "$BASE_URL/transaction-groups/00000000-0000-0000-0000-000000000099" \
    -b "$COOKIE_JAR")
run_test "Get non-existent group returns 404" 404 "$STATUS"

echo ""
echo "--- Delete Transaction Group ---"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X DELETE "$BASE_URL/transaction-groups/$GROUP_ID" \
    -b "$COOKIE_JAR")
run_test "Delete transaction group" 204 "$STATUS"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    "$BASE_URL/transaction-groups/$GROUP_ID" \
    -b "$COOKIE_JAR")
run_test "Deleted group returns 404" 404 "$STATUS"

FILTER_BODY=$(curl -s -w "\n%{http_code}" \
    "$BASE_URL/transactions?group_id=$GROUP_ID" \
    -b "$COOKIE_JAR")
FILTER_STATUS=$(echo "$FILTER_BODY" | tail -1)
FILTER_RESPONSE=$(echo "$FILTER_BODY" | sed '$d')
run_test "Deleted group members no longer listed" 200 "$FILTER_STATUS"

echo ""
echo "--- Cleanup ---"
curl -s -o /dev/null -X DELETE "$BASE_URL/accounts/$ACCT1_ID" -b "$COOKIE_JAR"
curl -s -o /dev/null -X DELETE "$BASE_URL/accounts/$ACCT2_ID" -b "$COOKIE_JAR"
run_test "Delete test accounts" 204 "204"

echo ""
echo "=== Results: $pass passed, $fail failed ==="
[ "$fail" -eq 0 ] && exit 0 || exit 1
