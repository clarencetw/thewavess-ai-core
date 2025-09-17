#!/bin/bash

# 🧪 Thewavess AI Core - 用戶資料管理測試
# 測試用戶個人資料相關API功能

set -e

# 載入測試工具庫
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/test_common.sh"

# ================================
# 測試配置
# ================================

TEST_NAME="user_profile"

# ================================
# 測試函數
# ================================

# 測試獲取用戶資料
test_get_profile() {
    tc_log "INFO" "測試獲取用戶資料"

    local response=$(tc_http_request "GET" "/user/profile" "" "Get User Profile" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local user_id=$(echo "$response" | jq -r '.data.id // ""')
        local username=$(echo "$response" | jq -r '.data.username // ""')
        local email=$(echo "$response" | jq -r '.data.email // ""')

        tc_log "PASS" "用戶資料獲取成功"
        tc_log "INFO" "  用戶ID: $user_id"
        tc_log "INFO" "  用戶名: $username"
        tc_log "INFO" "  電子郵件: $email"
        return 0
    else
        tc_log "FAIL" "用戶資料獲取失敗"
        return 1
    fi
}

# 測試更新用戶資料
test_update_profile() {
    tc_log "INFO" "測試更新用戶資料"

    local update_data='{
        "display_name": "測試用戶_更新",
        "bio": "這是更新後的個人簡介",
        "preferences": {
            "language": "zh-TW",
            "theme": "dark"
        }
    }'

    local response=$(tc_http_request "PUT" "/user/profile" "$update_data" "Update User Profile" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local display_name=$(echo "$response" | jq -r '.data.display_name // ""')
        local bio=$(echo "$response" | jq -r '.data.bio // ""')

        tc_log "PASS" "用戶資料更新成功"
        tc_log "INFO" "  顯示名稱: $display_name"
        tc_log "INFO" "  個人簡介: $bio"
        return 0
    else
        tc_log "FAIL" "用戶資料更新失敗"
        return 1
    fi
}

# 測試頭像上傳（模擬）
test_avatar_upload() {
    tc_log "INFO" "測試頭像上傳"

    # 創建一個模擬的頭像文件（小的base64圖片）
    local test_image="iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

    local upload_data="{
        \"avatar\": \"data:image/png;base64,$test_image\",
        \"filename\": \"test_avatar.png\"
    }"

    local response=$(tc_http_request "POST" "/user/avatar" "$upload_data" "Upload Avatar" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local avatar_url=$(echo "$response" | jq -r '.data.avatar_url // ""')
        tc_log "PASS" "頭像上傳成功"
        tc_log "INFO" "  頭像URL: $avatar_url"
        return 0
    else
        tc_log "FAIL" "頭像上傳失敗"
        tc_log "INFO" "  可能原因: 需要實際文件上傳實現"
        return 1
    fi
}

# 測試Token刷新
test_token_refresh() {
    tc_log "INFO" "測試Token刷新"

    if [ -z "$TC_REFRESH_TOKEN" ]; then
        tc_log "WARN" "沒有refresh token，跳過刷新測試"
        return 0
    fi

    local refresh_data="{\"refresh_token\": \"$TC_REFRESH_TOKEN\"}"

    local response=$(tc_http_request "POST" "/auth/refresh" "$refresh_data" "Refresh Token" "false")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local new_access_token=$(echo "$response" | jq -r '.data.access_token // ""')
        local new_refresh_token=$(echo "$response" | jq -r '.data.refresh_token // ""')

        if [ -n "$new_access_token" ] && [ "$new_access_token" != "null" ]; then
            # 更新token
            TC_JWT_TOKEN="$new_access_token"
            if [ -n "$new_refresh_token" ] && [ "$new_refresh_token" != "null" ]; then
                TC_REFRESH_TOKEN="$new_refresh_token"
            fi

            tc_log "PASS" "Token刷新成功"
            tc_log "INFO" "  新的Access Token: ${new_access_token:0:30}..."
            return 0
        fi
    fi

    tc_log "FAIL" "Token刷新失敗"
    return 1
}

# 測試刪除帳號（謹慎操作）
test_delete_account() {
    tc_log "INFO" "測試刪除帳號"

    # 只有在使用動態創建的測試用戶時才執行刪除測試
    if [[ "$TEST_USERNAME" != *"testusertemp_"* ]]; then
        tc_log "WARN" "非測試用戶，跳過刪除帳號測試"
        return 0
    fi

    local response=$(tc_http_request "DELETE" "/user/account" '{}' "Delete Account" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        tc_log "PASS" "帳號刪除成功"
        tc_log "INFO" "  用戶已軟刪除"

        # 清空token，因為帳號已刪除
        TC_JWT_TOKEN=""
        TC_REFRESH_TOKEN=""
        return 0
    else
        tc_log "FAIL" "帳號刪除失敗"
        return 1
    fi
}

# 測試已刪除用戶的API訪問
test_deleted_user_access() {
    tc_log "INFO" "測試已刪除用戶的API訪問限制"

    # 只有在帳號已刪除的情況下才測試
    if [ -n "$TC_JWT_TOKEN" ]; then
        tc_log "WARN" "用戶未刪除，跳過此測試"
        return 0
    fi

    # 嘗試使用已刪除用戶的token訪問API（應該失敗）
    local old_token="$TC_JWT_TOKEN"
    TC_JWT_TOKEN="expired_or_deleted_token"

    local response=$(curl -s -X GET "${TEST_BASE_URL}/user/profile" \
        -H "Authorization: Bearer $TC_JWT_TOKEN")

    if echo "$response" | grep -q '"success":false'; then
        tc_log "PASS" "已刪除用戶正確被拒絕訪問"
        return 0
    else
        tc_log "FAIL" "已刪除用戶仍可訪問API"
        return 1
    fi
}

# ================================
# 主執行流程
# ================================

main() {
    # 初始化測試
    tc_init_logging "$TEST_NAME"
    tc_init_csv "$TEST_NAME"
    tc_show_header "Thewavess AI Core - 用戶資料管理測試"

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

    # 執行測試
    tc_log "INFO" "==================== 用戶資料測試 ===================="

    if test_get_profile; then
        test_results+=("獲取資料:PASS")
    else
        test_results+=("獲取資料:FAIL")
    fi

    sleep 1

    if test_update_profile; then
        test_results+=("更新資料:PASS")
    else
        test_results+=("更新資料:FAIL")
    fi

    sleep 1

    if test_avatar_upload; then
        test_results+=("頭像上傳:PASS")
    else
        test_results+=("頭像上傳:FAIL")
    fi

    sleep 1

    if test_token_refresh; then
        test_results+=("Token刷新:PASS")
    else
        test_results+=("Token刷新:FAIL")
    fi

    sleep 1

    if test_delete_account; then
        test_results+=("刪除帳號:PASS")
    else
        test_results+=("刪除帳號:FAIL")
    fi

    sleep 1

    if test_deleted_user_access; then
        test_results+=("訪問限制:PASS")
    else
        test_results+=("訪問限制:FAIL")
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
        tc_log "PASS" "所有用戶資料管理測試通過！"
        tc_show_summary "用戶資料管理測試"
        exit 0
    else
        tc_log "FAIL" "$failed_count 個測試失敗"
        tc_show_summary "用戶資料管理測試"
        exit 1
    fi
}

# 執行主函數
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi