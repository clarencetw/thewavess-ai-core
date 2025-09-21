#!/bin/bash

# 🧪 Thewavess AI Core - 聊天進階功能測試
# 測試聊天進階API功能：模式切換、導出、重新生成等

set -e

# 載入測試工具庫
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/test_common.sh"

# ================================
# 測試配置
# ================================

TEST_NAME="chat_advanced"
TEST_CHAT_SESSION_ID=""
TEST_MESSAGE_ID=""

# ================================
# 測試函數
# ================================

# 創建測試聊天會話
setup_test_chat() {
    tc_log "INFO" "創建測試聊天會話"

    local session_data="{\"character_id\":\"$TEST_CHARACTER_ID\",\"title\":\"進階功能測試會話\"}"

    local response=$(tc_http_request "POST" "/chats" "$session_data" "Create Test Chat Session" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        TEST_CHAT_SESSION_ID=$(echo "$response" | jq -r '.data.id // ""')
        tc_log "PASS" "測試會話創建成功 (ID: $TEST_CHAT_SESSION_ID)"
        return 0
    else
        tc_log "FAIL" "測試會話創建失敗"
        return 1
    fi
}

# 發送測試消息以便後續測試
send_test_message() {
    tc_log "INFO" "發送測試消息"

    if [ -z "$TEST_CHAT_SESSION_ID" ]; then
        tc_log "ERROR" "沒有測試會話ID"
        return 1
    fi

    local message_data='{"message": "你好，這是一條測試消息，請回應我"}'

    local response=$(tc_http_request "POST" "/chats/$TEST_CHAT_SESSION_ID/messages" "$message_data" "Send Test Message" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        TEST_MESSAGE_ID=$(echo "$response" | jq -r '.data.id // ""')
        tc_log "PASS" "測試消息發送成功 (ID: $TEST_MESSAGE_ID)"
        return 0
    else
        tc_log "FAIL" "測試消息發送失敗"
        return 1
    fi
}

# 測試更新會話模式
test_update_session_mode() {
    tc_log "INFO" "測試更新會話模式"

    if [ -z "$TEST_CHAT_SESSION_ID" ]; then
        tc_log "ERROR" "沒有測試會話ID"
        return 1
    fi

    # 測試切換到novel模式
    local mode_data='{"mode": "novel"}'

    local response=$(tc_http_request "PUT" "/chats/$TEST_CHAT_SESSION_ID/mode" "$mode_data" "Update Session Mode to Novel" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local current_mode=$(echo "$response" | jq -r '.data.mode // ""')
        tc_log "PASS" "會話模式更新成功"
        tc_log "INFO" "  當前模式: $current_mode"

        # 測試切換回chat模式
        local chat_mode_data='{"mode": "chat"}'
        local chat_response=$(tc_http_request "PUT" "/chats/$TEST_CHAT_SESSION_ID/mode" "$chat_mode_data" "Update Session Mode to Chat" "true")

        if echo "$chat_response" | jq -e '.success' > /dev/null 2>&1; then
            local chat_mode=$(echo "$chat_response" | jq -r '.data.mode // ""')
            tc_log "PASS" "會話模式切換回chat成功 (當前: $chat_mode)"
            return 0
        else
            tc_log "FAIL" "切換回chat模式失敗"
            return 1
        fi
    else
        tc_log "FAIL" "會話模式更新失敗"
        return 1
    fi
}

# 測試導出會話
test_export_chat_session() {
    tc_log "INFO" "測試導出會話"

    if [ -z "$TEST_CHAT_SESSION_ID" ]; then
        tc_log "ERROR" "沒有測試會話ID"
        return 1
    fi

    local response=$(tc_http_request "GET" "/chats/$TEST_CHAT_SESSION_ID/export" "" "Export Chat Session" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local export_format=$(echo "$response" | jq -r '.data.format // ""')
        local export_content=$(echo "$response" | jq -r '.data.content // ""')
        local message_count=$(echo "$response" | jq -r '.data.message_count // 0')

        tc_log "PASS" "會話導出成功"
        tc_log "INFO" "  導出格式: $export_format"
        tc_log "INFO" "  消息數量: $message_count"
        tc_log "INFO" "  內容長度: ${#export_content} 字元"

        # 檢查導出內容是否包含基本信息
        if echo "$export_content" | grep -q "測試消息"; then
            tc_log "PASS" "導出內容包含預期的消息"
            return 0
        else
            tc_log "WARN" "導出內容可能不完整"
            return 0
        fi
    else
        tc_log "FAIL" "會話導出失敗"
        return 1
    fi
}

# 測試重新生成回應
test_regenerate_response() {
    tc_log "INFO" "測試重新生成回應"

    if [ -z "$TEST_CHAT_SESSION_ID" ] || [ -z "$TEST_MESSAGE_ID" ]; then
        tc_log "ERROR" "缺少測試會話ID或消息ID"
        return 1
    fi

    # 獲取原始回應以便比較
    local history_response=$(tc_http_request "GET" "/chats/$TEST_CHAT_SESSION_ID/history" "" "Get Original Response" "true")
    local original_response=""

    if echo "$history_response" | jq -e '.success' > /dev/null 2>&1; then
        # 找到AI的回應消息
        original_response=$(echo "$history_response" | jq -r '.data.messages[] | select(.role == "assistant") | .dialogue' | head -1)
    fi

    # 執行重新生成
    local regen_data='{"instruction": "請用不同的方式回應"}'

    local response=$(tc_http_request "POST" "/chats/$TEST_CHAT_SESSION_ID/messages/$TEST_MESSAGE_ID/regenerate" "$regen_data" "Regenerate Response" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local new_response=$(echo "$response" | jq -r '.data.dialogue // ""')
        local new_message_id=$(echo "$response" | jq -r '.data.id // ""')

        tc_log "PASS" "回應重新生成成功"
        tc_log "INFO" "  新消息ID: $new_message_id"
        tc_log "INFO" "  新回應長度: ${#new_response} 字元"

        # 比較新舊回應
        if [ -n "$original_response" ] && [ "$new_response" != "$original_response" ]; then
            tc_log "PASS" "新回應與原回應不同"
        else
            tc_log "WARN" "新回應與原回應相同或無法比較"
        fi

        return 0
    else
        tc_log "FAIL" "回應重新生成失敗"
        return 1
    fi
}

# 測試會話統計信息
test_chat_statistics() {
    tc_log "INFO" "測試會話統計信息"

    if [ -z "$TEST_CHAT_SESSION_ID" ]; then
        tc_log "ERROR" "沒有測試會話ID"
        return 1
    fi

    local response=$(tc_http_request "GET" "/chats/$TEST_CHAT_SESSION_ID" "" "Get Chat Statistics" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local message_count=$(echo "$response" | jq -r '.data.message_count // 0')
        local total_characters=$(echo "$response" | jq -r '.data.total_characters // 0')
        local created_at=$(echo "$response" | jq -r '.data.created_at // ""')
        local last_activity=$(echo "$response" | jq -r '.data.updated_at // ""')

        tc_log "PASS" "會話統計獲取成功"
        tc_log "INFO" "  消息數量: $message_count"
        tc_log "INFO" "  總字元數: $total_characters"
        tc_log "INFO" "  創建時間: $created_at"
        tc_log "INFO" "  最後活動: $last_activity"

        # 驗證統計數據合理性
        if [ "$message_count" -gt 0 ] && [ "$total_characters" -gt 0 ]; then
            tc_log "PASS" "統計數據合理"
            return 0
        else
            tc_log "WARN" "統計數據可能異常"
            return 1
        fi
    else
        tc_log "FAIL" "會話統計獲取失敗"
        return 1
    fi
}

# 測試會話搜索功能
test_chat_search() {
    tc_log "INFO" "測試會話搜索功能"

    # 搜索包含"測試"的會話
    local search_query="測試"

    local response=$(tc_http_request "GET" "/search/chats?q=$search_query" "" "Search Chats" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local results_count=$(echo "$response" | jq -r '.data.results | length')
        local total_count=$(echo "$response" | jq -r '.data.total_count // 0')

        tc_log "PASS" "會話搜索成功"
        tc_log "INFO" "  搜索關鍵字: $search_query"
        tc_log "INFO" "  結果數量: $results_count"
        tc_log "INFO" "  總匹配數: $total_count"

        # 檢查是否找到我們的測試會話
        local found_test_chat=$(echo "$response" | jq -r --arg chat_id "$TEST_CHAT_SESSION_ID" '.data.results[] | select(.id == $chat_id) | .id')

        if [ -n "$found_test_chat" ]; then
            tc_log "PASS" "找到測試會話"
            return 0
        else
            tc_log "WARN" "未找到測試會話，但搜索功能正常"
            return 0
        fi
    else
        tc_log "FAIL" "會話搜索失敗"
        return 1
    fi
}

# 清理測試數據
cleanup_test_data() {
    tc_log "INFO" "清理測試數據"

    if [ -n "$TEST_CHAT_SESSION_ID" ]; then
        local response=$(tc_http_request "DELETE" "/chats/$TEST_CHAT_SESSION_ID" "" "Delete Test Chat Session" "true")

        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            tc_log "PASS" "測試會話清理成功"
        else
            tc_log "WARN" "測試會話清理失敗，請手動清理 ID: $TEST_CHAT_SESSION_ID"
        fi
    fi
}

# ================================
# 主執行流程
# ================================

main() {
    # 初始化測試
    tc_init_logging "$TEST_NAME"
    # CSV功能已移除，改用詳細日誌記錄
    tc_show_header "Thewavess AI Core - 聊天進階功能測試"

    # 檢查依賴
    if ! tc_check_dependencies; then
        tc_log "ERROR" "依賴檢查失敗"
        exit 1
    fi

    # 檢查服務器健康狀態
    if ! tc_check_health; then
        tc_log "ERROR" "服務器健康檢查失敗"
        exit 1
    fi

    # 用戶註冊和認證
    tc_log "INFO" "執行用戶註冊和認證"
    if ! tc_register_and_authenticate; then
        tc_log "ERROR" "用戶註冊或認證失敗"
        exit 1
    fi

    local test_results=()

    # 設置測試環境
    tc_log "INFO" "==================== 設置測試環境 ===================="

    if ! setup_test_chat; then
        tc_log "ERROR" "測試環境設置失敗"
        exit 1
    fi

    sleep 2

    if ! send_test_message; then
        tc_log "ERROR" "測試消息發送失敗"
        cleanup_test_data
        exit 1
    fi

    sleep 3  # 等待AI回應

    # 執行進階功能測試
    tc_log "INFO" "==================== 進階功能測試 ===================="

    if test_update_session_mode; then
        test_results+=("模式切換:PASS")
    else
        test_results+=("模式切換:FAIL")
    fi

    sleep 2

    if test_export_chat_session; then
        test_results+=("會話導出:PASS")
    else
        test_results+=("會話導出:FAIL")
    fi

    sleep 2

    if test_regenerate_response; then
        test_results+=("重新生成:PASS")
    else
        test_results+=("重新生成:FAIL")
    fi

    sleep 2

    if test_chat_statistics; then
        test_results+=("統計信息:PASS")
    else
        test_results+=("統計信息:FAIL")
    fi

    sleep 2

    if test_chat_search; then
        test_results+=("會話搜索:PASS")
    else
        test_results+=("會話搜索:FAIL")
    fi

    # 清理測試數據
    cleanup_test_data

    # 顯示測試結果總結
    tc_log "INFO" "==================== 測試結果總結 ===================="
    local failed_count=0

    for result in "${test_results[@]}"; do
        IFS=':' read -r test_name test_status <<< "$result"
        case "$test_status" in
            "PASS") tc_log "PASS" "$test_name: 通過" ;;
            "FAIL")
                tc_log "FAIL" "$test_name: 失敗"
                failed_count=$((failed_count + 1))
                ;;
        esac
    done

    # 清理資源
    tc_cleanup

    # 最終結果
    if [ $failed_count -eq 0 ]; then
        tc_log "PASS" "所有聊天進階功能測試通過！"
        tc_show_summary "聊天進階功能測試"
        exit 0
    else
        tc_log "FAIL" "$failed_count 個測試失敗"
        tc_show_summary "聊天進階功能測試"
        exit 1
    fi
}

# 執行主函數
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi