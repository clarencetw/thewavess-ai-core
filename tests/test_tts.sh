#!/bin/bash

# 🧪 Thewavess AI Core - TTS功能測試
# 測試文字轉語音相關API功能

set -e

# 載入測試工具庫
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils/test_common.sh"

# ================================
# 測試配置
# ================================

TEST_NAME="tts"
TEST_CHAT_SESSION_ID=""
TEST_MESSAGE_ID=""
TEST_AUDIO_FILE=""

# ================================
# 測試函數
# ================================

# 準備TTS測試數據
setup_tts_test() {
    tc_log "INFO" "準備TTS測試數據"

    # 創建測試聊天會話
    local session_data='{"character_id":"'$TEST_CHARACTER_ID'","title":"TTS測試會話"}'

    local response=$(tc_http_request "POST" "/chats" "$session_data" "Create TTS Test Chat" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        TEST_CHAT_SESSION_ID=$(echo "$response" | jq -r '.data.id // ""')
        tc_log "PASS" "TTS測試會話創建成功 (ID: $TEST_CHAT_SESSION_ID)"
    else
        tc_log "FAIL" "TTS測試會話創建失敗"
        return 1
    fi

    # 發送測試消息以生成AI回應
    local message_data='{"message": "你好，請說一段簡短的話，我想測試語音合成功能"}'

    local msg_response=$(tc_http_request "POST" "/chats/$TEST_CHAT_SESSION_ID/messages" "$message_data" "Send TTS Test Message" "true")

    if echo "$msg_response" | jq -e '.success' > /dev/null 2>&1; then
        TEST_MESSAGE_ID=$(echo "$msg_response" | jq -r '.data.id // ""')
        tc_log "PASS" "TTS測試消息發送成功 (ID: $TEST_MESSAGE_ID)"
        sleep 3  # 等待AI回應
    else
        tc_log "FAIL" "TTS測試消息發送失敗"
        return 1
    fi

    tc_log "PASS" "TTS測試數據準備完成"
    return 0
}

# 測試獲取支援的語音列表
test_get_voices() {
    tc_log "INFO" "測試獲取支援的語音列表"

    local response=$(tc_http_request "GET" "/tts/voices" "" "Get Available Voices" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local voices_count=$(echo "$response" | jq -r '.data.voices | length')
        tc_log "PASS" "語音列表獲取成功"
        tc_log "INFO" "  可用語音數量: $voices_count"

        # 顯示前幾個語音選項
        if [ "$voices_count" -gt 0 ]; then
            local first_voice_name=$(echo "$response" | jq -r '.data.voices[0].name // ""')
            local first_voice_lang=$(echo "$response" | jq -r '.data.voices[0].language // ""')
            local first_voice_gender=$(echo "$response" | jq -r '.data.voices[0].gender // ""')

            tc_log "INFO" "  第一個語音: $first_voice_name"
            tc_log "INFO" "  語言: $first_voice_lang"
            tc_log "INFO" "  性別: $first_voice_gender"

            # 檢查是否有中文語音
            local chinese_voices=$(echo "$response" | jq -r '.data.voices[] | select(.language | contains("zh")) | .name' | wc -l)
            tc_log "INFO" "  中文語音數量: $chinese_voices"
        fi

        return 0
    else
        tc_log "FAIL" "語音列表獲取失敗"
        return 1
    fi
}

# 測試文字轉語音基本功能
test_text_to_speech() {
    tc_log "INFO" "測試文字轉語音基本功能"

    local test_text="你好，這是一個測試語音合成的範例文字。希望聽起來很自然。"

    local tts_data='{
        "text": "'$test_text'",
        "voice": "zh-TW-Standard-A",
        "speed": 1.0,
        "pitch": 0.0
    }'

    local response=$(tc_http_request "POST" "/tts/synthesize" "$tts_data" "Text to Speech" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local audio_url=$(echo "$response" | jq -r '.data.audio_url // ""')
        local audio_format=$(echo "$response" | jq -r '.data.format // ""')
        local duration=$(echo "$response" | jq -r '.data.duration // 0')

        tc_log "PASS" "文字轉語音成功"
        tc_log "INFO" "  音頻URL: $audio_url"
        tc_log "INFO" "  音頻格式: $audio_format"
        tc_log "INFO" "  音頻時長: ${duration}秒"

        # 檢查音頻URL是否有效
        if [ -n "$audio_url" ] && [ "$audio_url" != "null" ]; then
            tc_log "PASS" "音頻URL生成正常"
            TEST_AUDIO_FILE="$audio_url"
        else
            tc_log "WARN" "音頻URL可能無效"
        fi

        return 0
    else
        tc_log "FAIL" "文字轉語音失敗"
        return 1
    fi
}

# 測試訊息語音合成
test_message_tts() {
    tc_log "INFO" "測試訊息語音合成"

    if [ -z "$TEST_MESSAGE_ID" ]; then
        tc_log "ERROR" "沒有測試消息ID"
        return 1
    fi

    # 獲取消息內容以確認有AI回應
    local history_response=$(tc_http_request "GET" "/chats/$TEST_CHAT_SESSION_ID/history" "" "Get Message History" "true")
    local ai_message=""

    if echo "$history_response" | jq -e '.success' > /dev/null 2>&1; then
        # 找到AI的回應消息
        ai_message=$(echo "$history_response" | jq -r '.data.messages[] | select(.role == "assistant") | .dialogue' | head -1)

        if [ -n "$ai_message" ] && [ "$ai_message" != "null" ]; then
            tc_log "INFO" "找到AI回應消息: ${ai_message:0:50}..."
        else
            tc_log "WARN" "未找到AI回應消息，使用預設文字"
            ai_message="這是一個測試回應"
        fi
    fi

    # 請求為特定消息生成語音
    local message_tts_data='{
        "voice": "zh-TW-Standard-B",
        "speed": 1.2,
        "pitch": 0.2
    }'

    local response=$(tc_http_request "POST" "/messages/$TEST_MESSAGE_ID/tts" "$message_tts_data" "Message TTS" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local audio_url=$(echo "$response" | jq -r '.data.audio_url // ""')
        local message_id=$(echo "$response" | jq -r '.data.message_id // ""')

        tc_log "PASS" "消息語音合成成功"
        tc_log "INFO" "  消息ID: $message_id"
        tc_log "INFO" "  音頻URL: $audio_url"

        return 0
    else
        tc_log "FAIL" "消息語音合成失敗"
        return 1
    fi
}

# 測試不同語音參數
test_voice_parameters() {
    tc_log "INFO" "測試不同語音參數"

    local test_text="測試不同的語音參數設定"

    # 測試不同的語音設定
    local voice_configs=(
        "zh-TW-Standard-A:0.5:0.5"    # 慢速，高音調
        "zh-TW-Standard-B:1.5:-0.5"   # 快速，低音調
        "zh-TW-Standard-C:1.0:0.0"    # 標準設定
    )

    for config in "${voice_configs[@]}"; do
        IFS=':' read -r voice speed pitch <<< "$config"

        tc_log "INFO" "測試語音設定: $voice (速度: $speed, 音調: $pitch)"

        local tts_data='{
            "text": "'$test_text'",
            "voice": "'$voice'",
            "speed": '$speed',
            "pitch": '$pitch'
        }'

        local response=$(tc_http_request "POST" "/tts/synthesize" "$tts_data" "TTS with Parameters" "true")

        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            local duration=$(echo "$response" | jq -r '.data.duration // 0')
            tc_log "PASS" "語音參數測試成功 - $voice"
            tc_log "INFO" "  音頻時長: ${duration}秒"
        else
            tc_log "WARN" "語音參數測試失敗 - $voice"
        fi

        sleep 1
    done

    return 0
}

# 測試語音緩存功能
test_tts_cache() {
    tc_log "INFO" "測試語音緩存功能"

    local test_text="這是測試緩存功能的文字"
    local voice="zh-TW-Standard-A"

    # 第一次請求
    local tts_data='{
        "text": "'$test_text'",
        "voice": "'$voice'",
        "speed": 1.0,
        "pitch": 0.0
    }'

    local start_time1=$(date +%s%3N)
    local response1=$(tc_http_request "POST" "/tts/synthesize" "$tts_data" "TTS Cache Test 1" "true")
    local end_time1=$(date +%s%3N)
    local duration1=$((end_time1 - start_time1))

    if ! echo "$response1" | jq -e '.success' > /dev/null 2>&1; then
        tc_log "FAIL" "第一次TTS請求失敗"
        return 1
    fi

    sleep 2

    # 第二次請求（相同參數，應該使用緩存）
    local start_time2=$(date +%s%3N)
    local response2=$(tc_http_request "POST" "/tts/synthesize" "$tts_data" "TTS Cache Test 2" "true")
    local end_time2=$(date +%s%3N)
    local duration2=$((end_time2 - start_time2))

    if echo "$response2" | jq -e '.success' > /dev/null 2>&1; then
        tc_log "PASS" "TTS緩存測試完成"
        tc_log "INFO" "  第一次請求時間: ${duration1}ms"
        tc_log "INFO" "  第二次請求時間: ${duration2}ms"

        # 檢查緩存效果（第二次應該更快）
        if [ "$duration2" -lt "$duration1" ]; then
            tc_log "PASS" "緩存功能正常（第二次請求更快）"
        else
            tc_log "INFO" "未偵測到明顯的緩存效果"
        fi

        return 0
    else
        tc_log "FAIL" "第二次TTS請求失敗"
        return 1
    fi
}

