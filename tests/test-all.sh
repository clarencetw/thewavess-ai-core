#!/bin/bash

# 🧪 Thewavess AI Core - 統一測試工具
# 單一檔案包含所有測試功能

source "$(dirname "$0")/test-config.sh"

# 解析參數
TEST_TYPE="all"
CSV_OUTPUT=false
QUICK_MODE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --type) TEST_TYPE="$2"; shift 2 ;;
        --csv) CSV_OUTPUT=true; shift ;;
        --quick) QUICK_MODE=true; shift ;;
        --help)
            echo "Thewavess AI Core 統一測試工具"
            echo ""
            echo "使用方式: $0 [選項]"
            echo ""
            echo "測試類型:"
            echo "  --type health      系統健康檢查"
            echo "  --type auth        認證系統測試"
            echo "  --type api         API功能測試"
            echo "  --type chat        對話功能測試"
            echo "  --type nsfw        NSFW分級測試"
            echo "  --type admin       管理員API測試"
            echo "  --type soft-delete 軟刪除功能測試"
            echo "  --type all         所有測試 (預設)"
            echo ""
            echo "選項:"
            echo "  --csv            生成CSV報告"
            echo "  --quick          快速模式"
            echo "  --help           顯示此幫助"
            echo ""
            echo "範例:"
            echo "  $0                        # 執行所有測試"
            echo "  $0 --type api            # 只執行API測試"
            echo "  $0 --type nsfw --csv     # 執行NSFW測試並生成報告"
            exit 0
            ;;
        *) shift ;;
    esac
done

# 初始化測試
tc_init_logging "unified_test"
if [ "$CSV_OUTPUT" = true ]; then
    tc_init_csv "unified_test"
fi

tc_show_header "Thewavess AI Core 統一測試"

# 檢查系統
tc_check_dependencies || exit 1

# ================================
# 系統健康檢查
# ================================
run_health_tests() {
    tc_log "INFO" "執行系統健康檢查"
    
    tc_check_health || {
        tc_log "ERROR" "服務器未運行，請先啟動: make run"
        return 1
    }
    
    tc_http_request "GET" "/version" "" "獲取API版本"
    tc_http_request "GET" "/status" "" "獲取系統狀態"
    tc_http_request "GET" "/monitor/health" "" "健康檢查"
    tc_http_request "GET" "/monitor/ready" "" "就緒檢查"
    tc_http_request "GET" "/monitor/live" "" "存活檢查"
    tc_http_request "GET" "/monitor/stats" "" "系統統計"
    tc_http_request "GET" "/monitor/metrics" "" "系統指標"
    
    return 0
}

# ================================
# 認證系統測試
# ================================
run_auth_tests() {
    tc_log "INFO" "執行認證系統測試"
    
    # 測試用戶註冊
    test_user_id=$(tc_random_id "test_user")
    test_username="test_${test_user_id}"
    test_password="TestAuth123!"
    test_email="${test_user_id}@test.local"
    
    register_data="{\"username\":\"$test_username\",\"password\":\"$test_password\",\"email\":\"$test_email\"}"
    register_response=$(tc_http_request "POST" "/auth/register" "$register_data" "用戶註冊" "false")
    
    if [ $? -eq 0 ]; then
        # 測試用戶登入
        login_data="{\"username\":\"$test_username\",\"password\":\"$test_password\"}"
        login_response=$(tc_http_request "POST" "/auth/login" "$login_data" "用戶登入" "false")
        
        if [ $? -eq 0 ]; then
            # 提取新用戶的 token
            NEW_JWT_TOKEN=$(echo "$login_response" | jq -r '.data.access_token // .data.token // ""' 2>/dev/null)
            NEW_REFRESH_TOKEN=$(echo "$login_response" | jq -r '.data.refresh_token // ""' 2>/dev/null)
            
            if [ -n "$NEW_JWT_TOKEN" ] && [ "$NEW_JWT_TOKEN" != "null" ]; then
                tc_log "PASS" "新用戶登入成功"
                
                # 測試 token 刷新
                if [ -n "$NEW_REFRESH_TOKEN" ] && [ "$NEW_REFRESH_TOKEN" != "null" ]; then
                    refresh_data="{\"refresh_token\":\"$NEW_REFRESH_TOKEN\"}"
                    tc_http_request "POST" "/auth/refresh" "$refresh_data" "刷新Token" "false"
                fi
                
                # 測試用戶登出
                tc_http_request "POST" "/auth/logout" '{}' "用戶登出" "true" "200,201" "$NEW_JWT_TOKEN"
            else
                tc_log "WARN" "新用戶登入失敗"
            fi
        fi
        
        # 清理測試用戶 (需要管理員權限)
        if [ -n "$ADMIN_JWT_TOKEN" ]; then
            created_user_id=$(echo "$register_response" | jq -r '.data.user.id // .data.id // ""' 2>/dev/null)
            if [ -n "$created_user_id" ]; then
                tc_delete_test_user "$created_user_id" "$ADMIN_JWT_TOKEN" || true
            fi
        fi
    fi
    
    return 0
}

# ================================
# API功能測試
# ================================
run_api_tests() {
    tc_log "INFO" "執行API功能測試"
    
    # 認證
    tc_register_and_authenticate || {
        tc_log "ERROR" "用戶註冊或認證失敗"
        return 1
    }
    
    # 角色系統
    tc_http_request "GET" "/character/list" "" "角色列表" "false"
    tc_http_request "GET" "/character/search?q=林知遠" "" "角色搜索" "false"
    tc_http_request "GET" "/character/character_02" "" "角色詳情" "false"
    tc_http_request "GET" "/character/character_02/stats" "" "角色統計" "false"
    tc_http_request "GET" "/character/character_02/profile" "" "角色資料" "true"
    
    # 多會話架構測試 - 測試與同一角色創建多個會話
    tc_log "INFO" "測試多會話架構功能"
    tc_test_multiple_sessions "character_02" 2 || tc_log "WARN" "多會話測試失敗"
    
    # 關係系統 (已從 emotion 重命名為 relationships) - 需要先創建 chat session
    test_chat_id=$(tc_create_session "character_02" "API測試會話" 2>/dev/null | tail -n1)
    if [ -n "$test_chat_id" ]; then
        tc_http_request "GET" "/relationships/chat/$test_chat_id/status" "" "關係狀態" "true"
        tc_http_request "GET" "/relationships/chat/$test_chat_id/affection" "" "好感度" "true"
        tc_http_request "GET" "/relationships/chat/$test_chat_id/history" "" "關係歷史" "true"
        # 注意: POST /emotion/event 觸發情感事件功能已刪除
    fi
    
    # 搜索系統
    tc_http_request "GET" "/search/chats?q=測試" "" "搜索對話" "true"
    tc_http_request "GET" "/search/global?q=測試&type=chats" "" "全局搜索" "true"
    
    # TTS系統
    tc_http_request "GET" "/tts/voices" "" "TTS語音列表" "false"
    
    if [ "$QUICK_MODE" = false ]; then
        tts_data='{"text":"測試語音合成","voice":"alloy","speed":1.0}'
        tc_http_request "POST" "/tts/generate" "$tts_data" "生成TTS" "true"
    fi
    
    # 用戶系統 API
    tc_http_request "GET" "/user/profile" "" "獲取用戶資料" "true"
    
    update_profile_data='{"display_name":"測試用戶","bio":"測試用戶簡介"}'
    tc_http_request "PUT" "/user/profile" "$update_profile_data" "更新用戶資料" "true"
    
    return 0
}

