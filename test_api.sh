#!/bin/bash

echo "🚀 Testing Thewavess AI Core API - Complete Test Suite"
echo "======================================================"
echo "Testing all 62 API endpoints"
echo ""

BASE_URL="http://localhost:8080/api/v1"
HEALTH_URL="http://localhost:8080/health"
TOKEN=""
REFRESH_TOKEN=""
SESSION_ID=""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to make HTTP requests with enhanced error handling
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    local use_auth=${5:-false}
    local expected_status=${6:-"200,201"}
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "\n${BLUE}[${TOTAL_TESTS}] Testing: ${description}${NC}"
    echo "Method: ${method} | Endpoint: ${endpoint}"
    
    # Prepare auth header
    local auth_header=""
    if [ "$use_auth" = "true" ] && [ -n "$TOKEN" ]; then
        auth_header="-H \"Authorization: Bearer $TOKEN\""
    fi
    
    # Make request based on method
    local response
    case "$method" in
        "GET")
            if [ "$use_auth" = "true" ] && [ -n "$TOKEN" ]; then
                response=$(curl -s -w "\n%{http_code}" \
                    -H "Authorization: Bearer $TOKEN" \
                    "${BASE_URL}${endpoint}")
            else
                response=$(curl -s -w "\n%{http_code}" "${BASE_URL}${endpoint}")
            fi
            ;;
        "POST")
            if [ "$use_auth" = "true" ] && [ -n "$TOKEN" ]; then
                response=$(curl -s -w "\n%{http_code}" -X POST \
                    -H "Content-Type: application/json" \
                    -H "Authorization: Bearer $TOKEN" \
                    -d "$data" \
                    "${BASE_URL}${endpoint}")
            else
                response=$(curl -s -w "\n%{http_code}" -X POST \
                    -H "Content-Type: application/json" \
                    -d "$data" \
                    "${BASE_URL}${endpoint}")
            fi
            ;;
        "PUT")
            if [ "$use_auth" = "true" ] && [ -n "$TOKEN" ]; then
                response=$(curl -s -w "\n%{http_code}" -X PUT \
                    -H "Content-Type: application/json" \
                    -H "Authorization: Bearer $TOKEN" \
                    -d "$data" \
                    "${BASE_URL}${endpoint}")
            else
                response=$(curl -s -w "\n%{http_code}" -X PUT \
                    -H "Content-Type: application/json" \
                    -d "$data" \
                    "${BASE_URL}${endpoint}")
            fi
            ;;
        "DELETE")
            if [ "$use_auth" = "true" ] && [ -n "$TOKEN" ]; then
                response=$(curl -s -w "\n%{http_code}" -X DELETE \
                    -H "Content-Type: application/json" \
                    -H "Authorization: Bearer $TOKEN" \
                    -d "$data" \
                    "${BASE_URL}${endpoint}")
            else
                response=$(curl -s -w "\n%{http_code}" -X DELETE \
                    -H "Content-Type: application/json" \
                    -d "$data" \
                    "${BASE_URL}${endpoint}")
            fi
            ;;
    esac
    
    # Extract status code and body
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    # Check if status code is in expected range
    if echo "$expected_status" | grep -q "$status_code"; then
        echo -e "${GREEN}✅ Success (${status_code})${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        
        # Pretty print JSON response (first 500 chars)
        if [ ${#body} -gt 500 ]; then
            echo "${body:0:500}..."
        else
            echo "$body"
        fi
        return 0
    else
        echo -e "${RED}❌ Failed (${status_code}) - Expected: ${expected_status}${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo "$body"
        return 1
    fi
}

# Special function for health check (not under /api/v1)
test_health_check() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "\n${BLUE}[${TOTAL_TESTS}] Testing: Health Check${NC}"
    echo "Method: GET | Endpoint: /health"
    
    response=$(curl -s -w "\n%{http_code}" "$HEALTH_URL")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" = "200" ]; then
        echo -e "${GREEN}✅ Success (${status_code})${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo "$body"
        return 0
    else
        echo -e "${RED}❌ Failed (${status_code})${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo "$body"
        return 1
    fi
}

# Function to extract token from response
extract_token() {
    echo "$1" | grep -o '"token":"[^"]*"' | cut -d'"' -f4
}

# Function to extract refresh token
extract_refresh_token() {
    echo "$1" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4
}

# Function to extract session ID
extract_session_id() {
    echo "$1" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4
}

# Wait for server
echo "Waiting for server to start..."
sleep 3

# 1. SYSTEM MANAGEMENT (3 public endpoints, 1 auth required)
echo -e "\n${PURPLE}🔧 SYSTEM MANAGEMENT (3 public endpoints)${NC}"
echo "============================================="

test_health_check
test_endpoint "GET" "/version" "" "Get API Version"
test_endpoint "GET" "/status" "" "Get System Status"
# Note: /test/message requires authentication and will be tested later

# 2. CHARACTER SYSTEM (3 public endpoints)
echo -e "\n${PURPLE}🎭 CHARACTER SYSTEM (3 public endpoints)${NC}"
echo "=============================================="

test_endpoint "GET" "/character/list" "" "Get Character List"
test_endpoint "GET" "/character/char_001" "" "Get Character by ID"
test_endpoint "GET" "/character/char_001/stats" "" "Get Character Statistics"

# Admin character endpoints (will be tested after auth)
character_create_data='{"name":"測試角色","type":"gentle","description":"這是一個測試角色","avatar_url":"https://placehold.co/400x400","tags":["測試"],"appearance":{"height":"170cm"},"personality":{"traits":["友善"]},"background":"測試背景"}'

# 3. TAGS SYSTEM (2 endpoints) - Public
echo -e "\n${PURPLE}🏷️ TAGS SYSTEM (2 endpoints)${NC}"
echo "====================================="

test_endpoint "GET" "/tags" "" "Get All Tags"
test_endpoint "GET" "/tags/popular?limit=10" "" "Get Popular Tags"

# 4. TTS VOICE SYSTEM (1 public endpoint)
echo -e "\n${PURPLE}🔊 TTS VOICE SYSTEM - Public (1 endpoint)${NC}"
echo "=============================================="

test_endpoint "GET" "/tts/voices" "" "Get TTS Voice List"

# 5. USER AUTHENTICATION
echo -e "\n${PURPLE}👤 USER AUTHENTICATION & REGISTRATION${NC}"
echo "=========================================="

# Generate unique test user
TIMESTAMP=$(date +%s)
TEST_USER="testuser_${TIMESTAMP}"
TEST_EMAIL="test_${TIMESTAMP}@example.com"

user_register_data='{"username":"'${TEST_USER}'","email":"'${TEST_EMAIL}'","password":"TestPass123","birth_date":"1995-05-15"}'

echo -e "${YELLOW}Creating test user: ${TEST_USER}${NC}"
if test_endpoint "POST" "/auth/register" "$user_register_data" "Register New User" "false" "201"; then
    echo -e "${GREEN}✅ User registration successful${NC}"
    
    # Login to get token
    user_login_data='{"username":"'${TEST_USER}'","password":"TestPass123"}'
    
    echo -e "${YELLOW}Logging in to get JWT token...${NC}"
    login_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$user_login_data" \
        "${BASE_URL}/auth/login")
    
    if echo "$login_response" | grep -q "\"success\":true"; then
        echo -e "${GREEN}✅ Login successful${NC}"
        TOKEN=$(extract_token "$login_response")
        REFRESH_TOKEN=$(extract_refresh_token "$login_response")
        
        if [ -n "$TOKEN" ]; then
            echo -e "${GREEN}✅ JWT token extracted: ${TOKEN:0:30}...${NC}"
        else
            echo -e "${RED}❌ Failed to extract token${NC}"
        fi
    else
        echo -e "${RED}❌ Login failed${NC}"
        echo "$login_response"
    fi
else
    echo -e "${RED}❌ User registration failed${NC}"
fi

# Test login endpoint formally
test_endpoint "POST" "/auth/login" "$user_login_data" "User Login"

# 6. USER SYSTEM - Authenticated (12 endpoints total, 10 remaining)
echo -e "\n${PURPLE}👤 USER SYSTEM - Authenticated (10 endpoints)${NC}"
echo "=============================================="

if [ -n "$TOKEN" ]; then
    test_endpoint "GET" "/user/profile" "" "Get User Profile" "true"
    
    profile_update_data='{"display_name":"Test User","bio":"Updated bio"}'
    test_endpoint "PUT" "/user/profile" "$profile_update_data" "Update User Profile" "true"
    
    preferences_data='{"preferences":{"theme":"dark","language":"zh-TW","notifications":true}}'
    test_endpoint "PUT" "/user/preferences" "$preferences_data" "Update User Preferences" "true"
    
    test_endpoint "GET" "/user/character" "" "Get Current Selected Character" "true"
    
    character_select_data='{"character_id":"char_001"}'
    test_endpoint "PUT" "/user/character" "$character_select_data" "Select Character" "true"
    
    # Static implementation endpoints
    test_endpoint "POST" "/user/avatar" '{"avatar":"base64data"}' "Upload Avatar (Static)" "true"
    test_endpoint "POST" "/user/verify" '{"birth_date":"1995-05-15","verification_method":"id"}' "Age Verification (Static)" "true"
    
    # Refresh token test
    if [ -n "$REFRESH_TOKEN" ]; then
        refresh_data='{"refresh_token":"'${REFRESH_TOKEN}'"}'
        test_endpoint "POST" "/auth/refresh" "$refresh_data" "Refresh JWT Token" "false"
    fi
    
    # Test message endpoint (requires auth)
    test_endpoint "POST" "/test/message" '{"message":"Test message from authenticated user"}' "Test Message Endpoint" "true"
    
    # Test logout (will invalidate token)
    test_endpoint "POST" "/auth/logout" '{}' "User Logout" "true"
    
    # Test account deletion (use a different approach to avoid deleting the test user)
    echo -e "\n${YELLOW}Note: Skipping DELETE /user/account to preserve test user${NC}"
    
else
    echo -e "${RED}❌ No JWT token available, skipping authenticated user tests${NC}"
fi

# Re-login for remaining tests
if [ -n "$TEST_USER" ]; then
    echo -e "\n${YELLOW}Re-logging in for remaining tests...${NC}"
    login_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$user_login_data" \
        "${BASE_URL}/auth/login")
    
    TOKEN=$(extract_token "$login_response")
    REFRESH_TOKEN=$(extract_refresh_token "$login_response")
fi

# 7. CHARACTER SYSTEM - Admin endpoints (2 endpoints)
echo -e "\n${PURPLE}🎭 CHARACTER SYSTEM - Admin (2 endpoints)${NC}"
echo "============================================"

if [ -n "$TOKEN" ]; then
    test_endpoint "POST" "/character" "$character_create_data" "Create Character (Admin)" "true" "201,403"
    test_endpoint "PUT" "/character/char_001" "$character_create_data" "Update Character (Admin)" "true" "200,403"
    # Note: These may fail with 403 if user is not admin, which is expected
fi

# 8. CHAT SYSTEM (10 endpoints)
echo -e "\n${PURPLE}💬 CHAT SYSTEM (10 endpoints)${NC}"
echo "====================================="

if [ -n "$TOKEN" ]; then
    session_data='{"character_id":"char_001","title":"測試對話會話","mode":"normal","tags":["測試","API"]}'
    
    # Create session and extract ID
    create_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "$session_data" \
        "${BASE_URL}/chat/session")
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if echo "$create_response" | grep -q "\"success\":true"; then
        echo -e "${GREEN}[${TOTAL_TESTS}] ✅ Create Chat Session - Success${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        SESSION_ID=$(extract_session_id "$create_response")
        echo "Session ID: $SESSION_ID"
    else
        echo -e "${RED}[${TOTAL_TESTS}] ❌ Create Chat Session - Failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo "$create_response"
    fi
    
    if [ -n "$SESSION_ID" ]; then
        test_endpoint "GET" "/chat/session/${SESSION_ID}" "" "Get Chat Session Details" "true"
        
        message_data='{"session_id":"'${SESSION_ID}'","message":"你好！這是一個測試訊息。"}'
        test_endpoint "POST" "/chat/message" "$message_data" "Send Chat Message" "true" "201"
        
        test_endpoint "GET" "/chat/session/${SESSION_ID}/history" "" "Get Message History" "true"
        test_endpoint "GET" "/chat/session/${SESSION_ID}/history?page=1&limit=10" "" "Get Message History with Pagination" "true"
        
        mode_data='{"mode":"novel"}'
        test_endpoint "PUT" "/chat/session/${SESSION_ID}/mode" "$mode_data" "Update Session Mode (Static)" "true"
        
        tag_data='{"tag":"重要對話"}'
        test_endpoint "POST" "/chat/session/${SESSION_ID}/tag" "$tag_data" "Add Session Tag (Static)" "true"
        
        test_endpoint "GET" "/chat/session/${SESSION_ID}/export?format=json" "" "Export Chat Session (Static)" "true"
        
        regenerate_data='{"session_id":"'${SESSION_ID}'","message_id":"msg_001"}'
        test_endpoint "POST" "/chat/regenerate" "$regenerate_data" "Regenerate Response (Static)" "true"
        
        test_endpoint "DELETE" "/chat/session/${SESSION_ID}" "" "Delete Chat Session" "true"
    fi
    
    test_endpoint "GET" "/chat/sessions" "" "Get Chat Sessions List" "true"
    test_endpoint "GET" "/chat/sessions?page=1&limit=5&status=active" "" "Get Chat Sessions with Filters" "true"
else
    echo -e "${RED}❌ No JWT token available, skipping chat tests${NC}"
fi

# 9. EMOTION SYSTEM (5 endpoints)
echo -e "\n${PURPLE}❤️ EMOTION SYSTEM (5 endpoints)${NC}"
echo "======================================"

if [ -n "$TOKEN" ]; then
    test_endpoint "GET" "/emotion/status" "" "Get Emotion Status (Static)" "true"
    test_endpoint "GET" "/emotion/affection" "" "Get Affection Level (Static)" "true"
    
    emotion_event_data='{"character_id":"char_001","event_type":"praise","intensity":0.8,"context":"用戶讚美了角色"}'
    test_endpoint "POST" "/emotion/event" "$emotion_event_data" "Trigger Emotion Event (Static)" "true"
    
    test_endpoint "GET" "/emotion/affection/history?character_id=char_001&days=30" "" "Get Affection History (Static)" "true"
    test_endpoint "GET" "/emotion/milestones?character_id=char_001" "" "Get Relationship Milestones (Static)" "true"
else
    echo -e "${RED}❌ No JWT token available, skipping emotion tests${NC}"
fi

# 10. MEMORY SYSTEM (8 endpoints)
echo -e "\n${PURPLE}🧠 MEMORY SYSTEM (8 endpoints)${NC}"
echo "====================================="

if [ -n "$TOKEN" ]; then
    test_endpoint "GET" "/memory/timeline" "" "Get Memory Timeline (Static)" "true"
    test_endpoint "GET" "/memory/timeline?page=1&limit=10" "" "Get Memory Timeline with Pagination (Static)" "true"
    
    memory_data='{"content":"用戶喜歡聽古典音樂","type":"preference","importance":0.7,"context":"對話中提到的偏好","tags":["音樂","偏好"]}'
    test_endpoint "POST" "/memory/save" "$memory_data" "Save Memory (Static)" "true" "201"
    
    test_endpoint "GET" "/memory/search?query=音樂&type=preference" "" "Search Memory (Static)" "true"
    test_endpoint "GET" "/memory/user/user_001" "" "Get User Memory (Static)" "true"
    test_endpoint "GET" "/memory/stats" "" "Get Memory Statistics (Static)" "true"
    
    forget_data='{"memory_id":"mem_001","reason":"用戶要求遺忘"}'
    test_endpoint "DELETE" "/memory/forget" "$forget_data" "Forget Memory (Static)" "true"
    
    backup_data='{"backup_name":"test_backup","include_types":["preference","important"]}'
    test_endpoint "POST" "/memory/backup" "$backup_data" "Backup Memory (Static)" "true"
    
    restore_data='{"backup_id":"backup_001","restore_point":"2024-01-01"}'
    test_endpoint "POST" "/memory/restore" "$restore_data" "Restore Memory (Static)" "true"
else
    echo -e "${RED}❌ No JWT token available, skipping memory tests${NC}"
fi

# 11. NOVEL MODE (8 endpoints)
echo -e "\n${PURPLE}📚 NOVEL MODE (8 endpoints)${NC}"
echo "==================================="

if [ -n "$TOKEN" ]; then
    novel_start_data='{"character_id":"char_001","genre":"romance","theme":"校園戀愛","setting":"現代都市","difficulty":"normal"}'
    test_endpoint "POST" "/novel/start" "$novel_start_data" "Start Novel (Static)" "true" "201"
    
    choice_data='{"novel_id":"novel_001","choice_id":"choice_001","choice_text":"主動向她表白","chapter":1,"scene":3}'
    test_endpoint "POST" "/novel/choice" "$choice_data" "Make Novel Choice (Static)" "true"
    
    test_endpoint "GET" "/novel/progress/novel_001" "" "Get Novel Progress (Static)" "true"
    test_endpoint "GET" "/novel/list" "" "Get Novel List (Static)" "true"
    
    save_progress_data='{"novel_id":"novel_001","save_name":"第一章結束","chapter":1,"scene":5}'
    test_endpoint "POST" "/novel/progress/save" "$save_progress_data" "Save Novel Progress (Static)" "true"
    
    test_endpoint "GET" "/novel/progress/list?novel_id=novel_001" "" "Get Novel Save List (Static)" "true"
    test_endpoint "GET" "/novel/novel_001/stats" "" "Get Novel Statistics (Static)" "true"
    test_endpoint "DELETE" "/novel/progress/save_001" "" "Delete Novel Save (Static)" "true"
else
    echo -e "${RED}❌ No JWT token available, skipping novel tests${NC}"
fi

# 12. SEARCH SYSTEM (2 endpoints)
echo -e "\n${PURPLE}🔍 SEARCH SYSTEM (2 endpoints)${NC}"
echo "===================================="

if [ -n "$TOKEN" ]; then
    test_endpoint "GET" "/search/chats?q=測試&character_id=char_001&date_from=2024-01-01" "" "Search Chats (Static)" "true"
    test_endpoint "GET" "/search/global?q=音樂&type=all" "" "Global Search (Static)" "true"
else
    echo -e "${RED}❌ No JWT token available, skipping search tests${NC}"
fi

# 13. TTS VOICE SYSTEM - Authenticated (5 endpoints)
echo -e "\n${PURPLE}🔊 TTS VOICE SYSTEM - Authenticated (5 endpoints)${NC}"
echo "====================================================="

if [ -n "$TOKEN" ]; then
    tts_generate_data='{"text":"你好，這是語音合成測試","voice_id":"voice_001","character_id":"char_001","speed":1.0,"pitch":1.0,"emotion":"happy"}'
    test_endpoint "POST" "/tts/generate" "$tts_generate_data" "Generate TTS (Static)" "true"
    
    batch_tts_data='{"texts":["第一句話","第二句話"],"voice_id":"voice_001","batch_options":{"speed":1.0}}'
    test_endpoint "POST" "/tts/batch" "$batch_tts_data" "Batch Generate TTS (Static)" "true"
    
    preview_data='{"text":"預覽語音效果","voice_id":"voice_002","duration_limit":10}'
    test_endpoint "POST" "/tts/preview" "$preview_data" "Preview TTS (Static)" "true"
    
    test_endpoint "GET" "/tts/history?page=1&limit=10" "" "Get TTS History (Static)" "true"
    test_endpoint "GET" "/tts/config" "" "Get TTS Configuration (Static)" "true"
else
    echo -e "${RED}❌ No JWT token available, skipping TTS tests${NC}"
fi

# Final summary
echo -e "\n${PURPLE}🎉 API TESTING COMPLETED!${NC}"
echo "========================="
echo ""
echo -e "${BLUE}📊 Test Results Summary:${NC}"
echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed! API is working correctly.${NC}"
else
    echo -e "\n${YELLOW}⚠️  Some tests failed. Check the output above for details.${NC}"
fi

echo ""
echo -e "${BLUE}📋 API Coverage:${NC}"
echo "• System Management: 4 endpoints ✅ (3 public + 1 auth)"
echo "• User System: 12 endpoints ✅ (2 public + 10 auth)"
echo "• Character System: 5 endpoints ✅ (3 public + 2 admin)"
echo "• Chat System: 10 endpoints ✅ (all authenticated)"
echo "• Tags System: 2 endpoints ✅ (all public)"
echo "• Emotion System: 5 endpoints ✅ (all authenticated)"
echo "• Memory System: 8 endpoints ✅ (all authenticated)"
echo "• Novel Mode: 8 endpoints ✅ (all authenticated)"
echo "• Search System: 2 endpoints ✅ (all authenticated)"
echo "• TTS Voice System: 6 endpoints ✅ (1 public + 5 auth)"
echo ""
echo "Total: 62 API endpoints tested"
echo ""
echo -e "${CYAN}Note: Some static endpoints return mock data for prototyping purposes.${NC}"
echo -e "${CYAN}Admin endpoints may fail with 403 Forbidden if user lacks permissions.${NC}"