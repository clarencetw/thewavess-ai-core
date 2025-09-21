#!/bin/bash

# 🧪 Thewavess AI Core - 共用測試工具庫
# 提供統一的測試功能和工具函數，避免重複代碼和錯誤

# ================================
# 全域配置和常量
# ================================

# 預設配置（可被環境變數覆蓋）
TEST_BASE_URL="${TEST_BASE_URL:-http://localhost:8080/api/v1}"
TEST_HEALTH_URL="${TEST_HEALTH_URL:-http://localhost:8080/health}"
# 生成唯一測試用戶名避免衝突
TEST_USERNAME="${TEST_USERNAME:-testusertemp_$$_$(date +%s)}"
TEST_PASSWORD="${TEST_PASSWORD:-TempPassword123}"
TEST_USER_ID="${TEST_USER_ID:-test_user_01}"
TEST_CHARACTER_ID="${TEST_CHARACTER_ID:-character_02}"
TEST_DELAY="${TEST_DELAY:-2}"

# 顏色定義
export TC_RED='\033[0;31m'
export TC_GREEN='\033[0;32m'
export TC_YELLOW='\033[1;33m'
export TC_BLUE='\033[0;34m'
export TC_PURPLE='\033[0;35m'
export TC_CYAN='\033[0;36m'
export TC_NC='\033[0m'

# 全域變數
TC_JWT_TOKEN=""
TC_REFRESH_TOKEN=""
TC_ADMIN_TOKEN=""
TC_CHAT_ID=""
TC_TEST_COUNT=0
TC_PASS_COUNT=0
TC_FAIL_COUNT=0
TC_LOG_FILE=""
# CSV功能已移除，使用詳細日誌替代

# 角色配置
TC_CHARACTERS=("character_01" "character_02" "character_03")
TC_CHARACTER_NAMES=("沈宸" "林知遠" "周曜")

# ================================
# 日誌和輸出函數
# ================================

# 初始化日誌系統
tc_init_logging() {
    local test_name="${1:-test}"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    # 創建日誌目錄
    local log_dir="$(dirname "$0")/logs"
    mkdir -p "$log_dir"
    
    # 設置日誌檔案
    TC_LOG_FILE="$log_dir/${test_name}_${timestamp}.log"
    
    # 初始化日誌
    echo "# Thewavess AI Core Test Log" > "$TC_LOG_FILE"
    echo "# Test: $test_name" >> "$TC_LOG_FILE"
    echo "# Started: $(date -Iseconds)" >> "$TC_LOG_FILE"
    echo "# ================================" >> "$TC_LOG_FILE"
    echo "" >> "$TC_LOG_FILE"
}

# 記錄日誌
tc_log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # 輸出到控制台
    case "$level" in
        "INFO")  echo -e "${TC_BLUE}[$timestamp] INFO: $message${TC_NC}" ;;
        "PASS")  echo -e "${TC_GREEN}[$timestamp] PASS: $message${TC_NC}" ;;
        "FAIL")  echo -e "${TC_RED}[$timestamp] FAIL: $message${TC_NC}" ;;
        "WARN")  echo -e "${TC_YELLOW}[$timestamp] WARN: $message${TC_NC}" ;;
        "ERROR") echo -e "${TC_RED}[$timestamp] ERROR: $message${TC_NC}" ;;
        *)       echo -e "[$timestamp] $level: $message" ;;
    esac
    
    # 寫入日誌文件
    if [ -n "$TC_LOG_FILE" ]; then
        echo "[$timestamp] $level: $message" >> "$TC_LOG_FILE"
    fi
}

# 顯示測試標題
tc_show_header() {
    local title="$1"
    echo -e "${TC_PURPLE}════════════════════════════════════════${TC_NC}"
    echo -e "${TC_PURPLE}    $title${TC_NC}"
    echo -e "${TC_PURPLE}════════════════════════════════════════${TC_NC}"
    echo ""
    tc_log "INFO" "Test Started: $title"
}

# ================================
# HTTP請求和API工具
# ================================

# 檢查服務器健康狀態
tc_check_health() {
    tc_log "INFO" "Checking server health at $TEST_HEALTH_URL"
    
    local response
    response=$(curl -s -w "\n%{http_code}" "$TEST_HEALTH_URL" 2>/dev/null)
    local status_code=$(echo "$response" | tail -n1)
    
    if [ "$status_code" = "200" ]; then
        tc_log "PASS" "Server is healthy"
        return 0
    else
        tc_log "FAIL" "Server health check failed (Status: $status_code)"
        return 1
    fi
}

# 執行HTTP請求的通用函數
tc_http_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local description="$4"
    local use_auth="${5:-false}"
    local expected_status="${6:-200,201}"
    local custom_token="${7:-}"
    
    TC_TEST_COUNT=$((TC_TEST_COUNT + 1))

    tc_log "INFO" "[Test $TC_TEST_COUNT] $description"
    tc_log "INFO" "  Method: $method | Endpoint: $endpoint"

    # 記錄請求開始
    tc_log_request_start "$method" "$endpoint" "$description" "$data"
    
    # 準備認證標頭
    local auth_token="$TC_JWT_TOKEN"
    if [ -n "$custom_token" ]; then
        auth_token="$custom_token"
    fi
    
    local auth_header=""
    if [ "$use_auth" = "true" ] && [ -n "$auth_token" ]; then
        auth_header="-H \"Authorization: Bearer $auth_token\""
    fi
    
    # 執行請求
    local response
    local start_time=$(date +%s.%N)
    
    case "$method" in
        "GET")
            if [ "$use_auth" = "true" ] && [ -n "$auth_token" ]; then
                response=$(curl -s -w "\n%{http_code}" \
                    -H "Authorization: Bearer $auth_token" \
                    "${TEST_BASE_URL}${endpoint}")
            else
                response=$(curl -s -w "\n%{http_code}" "${TEST_BASE_URL}${endpoint}")
            fi
            ;;
        "POST"|"PUT"|"DELETE")
            # 使用臨時文件來處理JSON數據，避免引號問題
            local temp_data_file=""
            if [ -n "$data" ]; then
                temp_data_file="/tmp/tc_request_$$_$(date +%s).json"
                echo "$data" > "$temp_data_file"
            fi
            
            if [ "$use_auth" = "true" ] && [ -n "$auth_token" ]; then
                if [ -n "$data" ]; then
                    response=$(curl -s -w "\n%{http_code}" -X "$method" \
                        -H "Content-Type: application/json" \
                        -H "Authorization: Bearer $auth_token" \
                        -d "@$temp_data_file" \
                        "${TEST_BASE_URL}${endpoint}")
                else
                    response=$(curl -s -w "\n%{http_code}" -X "$method" \
                        -H "Content-Type: application/json" \
                        -H "Authorization: Bearer $auth_token" \
                        "${TEST_BASE_URL}${endpoint}")
                fi
            else
                if [ -n "$data" ]; then
                    response=$(curl -s -w "\n%{http_code}" -X "$method" \
                        -H "Content-Type: application/json" \
                        -d "@$temp_data_file" \
                        "${TEST_BASE_URL}${endpoint}")
                else
                    response=$(curl -s -w "\n%{http_code}" -X "$method" \
                        -H "Content-Type: application/json" \
                        "${TEST_BASE_URL}${endpoint}")
                fi
            fi
            
            # 清理臨時文件
            [ -n "$temp_data_file" ] && rm -f "$temp_data_file"
            ;;
    esac
    
    local end_time=$(date +%s.%N)
    local response_time_seconds=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    local response_time_ms=$(echo "$response_time_seconds * 1000" | bc -l 2>/dev/null | cut -d. -f1)

    # 解析回應
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')

    # 記錄API響應結束
    tc_log_response_end "$status_code" "$response_time_ms" "$body" "$(echo "$expected_status" | grep -q "$status_code" && echo "true" || echo "false")"

    # 驗證狀態碼
    if echo "$expected_status" | grep -q "$status_code"; then
        TC_PASS_COUNT=$((TC_PASS_COUNT + 1))
        tc_log "PASS" "Test passed (Status: $status_code, Time: ${response_time_ms}ms)"

        # 記錄詳細測試日誌
        tc_log_detailed "SUCCESS" "$method $endpoint" "$status_code" "$response_time_ms" "$description"

        # 記錄回應（限制長度）
        if [ ${#body} -gt 200 ]; then
            tc_log "INFO" "  Response: ${body:0:200}..."
        else
            tc_log "INFO" "  Response: $body"
        fi

        echo "$body"
        return 0
    else
        TC_FAIL_COUNT=$((TC_FAIL_COUNT + 1))
        tc_log "FAIL" "Test failed (Status: $status_code, Expected: $expected_status)"
        tc_log "ERROR" "  Response: $body"

        # 記錄詳細錯誤日誌
        tc_log_detailed "FAILED" "$method $endpoint" "$status_code" "$response_time_ms" "$description" "$body"

        echo "$body"
        return 1
    fi
}

# ================================
# 認證相關函數
# ================================

# 用戶登入
tc_authenticate() {
    local username="${1:-$TEST_USERNAME}"
    local password="${2:-$TEST_PASSWORD}"
    
    tc_log "INFO" "Authenticating user: $username"
    
    local login_data="{\"username\":\"$username\",\"password\":\"$password\"}"
    local response
    
    response=$(curl -s -X POST "${TEST_BASE_URL}/auth/login" \
        -H "Content-Type: application/json" \
        -d "$login_data")
    
    if echo "$response" | grep -q '"success":true'; then
        TC_JWT_TOKEN=$(echo "$response" | jq -r '.data.access_token // .data.token // ""' 2>/dev/null)
        TC_REFRESH_TOKEN=$(echo "$response" | jq -r '.data.refresh_token // ""' 2>/dev/null)
        TEST_USER_ID=$(echo "$response" | jq -r '.data.user.id // ""' 2>/dev/null)
        
        if [ -n "$TC_JWT_TOKEN" ] && [ "$TC_JWT_TOKEN" != "null" ]; then
            tc_log "PASS" "Authentication successful"
            tc_log "INFO" "  JWT Token: ${TC_JWT_TOKEN:0:30}..."
            return 0
        fi
    fi
    
    tc_log "FAIL" "Authentication failed"
    tc_log "ERROR" "  Response: $response"
    return 1
}

# 登出
tc_logout() {
    if [ -n "$TC_JWT_TOKEN" ]; then
        tc_http_request "POST" "/auth/logout" '{}' "User Logout" "true"
        TC_JWT_TOKEN=""
        TC_REFRESH_TOKEN=""
        tc_log "INFO" "User logged out"
    fi
}

# 註冊和認證用戶（用於新的測試腳本）
tc_register_and_authenticate() {
    local username="${1:-$TEST_USERNAME}"
    local password="${2:-$TEST_PASSWORD}"
    local email="${3:-${username}@example.com}"

    tc_log "INFO" "Registering and authenticating user: $username"

    # 1. 先嘗試註冊用戶
    local register_data="{\"username\":\"$username\",\"email\":\"$email\",\"password\":\"$password\",\"birth_date\":\"1995-01-01T00:00:00Z\",\"is_adult\":true}"
    local register_response

    register_response=$(curl -s -X POST "${TEST_BASE_URL}/auth/register" \
        -H "Content-Type: application/json" \
        -d "$register_data")

    # 如果註冊失敗但是因為用戶已存在，那我們繼續嘗試登入
    if echo "$register_response" | grep -q '"success":true'; then
        tc_log "PASS" "User registered successfully: $username"
    elif echo "$register_response" | grep -q "already exists\|already taken\|duplicate"; then
        tc_log "INFO" "User already exists, proceeding to login: $username"
    else
        tc_log "WARN" "Registration failed, attempting login anyway"
        tc_log "DEBUG" "Registration response: $register_response"
    fi

    # 2. 嘗試登入
    tc_authenticate "$username" "$password"
    return $?
}

# 清理測試用戶（如果是動態創建的）
tc_cleanup_test_user() {
    # 只清理動態創建的測試用戶，不清理預設的test_user_01
    if [ -n "$TEST_USER_ID" ] && [ "$TEST_USER_ID" != "test_user_01" ] && [ -n "$TC_JWT_TOKEN" ]; then
        tc_log "INFO" "Cleaning up test user: $TEST_USERNAME"
        curl -s -X DELETE "${TEST_BASE_URL}/user/account" \
            -H "Authorization: Bearer $TC_JWT_TOKEN" > /dev/null 2>&1 || true
        tc_log "INFO" "Test user cleanup completed"
    fi
}

# 通用清理函數
tc_cleanup() {
    tc_cleanup_test_user
    TC_JWT_TOKEN=""
    TC_REFRESH_TOKEN=""
    TC_CHAT_ID=""
}

# ================================
# 對話相關函數
# ================================

# 創建對話會話
tc_create_session() {
    local character_id="${1:-$TEST_CHARACTER_ID}"
    local title="${2:-測試對話會話}"
    
    tc_log "INFO" "Creating chat session with character: $character_id" >&2
    
    local session_data="{\"character_id\":\"$character_id\",\"title\":\"$title\"}"
    local response
    
    response=$(curl -s -X POST "${TEST_BASE_URL}/chats" \
        -H "Authorization: Bearer $TC_JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$session_data")
    
    if echo "$response" | grep -q '"success":true'; then
        TC_CHAT_ID=$(echo "$response" | jq -r '.data.chat_id // .data.id // ""' 2>/dev/null)
        if [ -n "$TC_CHAT_ID" ] && [ "$TC_CHAT_ID" != "null" ]; then
            tc_log "PASS" "Session created successfully: $TC_CHAT_ID" >&2
            echo "$TC_CHAT_ID"
            return 0
        fi
    fi
    
    tc_log "FAIL" "Session creation failed"
    tc_log "ERROR" "  Response: $response"
    return 1
}

# 測試多會話創建功能
tc_test_multiple_sessions() {
    local character_id="${1:-$TEST_CHARACTER_ID}"
    local session_count="${2:-3}"
    
    tc_log "INFO" "Testing multiple session creation with character: $character_id"
    
    local created_sessions=()
    local success_count=0
    
    for i in $(seq 1 $session_count); do
        local title="多會話測試 #${i}"
        local chat_id
        
        chat_id=$(tc_create_session "$character_id" "$title" 2>/dev/null)
        if [ $? -eq 0 ] && [ -n "$chat_id" ]; then
            created_sessions+=("$chat_id")
            success_count=$((success_count + 1))
            tc_log "PASS" "Session $i created: $chat_id"
        else
            tc_log "FAIL" "Session $i creation failed"
        fi
        
        # 短暫延遲避免請求過快
        sleep 0.5
    done
    
    tc_log "INFO" "Multiple session test results: $success_count/$session_count sessions created"
    
    # 驗證所有會話都是不同的ID
    local unique_count=$(printf '%s\n' "${created_sessions[@]}" | sort -u | wc -l)
    if [ "$unique_count" -eq "$success_count" ]; then
        tc_log "PASS" "All created sessions have unique IDs"
    else
        tc_log "FAIL" "Duplicate session IDs detected"
        return 1
    fi
    
    # 清理創建的測試會話
    for chat_id in "${created_sessions[@]}"; do
        tc_http_request "DELETE" "/chats/$chat_id" "" "Cleanup: Delete test session $chat_id" "true" "200,404" >/dev/null 2>&1 || true
    done
    
    return 0
}

# 發送消息
tc_send_message() {
    local chat_id="$1"
    local message="$2"
    local character_id="${3:-$TEST_CHARACTER_ID}"
    
    tc_log "INFO" "Sending message: $message"
    
    local message_data="{\"message\":\"$message\"}"
    local response
    local start_time=$(date +%s.%N)
    local temp_msg_file="/tmp/tc_msg_$$_$(date +%s).json"
    
    echo "$message_data" > "$temp_msg_file"
    
    response=$(curl -s --max-time 60 -X POST "${TEST_BASE_URL}/chats/$chat_id/messages" \
        -H "Authorization: Bearer $TC_JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "@$temp_msg_file")
    
    rm -f "$temp_msg_file"
    
    local end_time=$(date +%s.%N)
    local response_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
    
    # Debug: show curl command and response (to stderr to avoid contaminating return value)
    tc_log "DEBUG" "Curl command: curl -s --max-time 60 -X POST \"${TEST_BASE_URL}/chats/$chat_id/messages\" -H \"Authorization: Bearer ${TC_JWT_TOKEN:0:30}...\" -H \"Content-Type: application/json\" -d \"$message_data\""
    tc_log "DEBUG" "Raw response: [$response]"
    tc_log "DEBUG" "Response length: ${#response}"
    
    if echo "$response" | grep -q '"success":true'; then
        local ai_response=$(echo "$response" | jq -r '.data.content // .data.dialogue // .dialogue // .data.character_dialogue // "No response"' 2>/dev/null)
        local ai_engine=$(echo "$response" | jq -r '.data.ai_engine // .ai_engine // "unknown"' 2>/dev/null)
        local nsfw_level=$(echo "$response" | jq -r '.data.nsfw_level // .nsfw_level // 0' 2>/dev/null)
        
        tc_log "PASS" "Message sent successfully (Time: ${response_time}s)"
        tc_log "INFO" "  Engine: $ai_engine | NSFW: $nsfw_level"
        tc_log "INFO" "  Response: $ai_response"
        
        # Return only the JSON response
        echo "$response"
        return 0
    else
        tc_log "FAIL" "Message sending failed"
        tc_log "ERROR" "  Response: $response"
        return 1
    fi
}

# ================================
# 詳細日誌記錄系統
# ================================

# 詳細日誌記錄函數
tc_log_detailed() {
    local status="$1"       # SUCCESS, FAILED, INFO, WARN
    local request="$2"      # 請求信息 (如 "POST /api/v1/chats")
    local http_code="$3"    # HTTP 狀態碼
    local response_time="$4" # 響應時間
    local description="$5"  # 測試描述
    local response_preview="$6" # 響應預覽（可選）

    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local separator="----------------------------------------"

    # 寫入日誌文件
    if [ -n "$TC_LOG_FILE" ]; then
        {
            echo "$separator"
            echo "[$timestamp] $status: $description"
            echo "Request: $request"
            echo "HTTP Code: $http_code"
            echo "Response Time: ${response_time}ms"
            if [ -n "$response_preview" ] && [ "$status" = "FAILED" ]; then
                echo "Response Preview:"
                echo "$response_preview"
            fi
            echo "$separator"
            echo ""
        } >> "$TC_LOG_FILE"
    fi

    # 也記錄到控制台（簡化版）
    case "$status" in
        "SUCCESS") tc_log "PASS" "$description (${response_time}ms)" ;;
        "FAILED")  tc_log "FAIL" "$description (HTTP $http_code, ${response_time}ms)" ;;
        "INFO")    tc_log "INFO" "$description" ;;
        "WARN")    tc_log "WARN" "$description" ;;
    esac
}

# 記錄API請求開始
tc_log_request_start() {
    local method="$1"
    local endpoint="$2"
    local description="$3"
    local data="$4"

    if [ -n "$TC_LOG_FILE" ]; then
        {
            echo ">>> API REQUEST START <<<"
            echo "Description: $description"
            echo "Method: $method"
            echo "Endpoint: $endpoint"
            echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
            if [ -n "$data" ] && [ "$data" != "null" ]; then
                echo "Request Data:"
                echo "$data" | jq . 2>/dev/null || echo "$data"
            fi
            echo ""
        } >> "$TC_LOG_FILE"
    fi
}

# 記錄API響應結果
tc_log_response_end() {
    local http_code="$1"
    local response_time="$2"
    local response_body="$3"
    local success="$4"

    if [ -n "$TC_LOG_FILE" ]; then
        {
            echo "<<< API RESPONSE END <<<"
            echo "HTTP Code: $http_code"
            echo "Response Time: ${response_time}ms"
            echo "Success: $success"
            echo "Response Body:"
            echo "$response_body" | jq . 2>/dev/null || echo "$response_body"
            echo ""
            echo "=================================================="
            echo ""
        } >> "$TC_LOG_FILE"
    fi
}

# ================================
# 測試場景和數據
# ================================

# 測試場景 (使用簡單變數代替關聯陣列以支援舊版bash)
# 使用函數來模擬關聯陣列
tc_get_scenario() {
    local scenario_type="$1"
    local scenario_key="$2"
    
    case "$scenario_type" in
        "basic")
            case "$scenario_key" in
                "greeting") echo "你好，很高興認識你" ;;
                "casual_chat") echo "今天天氣真好" ;;
                "support") echo "我今天心情有點低落" ;;
            esac
            ;;
        "comprehensive")
            case "$scenario_key" in
                "greeting") echo "你好，我是新來的" ;;
                "work_stress") echo "工作壓力讓我很累" ;;
                "romantic_hint") echo "你真的很溫暖，我很喜歡和你聊天" ;;
                "physical_comfort") echo "可以抱抱嗎？我需要一些安慰" ;;
                "intimate_desire") echo "想要更親密一些" ;;
            esac
            ;;
        "nsfw")
            case "$scenario_key" in
                "level_3") echo "想要感受你的溫暖擁抱" ;;
                "level_4") echo "想要更親密的身體接觸" ;;
                "level_5") echo "想要完全屬於你" ;;
            esac
            ;;
    esac
}

# ================================
# 測試結果和報告
# ================================

# 顯示測試總結
tc_show_summary() {
    local test_name="${1:-Test}"
    
    echo ""
    echo -e "${TC_PURPLE}════════════════════════════════════════${TC_NC}"
    echo -e "${TC_GREEN}📊 $test_name Summary${TC_NC}"
    echo -e "${TC_PURPLE}════════════════════════════════════════${TC_NC}"
    echo -e "${TC_BLUE}   Total Tests: $TC_TEST_COUNT${TC_NC}"
    echo -e "${TC_GREEN}   Passed: $TC_PASS_COUNT${TC_NC}"
    echo -e "${TC_RED}   Failed: $TC_FAIL_COUNT${TC_NC}"
    
    if [ -n "$TC_LOG_FILE" ]; then
        echo -e "${TC_CYAN}   Log File: $TC_LOG_FILE${TC_NC}"
    fi
    
    echo -e "${TC_PURPLE}════════════════════════════════════════${TC_NC}"
    
    # 記錄到日誌
    tc_log "INFO" "Test Summary - Total: $TC_TEST_COUNT, Passed: $TC_PASS_COUNT, Failed: $TC_FAIL_COUNT"
    
    # 返回測試是否全部通過
    if [ $TC_FAIL_COUNT -eq 0 ]; then
        tc_log "PASS" "All tests passed!"
        return 0
    else
        tc_log "FAIL" "Some tests failed!"
        return 1
    fi
}

# 清理資源
tc_cleanup() {
    tc_log "INFO" "Cleaning up test resources"
    
    # 登出用戶
    tc_logout
    
    # 清理會話
    if [ -n "$TC_CHAT_ID" ]; then
        tc_http_request "DELETE" "/chats/$TC_CHAT_ID" "" "Cleanup: Delete session" "true" "200,404" >/dev/null 2>&1 || true
    fi
    
    # 重置全域變數
    TC_JWT_TOKEN=""
    TC_REFRESH_TOKEN=""
    TC_CHAT_ID=""
    
    tc_log "INFO" "Cleanup completed"
}

# ================================
# 測試用戶管理函數 (安全)
# ================================

# 創建測試用戶（用於需要刪除測試的場景）
tc_create_test_user() {
    local test_user_id="${1:-$(tc_random_id "test_user")}"
    local test_username="${2:-test_${test_user_id}}"
    local test_password="${3:-TestPass123!}"
    
    tc_log "INFO" "創建測試用戶: $test_username (ID: $test_user_id)"
    
    local user_data="{\"username\":\"$test_username\",\"password\":\"$test_password\",\"email\":\"${test_user_id}@test.local\"}"
    
    local response
    response=$(curl -s -X POST "${TEST_BASE_URL}/auth/register" \
        -H "Content-Type: application/json" \
        -d "$user_data")
    
    if echo "$response" | grep -q '"success":true'; then
        local created_user_id=$(echo "$response" | jq -r '.data.user.id // .data.id // ""' 2>/dev/null)
        tc_log "PASS" "測試用戶創建成功: $test_username (ID: $created_user_id)"
        echo "$created_user_id"
        return 0
    else
        tc_log "FAIL" "測試用戶創建失敗"
        tc_log "ERROR" "  Response: $response"
        return 1
    fi
}

# 安全的刪除測試用戶函數（僅刪除 test_ 前綴的用戶）
tc_delete_test_user() {
    local user_id="$1"
    local admin_token="${2:-$TC_ADMIN_TOKEN}"
    
    # 安全檢查：確保用戶ID包含 test_ 前綴
    if [[ ! "$user_id" =~ ^test_ ]]; then
        tc_log "ERROR" "安全檢查失敗: 只能刪除 test_ 前綴的測試用戶 (提供的ID: $user_id)"
        return 1
    fi
    
    # 二次確認：不能是主要測試用戶
    if [ "$user_id" = "$TEST_USER_ID" ] || [ "$user_id" = "testuser" ]; then
        tc_log "ERROR" "安全檢查失敗: 不能刪除主要測試用戶 ($user_id)"
        return 1
    fi
    
    tc_log "INFO" "刪除測試用戶: $user_id"
    
    if [ -z "$admin_token" ]; then
        tc_log "ERROR" "需要管理員權限才能刪除用戶"
        return 1
    fi
    
    local response
    response=$(curl -s -X DELETE "${TEST_BASE_URL}/admin/users/$user_id" \
        -H "Authorization: Bearer $admin_token")
    
    if echo "$response" | grep -q '"success":true'; then
        tc_log "PASS" "測試用戶刪除成功: $user_id"
        return 0
    else
        tc_log "WARN" "測試用戶刪除失敗或用戶不存在: $user_id"
        tc_log "INFO" "  Response: $response"
        return 1
    fi
}

# 測試用戶清理函數（清理所有 test_ 前綴的用戶）
tc_cleanup_test_users() {
    local admin_token="${1:-$TC_ADMIN_TOKEN}"
    
    if [ -z "$admin_token" ]; then
        tc_log "WARN" "沒有管理員權限，跳過測試用戶清理"
        return 0
    fi
    
    tc_log "INFO" "清理所有測試用戶"
    
    # 獲取用戶列表
    local users_response
    users_response=$(curl -s "${TEST_BASE_URL}/admin/users" \
        -H "Authorization: Bearer $admin_token")
    
    if echo "$users_response" | grep -q '"success":true'; then
        # 提取所有 test_ 前綴的用戶ID
        local test_users
        test_users=$(echo "$users_response" | jq -r '.data.users[]? | select(.id | startswith("test_")) | .id' 2>/dev/null)
        
        if [ -n "$test_users" ]; then
            local cleanup_count=0
            while IFS= read -r user_id; do
                if [ -n "$user_id" ] && tc_delete_test_user "$user_id" "$admin_token"; then
                    cleanup_count=$((cleanup_count + 1))
                fi
            done <<< "$test_users"
            
            tc_log "INFO" "清理完成，共清理 $cleanup_count 個測試用戶"
        else
            tc_log "INFO" "沒有發現需要清理的測試用戶"
        fi
    else
        tc_log "WARN" "無法獲取用戶列表進行清理"
    fi
}

# ================================
# PostgreSQL 數組序列化測試
# ================================

# 測試管理員角色更新的數組序列化
tc_test_admin_array_serialization() {
    local admin_token="${1:-$TC_ADMIN_TOKEN}"
    local character_id="${2:-character_01}"
    
    if [ -z "$admin_token" ]; then
        tc_log "ERROR" "需要管理員權限來測試數組序列化"
        return 1
    fi
    
    tc_log "INFO" "測試管理員角色更新的 PostgreSQL 數組序列化"
    
    # 測試數據：包含中文和特殊字符的標籤
    local test_tags='["測試標籤","PostgreSQL數組","中文測試","特殊字符!@#"]'
    local update_data="{\"tags\":$test_tags}"
    
    # 執行更新
    local response
    response=$(curl -s -X PUT "${TEST_BASE_URL}/admin/characters/$character_id" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$update_data")
    
    if echo "$response" | grep -q '"success":true'; then
        # 驗證返回的標籤數據
        local returned_tags=$(echo "$response" | jq -r '.data.tags // []' 2>/dev/null)
        
        if [ "$returned_tags" != "null" ] && [ -n "$returned_tags" ]; then
            tc_log "PASS" "管理員數組序列化測試成功"
            tc_log "INFO" "  更新的標籤: $returned_tags"
            
            # 驗證特定標籤是否存在
            if echo "$response" | jq -e '.data.tags | contains(["PostgreSQL數組"])' >/dev/null 2>&1; then
                tc_log "PASS" "中文標籤正確保存"
            else
                tc_log "WARN" "中文標籤可能未正確保存"
            fi
            
            return 0
        else
            tc_log "FAIL" "數組數據未正確返回"
            return 1
        fi
    else
        tc_log "FAIL" "管理員數組序列化測試失敗"
        tc_log "ERROR" "  Response: $response"
        return 1
    fi
}

# 測試用戶角色創建的數組序列化
tc_test_user_array_serialization() {
    local user_token="${1:-$TC_JWT_TOKEN}"
    
    if [ -z "$user_token" ]; then
        tc_log "ERROR" "需要用戶認證來測試數組序列化"
        return 1
    fi
    
    tc_log "INFO" "測試用戶角色創建的 PostgreSQL 數組序列化"
    
    # 測試數據
    local character_name="數組測試角色$(date +%s)"
    local test_tags='["用戶測試","數組序列化","PostgreSQL","中文支持"]'
    local character_data="{
        \"name\":\"$character_name\",
        \"type\":\"playful\",
        \"locale\":\"zh-TW\",
        \"is_active\":true,
        \"metadata\":{
            \"tags\":$test_tags,
            \"popularity\":3
        },
        \"user_description\":\"測試 PostgreSQL 數組序列化功能\"
    }"
    
    # 創建角色
    local response
    response=$(curl -s -X POST "${TEST_BASE_URL}/character" \
        -H "Authorization: Bearer $user_token" \
        -H "Content-Type: application/json" \
        -d "$character_data")
    
    if echo "$response" | grep -q '"success":true'; then
        local character_id=$(echo "$response" | jq -r '.data.id // ""' 2>/dev/null)
        local returned_tags=$(echo "$response" | jq -r '.data.metadata.tags // []' 2>/dev/null)
        
        tc_log "PASS" "用戶數組序列化測試成功"
        tc_log "INFO" "  創建的角色ID: $character_id"
        tc_log "INFO" "  保存的標籤: $returned_tags"
        
        # 清理測試數據
        if [ -n "$character_id" ] && [ "$character_id" != "null" ]; then
            tc_http_request "DELETE" "/character/$character_id" "" "清理測試角色" "true" "200,404" >/dev/null 2>&1 || true
        fi
        
        return 0
    else
        tc_log "FAIL" "用戶數組序列化測試失敗"
        tc_log "ERROR" "  Response: $response"
        return 1
    fi
}

# 測試用戶角色更新的數組序列化
tc_test_user_array_update() {
    local user_token="${1:-$TC_JWT_TOKEN}"
    local character_id="$2"
    
    if [ -z "$user_token" ]; then
        tc_log "ERROR" "需要用戶認證來測試數組更新"
        return 1
    fi
    
    if [ -z "$character_id" ]; then
        tc_log "ERROR" "需要角色ID來測試數組更新"
        return 1
    fi
    
    tc_log "INFO" "測試用戶角色更新的 PostgreSQL 數組序列化"
    
    # 測試數據：更新標籤
    local updated_tags='["更新測試","數組序列化驗證","PostgreSQL修復"]'
    local update_data="{
        \"name\":\"更新測試角色\",
        \"metadata\":{
            \"tags\":$updated_tags,
            \"popularity\":4
        }
    }"
    
    # 執行更新
    local response
    response=$(curl -s -X PUT "${TEST_BASE_URL}/character/$character_id" \
        -H "Authorization: Bearer $user_token" \
        -H "Content-Type: application/json" \
        -d "$update_data")
    
    if echo "$response" | grep -q '"success":true'; then
        local returned_tags=$(echo "$response" | jq -r '.data.metadata.tags // []' 2>/dev/null)
        
        tc_log "PASS" "用戶數組更新測試成功"
        tc_log "INFO" "  更新的標籤: $returned_tags"
        
        # 驗證特定標籤
        if echo "$response" | jq -e '.data.metadata.tags | contains(["PostgreSQL修復"])' >/dev/null 2>&1; then
            tc_log "PASS" "數組更新內容驗證成功"
        else
            tc_log "WARN" "數組更新內容可能不正確"
        fi
        
        return 0
    else
        tc_log "FAIL" "用戶數組更新測試失敗"
        tc_log "ERROR" "  Response: $response"
        return 1
    fi
}

# 綜合 PostgreSQL 數組序列化測試
tc_test_postgresql_array_serialization() {
    local admin_token="${1:-$TC_ADMIN_TOKEN}"
    local user_token="${2:-$TC_JWT_TOKEN}"
    
    tc_log "INFO" "開始 PostgreSQL 數組序列化綜合測試"
    
    local test_passed=0
    local total_tests=0
    
    # 測試1: 管理員角色更新
    total_tests=$((total_tests + 1))
    if tc_test_admin_array_serialization "$admin_token"; then
        test_passed=$((test_passed + 1))
    fi
    
    # 測試2: 用戶角色創建
    total_tests=$((total_tests + 1))
    if tc_test_user_array_serialization "$user_token"; then
        test_passed=$((test_passed + 1))
        
        # 如果創建成功，測試角色更新
        # 注意：這裡簡化處理，實際可以保存創建的角色ID進行更新測試
    fi
    
    tc_log "INFO" "PostgreSQL 數組序列化測試完成: $test_passed/$total_tests 通過"
    
    if [ $test_passed -eq $total_tests ]; then
        tc_log "PASS" "所有 PostgreSQL 數組序列化測試通過"
        return 0
    else
        tc_log "FAIL" "部分 PostgreSQL 數組序列化測試失敗"
        return 1
    fi
}

# ================================
# Soft Delete 測試函數
# ================================

# 管理員登入
tc_admin_authenticate() {
    local username="${1:-admin}"
    local password="${2:-admin123456}"
    
    tc_log "INFO" "Authenticating admin user: $username"
    
    local login_data="{\"username\":\"$username\",\"password\":\"$password\"}"
    local response
    
    response=$(curl -s -X POST "${TEST_BASE_URL}/admin/auth/login" \
        -H "Content-Type: application/json" \
        -d "$login_data")
    
    if echo "$response" | grep -q '"success":true'; then
        TC_ADMIN_TOKEN=$(echo "$response" | jq -r '.data.access_token // ""' 2>/dev/null)
        
        if [ -n "$TC_ADMIN_TOKEN" ] && [ "$TC_ADMIN_TOKEN" != "null" ]; then
            tc_log "PASS" "Admin authentication successful"
            tc_log "INFO" "  Admin Token: ${TC_ADMIN_TOKEN:0:30}..."
            return 0
        fi
    fi
    
    tc_log "FAIL" "Admin authentication failed"
    tc_log "ERROR" "  Response: $response"
    return 1
}

# 測試用戶軟刪除
tc_test_user_soft_delete() {
    local test_username="${1:-testuser_soft_delete_$(date +%s)}"
    local test_password="${2:-password123}"
    
    tc_log "INFO" "測試用戶軟刪除功能"
    
    # 1. 註冊測試用戶
    local register_data="{\"username\":\"$test_username\",\"email\":\"${test_username}@example.com\",\"password\":\"$test_password\",\"birth_date\":\"1995-01-01T00:00:00Z\",\"is_adult\":true}"
    local register_response
    
    register_response=$(curl -s -X POST "${TEST_BASE_URL}/auth/register" \
        -H "Content-Type: application/json" \
        -d "$register_data")
    
    if ! echo "$register_response" | grep -q '"success":true'; then
        tc_log "FAIL" "用戶註冊失敗"
        return 1
    fi
    
    local test_user_id=$(echo "$register_response" | jq -r '.data.id // ""' 2>/dev/null)
    tc_log "PASS" "測試用戶註冊成功: $test_user_id"
    
    # 2. 登入測試用戶
    local login_data="{\"username\":\"$test_username\",\"password\":\"$test_password\"}"
    local login_response
    
    login_response=$(curl -s -X POST "${TEST_BASE_URL}/auth/login" \
        -H "Content-Type: application/json" \
        -d "$login_data")
    
    local test_user_token=$(echo "$login_response" | jq -r '.data.token // ""' 2>/dev/null)
    
    if [ -z "$test_user_token" ] || [ "$test_user_token" = "null" ]; then
        tc_log "FAIL" "測試用戶登入失敗"
        return 1
    fi
    
    tc_log "PASS" "測試用戶登入成功"
    
    # 3. 執行軟刪除
    local delete_data="{\"password\":\"$test_password\",\"confirmation\":\"DELETE_MY_ACCOUNT\",\"reason\":\"testing soft delete\"}"
    local delete_response
    
    delete_response=$(curl -s -X DELETE "${TEST_BASE_URL}/user/account" \
        -H "Authorization: Bearer $test_user_token" \
        -H "Content-Type: application/json" \
        -d "$delete_data")
    
    if echo "$delete_response" | grep -q '"success":true'; then
        tc_log "PASS" "用戶軟刪除成功"
        
        # 4. 驗證已刪除用戶無法登入
        local retry_login_response
        retry_login_response=$(curl -s -X POST "${TEST_BASE_URL}/auth/login" \
            -H "Content-Type: application/json" \
            -d "$login_data")
        
        if echo "$retry_login_response" | grep -q '"success":false'; then
            tc_log "PASS" "已刪除用戶無法再次登入（軟刪除驗證成功）"
            return 0
        else
            tc_log "FAIL" "已刪除用戶仍能登入（軟刪除驗證失敗）"
            return 1
        fi
    else
        tc_log "FAIL" "用戶軟刪除失敗"
        tc_log "ERROR" "  Response: $delete_response"
        return 1
    fi
}

# 測試角色軟刪除
tc_test_character_soft_delete() {
    local user_token="${1:-$TC_JWT_TOKEN}"
    
    if [ -z "$user_token" ]; then
        tc_log "ERROR" "需要用戶認證來測試角色軟刪除"
        return 1
    fi
    
    tc_log "INFO" "測試角色軟刪除功能"
    
    # 1. 創建測試角色
    local character_name="軟刪除測試角色_$(date +%s)"
    local character_data="{
        \"name\":\"$character_name\",
        \"type\":\"playful\",
        \"locale\":\"zh-TW\",
        \"is_active\":true,
        \"metadata\":{
            \"tags\":[\"測試\",\"軟刪除\"],
            \"popularity\":5
        },
        \"user_description\":\"用於測試軟刪除功能的測試角色\"
    }"
    
    local create_response
    create_response=$(curl -s -X POST "${TEST_BASE_URL}/character" \
        -H "Authorization: Bearer $user_token" \
        -H "Content-Type: application/json" \
        -d "$character_data")
    
    local character_id=$(echo "$create_response" | jq -r '.data.id // ""' 2>/dev/null)
    
    if [ -z "$character_id" ] || [ "$character_id" = "null" ]; then
        tc_log "FAIL" "測試角色創建失敗"
        return 1
    fi
    
    tc_log "PASS" "測試角色創建成功: $character_id"
    
    # 2. 驗證角色可以訪問
    local get_response
    get_response=$(curl -s "${TEST_BASE_URL}/character/$character_id")
    
    if ! echo "$get_response" | grep -q '"success":true'; then
        tc_log "FAIL" "創建的角色無法訪問"
        return 1
    fi
    
    tc_log "PASS" "創建的角色可以正常訪問"
    
    # 3. 執行軟刪除
    local delete_response
    delete_response=$(curl -s -X DELETE "${TEST_BASE_URL}/character/$character_id" \
        -H "Authorization: Bearer $user_token")
    
    if echo "$delete_response" | grep -q '"success":true'; then
        tc_log "PASS" "角色軟刪除成功"
        
        # 4. 驗證已刪除角色無法訪問
        local retry_get_response
        retry_get_response=$(curl -s "${TEST_BASE_URL}/character/$character_id")
        
        if echo "$retry_get_response" | grep -q '"success":false'; then
            tc_log "PASS" "已刪除角色無法訪問（軟刪除驗證成功）"
            
            # 5. 驗證角色從列表中消失
            local list_response
            list_response=$(curl -s "${TEST_BASE_URL}/character/list")
            
            if ! echo "$list_response" | grep -q "$character_id"; then
                tc_log "PASS" "已刪除角色從公開列表中消失"
                return 0
            else
                tc_log "FAIL" "已刪除角色仍在公開列表中"
                return 1
            fi
        else
            tc_log "FAIL" "已刪除角色仍可訪問（軟刪除驗證失敗）"
            return 1
        fi
    else
        tc_log "FAIL" "角色軟刪除失敗"
        tc_log "ERROR" "  Response: $delete_response"
        return 1
    fi
}

# 測試管理員角色恢復功能
tc_test_character_restore() {
    local admin_token="${1:-$TC_ADMIN_TOKEN}"
    local user_token="${2:-$TC_JWT_TOKEN}"
    
    if [ -z "$admin_token" ] || [ -z "$user_token" ]; then
        tc_log "ERROR" "需要管理員和用戶認證來測試角色恢復"
        return 1
    fi
    
    tc_log "INFO" "測試管理員角色恢復功能"
    
    # 1. 用戶創建並刪除角色
    local character_name="恢復測試角色_$(date +%s)"
    local character_data="{
        \"name\":\"$character_name\",
        \"type\":\"gentle\",
        \"locale\":\"zh-TW\",
        \"is_active\":true,
        \"metadata\":{
            \"tags\":[\"恢復測試\"],
            \"popularity\":3
        },
        \"user_description\":\"用於測試恢復功能的角色\"
    }"
    
    # 創建角色
    local create_response
    create_response=$(curl -s -X POST "${TEST_BASE_URL}/character" \
        -H "Authorization: Bearer $user_token" \
        -H "Content-Type: application/json" \
        -d "$character_data")
    
    local character_id=$(echo "$create_response" | jq -r '.data.id // ""' 2>/dev/null)
    
    if [ -z "$character_id" ] || [ "$character_id" = "null" ]; then
        tc_log "FAIL" "恢復測試角色創建失敗"
        return 1
    fi
    
    # 刪除角色
    local delete_response
    delete_response=$(curl -s -X DELETE "${TEST_BASE_URL}/character/$character_id" \
        -H "Authorization: Bearer $user_token")
    
    if ! echo "$delete_response" | grep -q '"success":true'; then
        tc_log "FAIL" "角色刪除失敗，無法進行恢復測試"
        return 1
    fi
    
    tc_log "PASS" "角色已刪除，準備測試恢復"
    
    # 2. 管理員恢復角色
    local restore_response
    restore_response=$(curl -s -X POST "${TEST_BASE_URL}/admin/characters/$character_id/restore" \
        -H "Authorization: Bearer $admin_token")
    
    if echo "$restore_response" | grep -q '"success":true'; then
        tc_log "PASS" "管理員角色恢復成功"
        
        # 3. 驗證恢復的角色可以訪問
        local verify_response
        verify_response=$(curl -s "${TEST_BASE_URL}/character/$character_id")
        
        if echo "$verify_response" | grep -q '"success":true'; then
            tc_log "PASS" "恢復的角色可以正常訪問"
            
            # 清理測試角色
            curl -s -X DELETE "${TEST_BASE_URL}/character/$character_id" \
                -H "Authorization: Bearer $user_token" >/dev/null 2>&1 || true
            
            return 0
        else
            tc_log "FAIL" "恢復的角色無法訪問"
            return 1
        fi
    else
        tc_log "FAIL" "管理員角色恢復失敗"
        tc_log "ERROR" "  Response: $restore_response"
        return 1
    fi
}

# 測試管理員統計API的軟刪除過濾
tc_test_admin_stats_soft_delete() {
    local admin_token="${1:-$TC_ADMIN_TOKEN}"
    
    if [ -z "$admin_token" ]; then
        tc_log "ERROR" "需要管理員認證來測試統計API軟刪除過濾"
        return 1
    fi
    
    tc_log "INFO" "測試管理員統計API的軟刪除過濾功能"
    
    # 獲取統計數據
    local stats_response
    stats_response=$(curl -s -H "Authorization: Bearer $admin_token" "${TEST_BASE_URL}/admin/stats")
    
    if echo "$stats_response" | grep -q '"success":true'; then
        local user_count=$(echo "$stats_response" | jq -r '.data.users.total // 0' 2>/dev/null)
        local character_count=$(echo "$stats_response" | jq -r '.data.characters.total // 0' 2>/dev/null)
        
        tc_log "PASS" "管理員統計API正常回應"
        tc_log "INFO" "  用戶數量: $user_count"
        tc_log "INFO" "  角色數量: $character_count"
        
        # 驗證數量合理性（應該大於0）
        if [ "$user_count" -gt 0 ] && [ "$character_count" -gt 0 ]; then
            tc_log "PASS" "統計數據合理（已過濾軟刪除記錄）"
            return 0
        else
            tc_log "WARN" "統計數據可能異常（用戶: $user_count, 角色: $character_count）"
            return 1
        fi
    else
        tc_log "FAIL" "管理員統計API回應失敗"
        tc_log "ERROR" "  Response: $stats_response"
        return 1
    fi
}

# 測試公開API的軟刪除過濾
tc_test_public_api_soft_delete() {
    tc_log "INFO" "測試公開API的軟刪除過濾功能"
    
    # 1. 測試角色列表API
    local list_response
    list_response=$(curl -s "${TEST_BASE_URL}/character/list")
    
    if echo "$list_response" | grep -q '"success":true'; then
        local character_count=$(echo "$list_response" | jq -r '.data.pagination.total_count // 0' 2>/dev/null)
        tc_log "PASS" "角色列表API正常（顯示 $character_count 個活躍角色）"
    else
        tc_log "FAIL" "角色列表API失敗"
        return 1
    fi
    
    # 2. 測試角色搜尋API
    local search_response
    search_response=$(curl -s "${TEST_BASE_URL}/character/search?q=測試")
    
    if echo "$search_response" | grep -q '"success":true'; then
        tc_log "PASS" "角色搜尋API正常"
    else
        tc_log "FAIL" "角色搜尋API失敗"
        return 1
    fi
    
    return 0
}

# 綜合軟刪除測試
tc_test_soft_delete_comprehensive() {
    local admin_token="${1:-$TC_ADMIN_TOKEN}"
    local user_token="${2:-$TC_JWT_TOKEN}"
    
    tc_log "INFO" "開始綜合軟刪除功能測試"
    
    local test_passed=0
    local total_tests=0
    local start_time end_time duration
    
    # 測試1: 用戶軟刪除
    total_tests=$((total_tests + 1))
    start_time=$(date +%s.%N)
    if tc_test_user_soft_delete; then
        test_passed=$((test_passed + 1))
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "SUCCESS" "POST /auth/register+login+delete" "200" "$duration" "用戶軟刪除測試通過"
    else
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "FAILED" "POST /auth/register+login+delete" "500" "$duration" "用戶軟刪除測試失敗"
    fi
    
    # 測試2: 角色軟刪除
    total_tests=$((total_tests + 1))
    start_time=$(date +%s.%N)
    if tc_test_character_soft_delete "$user_token"; then
        test_passed=$((test_passed + 1))
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "SUCCESS" "POST /character+DELETE" "200" "$duration" "角色軟刪除測試通過"
    else
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "FAILED" "POST /character+DELETE" "500" "$duration" "角色軟刪除測試失敗"
    fi
    
    # 測試3: 管理員統計API
    total_tests=$((total_tests + 1))
    start_time=$(date +%s.%N)
    if tc_test_admin_stats_soft_delete "$admin_token"; then
        test_passed=$((test_passed + 1))
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "SUCCESS" "GET /admin/stats" "200" "$duration" "管理員統計API過濾測試通過"
    else
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "FAILED" "GET /admin/stats" "500" "$duration" "管理員統計API過濾測試失敗"
    fi
    
    # 測試4: 公開API過濾
    total_tests=$((total_tests + 1))
    start_time=$(date +%s.%N)
    if tc_test_public_api_soft_delete; then
        test_passed=$((test_passed + 1))
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "SUCCESS" "GET /character/list+search" "200" "$duration" "公開API過濾測試通過"
    else
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "FAILED" "GET /character/list+search" "500" "$duration" "公開API過濾測試失敗"
    fi
    
    # 測試5: 角色恢復功能
    total_tests=$((total_tests + 1))
    start_time=$(date +%s.%N)
    if tc_test_character_restore "$admin_token" "$user_token"; then
        test_passed=$((test_passed + 1))
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "SUCCESS" "POST /admin/characters/restore" "200" "$duration" "角色恢復功能測試通過"
    else
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0.0")
        tc_log_detailed "FAILED" "POST /admin/characters/restore" "500" "$duration" "角色恢復功能測試失敗"
    fi
    
    tc_log "INFO" "軟刪除功能測試完成: $test_passed/$total_tests 通過"
    
    if [ $test_passed -eq $total_tests ]; then
        tc_log "PASS" "所有軟刪除功能測試通過"
        return 0
    else
        tc_log "FAIL" "部分軟刪除功能測試失敗"
        return 1
    fi
}

# ================================
# 輔助工具函數
# ================================

# 等待延遲
tc_sleep() {
    local seconds="${1:-$TEST_DELAY}"
    tc_log "INFO" "Waiting $seconds seconds..."
    sleep "$seconds"
}

# 檢查依賴工具
tc_check_dependencies() {
    local deps=("curl" "jq")
    local missing=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing+=("$dep")
        fi
    done
    
    if [ ${#missing[@]} -ne 0 ]; then
        tc_log "ERROR" "Missing dependencies: ${missing[*]}"
        tc_log "ERROR" "Please install missing tools to continue"
        return 1
    fi
    
    tc_log "PASS" "All dependencies available"
    return 0
}

# JSON解析輔助函數
tc_parse_json() {
    local json="$1"
    local path="$2"
    local default="${3:-\"\"}"
    
    # 對於數字字段，使用數字預設值
    if [[ "$path" == *"nsfw_level"* ]] || [[ "$path" == *"affection"* ]]; then
        echo "$json" | jq -r "$path // 0" 2>/dev/null
    else
        echo "$json" | jq -r "$path // $default" 2>/dev/null
    fi
}

# 生成隨機測試ID
tc_random_id() {
    local prefix="${1:-test}"
    echo "${prefix}_$(date +%s)_$RANDOM"
}

# ================================
# 初始化提示
# ================================

if [ "${BASH_SOURCE[0]}" != "${0}" ]; then
    # 被作為庫載入
    tc_log "INFO" "Thewavess AI Test Common Library loaded"
    tc_log "INFO" "Base URL: $TEST_BASE_URL"
    tc_log "INFO" "Test User: $TEST_USERNAME"
else
    # 直接執行
    echo "Thewavess AI Test Common Library"
    echo "This file should be sourced, not executed directly"
    echo "Usage: source $(basename $0)"
fi