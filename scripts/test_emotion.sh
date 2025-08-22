#!/bin/bash

# 測試情感系統API的功能
BASE_URL="http://localhost:8080/api/v1"
USER_ID="test_user_01"
CHARACTER_ID="character_03"

echo "========== 情感系統API測試 =========="

echo "1. 測試情感狀態 API"
curl -s -H "Content-Type: application/json" \
     "${BASE_URL}/emotion/status?character_id=${CHARACTER_ID}&user_id=${USER_ID}" \
     | jq .

echo -e "\n2. 測試好感度 API"
curl -s -H "Content-Type: application/json" \
     "${BASE_URL}/emotion/affection?character_id=${CHARACTER_ID}&user_id=${USER_ID}" \
     | jq .

echo -e "\n3. 測試里程碑 API"  
curl -s -H "Content-Type: application/json" \
     "${BASE_URL}/emotion/milestones?character_id=${CHARACTER_ID}&user_id=${USER_ID}" \
     | jq .

echo -e "\n4. 測試好感度歷史 API"
curl -s -H "Content-Type: application/json" \
     "${BASE_URL}/emotion/affection/history?character_id=${CHARACTER_ID}&user_id=${USER_ID}&days=7" \
     | jq .

echo -e "\n5. 測試觸發情感事件 API"
curl -s -X POST -H "Content-Type: application/json" \
     -d '{"character_id":"'${CHARACTER_ID}'","user_id":"'${USER_ID}'","event_type":"compliment","intensity":5}' \
     "${BASE_URL}/emotion/trigger" \
     | jq .

echo -e "\n========== 測試完成 =========="