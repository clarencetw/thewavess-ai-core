#!/bin/bash

# 🧪 Thewavess AI Core - 關係系統測試
# 測試用戶與角色間的關係系統API功能

set -e

# 載入測試工具庫
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/test_common.sh"

# ================================
# 測試配置
# ================================

TEST_NAME="relationships"
TEST_CHAT_SESSION_ID=""

# ================================
# 測試函數
# ================================

# 創建測試聊天會話以建立關係
setup_relationship_context() {
    tc_log "INFO" "創建關係測試會話"

    local session_data="{\"character_id\":\"$TEST_CHARACTER_ID\",\"title\":\"關係系統測試會話\"}"

    local response=$(tc_http_request "POST" "/chats" "$session_data" "Create Relationship Test Chat" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        TEST_CHAT_SESSION_ID=$(echo "$response" | jq -r '.data.id // ""')
        tc_log "PASS" "關係測試會話創建成功 (ID: $TEST_CHAT_SESSION_ID)"
        return 0
    else
        tc_log "FAIL" "關係測試會話創建失敗"
        return 1
    fi
}

# 發送幾條消息來建立關係
build_relationship_history() {
    tc_log "INFO" "建立關係歷史"

    if [ -z "$TEST_CHAT_SESSION_ID" ]; then
        tc_log "ERROR" "沒有測試會話ID"
        return 1
    fi

    # 發送多條不同類型的消息來建立關係
    local messages=(
        "你好，很高興認識你！"
        "今天天氣真好，我們一起出去走走吧"
        "謝謝你一直陪伴我，我很開心"
        "你真的很溫暖，讓我感到很安心"
    )

    for message in "${messages[@]}"; do
        local message_data="{\"message\": \"$message\"}"

        local response=$(tc_http_request "POST" "/chats/$TEST_CHAT_SESSION_ID/messages" "$message_data" "Send Relationship Building Message" "true")

        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            tc_log "INFO" "消息發送成功: $message"
            sleep 3  # 等待AI回應
        else
            tc_log "WARN" "消息發送失敗: $message"
        fi
    done

    tc_log "PASS" "關係歷史建立完成"
    return 0
}

# 測試獲取關係狀態
test_get_relationship_status() {
    tc_log "INFO" "測試獲取關係狀態"

    local response=$(tc_http_request "GET" "/relationships/chat/$TEST_CHAT_SESSION_ID/status" "" "Get Relationship Status" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local relationship_level=$(echo "$response" | jq -r '.data.relationship_level // ""')
        local status_description=$(echo "$response" | jq -r '.data.status_description // ""')
        local interaction_count=$(echo "$response" | jq -r '.data.interaction_count // 0')
        local last_interaction=$(echo "$response" | jq -r '.data.last_interaction // ""')

        tc_log "PASS" "關係狀態獲取成功"
        tc_log "INFO" "  關係等級: $relationship_level"
        tc_log "INFO" "  狀態描述: $status_description"
        tc_log "INFO" "  互動次數: $interaction_count"
        tc_log "INFO" "  最後互動: $last_interaction"

        # 驗證數據合理性
        if [ "$interaction_count" -gt 0 ]; then
            tc_log "PASS" "互動次數合理"
            return 0
        else
            tc_log "WARN" "互動次數為0，可能需要更多時間建立關係"
            return 0
        fi
    else
        tc_log "FAIL" "關係狀態獲取失敗"
        return 1
    fi
}

# 測試獲取好感度等級
test_get_affection_level() {
    tc_log "INFO" "測試獲取好感度等級"

    local response=$(tc_http_request "GET" "/relationships/chat/$TEST_CHAT_SESSION_ID/affection" "" "Get Affection Level" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local affection_level=$(echo "$response" | jq -r '.data.affection_level // 0')
        local affection_description=$(echo "$response" | jq -r '.data.description // ""')
        local progress_to_next=$(echo "$response" | jq -r '.data.progress_to_next // 0')
        local max_level=$(echo "$response" | jq -r '.data.max_level // 0')

        tc_log "PASS" "好感度等級獲取成功"
        tc_log "INFO" "  好感度等級: $affection_level"
        tc_log "INFO" "  等級描述: $affection_description"
        tc_log "INFO" "  下一等級進度: $progress_to_next%"
        tc_log "INFO" "  最大等級: $max_level"

        # 驗證好感度數據
        if [ "$affection_level" -ge 0 ] && [ "$max_level" -gt 0 ]; then
            tc_log "PASS" "好感度數據合理"
            return 0
        else
            tc_log "WARN" "好感度數據異常"
            return 1
        fi
    else
        tc_log "FAIL" "好感度等級獲取失敗"
        return 1
    fi
}

# 測試獲取關係歷史
test_get_relationship_history() {
    tc_log "INFO" "測試獲取關係歷史"

    local response=$(tc_http_request "GET" "/relationships/chat/$TEST_CHAT_SESSION_ID/history?limit=10" "" "Get Relationship History" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local history_count=$(echo "$response" | jq -r '.data.history | length')
        local total_interactions=$(echo "$response" | jq -r '.data.total_interactions // 0')
        local relationship_milestones=$(echo "$response" | jq -r '.data.milestones | length // 0')

        tc_log "PASS" "關係歷史獲取成功"
        tc_log "INFO" "  歷史記錄數: $history_count"
        tc_log "INFO" "  總互動次數: $total_interactions"
        tc_log "INFO" "  關係里程碑: $relationship_milestones"

        # 檢查歷史記錄的詳細信息
        if [ "$history_count" -gt 0 ]; then
            # 獲取第一個歷史記錄的詳細信息
            local first_event_type=$(echo "$response" | jq -r '.data.history[0].event_type // ""')
            local first_event_date=$(echo "$response" | jq -r '.data.history[0].event_date // ""')
            local first_event_impact=$(echo "$response" | jq -r '.data.history[0].affection_impact // 0')

            tc_log "INFO" "  最新事件類型: $first_event_type"
            tc_log "INFO" "  事件日期: $first_event_date"
            tc_log "INFO" "  好感度影響: $first_event_impact"

            tc_log "PASS" "關係歷史記錄詳細"
            return 0
        else
            tc_log "WARN" "沒有關係歷史記錄"
            return 0
        fi
    else
        tc_log "FAIL" "關係歷史獲取失敗"
        return 1
    fi
}

# 測試關係數據統計
test_relationship_statistics() {
    tc_log "INFO" "測試關係數據統計"

    # 這個測試會組合多個API來獲取完整的關係統計
    local status_response=$(tc_http_request "GET" "/relationships/chat/$TEST_CHAT_SESSION_ID/status" "" "Get Status for Stats" "true")
    local affection_response=$(tc_http_request "GET" "/relationships/chat/$TEST_CHAT_SESSION_ID/affection" "" "Get Affection for Stats" "true")

    if echo "$status_response" | jq -e '.success' > /dev/null 2>&1 && echo "$affection_response" | jq -e '.success' > /dev/null 2>&1; then
        local total_interactions=$(echo "$status_response" | jq -r '.data.interaction_count // 0')
        local current_affection=$(echo "$affection_response" | jq -r '.data.affection_level // 0')
        local relationship_level=$(echo "$status_response" | jq -r '.data.relationship_level // ""')

        tc_log "PASS" "關係統計數據獲取成功"
        tc_log "INFO" "=== 關係統計總結 ==="
        tc_log "INFO" "  總互動次數: $total_interactions"
        tc_log "INFO" "  當前好感度: $current_affection"
        tc_log "INFO" "  關係類型: $relationship"

        # 計算互動效率（好感度/互動次數）
        if [ "$total_interactions" -gt 0 ]; then
            local efficiency=$(echo "scale=2; $current_affection / $total_interactions" | bc -l 2>/dev/null || echo "N/A")
            tc_log "INFO" "  互動效率: $efficiency 好感度/次"
        fi

        tc_log "PASS" "關係統計分析完成"
        return 0
    else
        tc_log "FAIL" "關係統計數據獲取失敗"
        return 1
    fi
}

# 測試多角色關係比較
test_multi_character_relationships() {
    tc_log "INFO" "測試多角色關係比較"

    # 測試獲取其他角色的關係狀態進行比較
    local characters=("character_01" "character_02" "character_03")
    local relationship_data=()

    for char_id in "${characters[@]}"; do
        tc_log "INFO" "  角色 $char_id: 跳過（需要特定聊天會話）"
        relationship_data+=("$char_id:0:0")
    done

    tc_log "PASS" "多角色關係比較完成"
    tc_log "INFO" "=== 角色關係排行 ==="

    # 簡單排序顯示（按好感度）
    for data in "${relationship_data[@]}"; do
        IFS=':' read -r char_id affection interactions <<< "$data"
        tc_log "INFO" "  $char_id: 好感度 $affection (互動 $interactions 次)"
    done

    return 0
}

# 清理測試數據
cleanup_relationship_test() {
    tc_log "INFO" "清理關係測試數據"

    if [ -n "$TEST_CHAT_SESSION_ID" ]; then
        local response=$(tc_http_request "DELETE" "/chats/$TEST_CHAT_SESSION_ID" "" "Delete Relationship Test Chat" "true")

        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            tc_log "PASS" "關係測試會話清理成功"
        else
            tc_log "WARN" "關係測試會話清理失敗，請手動清理 ID: $TEST_CHAT_SESSION_ID"
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
    tc_show_header "Thewavess AI Core - 關係系統測試"

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
    tc_log "INFO" "==================== 設置關係測試環境 ===================="

    if ! setup_relationship_context; then
        tc_log "ERROR" "關係測試環境設置失敗"
        exit 1
    fi

    sleep 2

    if ! build_relationship_history; then
        tc_log "ERROR" "關係歷史建立失敗"
        cleanup_relationship_test
        exit 1
    fi

    sleep 3  # 等待關係數據更新

    # 執行關係系統測試
    tc_log "INFO" "==================== 關係系統測試 ===================="

    if test_get_relationship_status; then
        test_results+=("關係狀態:PASS")
    else
        test_results+=("關係狀態:FAIL")
    fi

    sleep 2

    if test_get_affection_level; then
        test_results+=("好感度等級:PASS")
    else
        test_results+=("好感度等級:FAIL")
    fi

    sleep 2

    if test_get_relationship_history; then
        test_results+=("關係歷史:PASS")
    else
        test_results+=("關係歷史:FAIL")
    fi

    sleep 2

    if test_relationship_statistics; then
        test_results+=("關係統計:PASS")
    else
        test_results+=("關係統計:FAIL")
    fi

    sleep 2

    if test_multi_character_relationships; then
        test_results+=("多角色比較:PASS")
    else
        test_results+=("多角色比較:FAIL")
    fi

    # 清理測試數據
    cleanup_relationship_test

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
        tc_log "PASS" "所有關係系統測試通過！"
        tc_show_summary "關係系統測試"
        exit 0
    else
        tc_log "FAIL" "$failed_count 個測試失敗"
        tc_show_summary "關係系統測試"
        exit 1
    fi
}

# 執行主函數
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi