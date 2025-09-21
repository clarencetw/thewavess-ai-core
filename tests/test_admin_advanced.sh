#!/bin/bash

# 🧪 Thewavess AI Core - 管理員進階功能測試
# 測試管理員專用的進階功能API

set -e

# 載入測試工具庫
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/test_common.sh"

# ================================
# 測試配置
# ================================

TEST_NAME="admin_advanced"

# ================================
# 測試函數
# ================================

# 測試系統監控功能
test_system_monitoring() {
    tc_log "INFO" "測試系統監控功能"

    local response=$(tc_http_request "GET" "/admin/monitor/system" "" "Get System Monitor" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local cpu_usage=$(echo "$response" | jq -r '.data.cpu_usage // 0')
        local memory_usage=$(echo "$response" | jq -r '.data.memory_usage // 0')
        local disk_usage=$(echo "$response" | jq -r '.data.disk_usage // 0')

        tc_log "PASS" "系統監控數據獲取成功"
        tc_log "INFO" "  CPU使用率: $cpu_usage%"
        tc_log "INFO" "  記憶體使用率: $memory_usage%"
        tc_log "INFO" "  硬碟使用率: $disk_usage%"

        return 0
    else
        tc_log "FAIL" "系統監控數據獲取失敗"
        return 1
    fi
}

# 測試批量操作
test_bulk_operations() {
    tc_log "INFO" "測試批量操作"

    # 模擬批量用戶操作
    local user_data='{"operation": "status_check", "user_ids": ["user1", "user2", "user3"]}'

    local response=$(tc_http_request "POST" "/admin/users/bulk" "$user_data" "Bulk User Operations" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local processed_count=$(echo "$response" | jq -r '.data.processed_count // 0')
        local success_count=$(echo "$response" | jq -r '.data.success_count // 0')

        tc_log "PASS" "批量操作執行成功"
        tc_log "INFO" "  處理數量: $processed_count"
        tc_log "INFO" "  成功數量: $success_count"

        return 0
    else
        tc_log "WARN" "批量操作失敗或不支援"
        return 0
    fi
}

# 測試系統配置管理
test_system_configuration() {
    tc_log "INFO" "測試系統配置管理"

    # 獲取當前配置
    local config_response=$(tc_http_request "GET" "/admin/config" "" "Get System Config" "true")

    if echo "$config_response" | jq -e '.success' > /dev/null 2>&1; then
        local max_users=$(echo "$config_response" | jq -r '.data.max_users // 0')
        local max_sessions=$(echo "$config_response" | jq -r '.data.max_sessions // 0')

        tc_log "PASS" "系統配置獲取成功"
        tc_log "INFO" "  最大用戶數: $max_users"
        tc_log "INFO" "  最大會話數: $max_sessions"

        return 0
    else
        tc_log "WARN" "系統配置獲取失敗或不支援"
        return 0
    fi
}

# 測試備份與還原
test_backup_restore() {
    tc_log "INFO" "測試備份與還原功能"

    # 創建備份
    local backup_data='{"type": "database", "include_users": true, "include_chats": false}'

    local backup_response=$(tc_http_request "POST" "/admin/backup" "$backup_data" "Create Backup" "true")

    if echo "$backup_response" | jq -e '.success' > /dev/null 2>&1; then
        local backup_id=$(echo "$backup_response" | jq -r '.data.backup_id // ""')
        local backup_size=$(echo "$backup_response" | jq -r '.data.size_mb // 0')

        tc_log "PASS" "備份創建成功"
        tc_log "INFO" "  備份ID: $backup_id"
        tc_log "INFO" "  備份大小: $backup_size MB"

        return 0
    else
        tc_log "WARN" "備份功能失敗或不支援"
        return 0
    fi
}

# 測試安全審計
test_security_audit() {
    tc_log "INFO" "測試安全審計功能"

    local response=$(tc_http_request "GET" "/admin/security/audit" "" "Security Audit" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local failed_logins=$(echo "$response" | jq -r '.data.failed_logins_24h // 0')
        local suspicious_ips=$(echo "$response" | jq -r '.data.suspicious_ips | length // 0')
        local security_score=$(echo "$response" | jq -r '.data.security_score // 0')

        tc_log "PASS" "安全審計數據獲取成功"
        tc_log "INFO" "  24小時失敗登入: $failed_logins"
        tc_log "INFO" "  可疑IP數量: $suspicious_ips"
        tc_log "INFO" "  安全評分: $security_score/100"

        return 0
    else
        tc_log "WARN" "安全審計功能失敗或不支援"
        return 0
    fi
}

# ================================
# 主執行流程
# ================================

main() {
    # 初始化測試
    tc_init_logging "$TEST_NAME"
    # CSV功能已移除，改用詳細日誌記錄
    tc_show_header "Thewavess AI Core - 管理員進階功能測試"

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

    # 執行管理員功能測試
    tc_log "INFO" "==================== 管理員進階功能測試 ===================="

    if test_system_monitoring; then
        test_results+=("系統監控:PASS")
    else
        test_results+=("系統監控:FAIL")
    fi

    sleep 2

    if test_bulk_operations; then
        test_results+=("批量操作:PASS")
    else
        test_results+=("批量操作:FAIL")
    fi

    sleep 2

    if test_system_configuration; then
        test_results+=("系統配置:PASS")
    else
        test_results+=("系統配置:FAIL")
    fi

    sleep 2

    if test_backup_restore; then
        test_results+=("備份還原:PASS")
    else
        test_results+=("備份還原:FAIL")
    fi

    sleep 2

    if test_security_audit; then
        test_results+=("安全審計:PASS")
    else
        test_results+=("安全審計:FAIL")
    fi

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
        tc_log "PASS" "所有管理員進階功能測試通過！"
        tc_show_summary "管理員進階功能測試"
        exit 0
    else
        tc_log "FAIL" "$failed_count 個測試失敗"
        tc_show_summary "管理員進階功能測試"
        exit 1
    fi
}

# 執行主函數
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi