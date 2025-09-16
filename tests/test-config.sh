#!/bin/bash

# 🔧 Thewavess AI Core - 簡化測試配置
# 直接使用新的共用測試工具庫

# 載入共用工具庫
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/test_common.sh"

# 初始化完成提示
echo -e "${TC_CYAN}📋 Thewavess AI 測試工具已載入${TC_NC}"
echo -e "${TC_CYAN}  • API: $TEST_BASE_URL${TC_NC}"
echo -e "${TC_CYAN}  • 用戶: $TEST_USERNAME${TC_NC}"
echo -e "${TC_CYAN}  • 角色: $TEST_CHARACTER_ID${TC_NC}"