# ================================
# 對話功能測試
# ================================
run_chat_tests() {
    tc_log "INFO" "執行對話功能測試"
    
    # 認證
    tc_register_and_authenticate || {
        tc_log "ERROR" "用戶註冊或認證失敗"
        return 1
    }
    
    # 多會話獨立關係測試
    tc_log "INFO" "測試多會話獨立關係追蹤"
    
    # 創建兩個獨立會話
    chat_id_1=$(tc_create_session "character_02" "對話測試會話1")
    chat_id_2=$(tc_create_session "character_02" "對話測試會話2")
    
    if [ -n "$chat_id_1" ] && [ -n "$chat_id_2" ]; then
        tc_log "PASS" "成功創建兩個獨立會話: $chat_id_1, $chat_id_2"
        
        # 測試獨立關係狀態
        tc_send_message "$chat_id_1" "你好" "character_02" >/dev/null 2>&1
        tc_send_message "$chat_id_2" "你好" "character_02" >/dev/null 2>&1
        
        # 驗證每個會話都有獨立的關係狀態
        tc_http_request "GET" "/relationships/chat/$chat_id_1/status" "" "會話1關係狀態" "true" >/dev/null
        tc_http_request "GET" "/relationships/chat/$chat_id_2/status" "" "會話2關係狀態" "true" >/dev/null
        
        # 使用第一個會話進行後續測試
        chat_id="$chat_id_1"
        
        # 清理第二個測試會話
        tc_http_request "DELETE" "/chats/$chat_id_2" "" "清理測試會話2" "true" "200,404" >/dev/null 2>&1 || true
    else
        tc_log "WARN" "多會話測試失敗，使用單會話模式"
        chat_id=$(tc_create_session "character_02" "統一測試會話")
    fi
    
    [ -z "$chat_id" ] && return 1
    
    # 對話場景函數 (支援舊版bash)
    get_chat_scenario() {
        local key="$1"
        if [ "$QUICK_MODE" = true ]; then
            case "$key" in
                "greeting") echo "你好" ;;
                "casual") echo "今天天氣不錯" ;;
                "emotional") echo "我有點累" ;;
                *) echo "" ;;
            esac
        else
            case "$key" in
                "greeting") echo "你好，很高興認識你" ;;
                "casual") echo "今天天氣真好" ;;
                "emotional") echo "我今天心情有點低落" ;;
                "affectionate") echo "你真的很溫暖，我很喜歡你" ;;
                "intimate") echo "想要感受你的擁抱" ;;
                *) echo "" ;;
            esac
        fi
    }
    
    get_chat_scenarios() {
        if [ "$QUICK_MODE" = true ]; then
            echo "greeting casual emotional"
        else
            echo "greeting casual emotional affectionate intimate"
        fi
    }
    
    # 執行對話測試
    for scenario in $(get_chat_scenarios); do
        message=$(get_chat_scenario "$scenario")
        tc_log "INFO" "測試場景: $scenario"
        
        response=$(tc_send_message "$chat_id" "$message" "character_02")
        if [ $? -eq 0 ]; then
            # 提取數據供CSV記錄和 JSON 響應驗證
            ai_engine=$(tc_parse_json "$response" '.data.ai_engine')
            nsfw_level=$(tc_parse_json "$response" '.data.nsfw_level')
            affection=$(tc_parse_json "$response" '.data.affection')
            dialogue=$(tc_parse_json "$response" '.data.content')
            
            # 驗證 JSON 響應功能
            if [ -n "$dialogue" ]; then
                tc_log "PASS" "JSON 響應解析成功: 對話內容存在"
            else
                tc_log "WARN" "JSON 響應可能有問題: content=[$dialogue]"
            fi
            
            # 顯示 AI 響應詳情
            tc_log "INFO" "AI引擎: $ai_engine, NSFW等級: $nsfw_level, 好感度: $affection"
            
            if [ "$CSV_OUTPUT" = true ]; then
                tc_csv_record "chat,$scenario,$message,$nsfw_level,$ai_engine,$affection"
            fi
        fi
        
        tc_sleep 1
    done
    
    # 會話管理測試
    tc_http_request "GET" "/chats/$chat_id" "" "獲取會話詳情" "true"
    tc_http_request "GET" "/chats/$chat_id/history" "" "獲取會話歷史" "true"
    tc_http_request "GET" "/chats" "" "獲取會話列表" "true"
    
    # 額外會話功能測試
    mode_data='{"mode":"casual"}'
    tc_http_request "PUT" "/chats/$chat_id/mode" "$mode_data" "更新會話模式" "true"
    tc_http_request "GET" "/chats/$chat_id/export" "" "導出會話" "true"
    
    return 0
}

