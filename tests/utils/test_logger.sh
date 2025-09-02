#!/bin/bash

# 🪵 Thewavess AI Core - 測試日誌工具
# 提供統一的日誌記錄功能

# ================================
# 日誌配置
# ================================

# 日誌等級 (使用函數模擬關聯陣列以支援舊版bash)
get_log_level() {
    case "$1" in
        "DEBUG") echo 0 ;;
        "INFO") echo 1 ;;
        "WARN") echo 2 ;;
        "ERROR") echo 3 ;;
        "FAIL") echo 4 ;;
        "PASS") echo 5 ;;
        *) echo 1 ;;
    esac
}

# 預設日誌等級
LOG_LEVEL=${LOG_LEVEL:-1}  # INFO
LOG_FILE=""
LOG_TO_CONSOLE=true
LOG_WITH_TIMESTAMP=true

# 顏色配置函數
get_log_color() {
    case "$1" in
        "DEBUG") echo $'\033[0;37m' ;;   # Light Gray
        "INFO") echo $'\033[0;34m' ;;    # Blue
        "WARN") echo $'\033[1;33m' ;;    # Yellow
        "ERROR") echo $'\033[0;31m' ;;   # Red
        "FAIL") echo $'\033[1;31m' ;;    # Bold Red
        "PASS") echo $'\033[0;32m' ;;    # Green
        "NC") echo $'\033[0m' ;;         # No Color
        *) echo $'\033[0m' ;;
    esac
}

# ================================
# 核心日誌函數
# ================================

# 初始化日誌系統
log_init() {
    local test_name="${1:-test}"
    local log_dir="${2:-logs}"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    # 創建日誌目錄
    mkdir -p "$log_dir"
    
    # 設置日誌檔案
    LOG_FILE="$log_dir/${test_name}_${timestamp}.log"
    
    # 寫入日誌標頭
    {
        echo "# Thewavess AI Core Test Log"
        echo "# Test: $test_name"
        echo "# Started: $(date -Iseconds)"
        echo "# Log Level: $LOG_LEVEL"
        echo "# ================================"
        echo ""
    } > "$LOG_FILE"
    
    log_info "Log system initialized: $LOG_FILE"
}

# 核心日誌函數
log_write() {
    local level="$1"
    local message="$2"
    local level_num=$(get_log_level "$level")
    
    # 檢查日誌等級
    if [ $level_num -lt $LOG_LEVEL ]; then
        return 0
    fi
    
    # 生成時間戳
    local timestamp=""
    if [ "$LOG_WITH_TIMESTAMP" = true ]; then
        timestamp="[$(date '+%Y-%m-%d %H:%M:%S')] "
    fi
    
    # 格式化消息
    local log_message="${timestamp}${level}: ${message}"
    local color=$(get_log_color "$level")
    local nc=$(get_log_color "NC")
    local colored_message="${color}${log_message}${nc}"
    
    # 輸出到控制台
    if [ "$LOG_TO_CONSOLE" = true ]; then
        echo -e "$colored_message" >&2
    fi
    
    # 寫入檔案
    if [ -n "$LOG_FILE" ]; then
        echo "$log_message" >> "$LOG_FILE"
    fi
}

# ================================
# 便利日誌函數
# ================================

log_debug() { log_write "DEBUG" "$1"; }
log_info()  { log_write "INFO" "$1"; }
log_warn()  { log_write "WARN" "$1"; }
log_error() { log_write "ERROR" "$1"; }
log_fail()  { log_write "FAIL" "$1"; }
log_pass()  { log_write "PASS" "$1"; }

# ================================
# 特殊日誌函數
# ================================

# 日誌測試開始
log_test_start() {
    local test_name="$1"
    local description="$2"
    log_info "TEST START: $test_name - $description"
    echo "=== TEST START: $test_name ===" >> "$LOG_FILE" 2>/dev/null || true
}

# 日誌測試結束
log_test_end() {
    local test_name="$1"
    local result="$2"  # PASS/FAIL
    log_write "$result" "TEST END: $test_name"
    echo "=== TEST END: $test_name ($result) ===" >> "$LOG_FILE" 2>/dev/null || true
    echo "" >> "$LOG_FILE" 2>/dev/null || true
}

# 日誌HTTP請求
log_http_request() {
    local method="$1"
    local url="$2"
    local status="$3"
    local response_time="$4"
    
    log_info "HTTP $method $url -> $status (${response_time}ms)"
}

# 日誌API回應
log_api_response() {
    local endpoint="$1"
    local status="$2"
    local body="$3"
    local max_length="${4:-200}"
    
    local truncated_body="$body"
    if [ ${#body} -gt $max_length ]; then
        truncated_body="${body:0:$max_length}..."
    fi
    
    if [ "$status" -ge 200 ] && [ "$status" -lt 300 ]; then
        log_pass "API $endpoint: Status $status"
    else
        log_fail "API $endpoint: Status $status"
    fi
    
    if [ -n "$truncated_body" ]; then
        log_debug "Response: $truncated_body"
    fi
}

# ================================
# 測試統計和報告
# ================================

# 全域統計變數 (移除-g選項以支援舊版bash)
TEST_TOTAL=0
TEST_PASSED=0
TEST_FAILED=0
TEST_SKIPPED=0

# 重置統計
log_reset_stats() {
    TEST_TOTAL=0
    TEST_PASSED=0
    TEST_FAILED=0
    TEST_SKIPPED=0
    log_info "Test statistics reset"
}

# 記錄測試結果
log_test_result() {
    local result="$1"  # PASS/FAIL/SKIP
    
    TEST_TOTAL=$((TEST_TOTAL + 1))
    
    case "$result" in
        "PASS"|"pass")
            TEST_PASSED=$((TEST_PASSED + 1))
            ;;
        "FAIL"|"fail")
            TEST_FAILED=$((TEST_FAILED + 1))
            ;;
        "SKIP"|"skip")
            TEST_SKIPPED=$((TEST_SKIPPED + 1))
            ;;
    esac
}

# 顯示測試統計報告
log_test_report() {
    local test_suite="${1:-Test Suite}"
    
    echo ""
    log_info "================================"
    log_info "$test_suite - Final Report"
    log_info "================================"
    log_info "Total Tests: $TEST_TOTAL"
    log_pass "Passed: $TEST_PASSED"
    log_fail "Failed: $TEST_FAILED"
    
    if [ $TEST_SKIPPED -gt 0 ]; then
        log_warn "Skipped: $TEST_SKIPPED"
    fi
    
    local success_rate=0
    if [ $TEST_TOTAL -gt 0 ]; then
        success_rate=$((TEST_PASSED * 100 / TEST_TOTAL))
    fi
    
    log_info "Success Rate: ${success_rate}%"
    log_info "================================"
    
    # 寫入最終統計到日誌檔案
    if [ -n "$LOG_FILE" ]; then
        {
            echo ""
            echo "# Final Test Statistics"
            echo "# Total: $TEST_TOTAL, Passed: $TEST_PASSED, Failed: $TEST_FAILED, Skipped: $TEST_SKIPPED"
            echo "# Success Rate: ${success_rate}%"
            echo "# Completed: $(date -Iseconds)"
        } >> "$LOG_FILE"
    fi
    
    # 返回成功狀態
    [ $TEST_FAILED -eq 0 ]
}

# ================================
# CSV日誌功能
# ================================

CSV_FILE=""

# 初始化CSV記錄
log_csv_init() {
    local test_name="${1:-test}"
    local csv_dir="${2:-results}"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    mkdir -p "$csv_dir"
    CSV_FILE="$csv_dir/${test_name}_${timestamp}.csv"
    
    # CSV標頭
    echo "timestamp,test_name,method,endpoint,status_code,response_time_ms,success,notes" > "$CSV_FILE"
    log_info "CSV logging initialized: $CSV_FILE"
}

# 記錄CSV數據
log_csv_record() {
    if [ -n "$CSV_FILE" ]; then
        local timestamp=$(date -Iseconds)
        echo "$timestamp,$*" >> "$CSV_FILE"
    fi
}

# ================================
# 工具函數
# ================================

# 設置日誌等級
log_set_level() {
    local level="$1"
    local level_num=$(get_log_level "$level")
    if [ $level_num -ge 0 ]; then
        LOG_LEVEL=$level_num
        log_info "Log level set to: $level"
    else
        log_error "Invalid log level: $level"
        return 1
    fi
}

# 開啟/關閉控制台輸出
log_set_console() {
    local enabled="${1:-true}"
    LOG_TO_CONSOLE="$enabled"
    log_info "Console logging: $enabled"
}

# 開啟/關閉時間戳
log_set_timestamp() {
    local enabled="${1:-true}"
    LOG_WITH_TIMESTAMP="$enabled"
    log_info "Timestamp logging: $enabled"
}

# 獲取當前日誌檔案路徑
log_get_file() {
    echo "$LOG_FILE"
}

# ================================
# 初始化檢查
# ================================

# 如果直接執行此腳本，顯示使用說明
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "Thewavess AI Test Logger"
    echo "========================"
    echo "This is a logging utility library for test scripts."
    echo ""
    echo "Usage:"
    echo "  source $(basename "$0")"
    echo "  log_init \"my_test\""
    echo "  log_info \"Test message\""
    echo ""
    echo "Functions:"
    echo "  log_init <test_name> [log_dir]  - Initialize logging"
    echo "  log_debug <message>             - Debug level log"
    echo "  log_info <message>              - Info level log"
    echo "  log_warn <message>              - Warning level log"
    echo "  log_error <message>             - Error level log"
    echo "  log_fail <message>              - Failure log"
    echo "  log_pass <message>              - Success log"
    echo "  log_test_report [suite_name]    - Show test report"
    echo ""
    exit 0
fi