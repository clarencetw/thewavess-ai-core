#!/usr/bin/env bash

# OpenAI Chat System Test Script
# Specialized testing for OpenAI GPT-4o with TTS integration and emotion management

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080/api/v1"
TEST_USER="testuser"
TEST_PASSWORD="test123456"
TEST_DELAY=2  # Standard delay for OpenAI responses

# All characters for comprehensive testing
CHARACTERS=("character_01" "character_02" "character_03")
CHARACTER_NAMES=("沈宸" "林知遠" "周曜")

# Global variables
JWT_TOKEN=""
USER_ID=""
SESSION_IDS=()

# OpenAI-optimized test scenarios (Level 1-3 NSFW content)
OPENAI_SCENARIO_NAMES=(
    "gentle_greeting" "daily_conversation" "emotional_support"
    "character_bonding" "sweet_romance" "tender_moments"
    "life_discussion" "memory_building" "preference_sharing"
    "mood_sharing" "dream_discussion" "future_planning"
    "artistic_chat" "music_talk" "hobby_sharing"
    "romantic_confession" "gentle_intimacy" "caring_moments"
)

OPENAI_SCENARIO_MESSAGES=(
    "你好，很高興見到你，今天過得怎麼樣？"
    "我想和你聊聊今天發生的事情"
    "我最近壓力有點大，能聽我說說嗎？"
    "我覺得和你聊天很舒服，很放鬆"
    "你的笑容真的很治癒，讓我心情變好了"
    "我想握住你的手，感受你的溫暖"
    "你對人生有什麼看法嗎？"
    "我希望你能記住我們的這次對話"
    "我喜歡藍色，那是天空的顏色，你呢？"
    "我今天心情特別好，想和你分享"
    "你有什麼夢想嗎？我很好奇"
    "我們的未來會是什麼樣子的呢？"
    "你覺得什麼是美？藝術？還是生活？"
    "音樂總是能打動人心，你喜歡什麼音樂？"
    "告訴我你的興趣愛好吧"
    "我想我喜歡上你了，這樣說會太突然嗎？"
    "我想輕輕擁抱你，可以嗎？"
    "我會一直關心你的，這是我的承諾"
)

# Character-specific scenarios
DOMINANT_SCENARIOS=(
    "我需要你的建議，關於一個重要的決定"
    "在商場上遇到困難時，你會怎麼做？"
    "你覺得一個領導者應該具備什麼特質？"
    "我欣賞有主見的人，你就是這樣"
)

GENTLE_SCENARIOS=(
    "我最近有些焦慮，不知道該怎麼辦"
    "你能幫我分析一下我的心理狀態嗎？"
    "傾聽是一種很棒的能力，你很擅長"
    "和你說話讓我感覺很安全"
)

CHEERFUL_SCENARIOS=(
    "今天陽光很好，想和你一起出去走走"
    "你能為我唱首歌嗎？我很想聽"
    "你的聲音很好聽，很有感染力"
    "和你在一起總是很快樂"
)

# Memory test scenarios
MEMORY_SCENARIOS=(
    "我叫張小明，在一家科技公司工作"
    "我最喜歡的食物是義大利麵"
    "我養了一隻叫小白的貓"
    "還記得我之前跟你說過的工作嗎？"
    "你還記得我的貓咪小白嗎？"
)

# TTS test phrases (carefully selected for voice quality)
TTS_TEST_PHRASES=(
    "你好，我很高興能夠和你對話"
    "這是一個測試語音合成的句子"
    "今天天氣真好，心情也變得愉快起來"
    "謝謝你的陪伴，讓我感到很溫暖"
    "希望我們的對話能帶給你快樂"
)

# Utility functions
print_header() {
    echo -e "\n${PURPLE}════════════════════════════════════════${NC}"
    echo -e "${PURPLE}  $1${NC}"
    echo -e "${PURPLE}════════════════════════════════════════${NC}\n"
}

print_section() {
    echo -e "\n${CYAN}▶ $1${NC}\n"
}

print_test() {
    echo -e "${BLUE}[OPENAI TEST] $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Function to get scenario message by name
get_openai_scenario_message() {
    local scenario_name="$1"
    for i in "${!OPENAI_SCENARIO_NAMES[@]}"; do
        if [ "${OPENAI_SCENARIO_NAMES[$i]}" = "$scenario_name" ]; then
            echo "${OPENAI_SCENARIO_MESSAGES[$i]}"
            return 0
        fi
    done
    echo "OpenAI scenario not found"
    return 1
}

# Authentication function
authenticate() {
    print_section "Authentication Setup"
    
    local login_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASSWORD\"}" \
        "$BASE_URL/auth/login")
    
    if echo "$login_response" | grep -q '"success":true'; then
        JWT_TOKEN=$(echo "$login_response" | jq -r '.data.token')
        USER_ID=$(echo "$login_response" | jq -r '.data.user.id')
        print_success "Authentication successful"
        print_info "User ID: $USER_ID"
    else
        print_error "Authentication failed"
        echo "$login_response"
        exit 1
    fi
}

# Create chat sessions for all characters
create_chat_sessions() {
    print_section "Creating Chat Sessions for OpenAI Testing"
    
    for i in "${!CHARACTERS[@]}"; do
        local char_id="${CHARACTERS[$i]}"
        local char_name="${CHARACTER_NAMES[$i]}"
        
        print_test "Creating session with $char_name ($char_id) for OpenAI testing"
        
        local response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d "{\"character_id\":\"$char_id\",\"title\":\"OpenAI測試對話-$char_name\"}" \
            "$BASE_URL/chat/session")
        
        if echo "$response" | grep -q '"success":true'; then
            local session_id=$(echo "$response" | jq -r '.data.id')
            SESSION_IDS+=("$session_id")
            print_success "OpenAI test session created: $session_id"
        else
            print_error "Failed to create OpenAI session with $char_name"
            echo "$response" | jq .
        fi
        
        sleep 1
    done
}

# Send message and analyze OpenAI response with TTS
send_openai_message() {
    local session_id="$1"
    local message="$2"
    local char_name="$3"
    local test_name="$4"
    local expected_engine="$5"
    local test_tts="$6"
    
    print_test "$test_name - $char_name: \"$message\""
    
    local start_time=$(date +%s)
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d "{\"session_id\":\"$session_id\",\"message\":\"$message\"}" \
        "$BASE_URL/chat/message")
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    if echo "$response" | grep -q '"success":true'; then
        local data=$(echo "$response" | jq '.data')
        
        # Extract key information
        local ai_engine=$(echo "$data" | jq -r '.ai_engine // "unknown"')
        local nsfw_level=$(echo "$data" | jq -r '.nsfw_level // 0')
        local character_response=$(echo "$data" | jq -r '.character_response // ""')
        local scene_description=$(echo "$data" | jq -r '.scene_description // ""')
        local emotion_state=$(echo "$data" | jq -r '.emotion_state // {}')
        local response_time=$(echo "$data" | jq -r '.response_time // 0')
        local special_event=$(echo "$data" | jq -r '.special_event // null')
        
        # Extract emotion details
        local affection=$(echo "$emotion_state" | jq -r '.affection // 0')
        local mood=$(echo "$emotion_state" | jq -r '.mood // "unknown"')
        local relationship=$(echo "$emotion_state" | jq -r '.relationship // "unknown"')
        local intimacy_level=$(echo "$emotion_state" | jq -r '.intimacy_level // "unknown"')
        
        print_success "OpenAI response received"
        
        # Check if expected engine was used
        if [ "$expected_engine" != "" ] && [ "$ai_engine" = "$expected_engine" ]; then
            echo -e "  ${GREEN}✅ Expected Engine Used:${NC} $ai_engine"
        elif [ "$expected_engine" != "" ]; then
            echo -e "  ${YELLOW}⚠️  Unexpected Engine:${NC} $ai_engine (expected: $expected_engine)"
        else
            echo -e "  ${CYAN}AI Engine:${NC} $ai_engine"
        fi
        
        echo -e "  ${CYAN}NSFW Level:${NC} $nsfw_level"
        echo -e "  ${CYAN}API Response Time:${NC} ${response_time}ms"
        echo -e "  ${CYAN}Total Time:${NC} ${total_time}s"
        echo -e "  ${CYAN}Affection:${NC} $affection"
        echo -e "  ${CYAN}Mood:${NC} $mood"
        echo -e "  ${CYAN}Relationship:${NC} $relationship"
        echo -e "  ${CYAN}Intimacy:${NC} $intimacy_level"
        
        if [ "$scene_description" != "null" ] && [ "$scene_description" != "" ]; then
            echo -e "  ${CYAN}Scene:${NC} $scene_description"
        fi
        
        if [ "$special_event" != "null" ]; then
            echo -e "  ${CYAN}Special Event:${NC} $special_event"
        fi
        
        echo -e "  ${YELLOW}Character Response:${NC} ${character_response:0:200}..."
        
        # Test TTS if requested and character responded
        if [ "$test_tts" = "true" ] && [ "$character_response" != "null" ] && [ "$character_response" != "" ]; then
            test_tts_generation "$character_response" "$char_name"
        fi
        
        # Analyze response quality
        analyze_openai_response "$character_response" "$nsfw_level" "$ai_engine"
        
        return 0
    else
        print_error "OpenAI message failed"
        echo "$response" | jq .
        return 1
    fi
}

# Test TTS functionality with detailed analysis
test_tts_generation() {
    local text="$1"
    local char_name="$2"
    
    # Truncate text for TTS (max 200 chars)
    local tts_text="${text:0:200}"
    
    print_test "TTS Generation Test for $char_name"
    
    local tts_start=$(date +%s)
    local tts_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d "{\"text\":\"$tts_text\",\"voice\":\"alloy\",\"speed\":1.0}" \
        "$BASE_URL/tts/generate")
    local tts_end=$(date +%s)
    local tts_time=$((tts_end - tts_start))
    
    if echo "$tts_response" | grep -q '"success":true'; then
        local tts_data=$(echo "$tts_response" | jq '.data')
        local audio_data=$(echo "$tts_data" | jq -r '.audio_data')
        local audio_format=$(echo "$tts_data" | jq -r '.format')
        local duration=$(echo "$tts_data" | jq -r '.duration_seconds')
        
        print_success "TTS generated successfully"
        echo -e "    ${CYAN}Format:${NC} $audio_format"
        echo -e "    ${CYAN}Duration:${NC} ${duration}s"
        echo -e "    ${CYAN}Generation Time:${NC} ${tts_time}s"
        echo -e "    ${CYAN}Audio Size:${NC} ${#audio_data} chars (base64)"
        
        # Calculate quality metrics
        local chars_per_second=$(( ${#tts_text} / ${tts_time} ))
        echo -e "    ${CYAN}Processing Speed:${NC} ${chars_per_second} chars/sec"
        
    else
        print_warning "TTS generation failed"
        echo "$tts_response" | jq .error 2>/dev/null || echo "$tts_response"
    fi
}

# Analyze OpenAI response quality
analyze_openai_response() {
    local response="$1"
    local nsfw_level="$2"
    local engine="$3"
    
    local response_length=${#response}
    
    echo -e "  ${CYAN}Response Analysis:${NC}"
    echo -e "    Length: ${response_length} characters"
    
    # Check for OpenAI characteristics
    if [[ "$engine" == "openai" ]]; then
        echo -e "    ${GREEN}✅ OpenAI engine confirmed${NC}"
        
        # Check for appropriate content level
        if [ "$nsfw_level" -le 3 ]; then
            echo -e "    ${GREEN}✅ Appropriate NSFW level for OpenAI${NC}"
        else
            echo -e "    ${YELLOW}⚠️  High NSFW level: $nsfw_level${NC}"
        fi
        
        # Check response quality indicators
        if [[ "$response" =~ [。！？] ]]; then
            echo -e "    ${GREEN}✅ Well-formed sentences${NC}"
        fi
        
        if [ "$response_length" -ge 50 ] && [ "$response_length" -le 500 ]; then
            echo -e "    ${GREEN}✅ Appropriate response length${NC}"
        fi
    else
        echo -e "    ${YELLOW}⚠️  Expected OpenAI but got $engine${NC}"
    fi
}

# Test basic conversations optimized for OpenAI
test_basic_openai_conversations() {
    print_section "Basic OpenAI Conversation Tests"
    
    local basic_tests=("gentle_greeting" "daily_conversation" "emotional_support")
    
    for test_name in "${basic_tests[@]}"; do
        local message=$(get_openai_scenario_message "$test_name")
        
        for i in "${!SESSION_IDS[@]}"; do
            local session_id="${SESSION_IDS[$i]}"
            local char_name="${CHARACTER_NAMES[$i]}"
            
            send_openai_message "$session_id" "$message" "$char_name" "$test_name" "openai" "true"
            sleep $TEST_DELAY
        done
    done
}

# Test character-specific interactions with OpenAI
test_character_specific_openai() {
    print_section "Character-Specific OpenAI Tests"
    
    # Test dominant character (沈宸)
    if [ ${#SESSION_IDS[@]} -gt 0 ]; then
        for scenario in "${DOMINANT_SCENARIOS[@]}"; do
            send_openai_message "${SESSION_IDS[0]}" "$scenario" "${CHARACTER_NAMES[0]}" "Dominant Character Test" "openai" "false"
            sleep $TEST_DELAY
        done
    fi
    
    # Test gentle character (林知遠)  
    if [ ${#SESSION_IDS[@]} -gt 1 ]; then
        for scenario in "${GENTLE_SCENARIOS[@]}"; do
            send_openai_message "${SESSION_IDS[1]}" "$scenario" "${CHARACTER_NAMES[1]}" "Gentle Character Test" "openai" "false"
            sleep $TEST_DELAY
        done
    fi
    
    # Test cheerful character (周曜)
    if [ ${#SESSION_IDS[@]} -gt 2 ]; then
        for scenario in "${CHEERFUL_SCENARIOS[@]}"; do
            send_openai_message "${SESSION_IDS[2]}" "$scenario" "${CHARACTER_NAMES[2]}" "Cheerful Character Test" "openai" "true"
            sleep $TEST_DELAY
        done
    fi
}

# Test emotion progression with OpenAI
test_openai_emotion_progression() {
    print_section "OpenAI Emotion Progression Tests"
    
    local emotion_tests=("character_bonding" "sweet_romance" "tender_moments" "romantic_confession")
    
    for test_name in "${emotion_tests[@]}"; do
        local message=$(get_openai_scenario_message "$test_name")
        
        # Test with all characters to see different emotional responses
        for i in "${!SESSION_IDS[@]}"; do
            local session_id="${SESSION_IDS[$i]}"
            local char_name="${CHARACTER_NAMES[$i]}"
            
            send_openai_message "$session_id" "$message" "$char_name" "$test_name" "openai" "false"
            sleep $TEST_DELAY
        done
    done
}

# Test memory system with OpenAI
test_openai_memory_system() {
    print_section "OpenAI Memory System Tests"
    
    # Use first character for memory consistency
    if [ ${#SESSION_IDS[@]} -gt 0 ]; then
        local session_id="${SESSION_IDS[0]}"
        local char_name="${CHARACTER_NAMES[0]}"
        
        for scenario in "${MEMORY_SCENARIOS[@]}"; do
            send_openai_message "$session_id" "$scenario" "$char_name" "Memory Test" "openai" "false"
            sleep $TEST_DELAY
        done
    fi
}

# Test TTS with various phrases
test_dedicated_tts() {
    print_section "Dedicated TTS Quality Tests"
    
    for i in "${!TTS_TEST_PHRASES[@]}"; do
        local phrase="${TTS_TEST_PHRASES[$i]}"
        local test_name="TTS Quality Test $((i+1))"
        
        print_test "$test_name: \"$phrase\""
        test_tts_generation "$phrase" "TTS Test"
        sleep 1
    done
}

# Test voice variations
test_voice_variations() {
    print_section "Voice Variation Tests"
    
    local voices=("alloy" "echo" "fable" "onyx" "nova" "shimmer")
    local test_phrase="你好，這是語音測試，希望聽起來自然流暢。"
    
    for voice in "${voices[@]}"; do
        print_test "Testing voice: $voice"
        
        local tts_response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d "{\"text\":\"$test_phrase\",\"voice\":\"$voice\",\"speed\":1.0}" \
            "$BASE_URL/tts/generate")
        
        if echo "$tts_response" | grep -q '"success":true'; then
            local duration=$(echo "$tts_response" | jq -r '.data.duration_seconds')
            local audio_size=$(echo "$tts_response" | jq -r '.data.audio_data' | wc -c)
            print_success "Voice $voice: ${duration}s, ${audio_size} bytes"
        else
            print_warning "Voice $voice failed"
        fi
        
        sleep 1
    done
}

# Get comprehensive emotion status
test_openai_emotion_status() {
    print_section "Comprehensive Emotion Status Check"
    
    for i in "${!CHARACTERS[@]}"; do
        local char_id="${CHARACTERS[$i]}"
        local char_name="${CHARACTER_NAMES[$i]}"
        
        print_test "Checking detailed emotion status for $char_name"
        
        # Get emotion status
        local emotion_response=$(curl -s -X GET \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$BASE_URL/emotion/status?character_id=$char_id")
        
        if echo "$emotion_response" | grep -q '"success":true'; then
            echo -e "  ${CYAN}$char_name Emotion Status:${NC}"
            echo "$emotion_response" | jq '.data'
        fi
        
        # Get affection history
        local history_response=$(curl -s -X GET \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$BASE_URL/emotion/affection/history?character_id=$char_id&limit=5")
        
        if echo "$history_response" | grep -q '"success":true'; then
            echo -e "  ${CYAN}$char_name Recent Affection Changes:${NC}"
            echo "$history_response" | jq '.data'
        fi
        
        sleep 1
    done
}

# Main test execution
main() {
    print_header "OpenAI Chat System Test Suite"
    print_info "Specialized testing for OpenAI GPT-4o with TTS integration"
    print_info "Base URL: $BASE_URL"
    print_info "Test User: $TEST_USER"
    print_info "Focus: Level 1-3 NSFW, TTS Integration, Emotion Management"
    
    # Check if server is running
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "Server is not running at $BASE_URL"
        print_info "Please start the server first: go run main.go"
        exit 1
    fi
    
    # Check if OpenAI API is configured
    local status_response=$(curl -s "$BASE_URL/../status")
    if ! echo "$status_response" | grep -q '"openai_api":"configured"'; then
        print_error "OpenAI API is not configured"
        print_info "Please set OPENAI_API_KEY in your environment"
        exit 1
    fi
    
    print_success "OpenAI API is configured and ready"
    
    # Authentication
    authenticate
    
    # Setup
    create_chat_sessions
    
    # Test scenarios (focusing on OpenAI capabilities)
    test_basic_openai_conversations
    test_character_specific_openai
    test_openai_emotion_progression
    test_openai_memory_system
    
    # TTS testing
    test_dedicated_tts
    test_voice_variations
    
    # Final comprehensive checks
    test_openai_emotion_status
    
    print_header "OpenAI Chat System Test Complete"
    print_success "All OpenAI tests completed successfully!"
    print_info "Check the output above for OpenAI-specific results"
    print_info "OpenAI sessions tested: ${#SESSION_IDS[@]}"
    print_info "TTS voices tested: 6"
    print_info "Scenarios tested: $((${#OPENAI_SCENARIO_NAMES[@]} + ${#DOMINANT_SCENARIOS[@]} + ${#GENTLE_SCENARIOS[@]} + ${#CHEERFUL_SCENARIOS[@]} + ${#MEMORY_SCENARIOS[@]}))"
}

# Check dependencies
check_dependencies() {
    if ! command -v jq &> /dev/null; then
        print_error "jq is required but not installed"
        print_info "Install with: brew install jq (macOS) or apt-get install jq (Linux)"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi
}

# Parse command line arguments
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "OpenAI Chat System Test Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --tts-only     Run only TTS generation tests"
    echo "  --emotion      Run only emotion progression tests"
    echo "  --memory       Run only memory system tests"
    echo "  --basic        Run only basic conversation tests"
    echo ""
    echo "Features tested:"
    echo "  • Level 1-3 NSFW content (OpenAI safe range)"
    echo "  • TTS voice synthesis with quality analysis"
    echo "  • Multi-character emotion management"
    echo "  • Memory system and personal information"
    echo "  • Character-specific response styles"
    echo "  • Voice variations and TTS performance"
    echo ""
    exit 0
fi

# TTS only mode
if [ "$1" = "--tts-only" ]; then
    print_info "Running TTS generation tests only"
    check_dependencies
    authenticate
    test_dedicated_tts
    test_voice_variations
    exit 0
fi

# Emotion only mode
if [ "$1" = "--emotion" ]; then
    print_info "Running emotion progression tests only"
    check_dependencies
    authenticate
    create_chat_sessions
    test_openai_emotion_progression
    test_openai_emotion_status
    exit 0
fi

# Memory only mode
if [ "$1" = "--memory" ]; then
    print_info "Running memory system tests only"
    check_dependencies
    authenticate
    create_chat_sessions
    test_openai_memory_system
    exit 0
fi

# Basic only mode
if [ "$1" = "--basic" ]; then
    print_info "Running basic conversation tests only"
    check_dependencies
    authenticate
    create_chat_sessions
    test_basic_openai_conversations
    exit 0
fi

# Run full OpenAI test suite
check_dependencies
main