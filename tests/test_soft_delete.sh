#!/bin/bash

# 🧪 Thewavess AI Core - Soft Delete 功能測試
# 測試所有軟刪除相關的API和功能

# 導入測試工具庫
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$script_dir/utils/test_common.sh"

# ================================
# 測試設定
# ================================

TEST_NAME="soft_delete"
ADMIN_JWT_TOKEN=""

# ================================
# 主要測試函數
# ================================

# 執行所有軟刪除測試
run_soft_delete_tests() {
    tc_show_header "🗂️  Soft Delete 功能全面測試"
    
    # 1. 檢查服務器狀態
    tc_log "INFO" "檢查服務器健康狀態"
    if ! tc_check_health; then
        tc_log "ERROR" "服務器不可用，無法執行測試"
        return 1
    fi
    
    # 2. 管理員認證
    tc_log "INFO" "執行管理員認證"
    if ! tc_admin_authenticate; then
        tc_log "ERROR" "管理員認證失敗，無法執行管理員測試"
        return 1
    fi
    
    # 3. 用戶認證
    tc_log "INFO" "執行用戶認證"
    if ! tc_authenticate; then
        tc_log "ERROR" "用戶認證失敗，無法執行用戶測試"
        return 1
    fi
    
    # 4. 執行綜合軟刪除測試
    tc_log "INFO" "開始執行綜合軟刪除測試"
    if tc_test_soft_delete_comprehensive "$ADMIN_JWT_TOKEN" "$TC_JWT_TOKEN"; then
        tc_log "PASS" "軟刪除功能測試全部通過！"
    else
        tc_log "FAIL" "軟刪除功能測試存在失敗"
        return 1
    fi
    
    return 0
}

# 分別執行各項測試（詳細模式）
run_detailed_soft_delete_tests() {
    tc_show_header "🔍 Soft Delete 功能詳細測試"
    
    # 檢查服務器和認證
    if ! tc_check_health; then
        tc_log "ERROR" "服務器不可用"
        return 1
    fi
    
    if ! tc_admin_authenticate || ! tc_authenticate; then
        tc_log "ERROR" "認證失敗"
        return 1
    fi
    
    # 測試1: 用戶軟刪除
    echo ""
    tc_log "INFO" "=== 測試 1: 用戶軟刪除功能 ==="
    if tc_test_user_soft_delete; then
        tc_log "PASS" "✅ 用戶軟刪除測試通過"
    else
        tc_log "FAIL" "❌ 用戶軟刪除測試失敗"
    fi
    
    # 測試2: 角色軟刪除
    echo ""
    tc_log "INFO" "=== 測試 2: 角色軟刪除功能 ==="
    if tc_test_character_soft_delete "$TC_JWT_TOKEN"; then
        tc_log "PASS" "✅ 角色軟刪除測試通過"
    else
        tc_log "FAIL" "❌ 角色軟刪除測試失敗"
    fi
    
    # 測試3: 管理員統計API軟刪除過濾
    echo ""
    tc_log "INFO" "=== 測試 3: 管理員統計API軟刪除過濾 ==="
    if tc_test_admin_stats_soft_delete "$ADMIN_JWT_TOKEN"; then
        tc_log "PASS" "✅ 管理員統計API測試通過"
    else
        tc_log "FAIL" "❌ 管理員統計API測試失敗"
    fi
    
    # 測試4: 公開API軟刪除過濾
    echo ""
    tc_log "INFO" "=== 測試 4: 公開API軟刪除過濾 ==="
    if tc_test_public_api_soft_delete; then
        tc_log "PASS" "✅ 公開API軟刪除過濾測試通過"
    else
        tc_log "FAIL" "❌ 公開API軟刪除過濾測試失敗"
    fi
    
    # 測試5: 管理員角色恢復功能
    echo ""
    tc_log "INFO" "=== 測試 5: 管理員角色恢復功能 ==="
    if tc_test_character_restore "$ADMIN_JWT_TOKEN" "$TC_JWT_TOKEN"; then
        tc_log "PASS" "✅ 角色恢復功能測試通過"
    else
        tc_log "FAIL" "❌ 角色恢復功能測試失敗"
    fi
}

# 測試 PostgreSQL 數組序列化（與軟刪除相關）
run_array_serialization_tests() {
    tc_show_header "🔧 PostgreSQL 數組序列化測試"
    
    if ! tc_check_health; then
        tc_log "ERROR" "服務器不可用"
        return 1
    fi
    
    if ! tc_admin_authenticate || ! tc_authenticate; then
        tc_log "ERROR" "認證失敗"
        return 1
    fi
    
    # 執行數組序列化測試
    if tc_test_postgresql_array_serialization "$ADMIN_JWT_TOKEN" "$TC_JWT_TOKEN"; then
        tc_log "PASS" "✅ PostgreSQL 數組序列化測試全部通過"
    else
        tc_log "FAIL" "❌ PostgreSQL 數組序列化測試存在失敗"
        return 1
    fi
}

# ================================
# 主程序
# ================================

main() {
    # 初始化測試環境
    tc_init_logging "$TEST_NAME"
    tc_init_csv "$TEST_NAME"
    
    # 檢查依賴
    if ! tc_check_dependencies; then
        exit 1
    fi
    
    # 解析參數
    local test_mode="${1:-comprehensive}"
    
    case "$test_mode" in
        "comprehensive"|"comp")
            if run_soft_delete_tests; then
                tc_show_summary "Soft Delete Tests"
                exit 0
            else
                tc_show_summary "Soft Delete Tests"
                exit 1
            fi
            ;;
        "detailed"|"detail")
            run_detailed_soft_delete_tests
            tc_show_summary "Detailed Soft Delete Tests"
            ;;
        "array"|"serialization")
            if run_array_serialization_tests; then
                tc_show_summary "Array Serialization Tests"
                exit 0
            else
                tc_show_summary "Array Serialization Tests"
                exit 1
            fi
            ;;
        "help"|"-h"|"--help")
            echo "🧪 Thewavess AI Core - Soft Delete 測試腳本"
            echo ""
            echo "用法: $0 [mode]"
            echo ""
            echo "測試模式:"
            echo "  comprehensive, comp    - 執行綜合軟刪除測試 (默認)"
            echo "  detailed, detail       - 執行詳細分項測試"
            echo "  array, serialization   - 執行數組序列化測試"
            echo "  help, -h, --help      - 顯示此幫助信息"
            echo ""
            echo "示例:"
            echo "  $0                     # 執行綜合測試"
            echo "  $0 detailed           # 執行詳細測試"
            echo "  $0 array              # 執行數組序列化測試"
            exit 0
            ;;
        *)
            tc_log "ERROR" "未知的測試模式: $test_mode"
            tc_log "INFO" "使用 '$0 help' 查看可用選項"
            exit 1
            ;;
    esac
    
    # 清理資源
    tc_cleanup
}

# 腳本執行入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi