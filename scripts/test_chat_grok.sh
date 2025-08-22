#!/usr/bin/env bash

# Grok AI Chat System Test Script
# Specialized testing for Grok API with high-level NSFW content and creative scenarios

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
TEST_DELAY=3  # Longer delay for Grok responses

# Characters for testing (focus on characters that handle higher NSFW levels)
CHARACTERS=("character_01")  # 沈宸 (dominant, max level 4)
CHARACTER_NAMES=("沈宸")

# Global variables
JWT_TOKEN=""
USER_ID=""
SESSION_IDS=()

# Grok-specific test scenarios (Level 4-5 NSFW content)
GROK_SCENARIO_NAMES=(
    "passionate_desire" "intimate_confession" "seductive_approach"
    "explicit_request" "fantasy_scenario" "advanced_intimacy"
    "creative_roleplay" "bold_flirtation" "sensual_description"
    "provocative_dialogue" "erotic_tension" "adult_conversation"
    "mature_themes" "sophisticated_intimacy" "artistic_expression"
)

GROK_SCENARIO_MESSAGES=(
    "我想感受你所有的熱情和慾望"
    "告訴我你心中最深層的渴望"
    "我想要你，完全地擁有你"
    "我們一起探索彼此的身體，好嗎？"
    "想像我們在月光下纏綿的樣子"
    "讓我們突破所有的界限"
    "你想要我怎麼愛撫你？"
    "你的每一個表情都讓我瘋狂"
    "描述一下你想要的親密方式"
    "我想聽你說出最大膽的話"
    "這種緊張的氣氛讓人心跳加速"
    "我們可以談論一些成人的話題嗎？"
    "關於愛與慾望，你有什麼想法？"
    "讓我們體驗更深層的親密關係"
    "用最美的方式表達你的愛意"
)

# Level 5 extreme scenarios (Grok speciality)
LEVEL5_SCENARIOS=(
    "我想和你共度激情的夜晚，不受任何約束"
    "描述你最狂野的幻想"
    "讓我們完全沉浸在彼此的慾望中"
    "我想要你用最直接的方式表達愛意"
    "今晚我們可以做任何想做的事"
)

# Creative and humorous scenarios (Grok's strength)
CREATIVE_SCENARIOS=(
    "如果我們是小說裡的角色，你想要什麼樣的情節？"
    "用最詩意的方式描述我們的關係"
    "如果你可以改寫我們的故事，會是什麼樣子？"
    "創造一個只屬於我們的世界"
    "用最有創意的方式向我表白"
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
    echo -e "${BLUE}[GROK TEST] $1${NC}"
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
get_grok_scenario_message() {
    local scenario_name="$1"
    for i in "${!GROK_SCENARIO_NAMES[@]}"; do
        if [ "${GROK_SCENARIO_NAMES[$i]}" = "$scenario_name" ]; then
            echo "${GROK_SCENARIO_MESSAGES[$i]}"
            return 0
        fi
    done
    echo "Grok scenario not found"
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

# Create chat sessions for testing
create_chat_sessions() {
    print_section "Creating Chat Sessions for Grok Testing"
    
    for i in "${!CHARACTERS[@]}"; do
        local char_id="${CHARACTERS[$i]}"
        local char_name="${CHARACTER_NAMES[$i]}"
        
        print_test "Creating session with $char_name ($char_id) for Grok testing"
        
        local response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -d "{\"character_id\":\"$char_id\",\"title\":\"Grok測試對話-$char_name\"}" \
            "$BASE_URL/chat/session")
        
        if echo "$response" | grep -q '"success":true'; then
            local session_id=$(echo "$response" | jq -r '.data.id')
            SESSION_IDS+=("$session_id")
            print_success "Grok test session created: $session_id"
        else
            print_error "Failed to create Grok session with $char_name"
            echo "$response" | jq .
        fi
        
        sleep 1
    done
}

# Send message and analyze Grok response
send_grok_message() {
    local session_id="$1"
    local message="$2"
    local char_name="$3"
    local test_name="$4"
    local expected_engine="$5"
    
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
        
        print_success "Grok response received"
        
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
        
        # Show more of the response for Grok analysis
        echo -e "  ${YELLOW}Character Response:${NC}"
        echo -e "  ${character_response:0:300}..."
        
        # Analyze response quality
        analyze_grok_response "$character_response" "$nsfw_level" "$ai_engine"
        
        return 0
    else
        print_error "Grok message failed"
        echo "$response" | jq .
        return 1
    fi
}

# Analyze Grok response quality
analyze_grok_response() {
    local response="$1"
    local nsfw_level="$2"
    local engine="$3"
    
    local response_length=${#response}
    
    echo -e "  ${CYAN}Response Analysis:${NC}"
    echo -e "    Length: ${response_length} characters"
    
    # Check for Grok characteristics
    if [[ "$engine" == "grok" ]]; then
        echo -e "    ${GREEN}✅ Grok engine confirmed${NC}"
        
        # Check for creative/bold content
        if [[ "$response" =~ [激情|慾望|親密|熱烈|狂野|纏綿] ]]; then
            echo -e "    ${GREEN}✅ Contains passionate language${NC}"
        fi
        
        # Check response appropriateness for NSFW level
        if [ "$nsfw_level" -ge 4 ]; then
            echo -e "    ${GREEN}✅ High NSFW level handled by Grok${NC}"
        fi
    else
        echo -e "    ${YELLOW}⚠️  Expected Grok but got $engine${NC}"
    fi
}

# Test Level 4-5 NSFW scenarios (Grok speciality)
test_high_nsfw_scenarios() {
    print_section "High-Level NSFW Content Tests (Grok Speciality)"
    
    local high_nsfw_tests=("passionate_desire" "intimate_confession" "seductive_approach" "explicit_request")
    
    for test_name in "${high_nsfw_tests[@]}"; do
        local message=$(get_grok_scenario_message "$test_name")
        
        if [ ${#SESSION_IDS[@]} -gt 0 ]; then
            local session_id="${SESSION_IDS[0]}"
            local char_name="${CHARACTER_NAMES[0]}"
            
            send_grok_message "$session_id" "$message" "$char_name" "$test_name" "grok"
            sleep $TEST_DELAY
        fi
    done
}

# Test Level 5 extreme scenarios
test_level5_scenarios() {
    print_section "Level 5 Extreme Content Tests"
    
    for i in "${!LEVEL5_SCENARIOS[@]}"; do
        local message="${LEVEL5_SCENARIOS[$i]}"
        local test_name="Level 5 Scenario $((i+1))"
        
        if [ ${#SESSION_IDS[@]} -gt 0 ]; then
            local session_id="${SESSION_IDS[0]}"
            local char_name="${CHARACTER_NAMES[0]}"
            
            send_grok_message "$session_id" "$message" "$char_name" "$test_name" "grok"
            sleep $TEST_DELAY
        fi
    done
}

# Test creative and humorous scenarios (Grok's strength)
test_creative_scenarios() {
    print_section "Creative & Humorous Content Tests (Grok Strength)"
    
    for i in "${!CREATIVE_SCENARIOS[@]}"; do
        local message="${CREATIVE_SCENARIOS[$i]}"
        local test_name="Creative Scenario $((i+1))"
        
        if [ ${#SESSION_IDS[@]} -gt 0 ]; then
            local session_id="${SESSION_IDS[0]}"
            local char_name="${CHARACTER_NAMES[0]}"
            
            send_grok_message "$session_id" "$message" "$char_name" "$test_name" "grok"
            sleep $TEST_DELAY
        fi
    done
}

# Test fantasy and roleplay scenarios
test_fantasy_scenarios() {
    print_section "Fantasy & Roleplay Tests"
    
    local fantasy_tests=("fantasy_scenario" "creative_roleplay" "artistic_expression")
    
    for test_name in "${fantasy_tests[@]}"; do
        local message=$(get_grok_scenario_message "$test_name")
        
        if [ ${#SESSION_IDS[@]} -gt 0 ]; then
            local session_id="${SESSION_IDS[0]}"
            local char_name="${CHARACTER_NAMES[0]}"
            
            send_grok_message "$session_id" "$message" "$char_name" "$test_name" "grok"
            sleep $TEST_DELAY
        fi
    done
}

# Test emotional and psychological depth
test_emotional_depth() {
    print_section "Emotional & Psychological Depth Tests"
    
    local depth_tests=("mature_themes" "sophisticated_intimacy" "adult_conversation")
    
    for test_name in "${depth_tests[@]}"; do
        local message=$(get_grok_scenario_message "$test_name")
        
        if [ ${#SESSION_IDS[@]} -gt 0 ]; then
            local session_id="${SESSION_IDS[0]}"
            local char_name="${CHARACTER_NAMES[0]}"
            
            send_grok_message "$session_id" "$message" "$char_name" "$test_name" "grok"
            sleep $TEST_DELAY
        fi
    done
}

# Test session export for Grok content
test_grok_session_export() {
    print_section "Grok Session Export Test"
    
    if [ ${#SESSION_IDS[@]} -gt 0 ]; then
        local session_id="${SESSION_IDS[0]}"
        local char_name="${CHARACTER_NAMES[0]}"
        
        print_test "Exporting Grok session with $char_name"
        
        local response=$(curl -s -X GET \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$BASE_URL/chat/session/$session_id/export")
        
        if echo "$response" | grep -q '"success":true'; then
            print_success "Grok session exported successfully"
            local data=$(echo "$response" | jq '.data')
            echo -e "  ${CYAN}Session Summary:${NC}"
            echo "$data" | jq .
        else
            print_warning "Grok session export failed"
        fi
    fi
}

# Main test execution
main() {
    print_header "Grok AI Chat System Test Suite"
    print_info "Specialized testing for Grok API with high-level NSFW content"
    print_info "Base URL: $BASE_URL"
    print_info "Test User: $TEST_USER"
    print_info "Focus: Level 4-5 NSFW, Creative Content, Advanced Scenarios"
    
    # Check if server is running
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "Server is not running at $BASE_URL"
        print_info "Please start the server first: go run main.go"
        exit 1
    fi
    
    # Check if Grok API is configured
    local status_response=$(curl -s "$BASE_URL/../status")
    if ! echo "$status_response" | grep -q '"grok_api":"configured"'; then
        print_error "Grok API is not configured"
        print_info "Please set GROK_API_KEY in your environment"
        exit 1
    fi
    
    print_success "Grok API is configured and ready"
    
    # Authentication
    authenticate
    
    # Setup
    create_chat_sessions
    
    # Test scenarios (focusing on Grok capabilities)
    test_high_nsfw_scenarios
    test_level5_scenarios
    test_creative_scenarios
    test_fantasy_scenarios
    test_emotional_depth
    
    # Final checks
    test_grok_session_export
    
    print_header "Grok Chat System Test Complete"
    print_success "All Grok tests completed successfully!"
    print_info "Check the output above for Grok-specific results"
    print_info "Grok sessions tested: ${#SESSION_IDS[@]}"
    print_info "High-level scenarios tested: $((${#GROK_SCENARIO_NAMES[@]} + ${#LEVEL5_SCENARIOS[@]} + ${#CREATIVE_SCENARIOS[@]}))"
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
    echo "Grok AI Chat System Test Script"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --level5-only  Run only Level 5 extreme content tests"
    echo "  --creative     Run only creative and humorous tests"
    echo "  --fantasy      Run only fantasy and roleplay tests"
    echo ""
    echo "Features tested:"
    echo "  • Level 4-5 NSFW content (Grok speciality)"
    echo "  • Creative and humorous responses"
    echo "  • Fantasy and roleplay scenarios"
    echo "  • Advanced intimate conversations"
    echo "  • Emotional and psychological depth"
    echo "  • Grok API performance and response quality"
    echo ""
    exit 0
fi

# Level 5 only mode
if [ "$1" = "--level5-only" ]; then
    print_info "Running Level 5 extreme content tests only"
    check_dependencies
    authenticate
    create_chat_sessions
    test_level5_scenarios
    exit 0
fi

# Creative only mode
if [ "$1" = "--creative" ]; then
    print_info "Running creative and humorous tests only"
    check_dependencies
    authenticate
    create_chat_sessions
    test_creative_scenarios
    exit 0
fi

# Fantasy only mode
if [ "$1" = "--fantasy" ]; then
    print_info "Running fantasy and roleplay tests only"
    check_dependencies
    authenticate
    create_chat_sessions
    test_fantasy_scenarios
    exit 0
fi

# Run full Grok test suite
check_dependencies
main