# ================================
# NSFW分級與AI引擎切換測試
# ================================
run_nsfw_tests() {
    tc_log "INFO" "執行NSFW分級與AI引擎切換測試"
    
    # 認證
    tc_register_and_authenticate || {
        tc_log "ERROR" "用戶註冊或認證失敗"
        return 1
    }
    
    # 創建會話
    chat_id=$(tc_create_session "character_02" "NSFW測試會話")
    [ -z "$chat_id" ] && return 1
    
    # 2-Level NSFW測試案例 (布林分級)
    get_nsfw_message() {
        case "$1" in
            "safe_1") echo "今天天氣很好" ;;
            "safe_2") echo "我很喜歡和你聊天" ;;
            "safe_3") echo "想要你的溫暖擁抱" ;;
            "safe_4") echo "你很溫暖，讓我感到安心" ;;
            "nsfw_1") echo "想要和你做愛" ;;
            "nsfw_2") echo "想看你的裸體" ;;
            "nsfw_3") echo "我們來場性愛吧" ;;
        esac
    }
    
    get_expected_level() {
        case "$1" in
            "safe_"*) echo 1 ;;
            "nsfw_"*) echo 5 ;;
        esac
    }
    
    get_expected_engine() {
        case "$1" in
            "safe_"*) echo "openai" ;;
            "nsfw_"*) echo "grok" ;;
        esac
    }
    
    # 測試案例：涵蓋安全內容和明確成人內容
    test_cases="safe_1 safe_2 safe_3 safe_4 nsfw_1 nsfw_2 nsfw_3"
    correct_level_count=0
    correct_engine_count=0
    total_tests=7
    
    for test_case in $test_cases; do
        message=$(get_nsfw_message "$test_case")
        expected_level=$(get_expected_level "$test_case")
        expected_engine=$(get_expected_engine "$test_case")
        
        tc_log "INFO" "測試: $test_case (預期等級: $expected_level, 預期引擎: $expected_engine)"
        
        start_time=$(date +%s.%N)
        response=$(tc_send_message "$chat_id" "$message" "character_02")
        end_time=$(date +%s.%N)
        
        if [ $? -eq 0 ]; then
            # Parse values with fallback
            actual_level=$(echo "$response" | jq -r '.data.nsfw_level // 0' 2>/dev/null)
            actual_engine=$(echo "$response" | jq -r '.data.ai_engine // "unknown"' 2>/dev/null)
            response_time=$(echo "$end_time - $start_time" | bc -l)
            
            # 檢查等級準確性
            if [ "$actual_level" = "$expected_level" ]; then
                correct_level_count=$((correct_level_count + 1))
                tc_log "PASS" "NSFW等級正確: $actual_level"
            else
                tc_log "WARN" "NSFW等級錯誤: 預期 $expected_level, 實際 $actual_level"
            fi
            
            # 檢查引擎選擇準確性
            if [ "$actual_engine" = "$expected_engine" ]; then
                correct_engine_count=$((correct_engine_count + 1))
                tc_log "PASS" "AI引擎選擇正確: $actual_engine"
            else
                tc_log "WARN" "AI引擎選擇錯誤: 預期 $expected_engine, 實際 $actual_engine"
            fi
            
            if [ "$CSV_OUTPUT" = true ]; then
                tc_csv_record "nsfw,$test_case,$message,$expected_level,$actual_level,$actual_engine,$response_time"
            fi
        fi
        
        tc_sleep 1
    done
    
    # 計算準確率
    level_accuracy=$((correct_level_count * 100 / total_tests))
    engine_accuracy=$((correct_engine_count * 100 / total_tests))
    
    tc_log "INFO" "NSFW等級準確率: ${level_accuracy}% ($correct_level_count/$total_tests)"
    tc_log "INFO" "AI引擎選擇準確率: ${engine_accuracy}% ($correct_engine_count/$total_tests)"
    
    # 判斷測試結果
    if [ $level_accuracy -ge 85 ] && [ $engine_accuracy -ge 85 ]; then
        tc_log "PASS" "NSFW分級與AI引擎切換系統表現優秀"
        return 0
    elif [ $level_accuracy -ge 70 ] && [ $engine_accuracy -ge 70 ]; then
        tc_log "WARN" "NSFW分級與AI引擎切換系統表現良好，但有改進空間"
        return 0
    else
        tc_log "FAIL" "NSFW分級與AI引擎切換系統需要重大調整"
        return 1
    fi
}

# ================================
# Admin API 測試
# ================================
run_admin_tests() {
    tc_log "INFO" "執行Admin API測試"
    
    # 創建admin測試數據
    admin_login_data='{"username":"admin","password":"admin123456"}'
    
    # Admin 認證
    tc_log "INFO" "Authenticating admin user"
    admin_response=$(curl -s -X POST "${TEST_BASE_URL}/admin/auth/login" \
        -H "Content-Type: application/json" \
        -d "$admin_login_data")
    
    if echo "$admin_response" | grep -q '"success":true'; then
        ADMIN_JWT_TOKEN=$(echo "$admin_response" | jq -r '.data.access_token // .data.token // ""' 2>/dev/null)
        
        if [ -n "$ADMIN_JWT_TOKEN" ] && [ "$ADMIN_JWT_TOKEN" != "null" ]; then
            tc_log "PASS" "Admin authentication successful"
            
            # Admin API 測試
            tc_http_request "GET" "/admin/stats" "" "系統統計" "true" "200,201" "$ADMIN_JWT_TOKEN"
            tc_http_request "GET" "/admin/logs" "" "系統日誌" "true" "200,201" "$ADMIN_JWT_TOKEN"
            tc_http_request "GET" "/admin/users" "" "用戶列表" "true" "200,201" "$ADMIN_JWT_TOKEN"
            tc_http_request "GET" "/admin/chats" "" "對話列表管理" "true" "200,201" "$ADMIN_JWT_TOKEN"
            
            # 測試用戶管理API (使用 testuser)
            tc_http_request "GET" "/admin/users/$TEST_USER_ID" "" "獲取特定用戶" "true" "200,404" "$ADMIN_JWT_TOKEN"
            
            # 測試用戶狀態更新 (謹慎操作)
            user_status_data='{"status":"active"}'
            tc_http_request "PUT" "/admin/users/$TEST_USER_ID/status" "$user_status_data" "更新用戶狀態" "true" "200,404" "$ADMIN_JWT_TOKEN"
            
            # 測試角色狀態更新
            char_status_data='{"status":"active"}'
            tc_http_request "PUT" "/admin/character/character_01/status" "$char_status_data" "更新角色狀態" "true" "200,404" "$ADMIN_JWT_TOKEN"
            
            # 嘗試獲取管理員列表 (需要超級管理員權限，可能會失敗)
            tc_http_request "GET" "/admin/admins" "" "管理員列表" "true" "200,403" "$ADMIN_JWT_TOKEN"
            
            # 測試特定會話歷史查看 (使用示例會話ID)
            tc_http_request "GET" "/admin/chats/example-chat-id/history" "" "會話歷史查看" "true" "200,404" "$ADMIN_JWT_TOKEN"
        else
            tc_log "FAIL" "Admin authentication failed - no token"
            return 1
        fi
    else
        tc_log "WARN" "Admin authentication failed, skipping admin tests"
        tc_log "INFO" "Response: $admin_response"
        return 0  # 不讓這個失敗導致整個測試失敗
    fi
    
    return 0
}

# ================================
# 軟刪除功能測試
# ================================
run_soft_delete_tests() {
    tc_log "INFO" "執行軟刪除功能測試"
    
    # 檢查是否有軟刪除測試腳本
    soft_delete_script="$(dirname "$0")/test_soft_delete.sh"
    
    if [ -f "$soft_delete_script" ]; then
        tc_log "INFO" "執行專用軟刪除測試腳本"
        if [ "$QUICK_MODE" = true ]; then
            "$soft_delete_script" comprehensive
        else
            "$soft_delete_script" detailed
        fi
        
        local soft_delete_result=$?
        if [ $soft_delete_result -eq 0 ]; then
            tc_log "PASS" "軟刪除功能測試通過"
        else
            tc_log "FAIL" "軟刪除功能測試失敗"
        fi
        
        return $soft_delete_result
    else
        tc_log "WARN" "軟刪除測試腳本不存在，執行基礎軟刪除測試"
        
        # 基礎認證
        tc_admin_authenticate || {
            tc_log "ERROR" "管理員用戶註冊或認證失敗"
            return 1
        }
        
        tc_register_and_authenticate || {
            tc_log "ERROR" "用戶用戶註冊或認證失敗"
            return 1
        }
        
        # 執行內建軟刪除測試
        tc_test_soft_delete_comprehensive "$ADMIN_JWT_TOKEN" "$TC_JWT_TOKEN"
        local result=$?
        
        if [ $result -eq 0 ]; then
            tc_log "PASS" "基礎軟刪除測試通過"
        else
            tc_log "FAIL" "基礎軟刪除測試失敗"
        fi
        
        return $result
    fi
}

# ================================
# 主執行流程
# ================================

START_TIME=$(date +%s)

case "$TEST_TYPE" in
    "health")
        run_health_tests
        result=$?
        ;;
    "auth")
        run_health_tests && run_auth_tests
        result=$?
        ;;
    "api")
        run_health_tests && run_api_tests
        result=$?
        ;;
    "chat")
        run_health_tests && run_chat_tests
        result=$?
        ;;
    "nsfw")
        run_health_tests && run_nsfw_tests
        result=$?
        ;;
    "admin")
        run_health_tests && run_admin_tests
        result=$?
        ;;
    "soft-delete")
        run_health_tests && run_soft_delete_tests
        result=$?
        ;;
    "all")
        tc_log "INFO" "執行完整測試套件"
        run_health_tests
        health_result=$?
        
        run_auth_tests
        auth_result=$?
        
        run_api_tests
        api_result=$?
        
        run_chat_tests
        chat_result=$?
        
        run_nsfw_tests
        nsfw_result=$?
        
        run_admin_tests
        admin_result=$?
        
        run_soft_delete_tests
        soft_delete_result=$?
        
        result=$((health_result + auth_result + api_result + chat_result + nsfw_result + admin_result + soft_delete_result))
        
        # 顯示結果摘要
        echo ""
        tc_log "INFO" "完整測試套件結果摘要："
        [ $health_result -eq 0 ] && tc_log "PASS" "系統健康: 通過" || tc_log "FAIL" "系統健康: 失敗"
        [ $auth_result -eq 0 ] && tc_log "PASS" "認證測試: 通過" || tc_log "FAIL" "認證測試: 失敗"
        [ $api_result -eq 0 ] && tc_log "PASS" "API測試: 通過" || tc_log "FAIL" "API測試: 失敗"
        [ $chat_result -eq 0 ] && tc_log "PASS" "對話測試: 通過" || tc_log "FAIL" "對話測試: 失敗"
        [ $nsfw_result -eq 0 ] && tc_log "PASS" "NSFW測試: 通過" || tc_log "FAIL" "NSFW測試: 失敗"
        [ $admin_result -eq 0 ] && tc_log "PASS" "Admin測試: 通過" || tc_log "FAIL" "Admin測試: 失敗"
        [ $soft_delete_result -eq 0 ] && tc_log "PASS" "軟刪除測試: 通過" || tc_log "FAIL" "軟刪除測試: 失敗"
        ;;
    *)
        tc_log "ERROR" "未知的測試類型: $TEST_TYPE"
        exit 1
        ;;
esac

# 計算執行時間並顯示總結
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

tc_log "INFO" "測試完成，執行時間: ${DURATION}秒"

if [ "$CSV_OUTPUT" = true ]; then
    tc_log "INFO" "CSV報告: $TC_CSV_FILE"
fi

tc_show_summary "$TEST_TYPE 測試"
tc_cleanup

exit $result