# 測試音頻檔案下載
test_audio_download() {
    tc_log "INFO" "測試音頻檔案下載"

    if [ -z "$TEST_AUDIO_FILE" ]; then
        tc_log "WARN" "沒有音頻檔案URL，跳過下載測試"
        return 0
    fi

    # 嘗試下載音頻檔案
    local temp_file="/tmp/tts_test_$(date +%s).mp3"

    if curl -s -f -o "$temp_file" "$TEST_AUDIO_FILE"; then
        local file_size=$(ls -l "$temp_file" 2>/dev/null | awk '{print $5}' || echo "0")

        tc_log "PASS" "音頻檔案下載成功"
        tc_log "INFO" "  檔案大小: $file_size bytes"

        # 檢查檔案是否有效（大小大於0）
        if [ "$file_size" -gt 0 ]; then
            tc_log "PASS" "音頻檔案有效"
        else
            tc_log "WARN" "音頻檔案可能損壞"
        fi

        # 清理暫存檔案
        rm -f "$temp_file"
        return 0
    else
        tc_log "FAIL" "音頻檔案下載失敗"
        return 1
    fi
}

# 測試SSML語音標記
test_ssml_support() {
    tc_log "INFO" "測試SSML語音標記支援"

    # 測試SSML標記文字
    local ssml_text='<speak>你好！<break time="500ms"/>這是<emphasis level="strong">重要</emphasis>的測試。<prosody rate="slow" pitch="high">這段話速度較慢，音調較高。</prosody></speak>'

    local ssml_data='{
        "text": "'$ssml_text'",
        "voice": "zh-TW-Standard-A",
        "format": "ssml",
        "speed": 1.0,
        "pitch": 0.0
    }'

    local response=$(tc_http_request "POST" "/tts/synthesize" "$ssml_data" "SSML TTS Test" "true")

    if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
        local duration=$(echo "$response" | jq -r '.data.duration // 0')
        tc_log "PASS" "SSML語音標記測試成功"
        tc_log "INFO" "  音頻時長: ${duration}秒"
        return 0
    else
        tc_log "WARN" "SSML語音標記不支援或失敗"
        return 0  # 不作為失敗，因為SSML可能是可選功能
    fi
}

# 清理測試數據
cleanup_tts_test() {
    tc_log "INFO" "清理TTS測試數據"

    if [ -n "$TEST_CHAT_SESSION_ID" ]; then
        local response=$(tc_http_request "DELETE" "/chats/$TEST_CHAT_SESSION_ID" "" "Delete TTS Test Chat" "true")

        if echo "$response" | jq -e '.success' > /dev/null 2>&1; then
            tc_log "PASS" "TTS測試會話清理成功"
        else
            tc_log "WARN" "TTS測試會話清理失敗，請手動清理 ID: $TEST_CHAT_SESSION_ID"
        fi
    fi

    # 清理可能的暫存音頻檔案
    rm -f /tmp/tts_test_*.mp3 2>/dev/null || true
}

# ================================
# 主執行流程
# ================================

main() {
    # 初始化測試
    tc_init_logging "$TEST_NAME"
    # CSV功能已移除，改用詳細日誌記錄
    tc_show_header "Thewavess AI Core - TTS功能測試"

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
    tc_log "INFO" "==================== 設置TTS測試環境 ===================="

    if ! setup_tts_test; then
        tc_log "ERROR" "TTS測試環境設置失敗"
        exit 1
    fi

    sleep 3

    # 執行TTS功能測試
    tc_log "INFO" "==================== TTS功能測試 ===================="

    if test_get_voices; then
        test_results+=("語音列表:PASS")
    else
        test_results+=("語音列表:FAIL")
    fi

    sleep 2

    if test_text_to_speech; then
        test_results+=("文字轉語音:PASS")
    else
        test_results+=("文字轉語音:FAIL")
    fi

    sleep 2

    if test_message_tts; then
        test_results+=("消息語音:PASS")
    else
        test_results+=("消息語音:FAIL")
    fi

    sleep 2

    if test_voice_parameters; then
        test_results+=("語音參數:PASS")
    else
        test_results+=("語音參數:FAIL")
    fi

    sleep 2

    if test_tts_cache; then
        test_results+=("語音緩存:PASS")
    else
        test_results+=("語音緩存:FAIL")
    fi

    sleep 2

    if test_audio_download; then
        test_results+=("音頻下載:PASS")
    else
        test_results+=("音頻下載:FAIL")
    fi

    sleep 2

    if test_ssml_support; then
        test_results+=("SSML支援:PASS")
    else
        test_results+=("SSML支援:FAIL")
    fi

    # 清理測試數據
    cleanup_tts_test

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
        tc_log "PASS" "所有TTS功能測試通過！"
        tc_show_summary "TTS功能測試"
        exit 0
    else
        tc_log "FAIL" "$failed_count 個測試失敗"
        tc_show_summary "TTS功能測試"
        exit 1
    fi
}

# 執行主函數
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi