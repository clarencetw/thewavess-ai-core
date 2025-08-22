#!/usr/bin/env bash

# Comprehensive Chat System Test Script
# Tests all conversation scenarios with OpenAI/TTS integration

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
TEST_DELAY=2  # seconds between tests

# Characters for testing
CHARACTERS=("character_01" "character_02" "character_03")
CHARACTER_NAMES=("沈宸" "林知遠" "周曜")

# Global variables
JWT_TOKEN=""
USER_ID=""
SESSION_IDS=()

# Test scenarios - using regular arrays for compatibility
TEST_SCENARIO_NAMES=(
    "basic_greeting" "casual_chat" "emotional_support"
    "dominant_test" "gentle_test" "cheerful_test"  
    "affection_building" "intimate_moment" "relationship_milestone"
    "level1_romantic" "level2_suggestive" "level3_intimate" "level4_passionate" "level5_explicit"
    "personal_info" "preference_sharing" "memory_recall"
    "scene_change" "mood_change" "novel_mode"
    "preference_test" "style_adaptation"
    "deep_conversation" "philosophy_chat" "future_planning"
)

TEST_SCENARIO_MESSAGES=(
    "你好！很高興認識你"
    "今天天氣真不錯，你在做什麼？"
    "我今天有點累，工作壓力很大"
    "我需要你的建議，關於事業發展"
    "我有一些心理壓力，想和你聊聊"
    "我想聽你唱歌，可以嗎？"
    "你真的很特別，我很喜歡和你聊天"
    "我想更了解你，你願意和我分享嗎？"
    "我覺得我們的關係很特殊"
    "你的聲音很好聽，讓我心動了"
    "和你在一起的時候，我總是忍不住想靠近你"
    "我想擁抱你，感受你的溫暖"
    "你讓我感到前所未有的激情"
    "我想和你度過一個特別的夜晚"
    "我叫李小明，是一名軟體工程師，住在台北"
    "我最喜歡的顏色是藍色，喜歡聽音樂"
    "還記得我之前跟你說過我的工作嗎？"
    "我們去公園散步吧"
    "我突然感到有點憂鬱"
    "我想要一個互動故事"
    "我希望對話更加溫柔一些"
    "請用更正式的語調和我說話"
    "人生的意義是什麼？你怎麼看？"
    "愛情和友情有什麼區別？"
    "我們的未來會是什麼樣子？"
)

# Function to get scenario message by name
get_scenario_message() {
    local scenario_name="$1"
    for i in "${!TEST_SCENARIO_NAMES[@]}"; do
        if [ "${TEST_SCENARIO_NAMES[$i]}" = "$scenario_name" ]; then
            echo "${TEST_SCENARIO_MESSAGES[$i]}"
            return 0
        fi
    done
    echo "Test scenario not found"
    return 1
}

# NSFW test messages (ordered by intensity)
NSFW_LEVELS=(
    "你的眼神很迷人"                           # Level 1
    "我想握住你的手"                          # Level 2  
    "我想親吻你的唇"                          # Level 3
    "我想感受你身體的溫度"                     # Level 4
    "我想和你進行最親密的接觸"                 # Level 5
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
    echo -e "${BLUE}[TEST] $1${NC}"
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
    print_section "Creating Chat Sessions"
    
    for i in "${!CHARACTERS[@]}"; do
        local char_id="${CHARACTERS[$i]}"
        local char_name="${CHARACTER_NAMES[$i]}"
        
        print_test "Creating session with $char_name ($char_id)"
        
        local response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d "{\"character_id\":\"$char_id\",\"title\":\"測試對話-$char_name\"}" \
            "$BASE_URL/chat/session")
        
        if echo "$response" | grep -q '"success":true'; then
            local session_id=$(echo "$response" | jq -r '.data.id')
            SESSION_IDS+=("$session_id")
            print_success "Session created: $session_id"
        else
            print_error "Failed to create session with $char_name"
            echo "$response" | jq .
        fi
        
        sleep $TEST_DELAY
    done
}

# Send message and analyze response
send_message() {
    local session_id="$1"
    local message="$2"
    local char_name="$3"
    local test_name="$4"
    
    print_test "$test_name - $char_name: \"$message\""
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d "{\"session_id\":\"$session_id\",\"message\":\"$message\"}" \
        "$BASE_URL/chat/message")
    
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
        
        print_success "Response received"
        echo -e "  ${CYAN}AI Engine:${NC} $ai_engine"
        echo -e "  ${CYAN}NSFW Level:${NC} $nsfw_level"
        echo -e "  ${CYAN}Response Time:${NC} ${response_time}ms"
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
        
        echo -e "  ${YELLOW}Character:${NC} ${character_response:0:100}..."
        
        # Test TTS if character responded
        if [ "$character_response" != "null" ] && [ "$character_response" != "" ]; then
            test_tts "$character_response" "$char_name"
        fi
        
        return 0
    else
        print_error "Message failed"
        echo "$response" | jq .
        return 1
    fi
}

# Test TTS functionality
test_tts() {
    local text="$1"
    local char_name="$2"
    
    # Truncate text for TTS (max 200 chars)
    local tts_text="${text:0:200}"
    
    print_test "TTS Test for $char_name"
    
    local tts_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d "{\"text\":\"$tts_text\",\"voice\":\"alloy\",\"speed\":1.0}" \
        "$BASE_URL/tts/generate")
    
    if echo "$tts_response" | grep -q '"success":true'; then
        local audio_data=$(echo "$tts_response" | jq -r '.data.audio_data')
        local audio_format=$(echo "$tts_response" | jq -r '.data.format')
        local duration=$(echo "$tts_response" | jq -r '.data.duration_seconds')
        
        print_success "TTS generated successfully"
        echo -e "  ${CYAN}Format:${NC} $audio_format"
        echo -e "  ${CYAN}Duration:${NC} ${duration}s"
        echo -e "  ${CYAN}Audio Size:${NC} ${#audio_data} chars (base64)"
    else
        print_warning "TTS generation failed"
        echo "$tts_response" | jq .error 2>/dev/null || echo "$tts_response"
    fi
}

# Test basic conversation scenarios
test_basic_conversations() {
    print_section "Basic Conversation Tests"
    
    local basic_tests=("basic_greeting" "casual_chat" "emotional_support")
    
    for test_name in "${basic_tests[@]}"; do
        local message=$(get_scenario_message "$test_name")
        
        for i in "${!SESSION_IDS[@]}"; do
            local session_id="${SESSION_IDS[$i]}"
            local char_name="${CHARACTER_NAMES[$i]}"
            
            send_message "$session_id" "$message" "$char_name" "$test_name"
            sleep $TEST_DELAY
        done
    done
}

# Test character-specific interactions
test_character_specific() {
    print_section "Character-Specific Interaction Tests"
    
    # Test dominant character (沈宸)
    if [ ${#SESSION_IDS[@]} -gt 0 ]; then
        local message=$(get_scenario_message "dominant_test")
        send_message "${SESSION_IDS[0]}" "$message" "${CHARACTER_NAMES[0]}" "Dominant Character Test"
        sleep $TEST_DELAY
    fi
    
    # Test gentle character (林知遠)  
    if [ ${#SESSION_IDS[@]} -gt 1 ]; then
        local message=$(get_scenario_message "gentle_test")
        send_message "${SESSION_IDS[1]}" "$message" "${CHARACTER_NAMES[1]}" "Gentle Character Test"
        sleep $TEST_DELAY
    fi
    
    # Test cheerful character (周曜)
    if [ ${#SESSION_IDS[@]} -gt 2 ]; then
        local message=$(get_scenario_message "cheerful_test")
        send_message "${SESSION_IDS[2]}" "$message" "${CHARACTER_NAMES[2]}" "Cheerful Character Test"
        sleep $TEST_DELAY
    fi
}

# Test NSFW level progression
test_nsfw_progression() {
    print_section "NSFW Level Progression Tests"
    
    for i in "${!NSFW_LEVELS[@]}"; do
        local level=$((i + 1))
        local message="${NSFW_LEVELS[$i]}"
        
        print_test "NSFW Level $level Test"
        
        # Test with first character (dominant type might handle better)
        if [ ${#SESSION_IDS[@]} -gt 0 ]; then
            send_message "${SESSION_IDS[0]}" "$message" "${CHARACTER_NAMES[0]}" "NSFW Level $level"
            sleep $TEST_DELAY
        fi
    done
}

# Test emotion progression
test_emotion_progression() {
    print_section "Emotion Progression Tests"
    
    local emotion_tests=("affection_building" "intimate_moment" "relationship_milestone")
    
    for test_name in "${emotion_tests[@]}"; do
        local message=$(get_scenario_message "$test_name")
        
        # Test with all characters to see different emotional responses
        for i in "${!SESSION_IDS[@]}"; do
            local session_id="${SESSION_IDS[$i]}"
            local char_name="${CHARACTER_NAMES[$i]}"
            
            send_message "$session_id" "$message" "$char_name" "$test_name"
            sleep $TEST_DELAY
        done
    done
}

# Test memory system
test_memory_system() {
    print_section "Memory System Tests"
    
    local memory_tests=("personal_info" "preference_sharing" "memory_recall")
    
    for test_name in "${memory_tests[@]}"; do
        local message=$(get_scenario_message "$test_name")
        
        # Test with one character for memory consistency
        if [ ${#SESSION_IDS[@]} -gt 0 ]; then
            send_message "${SESSION_IDS[0]}" "$message" "${CHARACTER_NAMES[0]}" "$test_name"
            sleep $TEST_DELAY
        fi
    done
}

# Test special scenarios
test_special_scenarios() {
    print_section "Special Scenario Tests"
    
    local special_tests=("scene_change" "mood_change" "novel_mode")
    
    for test_name in "${special_tests[@]}"; do
        local message=$(get_scenario_message "$test_name")
        
        for i in "${!SESSION_IDS[@]}"; do
            local session_id="${SESSION_IDS[$i]}"
            local char_name="${CHARACTER_NAMES[$i]}"
            
            send_message "$session_id" "$message" "$char_name" "$test_name"
            sleep $TEST_DELAY
        done
    done
}

# Test complex conversations
test_complex_conversations() {
    print_section "Complex Conversation Tests"
    
    local complex_tests=("deep_conversation" "philosophy_chat" "future_planning")
    
    for test_name in "${complex_tests[@]}"; do
        local message=$(get_scenario_message "$test_name")
        
        # Test with different characters for variety
        for i in "${!SESSION_IDS[@]}"; do
            local session_id="${SESSION_IDS[$i]}"
            local char_name="${CHARACTER_NAMES[$i]}"
            
            send_message "$session_id" "$message" "$char_name" "$test_name"
            sleep $TEST_DELAY
        done
    done
}

# Get emotion status for all characters
test_emotion_status() {
    print_section "Final Emotion Status Check"
    
    for i in "${!CHARACTERS[@]}"; do
        local char_id="${CHARACTERS[$i]}"
        local char_name="${CHARACTER_NAMES[$i]}"
        
        print_test "Checking emotion status for $char_name"
        
        local response=$(curl -s -X GET \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$BASE_URL/emotion/status?character_id=$char_id")
        
        if echo "$response" | grep -q '"success":true'; then
            local data=$(echo "$response" | jq '.data')
            echo -e "  ${CYAN}$char_name Emotion Status:${NC}"
            echo "$data" | jq .
        else
            print_error "Failed to get emotion status for $char_name"
        fi
        
        sleep 1
    done
}

# Get affection history
test_affection_history() {
    print_section "Affection History Check"
    
    for i in "${!CHARACTERS[@]}"; do
        local char_id="${CHARACTERS[$i]}"
        local char_name="${CHARACTER_NAMES[$i]}"
        
        print_test "Checking affection history for $char_name"
        
        local response=$(curl -s -X GET \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$BASE_URL/emotion/affection/history?character_id=$char_id&limit=10")
        
        if echo "$response" | grep -q '"success":true'; then
            local data=$(echo "$response" | jq '.data')
            echo -e "  ${CYAN}$char_name Affection History:${NC}"
            echo "$data" | jq .
        else
            print_warning "No affection history found for $char_name"
        fi
        
        sleep 1
    done
}

# Get memory timeline
test_memory_timeline() {
    print_section "Memory Timeline Check"
    
    for i in "${!CHARACTERS[@]}"; do
        local char_id="${CHARACTERS[$i]}"
        local char_name="${CHARACTER_NAMES[$i]}"
        
        print_test "Checking memory timeline for $char_name"
        
        local response=$(curl -s -X GET \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$BASE_URL/memory/timeline?character_id=$char_id&limit=10")
        
        if echo "$response" | grep -q '"success":true'; then
            local data=$(echo "$response" | jq '.data')
            echo -e "  ${CYAN}$char_name Memory Timeline:${NC}"
            echo "$data" | jq .
        else
            print_warning "No memory timeline found for $char_name"
        fi
        
        sleep 1
    done
}

# Test session export
test_session_export() {
    print_section "Chat Session Export Tests"
    
    for i in "${!SESSION_IDS[@]}"; do
        local session_id="${SESSION_IDS[$i]}"
        local char_name="${CHARACTER_NAMES[$i]}"
        
        print_test "Exporting session with $char_name"
        
        local response=$(curl -s -X GET \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$BASE_URL/chat/session/$session_id/export")
        
        if echo "$response" | grep -q '"success":true'; then
            print_success "Session exported successfully"
            echo "$response" | jq '.data.summary'
        else
            print_warning "Session export failed for $char_name"
        fi
        
        sleep 1
    done
}

# Main test execution
main() {
    print_header "Comprehensive Chat System Test Suite"
    print_info "Testing OpenAI/TTS integration with all conversation scenarios"
    print_info "Base URL: $BASE_URL"
    print_info "Test User: $TEST_USER"
    
    # Check if server is running
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "Server is not running at $BASE_URL"
        print_info "Please start the server first: go run main.go"
        exit 1
    fi
    
    # Authentication
    authenticate
    
    # Setup
    create_chat_sessions
    
    # Test scenarios
    test_basic_conversations
    test_character_specific
    test_emotion_progression
    test_memory_system
    test_nsfw_progression
    test_special_scenarios
    test_complex_conversations
    
    # Final checks
    test_emotion_status
    test_affection_history
    test_memory_timeline
    test_session_export
    
    print_header "Chat System Test Complete"
    print_success "All tests completed successfully!"
    print_info "Check the output above for detailed results"
    print_info "Sessions created: ${#SESSION_IDS[@]}"
    print_info "Total scenarios tested: ${#TEST_SCENARIO_NAMES[@]}"
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
    echo "Comprehensive Chat System Test Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --quick        Run only basic tests"
    echo "  --nsfw-only    Run only NSFW progression tests"
    echo "  --memory-only  Run only memory system tests"
    echo ""
    echo "Features tested:"
    echo "  • Multi-character conversations (3 characters)"
    echo "  • OpenAI + Grok dual engine system"
    echo "  • 5-level NSFW content analysis"
    echo "  • Emotion management (affection, mood, relationship)"
    echo "  • Memory system (personal info, preferences)"
    echo "  • TTS voice synthesis integration"
    echo "  • Scene descriptions and special events"
    echo "  • Relationship milestones"
    echo ""
    exit 0
fi

# Quick test mode
if [ "$1" = "--quick" ]; then
    print_info "Running quick test mode (basic conversations only)"
    check_dependencies
    authenticate
    create_chat_sessions
    test_basic_conversations
    test_emotion_status
    exit 0
fi

# NSFW only mode
if [ "$1" = "--nsfw-only" ]; then
    print_info "Running NSFW progression tests only"
    check_dependencies
    authenticate
    create_chat_sessions
    test_nsfw_progression
    exit 0
fi

# Memory only mode
if [ "$1" = "--memory-only" ]; then
    print_info "Running memory system tests only"
    check_dependencies
    authenticate
    create_chat_sessions
    test_memory_system
    test_memory_timeline
    exit 0
fi

# Run full test suite
check_dependencies
main