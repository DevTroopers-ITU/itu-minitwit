#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://64.226.116.162:8080}"
AUTH="Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh"
PASS=0
FAIL=0
LATEST=1000

bold()  { printf "\033[1m%s\033[0m\n" "$*"; }
green() { printf "\033[32m  ✓ %s\033[0m\n" "$*"; }
red()   { printf "\033[31m  ✗ %s\033[0m\n" "$*"; }

check() {
    local desc="$1" expected="$2" actual="$3"
    if [ "$actual" = "$expected" ]; then
        green "$desc"
        ((PASS++))
    else
        red "$desc (expected $expected, got $actual)"
        ((FAIL++))
    fi
}

check_contains() {
    local desc="$1" needle="$2" haystack="$3"
    if echo "$haystack" | grep -q "$needle"; then
        green "$desc"
        ((PASS++))
    else
        red "$desc (expected to contain '$needle')"
        ((FAIL++))
    fi
}

UNIQUE=$(date +%s)

bold "=== Testing DO Swarm at $BASE_URL ==="
echo ""

# --- Health check ---
bold "1. Health check"
HTTP=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/latest")
check "/latest reachable" "200" "$HTTP"

# --- Register users ---
bold "2. Register test users"
for suffix in a b c; do
    USER="simtest_${suffix}_${UNIQUE}"
    ((LATEST++))
    HTTP=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$BASE_URL/register?latest=$LATEST" \
        -H "Content-Type: application/json" \
        -H "Authorization: $AUTH" \
        -d "{\"username\":\"$USER\",\"email\":\"$USER@test.com\",\"pwd\":\"testpwd\"}")
    check "Register $USER" "204" "$HTTP"
done

USER_A="simtest_a_${UNIQUE}"
USER_B="simtest_b_${UNIQUE}"
USER_C="simtest_c_${UNIQUE}"

# --- Verify /latest tracking ---
bold "3. Latest counter"
BODY=$(curl -s "$BASE_URL/latest" -H "Authorization: $AUTH")
GOT_LATEST=$(echo "$BODY" | grep -o '"latest":[0-9]*' | cut -d: -f2)
check "/latest updated to $LATEST" "$LATEST" "$GOT_LATEST"

# --- Post messages ---
bold "4. Post messages"
for i in 1 2 3; do
    ((LATEST++))
    HTTP=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$BASE_URL/msgs/$USER_A?latest=$LATEST" \
        -H "Content-Type: application/json" \
        -H "Authorization: $AUTH" \
        -d "{\"content\":\"Test message $i from $USER_A\"}")
    check "Post message $i" "204" "$HTTP"
done

# --- Read messages ---
bold "5. Read messages"
BODY=$(curl -s "$BASE_URL/msgs/$USER_A?no=20&latest=$((++LATEST))" \
    -H "Authorization: $AUTH")
check_contains "GET /msgs/$USER_A contains posted message" "Test message 1" "$BODY"

BODY=$(curl -s "$BASE_URL/msgs?no=100&latest=$((++LATEST))" \
    -H "Authorization: $AUTH")
check_contains "GET /msgs (public) contains posted message" "$USER_A" "$BODY"

# --- Follow / unfollow ---
bold "6. Follow & unfollow"
((LATEST++))
HTTP=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/fllws/$USER_A?latest=$LATEST" \
    -H "Content-Type: application/json" \
    -H "Authorization: $AUTH" \
    -d "{\"follow\":\"$USER_B\"}")
check "Follow $USER_B" "204" "$HTTP"

((LATEST++))
BODY=$(curl -s "$BASE_URL/fllws/$USER_A?no=20&latest=$LATEST" \
    -H "Authorization: $AUTH")
check_contains "GET /fllws shows $USER_B" "$USER_B" "$BODY"

((LATEST++))
HTTP=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/fllws/$USER_A?latest=$LATEST" \
    -H "Content-Type: application/json" \
    -H "Authorization: $AUTH" \
    -d "{\"unfollow\":\"$USER_B\"}")
check "Unfollow $USER_B" "204" "$HTTP"

((LATEST++))
BODY=$(curl -s "$BASE_URL/fllws/$USER_A?no=20&latest=$LATEST" \
    -H "Authorization: $AUTH")
if echo "$BODY" | grep -q "$USER_B"; then
    red "Unfollow verified ($USER_B still in follows)"
    ((FAIL++))
else
    green "Unfollow verified ($USER_B removed)"
    ((PASS++))
fi

# --- Duplicate registration (should fail) ---
bold "7. Edge cases"
((LATEST++))
HTTP=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/register?latest=$LATEST" \
    -H "Content-Type: application/json" \
    -H "Authorization: $AUTH" \
    -d "{\"username\":\"$USER_A\",\"email\":\"dup@test.com\",\"pwd\":\"testpwd\"}")
check "Duplicate register returns 400" "400" "$HTTP"

# --- Concurrent load test ---
bold "8. Concurrent load (20 parallel requests)"
TMPDIR=$(mktemp -d)
for i in $(seq 1 20); do
    curl -s -o /dev/null -w "%{http_code}" \
        "$BASE_URL/msgs?no=10&latest=$((LATEST + i))" \
        -H "Authorization: $AUTH" > "$TMPDIR/$i" &
done
wait

LOAD_OK=0
LOAD_FAIL=0
for i in $(seq 1 20); do
    if [ "$(cat "$TMPDIR/$i")" = "200" ]; then
        ((LOAD_OK++))
    else
        ((LOAD_FAIL++))
    fi
done
rm -rf "$TMPDIR"
check "All 20 concurrent requests returned 200" "20" "$LOAD_OK"

# --- Summary ---
echo ""
bold "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && echo "DO Swarm is ready for simulator traffic." || exit 1
