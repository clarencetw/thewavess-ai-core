#!/bin/bash

# 測試雙引擎 AI 架構 (OpenAI + Grok)
# 使用共用測試工具庫

# 載入測試工具
source "$(dirname "$0")/utils/test_common.sh"

echo "🧪 雙引擎 AI 架構測試 (OpenAI + Grok)"
echo "======================================="

# 初始化測試環境
tc_init_logging "dual_engine"

# 函數：創建測試會話
create_test_chat() {
    tc_log "INFO" "創建測試會話..."

    local response=$(tc_http_request "POST" "/chats" '{
        "character_id": "character_02",
        "title": "雙引擎 AI 測試"
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
        tc_log "PASS" "引擎選擇正確: $actual_engine (L$actual_level)"
    else
        tc_log "FAIL" "引擎選擇錯誤 - 實際: $actual_engine (L$actual_level), 預期: $expected_engine (L$expected_level)"
        return 1
    fi

    return 0
}

# 主測試流程
run_tests() {
    tc_log "INFO" "==================== 雙引擎 AI 架構測試 ===================="

    # 檢查服務器健康狀態
    if ! tc_check_health; then
        tc_log "ERROR" "服務器健康檢查失敗"
        return 1
    fi

    # 用戶註冊和認證
    if ! tc_register_and_authenticate; then
        tc_log "ERROR" "用戶註冊或認證失敗"
        return 1
    fi

    # 創建測試會話
    if ! create_test_chat; then
        tc_log "ERROR" "無法創建測試會話"
        return 1
    fi

    # 測試不同 NSFW 等級的訊息路由
    tc_log "INFO" "==================== 測試 OpenAI (L1-L3) ===================="

    # L1 安全內容 -> OpenAI
    test_message "今天天氣真好，我們去公園散步吧" "1" "openai" "L1 安全內容"
    sleep 2

    # L2 輕度內容 -> OpenAI
    test_message "你今天看起來很性感" "2" "openai" "L2 輕度內容"
    sleep 2

    # L3 中度內容 -> OpenAI
    test_message "我想看你的裸體照片" "3" "openai" "L3 中度內容"
    sleep 2

    tc_log "INFO" "==================== 測試 Grok (L4-L5) ===================="

    # L4 露骨內容 -> Grok
    test_message "我想要和你做愛，感受你的身體" "4" "grok" "L4 露骨內容"
    sleep 2

    # L5 極度露骨 -> Grok
    test_message "我要插入你的陰道，讓你高潮" "5" "grok" "L5 極度露骨"
    sleep 2

    tc_log "INFO" "==================== 測試 Sticky Session ===================="

    # 測試 sticky session（L4+ 後維持 Grok 5分鐘）
    tc_log "INFO" "測試 sticky session 機制..."
    test_message "你好，今天過得怎麼樣？" "1" "grok" "Sticky Session 測試（應該還是 Grok）"

    tc_log "PASS" "雙引擎 AI 架構測試完成"

    # 清理測試會話
    if [ -n "$TC_CHAT_ID" ]; then
        tc_http_request "DELETE" "/chats/$TC_CHAT_ID" "" "Delete Test Chat"
    fi

    return 0
}

# 執行主函數
main() {
    # CSV功能已移除，改用詳細日誌記錄
    tc_show_header "雙引擎 AI 架構測試"

    # 檢查依賴
    if ! tc_check_dependencies; then
        tc_log "ERROR" "依賴檢查失敗"
        exit 1
    fi

    # 執行測試
    if run_tests; then
        tc_log "PASS" "所有雙引擎測試通過！"
        tc_show_summary "雙引擎 AI 架構測試"
        exit 0
    else
        tc_log "FAIL" "部分測試失敗"
        tc_show_summary "雙引擎 AI 架構測試"
        exit 1
    fi
}

if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi