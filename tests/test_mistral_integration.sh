#!/bin/bash

# 測試 Mistral 三層 AI 引擎架構
# 使用共用測試工具庫

# 載入測試工具
source "$(dirname "$0")/utils/test_common.sh"

echo "🧪 Mistral 三層 AI 引擎架構測試"
echo "================================="

# 初始化測試環境
tc_init_logging "mistral_integration"

# 函數：創建測試會話
create_test_chat() {
    tc_log "INFO" "創建測試會話..."

    local response=$(tc_http_request "POST" "/chats" '{
        "character_id": "character_02",
        "title": "Mistral 三層引擎測試"
    }' "Create Test Chat")

    TC_CHAT_ID=$(echo "$response" | jq -r '.data.id // ""')

    if [ -z "$TC_CHAT_ID" ] || [ "$TC_CHAT_ID" = "null" ]; then
        tc_log "FAIL" "會話創建失敗"
        return 1
    fi
    tc_log "PASS" "會話創建成功 (ID: $TC_CHAT_ID)"
    return 0
}

# 函數：測試訊息並檢查引擎選擇
test_message() {
    local message="$1"
    local expected_level="$2"
    local expected_engine="$3"
    local test_name="$4"

    tc_log "INFO" "🔍 測試: $test_name"
    tc_log "INFO" "訊息: $message"
    tc_log "INFO" "預期等級: L$expected_level | 預期引擎: $expected_engine"

    local response=$(tc_http_request "POST" "/chats/$TC_CHAT_ID/messages" "{\"message\": \"$message\"}" "$test_name")

    local actual_engine=$(echo "$response" | jq -r '.data.ai_engine // ""')
    local actual_level=$(echo "$response" | jq -r '.data.nsfw_level // ""')

    if [ "$actual_engine" = "$expected_engine" ] && [ "$actual_level" = "$expected_level" ]; then
        tc_log "PASS" "引擎=$actual_engine，等級=L$actual_level"
        return 0
    else
        tc_log "FAIL" "預期引擎=$expected_engine/等級=L$expected_level，實際引擎=$actual_engine/等級=L$actual_level"
        return 1
    fi
}

# 函數：清理測試會話
cleanup_chat() {
    if [ -n "$TC_CHAT_ID" ] && [ "$TC_CHAT_ID" != "null" ]; then
        tc_log "INFO" "清理測試會話..."
        tc_http_request "DELETE" "/chats/$TC_CHAT_ID" '{}' "Delete Test Chat" > /dev/null
        tc_log "PASS" "測試會話已清理"
    fi
}

# 主測試流程
main() {
    tc_log "INFO" "開始 Mistral 三層 AI 引擎架構測試"

    # 使用預設測試用戶登入
    if ! tc_authenticate; then
        tc_log "FAIL" "用戶驗證失敗，結束測試"
        exit 1
    fi

    # 創建測試會話
    if ! create_test_chat; then
        tc_log "FAIL" "測試會話創建失敗，結束測試"
        exit 1
    fi

    tc_log "INFO" "🎯 開始三層 AI 引擎測試..."

    local total_tests=0
    local passed_tests=0

    # L1 測試 (OpenAI)
    total_tests=$((total_tests + 1))
    if test_message "今天天氣真好！" "1" "openai" "L1 - 安全內容"; then
        passed_tests=$((passed_tests + 1))
    fi

    # L2 測試 (Mistral)
    total_tests=$((total_tests + 1))
    if test_message "你的身材真好" "2" "mistral" "L2 - 身體描述"; then
        passed_tests=$((passed_tests + 1))
    fi

    # L3 測試 (Mistral)
    total_tests=$((total_tests + 1))
    if test_message "我想看你的裸體" "3" "mistral" "L3 - 裸體相關"; then
        passed_tests=$((passed_tests + 1))
    fi

    # L4 測試 (Grok)
    total_tests=$((total_tests + 1))
    if test_message "我想要口交" "4" "grok" "L4 - 明確性行為"; then
        passed_tests=$((passed_tests + 1))
    fi

    # L5 測試 (Grok)
    total_tests=$((total_tests + 1))
    if test_message "我要強姦你" "5" "grok" "L5 - 性暴力內容"; then
        passed_tests=$((passed_tests + 1))
    fi

    # 清理
    cleanup_chat

    # 測試結果摘要
    tc_log "INFO" "🎉 Mistral 三層 AI 引擎架構測試完成！"
    tc_log "INFO" "測試結果：$passed_tests/$total_tests 通過"

    if [ "$passed_tests" -eq "$total_tests" ]; then
        tc_log "PASS" "所有測試都通過！"
        exit 0
    else
        tc_log "FAIL" "有 $((total_tests - passed_tests)) 個測試失敗"
        exit 1
    fi
}

# 執行主函數
main "$@"