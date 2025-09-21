#!/bin/bash

# 🧪 Thewavess AI Core - 全域搜索功能測試
# 測試全域搜索相關API功能

set -e

# 載入測試工具庫
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/test_common.sh"

# ================================
# 測試配置
# ================================

TEST_NAME="search"
TEST_CHAT_SESSION_ID=""
TEST_CHARACTER_ID_FOR_SEARCH=""

# ================================
# 測試函數
# ================================

# 準備搜索測試數據
setup_search_data() {
    tc_log "INFO" "準備搜索測試數據"

    # 創建測試聊天會話
    local session_data='{"character_id":"'$TEST_CHARACTER_ID'","title":"搜索測試會話_特殊關鍵字_技術討論"}'

    local response=$(tc_http_request "POST" "/chats" "$session_data" "Create Search Test Chat" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        TEST_CHAT_SESSION_ID=$(echo "$response" | jq -r '.data.id // ""')
        tc_log "PASS" "搜索測試會話創建成功 (ID: $TEST_CHAT_SESSION_ID)"
    else
        tc_log "FAIL" "搜索測試會話創建失敗"
        return 1
    fi

    # 發送幾條具有關鍵字的測試消息
    local test_messages=(
        "我正在學習人工智能和機器學習"
        "今天討論了深度學習和神經網絡"
        "區塊鏈技術很有趣，特別是智能合約"
        "Python編程在數據科學中很重要"
        "雲計算和容器化技術正在改變軟件開發"
    )

    for message in "${test_messages[@]}"; do
        local message_data='{"message": "'$message'"}'
        local msg_response=$(tc_http_request "POST" "/chats/$TEST_CHAT_SESSION_ID/messages" "$message_data" "Send Search Test Message" "true")

        if echo "$msg_response" | jq -e '.success' > /dev/null 2>&1; then
            tc_log "INFO" "測試消息發送成功: $message"
            sleep 2  # 等待AI回應
        else
            tc_log "WARN" "測試消息發送失敗: $message"
        fi
    done

    tc_log "PASS" "搜索測試數據準備完成"
    return 0
}

# 測試全域聊天搜索
test_global_chat_search() {
    tc_log "INFO" "測試全域聊天搜索"

    # 測試不同關鍵字搜索
    local search_terms=("學習" "技術" "Python" "人工智能")

    for term in "${search_terms[@]}"; do
        tc_log "INFO" "搜索關鍵字: $term"

        local response=$(tc_http_request "GET" "/search/chats?q=$term" "" "Search Chats: $term" "true")

        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            local results_count=$(echo "$response" | jq -r '.data.results | length')
            local total_count=$(echo "$response" | jq -r '.data.total_count // 0')
            local has_highlights=$(echo "$response" | jq -r '.data.results[0].highlights // [] | length')

            tc_log "PASS" "聊天搜索成功 - 關鍵字: $term"
            tc_log "INFO" "  結果數量: $results_count"
            tc_log "INFO" "  總匹配數: $total_count"
            tc_log "INFO" "  高亮片段: $has_highlights"

            # 檢查是否找到我們的測試會話
            local found_test_session=$(echo "$response" | jq -r --arg chat_id "$TEST_CHAT_SESSION_ID" '.data.results[] | select(.id == $chat_id) | .id')

            if [ -n "$found_test_session" ]; then
                tc_log "PASS" "找到測試會話，搜索功能正常"
            else
                tc_log "INFO" "未找到測試會話，可能關鍵字不匹配"
            fi
        else
            tc_log "FAIL" "聊天搜索失敗 - 關鍵字: $term"
            return 1
        fi

        sleep 1
    done

    return 0
}

# 測試全域搜索（包含角色）
test_global_search() {
    tc_log "INFO" "測試全域搜索"

    # 搜索包含聊天和角色的全域搜索
    local search_terms=("測試" "學習" "人工智能")

    for term in "${search_terms[@]}"; do
        tc_log "INFO" "全域搜索關鍵字: $term"

        local response=$(tc_http_request "GET" "/search/global?q=$term" "" "Global Search: $term" "true")

        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            local total_results=$(echo "$response" | jq -r '.data.total_results // 0')
            local chat_count=$(echo "$response" | jq -r '.data.results.chats.count // 0')
            local character_count=$(echo "$response" | jq -r '.data.results.characters.count // 0')

            tc_log "PASS" "全域搜索成功 - 關鍵字: $term"
            tc_log "INFO" "  總結果數: $total_results"
            tc_log "INFO" "  聊天結果: $chat_count"
            tc_log "INFO" "  角色結果: $character_count"

            # 檢查聊天結果
            if [ "$chat_count" -gt 0 ]; then
                local first_chat_title=$(echo "$response" | jq -r '.data.results.chats.results[0].title // ""')
                tc_log "INFO" "  第一個聊天: $first_chat_title"
            fi

            # 檢查角色結果
            if [ "$character_count" -gt 0 ]; then
                local first_character_name=$(echo "$response" | jq -r '.data.results.characters.results[0].name // ""')
                tc_log "INFO" "  第一個角色: $first_character_name"
            fi
        else
            tc_log "FAIL" "全域搜索失敗 - 關鍵字: $term"
            return 1
        fi

        sleep 1
    done

    return 0
}

# 測試分頁搜索
test_paginated_search() {
    tc_log "INFO" "測試分頁搜索"

    # 測試第一頁
    local response_page1=$(tc_http_request "GET" "/search/chats?q=測試&limit=5&offset=0" "" "Search with Pagination Page 1" "true")

    if echo "$response_page1" | jq -e '.success' > /dev/null 2>&1; then
        local page1_count=$(echo "$response_page1" | jq -r '.data.results | length')
        local total_count=$(echo "$response_page1" | jq -r '.data.total_count // 0')

        tc_log "PASS" "分頁搜索第一頁成功"
        tc_log "INFO" "  第一頁結果: $page1_count"
        tc_log "INFO" "  總結果數: $total_count"

        # 如果總結果大於5，測試第二頁
        if [ "$total_count" -gt 5 ]; then
            local response_page2=$(tc_http_request "GET" "/search/chats?q=測試&limit=5&offset=5" "" "Search with Pagination Page 2" "true")

            if echo "$response_page2" | jq -e '.success' > /dev/null 2>&1; then
                local page2_count=$(echo "$response_page2" | jq -r '.data.results | length')
                tc_log "PASS" "分頁搜索第二頁成功"
                tc_log "INFO" "  第二頁結果: $page2_count"
            else
                tc_log "FAIL" "分頁搜索第二頁失敗"
                return 1
            fi
        else
            tc_log "INFO" "總結果數不足，跳過第二頁測試"
        fi
    else
        tc_log "FAIL" "分頁搜索失敗"
        return 1
    fi

    return 0
}

# 測試高級搜索篩選
test_advanced_search_filters() {
    tc_log "INFO" "測試高級搜索篩選"

    # 測試日期範圍篩選
    local today=$(date "+%Y-%m-%d")
    local yesterday=$(date -d "yesterday" "+%Y-%m-%d" 2>/dev/null || date -j -v-1d "+%Y-%m-%d" 2>/dev/null || echo "2024-01-01")

    local response=$(tc_http_request "GET" "/search/chats?q=測試&date_from=$yesterday&date_to=$today" "" "Search with Date Filter" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local filtered_count=$(echo "$response" | jq -r '.data.results | length')
        tc_log "PASS" "日期篩選搜索成功"
        tc_log "INFO" "  篩選後結果: $filtered_count"
    else
        tc_log "FAIL" "日期篩選搜索失敗"
        return 1
    fi

    # 測試角色篩選
    if [ -n "$TEST_CHARACTER_ID" ]; then
        local char_response=$(tc_http_request "GET" "/search/chats?q=測試&character_id=$TEST_CHARACTER_ID" "" "Search with Character Filter" "true")

        if echo "$char_response" | jq -e '.success' > /dev/null 2>&1; then
            local char_filtered_count=$(echo "$char_response" | jq -r '.data.results | length')
            tc_log "PASS" "角色篩選搜索成功"
            tc_log "INFO" "  角色篩選後結果: $char_filtered_count"
        else
            tc_log "FAIL" "角色篩選搜索失敗"
            return 1
        fi
    fi

    return 0
}

# 測試搜索篩選功能
test_search_with_filters() {
    tc_log "INFO" "測試搜索篩選功能"

    # 測試不同的篩選條件
    local response=$(tc_http_request "GET" "/search/chats?q=測試&limit=5" "" "Search with Limit Filter" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local results_count=$(echo "$response" | jq -r '.data.results | length')
        tc_log "PASS" "限制數量篩選成功"
        tc_log "INFO" "  返回結果數: $results_count (應該 <= 5)"

        # 驗證返回結果不超過限制
        if [ "$results_count" -le 5 ]; then
            tc_log "PASS" "結果數量符合限制"
        else
            tc_log "FAIL" "結果數量超過限制"
            return 1
        fi
    else
        tc_log "FAIL" "搜索篩選失敗"
        return 1
    fi

    return 0
}

# 測試搜索性能
test_search_performance() {
    tc_log "INFO" "測試搜索性能"

    local search_term="測試"
    local start_time=$(date +%s%3N)

    local response=$(tc_http_request "GET" "/search/chats?q=$search_term" "" "Search Performance Test" "true")

    local end_time=$(date +%s%3N)
    local duration=$((end_time - start_time))

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local results_count=$(echo "$response" | jq -r '.data.results | length')
        tc_log "PASS" "搜索性能測試完成"
        tc_log "INFO" "  搜索時間: ${duration}ms"
        tc_log "INFO" "  結果數量: $results_count"

        # 性能基準：搜索應在3秒內完成
        if [ "$duration" -lt 3000 ]; then
            tc_log "PASS" "搜索性能良好 (< 3s)"
        else
            tc_log "WARN" "搜索性能較慢 (> 3s)"
        fi
    else
        tc_log "FAIL" "搜索性能測試失敗"
        return 1
    fi

    return 0
}

# 清理測試數據
cleanup_search_test() {
    tc_log "INFO" "清理搜索測試數據"

    if [ -n "$TEST_CHAT_SESSION_ID" ]; then
        local response=$(tc_http_request "DELETE" "/chats/$TEST_CHAT_SESSION_ID" "" "Delete Search Test Chat" "true")

        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            tc_log "PASS" "搜索測試會話清理成功"
        else
            tc_log "WARN" "搜索測試會話清理失敗，請手動清理 ID: $TEST_CHAT_SESSION_ID"
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
    tc_show_header "Thewavess AI Core - 全域搜索功能測試"

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
    tc_log "INFO" "==================== 設置搜索測試環境 ===================="

    if ! setup_search_data; then
        tc_log "ERROR" "搜索測試環境設置失敗"
        exit 1
    fi

    sleep 5  # 等待數據索引

    # 執行搜索功能測試
    tc_log "INFO" "==================== 搜索功能測試 ===================="

    if test_global_chat_search; then
        test_results+=("聊天搜索:PASS")
    else
        test_results+=("聊天搜索:FAIL")
    fi

    sleep 2

    if test_global_search; then
        test_results+=("全域搜索:PASS")
    else
        test_results+=("全域搜索:FAIL")
    fi

    sleep 2

    if test_paginated_search; then
        test_results+=("分頁搜索:PASS")
    else
        test_results+=("分頁搜索:FAIL")
    fi

    sleep 2

    if test_advanced_search_filters; then
        test_results+=("高級篩選:PASS")
    else
        test_results+=("高級篩選:FAIL")
    fi

    sleep 2

    if test_search_with_filters; then
        test_results+=("搜索篩選:PASS")
    else
        test_results+=("搜索篩選:FAIL")
    fi

    sleep 2

    if test_search_performance; then
        test_results+=("搜索性能:PASS")
    else
        test_results+=("搜索性能:FAIL")
    fi

    # 清理測試數據
    cleanup_search_test

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
        tc_log "PASS" "所有搜索功能測試通過！"
        tc_show_summary "搜索功能測試"
        exit 0
    else
        tc_log "FAIL" "$failed_count 個測試失敗"
        tc_show_summary "搜索功能測試"
        exit 1
    fi
}

# 執行主函數
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi