#!/bin/bash

# 關鍵字管理工具集
# 使用方法: ./keyword_tools.sh [analyze|clean|both|help]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日誌函數
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 顯示幫助
show_help() {
    echo "🔧 關鍵字管理工具集"
    echo ""
    echo "使用方法:"
    echo "  $0 [command]"
    echo ""
    echo "命令:"
    echo "  analyze     - 分析關鍵字統計和衝突"
    echo "  clean       - 清理重複關鍵字"
    echo "  both        - 先分析再清理"
    echo "  help        - 顯示此幫助"
    echo ""
    echo "範例:"
    echo "  $0 analyze  # 只進行分析"
    echo "  $0 clean    # 只進行清理"
    echo "  $0 both     # 完整流程"
}

# 檢查Go環境
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go 未安裝或不在 PATH 中"
        exit 1
    fi
    log_info "Go 版本: $(go version)"
}

# 運行關鍵字分析
run_analysis() {
    log_info "開始關鍵字分析..."

    cd "$PROJECT_ROOT"

    # 編譯並運行分析工具
    go run scripts/keyword_analyzer.go

    if [ $? -eq 0 ]; then
        log_success "分析完成"
    else
        log_error "分析失敗"
        exit 1
    fi
}

# 運行關鍵字清理
run_cleanup() {
    log_info "開始關鍵字清理..."

    cd "$PROJECT_ROOT"

    # 檢查是否有備份檔案
    if ls services/*.backup 1> /dev/null 2>&1; then
        log_warning "發現備份檔案，是否覆蓋？"
        read -p "繼續? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "操作已取消"
            exit 0
        fi
    fi

    # 編譯並運行清理工具
    go run scripts/keyword_cleaner.go --auto

    if [ $? -eq 0 ]; then
        log_success "清理完成"
    else
        log_error "清理失敗"
        exit 1
    fi
}

# 驗證分類器
validate_classifier() {
    log_info "驗證關鍵字分類器..."

    cd "$PROJECT_ROOT"

    # 運行測試
    if go test -run TestMassiveVocabularyGaps tests/comprehensive_coverage_test.go -v > /dev/null 2>&1; then
        log_success "分類器驗證通過"
    else
        log_warning "分類器驗證未完全通過，建議檢查測試結果"
    fi
}

# 生成統計報告
generate_stats() {
    log_info "生成統計摘要..."

    cd "$PROJECT_ROOT"

    echo ""
    echo "📊 快速統計:"
    echo "----------------------------------------"

    for level in {1..5}; do
        file="services/keyword_classifier_l${level}.go"
        if [ -f "$file" ]; then
            count=$(grep -o '"[^"]*"' "$file" | wc -l | tr -d ' ')
            echo "L${level}: ${count} 個關鍵字"
        fi
    done

    total=$(grep -o '"[^"]*"' services/keyword_classifier_l*.go | wc -l | tr -d ' ')
    echo "----------------------------------------"
    echo "總計: ${total} 個關鍵字"
    echo ""
}

# 主函數
main() {
    case "${1:-help}" in
        "analyze")
            check_go
            generate_stats
            run_analysis
            ;;
        "clean")
            check_go
            generate_stats
            run_cleanup
            validate_classifier
            generate_stats
            ;;
        "both")
            check_go
            log_info "執行完整關鍵字管理流程..."
            generate_stats
            run_analysis
            echo ""
            read -p "是否繼續進行清理? (y/N): " -n 1 -r
            echo ""
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                run_cleanup
                validate_classifier
                generate_stats
            else
                log_info "僅完成分析，跳過清理"
            fi
            ;;
        "help")
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 執行主函數
main "$@"