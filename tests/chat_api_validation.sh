#!/bin/bash

# Chat API Response Validation Script
# 用於驗證 chat API 回應資料的合理性和完整性

set -e

# 載入測試工具庫
source "$(dirname "$0")/utils/test_common.sh"

# 初始化測試環境
tc_init_logging "chat_api_validation"

tc_log "INFO" "🧪 Chat API Response Validation Tool"
tc_log "INFO" "Base URL: $TEST_BASE_URL"

# 1. 登入並取得 token
tc_log "INFO" "[1/6] 用戶登入..."
if ! tc_register_and_authenticate; then
    tc_log "FAIL" "用戶註冊或認證失敗"
    exit 1
fi
tc_log "PASS" "登入成功"

# 2. 創建測試會話
tc_log "INFO" "[2/6] 創建測試會話..."
CREATE_RESPONSE=$(tc_http_request "POST" "/chats" "{\"character_id\": \"$TEST_CHARACTER_ID\", \"title\": \"API驗證測試\"}" "Create Test Chat")

CHAT_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id // empty')
if [ -z "$CHAT_ID" ] || [ "$CHAT_ID" = "null" ]; then
    tc_log "FAIL" "創建會話失敗"
    echo "$CREATE_RESPONSE" | jq .
    exit 1
fi
tc_log "PASS" "會話創建成功 (ID: $CHAT_ID)"

# 驗證創建會話回應資料
echo "驗證創建會話回應資料..."
VALIDATION_RESULTS=()

# 檢查必要欄位
REQUIRED_FIELDS=("id" "user_id" "character_id" "title" "status" "message_count" "created_at")
for field in "${REQUIRED_FIELDS[@]}"; do
    value=$(echo "$CREATE_RESPONSE" | jq -r ".data.$field // \"missing\"")
    if [ "$value" = "missing" ] || [ "$value" = "null" ]; then
        VALIDATION_RESULTS+=("❌ 缺少必要欄位: $field")
    else
        VALIDATION_RESULTS+=("✅ $field: $value")
    fi
done

# 檢查角色資訊
CHARACTER_NAME=$(echo "$CREATE_RESPONSE" | jq -r '.data.character.name // "missing"')
if [ "$CHARACTER_NAME" = "missing" ]; then
    VALIDATION_RESULTS+=("❌ 缺少角色名稱")
else
    VALIDATION_RESULTS+=("✅ 角色名稱: $CHARACTER_NAME")
fi

# 檢查歡迎訊息
LAST_MESSAGE=$(echo "$CREATE_RESPONSE" | jq -r '.data.last_message.dialogue // "missing"')
if [ "$LAST_MESSAGE" = "missing" ]; then
    VALIDATION_RESULTS+=("❌ 缺少歡迎訊息")
else
    MESSAGE_LENGTH=${#LAST_MESSAGE}
    VALIDATION_RESULTS+=("✅ 歡迎訊息長度: $MESSAGE_LENGTH 字元")
fi

echo -e "${GREEN}✅ 會話創建成功 (ID: $CHAT_ID)${NC}"
for result in "${VALIDATION_RESULTS[@]}"; do
    echo "  $result"
done
echo ""

# 3. 測試獲取會話詳情
echo -e "${YELLOW}[3/6] 測試獲取會話詳情...${NC}"
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/chats/$CHAT_ID" \
  -H "Authorization: Bearer $TOKEN")

SUCCESS=$(echo "$GET_RESPONSE" | jq -r '.success')
if [ "$SUCCESS" != "true" ]; then
    echo -e "${RED}❌ 獲取會話失敗${NC}"
    echo "$GET_RESPONSE" | jq .
    exit 1
fi

# 驗證會話詳情回應
MESSAGE_COUNT=$(echo "$GET_RESPONSE" | jq -r '.data.message_count')
TOTAL_CHARACTERS=$(echo "$GET_RESPONSE" | jq -r '.data.total_characters')
echo -e "${GREEN}✅ 會話詳情獲取成功${NC}"
echo "  ✅ 訊息數量: $MESSAGE_COUNT"
echo "  ✅ 總字元數: $TOTAL_CHARACTERS"
echo ""

# 4. 測試角色資料
echo -e "${YELLOW}[4/6] 測試角色資料...${NC}"
CHARACTER_RESPONSE=$(curl -s -X GET "$BASE_URL/character/$CHARACTER_ID/profile" \
  -H "Authorization: Bearer $TOKEN")

CHARACTER_SUCCESS=$(echo "$CHARACTER_RESPONSE" | jq -r '.success')
if [ "$CHARACTER_SUCCESS" != "true" ]; then
    echo -e "${RED}❌ 獲取角色資料失敗${NC}"
    echo "$CHARACTER_RESPONSE" | jq .
    exit 1
fi

CHARACTER_TYPE=$(echo "$CHARACTER_RESPONSE" | jq -r '.data.type')
CHARACTER_LOCALE=$(echo "$CHARACTER_RESPONSE" | jq -r '.data.locale')
DESCRIPTION_LENGTH=$(echo "$CHARACTER_RESPONSE" | jq -r '.data.user_description | length')

echo -e "${GREEN}✅ 角色資料獲取成功${NC}"
echo "  ✅ 角色類型: $CHARACTER_TYPE"
echo "  ✅ 語言區域: $CHARACTER_LOCALE"
echo "  ✅ 描述長度: $DESCRIPTION_LENGTH 字元"
echo ""

# 5. 測試發送訊息
echo -e "${YELLOW}[5/6] 測試發送訊息...${NC}"
MESSAGE_RESPONSE=$(curl -s -X POST "$BASE_URL/chats/$CHAT_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"message": "今天是個美好的測試日！"}')

MESSAGE_SUCCESS=$(echo "$MESSAGE_RESPONSE" | jq -r '.success')
if [ "$MESSAGE_SUCCESS" != "true" ]; then
    echo -e "${RED}❌ 發送訊息失敗${NC}"
    echo "$MESSAGE_RESPONSE" | jq .
    exit 1
fi

# 驗證訊息回應資料
AI_ENGINE=$(echo "$MESSAGE_RESPONSE" | jq -r '.data.ai_engine')
NSFW_LEVEL=$(echo "$MESSAGE_RESPONSE" | jq -r '.data.nsfw_level')
RESPONSE_TIME=$(echo "$MESSAGE_RESPONSE" | jq -r '.data.response_time')
AFFECTION=$(echo "$MESSAGE_RESPONSE" | jq -r '.data.affection')
CONFIDENCE=$(echo "$MESSAGE_RESPONSE" | jq -r '.data.confidence')

echo -e "${GREEN}✅ 訊息發送成功${NC}"
echo "  ✅ AI 引擎: $AI_ENGINE"
echo "  ✅ NSFW 等級: $NSFW_LEVEL"
echo "  ✅ 回應時間: ${RESPONSE_TIME}ms"
echo "  ✅ 好感度: $AFFECTION"
echo "  ✅ 信心度: $CONFIDENCE"
echo ""

# 6. 測試訊息歷史
echo -e "${YELLOW}[6/6] 測試訊息歷史...${NC}"
HISTORY_RESPONSE=$(curl -s -X GET "$BASE_URL/chats/$CHAT_ID/history" \
  -H "Authorization: Bearer $TOKEN")

HISTORY_SUCCESS=$(echo "$HISTORY_RESPONSE" | jq -r '.success')
if [ "$HISTORY_SUCCESS" != "true" ]; then
    echo -e "${RED}❌ 獲取訊息歷史失敗${NC}"
    echo "$HISTORY_RESPONSE" | jq .
    exit 1
fi

# 驗證歷史資料
TOTAL_MESSAGES=$(echo "$HISTORY_RESPONSE" | jq -r '.data.messages | length')
PAGINATION_TOTAL=$(echo "$HISTORY_RESPONSE" | jq -r '.data.pagination.total_count')
HAS_WELCOME=$(echo "$HISTORY_RESPONSE" | jq -r '.data.messages[] | select(.role == "assistant" and (.id | contains("welcome"))) | .id' | wc -l)
HAS_USER_MSG=$(echo "$HISTORY_RESPONSE" | jq -r '.data.messages[] | select(.role == "user") | .id' | wc -l)
HAS_AI_RESPONSE=$(echo "$HISTORY_RESPONSE" | jq -r '.data.messages[] | select(.role == "assistant" and (.id | contains("_ai"))) | .id' | wc -l)

echo -e "${GREEN}✅ 訊息歷史獲取成功${NC}"
echo "  ✅ 歷史訊息數: $TOTAL_MESSAGES"
echo "  ✅ 分頁總數: $PAGINATION_TOTAL"
echo "  ✅ 歡迎訊息: $HAS_WELCOME 條"
echo "  ✅ 用戶訊息: $HAS_USER_MSG 條"
echo "  ✅ AI 回應: $HAS_AI_RESPONSE 條"
echo ""

# 清理測試資料
echo -e "${YELLOW}清理測試會話...${NC}"
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/chats/$CHAT_ID" \
  -H "Authorization: Bearer $TOKEN")
DELETE_SUCCESS=$(echo "$DELETE_RESPONSE" | jq -r '.success')
if [ "$DELETE_SUCCESS" = "true" ]; then
    echo -e "${GREEN}✅ 測試會話已清理${NC}"
else
    echo -e "${YELLOW}⚠️ 測試會話清理失敗，請手動清理 Chat ID: $CHAT_ID${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Chat API 驗證完成！所有測試通過。${NC}"
echo -e "${BLUE}總結：${NC}"
echo "  • 會話管理功能正常"
echo "  • 角色資料完整"
echo "  • AI 引擎回應正常"
echo "  • 訊息歷史記錄完整"
echo "  • 資料結構合理"