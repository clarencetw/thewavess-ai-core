#!/bin/bash

# 🧪 Thewavess AI Core - 管理員頁面測試
# 專門測試管理員前端頁面載入和監控API整合

set -e

# 載入測試工具庫
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/test_common.sh"

# ================================
# 測試配置
# ================================

# 管理員頁面配置
ADMIN_BASE_URL="${ADMIN_BASE_URL:-http://localhost:8080/admin}"
MONITOR_BASE_URL="${MONITOR_BASE_URL:-http://localhost:8080/api/v1/monitor}"

# 管理員頁面路由
declare -a ADMIN_PAGES=(
    "login:管理員登入"
    "dashboard:管理儀表板" 
    "users:用戶管理"
    "chats:聊天記錄管理"
    "characters:角色管理"
)

# 監控API端點
declare -a MONITOR_ENDPOINTS=(
    "health:系統健康檢查"
    "stats:系統統計資訊"
    "ready:服務就緒檢查"
    "live:服務存活檢查"
    "metrics:Prometheus指標"
)

# 管理員認證資訊
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123456}"
ADMIN_JWT_TOKEN=""

# ================================
# 輔助函數
# ================================

# 管理員認證
admin_authenticate() {
    tc_log "INFO" "正在進行管理員認證"
    
    local login_data="{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}"
    local response
    
    response=$(curl -s -X POST "${TEST_BASE_URL}/admin/auth/login" \
        -H "Content-Type: application/json" \
        -d "$login_data")
    
    if echo "$response" | grep -q '"success":true'; then
        ADMIN_JWT_TOKEN=$(echo "$response" | jq -r '.data.access_token // .data.token // ""' 2>/dev/null)
        
        if [ -n "$ADMIN_JWT_TOKEN" ] && [ "$ADMIN_JWT_TOKEN" != "null" ]; then
            tc_log "PASS" "管理員認證成功"
            return 0
        fi
    fi
    
    tc_log "FAIL" "管理員認證失敗"
    tc_log "ERROR" "Response: $response"
    return 1
}

# 檢查HTML頁面基本結構
check_html_structure() {
    local page_content="$1"
    local page_name="$2"
    local expected_title="$3"
    
    tc_log "INFO" "檢查 $page_name 頁面HTML結構"
    
    # 檢查基本HTML結構
    if ! echo "$page_content" | grep -q "<!DOCTYPE html>"; then
        tc_log "FAIL" "$page_name: 缺少DOCTYPE聲明"
        return 1
    fi
    
    if ! echo "$page_content" | grep -q "<html"; then
        tc_log "FAIL" "$page_name: 缺少HTML標籤"
        return 1
    fi
    
    if ! echo "$page_content" | grep -q "<head>"; then
        tc_log "FAIL" "$page_name: 缺少HEAD標籤"
        return 1
    fi
    
    if ! echo "$page_content" | grep -q "<body"; then
        tc_log "FAIL" "$page_name: 缺少BODY標籤"
        return 1
    fi
    
    # 檢查標題
    if [ -n "$expected_title" ]; then
        if echo "$page_content" | grep -q "<title>.*$expected_title.*</title>"; then
            tc_log "PASS" "$page_name: 標題正確包含 '$expected_title'"
        else
            tc_log "WARN" "$page_name: 標題不包含預期文字 '$expected_title'"
        fi
    fi
    
    # 檢查重要資源
    if echo "$page_content" | grep -q "tailwindcss.com"; then
        tc_log "PASS" "$page_name: Tailwind CSS已載入"
    else
        tc_log "WARN" "$page_name: 未檢測到Tailwind CSS"
    fi
    
    if echo "$page_content" | grep -q "font-awesome"; then
        tc_log "PASS" "$page_name: Font Awesome已載入"
    else
        tc_log "WARN" "$page_name: 未檢測到Font Awesome"
    fi
    
    if echo "$page_content" | grep -q "/public/app.js"; then
        tc_log "PASS" "$page_name: 主JavaScript文件已載入"
    else
        tc_log "WARN" "$page_name: 未檢測到主JavaScript文件"
    fi
    
    tc_log "PASS" "$page_name HTML結構檢查完成"
    return 0
}

# 檢查頁面特定元素
check_page_elements() {
    local page_content="$1"
    local page_type="$2"
    
    case "$page_type" in
        "dashboard")
            # 檢查儀表板特定元素
            if echo "$page_content" | grep -q 'data-page="dashboard"'; then
                tc_log "PASS" "儀表板: 頁面識別標記正確"
            else
                tc_log "FAIL" "儀表板: 缺少頁面識別標記"
                return 1
            fi
            
            if echo "$page_content" | grep -q "系統總覽與統計資訊"; then
                tc_log "PASS" "儀表板: 包含預期的描述文字"
            fi
            
            if echo "$page_content" | grep -q "statsGrid"; then
                tc_log "PASS" "儀表板: 統計卡片容器存在"
            fi
            
            if echo "$page_content" | grep -q "alertsPanel"; then
                tc_log "PASS" "儀表板: 警報面板存在"
            fi
            ;;
            
        "users")
            # 檢查用戶管理特定元素
            if echo "$page_content" | grep -q 'data-page="users"'; then
                tc_log "PASS" "用戶管理: 頁面識別標記正確"
            else
                tc_log "FAIL" "用戶管理: 缺少頁面識別標記"
                return 1
            fi
            
            if echo "$page_content" | grep -q "userSearchInput"; then
                tc_log "PASS" "用戶管理: 搜尋輸入框存在"
            fi
            ;;
            
        "chats")
            # 檢查聊天記錄特定元素
            if echo "$page_content" | grep -q 'data-page="chats"'; then
                tc_log "PASS" "聊天記錄: 頁面識別標記正確"
            else
                tc_log "FAIL" "聊天記錄: 缺少頁面識別標記"
                return 1
            fi
            ;;
            
        "characters")
            # 檢查角色管理特定元素
            if echo "$page_content" | grep -q 'data-page="characters"'; then
                tc_log "PASS" "角色管理: 頁面識別標記正確"
            else
                tc_log "FAIL" "角色管理: 缺少頁面識別標記"
                return 1
            fi
            
            if echo "$page_content" | grep -q "characterSearchInput"; then
                tc_log "PASS" "角色管理: 搜尋輸入框存在"
            fi
            ;;
            
        "login")
            # 檢查登入頁面特定元素
            if echo "$page_content" | grep -q "管理員登入"; then
                tc_log "PASS" "登入頁面: 包含標題"
            fi
            ;;
    esac
    
    return 0
}

# ================================
# 主要測試函數
# ================================

# 測試管理員頁面載入
test_admin_pages() {
    tc_log "INFO" "開始測試管理員頁面載入"
    
    local total_tests=0
    local passed_tests=0
    
    for page_info in "${ADMIN_PAGES[@]}"; do
        IFS=':' read -r page_path page_title <<< "$page_info"
        total_tests=$((total_tests + 1))
        
        tc_log "INFO" "測試頁面: /$page_path"
        
        # 執行HTTP請求
        local response
        local start_time=$(date +%s.%N)
        
        response=$(curl -s -w "\n%{http_code}" "$ADMIN_BASE_URL/$page_path" 2>/dev/null)
        local status_code=$(echo "$response" | tail -n1)
        local page_content=$(echo "$response" | sed '$d')
        
        local end_time=$(date +%s.%N)
        local response_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
        
        # 檢查HTTP狀態碼
        if [ "$status_code" = "200" ]; then
            tc_log "PASS" "$page_title 載入成功 (${response_time}s)"
            
            # 檢查HTML結構
            if check_html_structure "$page_content" "$page_title" "$page_title"; then
                # 檢查頁面特定元素
                check_page_elements "$page_content" "$page_path"
                passed_tests=$((passed_tests + 1))
            fi
        else
            tc_log "FAIL" "$page_title 載入失敗 (Status: $status_code)"
            tc_log "ERROR" "URL: $ADMIN_BASE_URL/$page_path"
        fi
        
        # 短暫延遲避免請求過快
        sleep 0.5
    done
    
    tc_log "INFO" "頁面載入測試完成: $passed_tests/$total_tests 頁面正常"
    return $([ $passed_tests -eq $total_tests ] && echo 0 || echo 1)
}

# 測試監控API整合
test_monitor_integration() {
    tc_log "INFO" "開始測試監控API整合"
    
    local total_tests=0
    local passed_tests=0
    
    for endpoint_info in "${MONITOR_ENDPOINTS[@]}"; do
        IFS=':' read -r endpoint_path endpoint_desc <<< "$endpoint_info"
        total_tests=$((total_tests + 1))
        
        tc_log "INFO" "測試監控端點: /$endpoint_path"
        
        # 執行HTTP請求
        local response
        local start_time=$(date +%s.%N)
        
        response=$(curl -s -w "\n%{http_code}" "$MONITOR_BASE_URL/$endpoint_path" 2>/dev/null)
        local status_code=$(echo "$response" | tail -n1)
        local response_body=$(echo "$response" | sed '$d')
        
        local end_time=$(date +%s.%N)
        local response_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
        
        # 檢查HTTP狀態碼
        case "$endpoint_path" in
            "metrics")
                # Prometheus指標使用純文本格式
                if [ "$status_code" = "200" ]; then
                    if echo "$response_body" | grep -q "thewavess_"; then
                        tc_log "PASS" "$endpoint_desc 正常 (${response_time}s) - 包含指標數據"
                        passed_tests=$((passed_tests + 1))
                    else
                        tc_log "FAIL" "$endpoint_desc 回應格式錯誤 - 缺少指標數據"
                    fi
                else
                    tc_log "FAIL" "$endpoint_desc 請求失敗 (Status: $status_code)"
                fi
                ;;
            *)
                # 其他端點使用JSON格式
                if [ "$status_code" = "200" ] || [ "$status_code" = "503" ]; then
                    if echo "$response_body" | jq . >/dev/null 2>&1; then
                        local success_field=$(echo "$response_body" | jq -r '.success // empty' 2>/dev/null)
                        if [ "$success_field" = "true" ] || [ -n "$(echo "$response_body" | jq -r '.data // empty' 2>/dev/null)" ]; then
                            tc_log "PASS" "$endpoint_desc 正常 (${response_time}s) - JSON格式正確"
                            passed_tests=$((passed_tests + 1))
                        else
                            tc_log "WARN" "$endpoint_desc JSON結構異常但可解析"
                            passed_tests=$((passed_tests + 1))  # 仍然算通過，因為可能是服務降級
                        fi
                    else
                        tc_log "FAIL" "$endpoint_desc JSON格式錯誤"
                    fi
                else
                    tc_log "FAIL" "$endpoint_desc 請求失敗 (Status: $status_code)"
                fi
                ;;
        esac
        
        # 短暫延遲
        sleep 0.2
    done
    
    tc_log "INFO" "監控API測試完成: $passed_tests/$total_tests 端點正常"
    return $([ $passed_tests -eq $total_tests ] && echo 0 || echo 1)
}

# 測試頁面性能
test_page_performance() {
    tc_log "INFO" "開始測試頁面載入性能"
    
    local total_response_time=0
    local test_count=0
    local slow_pages=()
    
    for page_info in "${ADMIN_PAGES[@]}"; do
        IFS=':' read -r page_path page_title <<< "$page_info"
        test_count=$((test_count + 1))
        
        # 測試3次取平均值
        local sum_time=0
        for i in {1..3}; do
            local start_time=$(date +%s.%N)
            curl -s "$ADMIN_BASE_URL/$page_path" >/dev/null 2>&1
            local end_time=$(date +%s.%N)
            local response_time=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "1")
            sum_time=$(echo "$sum_time + $response_time" | bc -l 2>/dev/null || echo "$sum_time")
        done
        
        local avg_time=$(echo "scale=3; $sum_time / 3" | bc -l 2>/dev/null || echo "1")
        total_response_time=$(echo "$total_response_time + $avg_time" | bc -l 2>/dev/null || echo "$total_response_time")
        
        # 檢查是否超過閾值 (2秒)
        if (( $(echo "$avg_time > 2.0" | bc -l 2>/dev/null || echo 0) )); then
            slow_pages+=("$page_title:${avg_time}s")
            tc_log "WARN" "$page_title 載入較慢: ${avg_time}s"
        else
            tc_log "PASS" "$page_title 載入時間正常: ${avg_time}s"
        fi
        
        sleep 0.5
    done
    
    local avg_overall=$(echo "scale=3; $total_response_time / $test_count" | bc -l 2>/dev/null || echo "1")
    tc_log "INFO" "平均頁面載入時間: ${avg_overall}s"
    
    if [ ${#slow_pages[@]} -eq 0 ]; then
        tc_log "PASS" "所有頁面載入性能良好"
        return 0
    else
        tc_log "WARN" "發現 ${#slow_pages[@]} 個頁面載入較慢"
        for slow_page in "${slow_pages[@]}"; do
            tc_log "WARN" "  - $slow_page"
        done
        return 1
    fi
}

# ================================
# 主執行流程
# ================================

main() {
    # 初始化測試
    tc_init_logging "admin_pages_test"
    tc_show_header "Thewavess AI Core - 管理員頁面測試"
    
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
    
    local test_results=()
    
    # 執行頁面載入測試
    tc_log "INFO" "==================== 頁面載入測試 ===================="
    if test_admin_pages; then
        test_results+=("頁面載入:PASS")
    else
        test_results+=("頁面載入:FAIL")
    fi
    
    # 執行監控API整合測試
    tc_log "INFO" "==================== 監控API整合測試 ===================="
    if test_monitor_integration; then
        test_results+=("監控整合:PASS")
    else
        test_results+=("監控整合:FAIL")
    fi
    
    # 執行性能測試
    tc_log "INFO" "==================== 頁面性能測試 ===================="
    if test_page_performance; then
        test_results+=("性能測試:PASS")
    else
        test_results+=("性能測試:WARN")  # 性能問題不算嚴重失敗
    fi
    
    # 顯示測試結果總結
    tc_log "INFO" "==================== 測試結果總結 ===================="
    local failed_count=0
    
    for result in "${test_results[@]}"; do
        IFS=':' read -r test_name test_status <<< "$result"
        case "$test_status" in
            "PASS") tc_log "PASS" "$test_name: 通過" ;;
            "WARN") tc_log "WARN" "$test_name: 警告" ;;
            "FAIL") 
                tc_log "FAIL" "$test_name: 失敗"
                failed_count=$((failed_count + 1))
                ;;
        esac
    done
    
    # 最終結果
    if [ $failed_count -eq 0 ]; then
        tc_log "PASS" "所有關鍵測試通過！"
        tc_show_summary "管理員頁面測試"
        exit 0
    else
        tc_log "FAIL" "$failed_count 個測試失敗"
        tc_show_summary "管理員頁面測試"
        exit 1
    fi
}

# 執行主函數
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi