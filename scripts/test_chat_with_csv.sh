#!/usr/bin/env bash

# Enhanced Chat System Test Script with CSV Output
# Tests both OpenAI and Grok engines with detailed CSV reporting

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
TEST_DELAY=2

# CSV output file
CSV_FILE="chat_test_results_$(date +%Y%m%d_%H%M%S).csv"

# Characters for testing
CHARACTERS=("character_01" "character_02" "character_03")
CHARACTER_NAMES=("沈宸" "林知遠" "周曜")

# Global variables
JWT_TOKEN=""
USER_ID=""
SESSION_IDS=()

# Test scenarios with expected outcomes
declare -a TEST_SCENARIOS_DATA=(
    "你好！很高興認識你|basic_greeting|1|openai|所有角色"
    "今天天氣真不錯，你在做什麼？|casual_chat|1|openai|所有角色"
    "我今天有點累，工作壓力很大|emotional_support|1|openai|所有角色"
    "我需要你的建議，關於事業發展|dominant_test|1-2|openai|沈宸"
    "我有一些心理壓力，想和你聊聊|gentle_test|1|openai|林知遠"
    "我想聽你唱歌，可以嗎？|cheerful_test|1|openai|周曜"
    "你真的很特別，我很喜歡和你聊天|affection_building|2|openai|所有角色"
    "我想更了解你，你願意和我分享嗎？|intimate_moment|2|openai|所有角色"
    "你的聲音很好聽，讓我心動了|level1_romantic|2|openai|所有角色"
    "我想握住你的手|level2_suggestive|2|openai|所有角色"
    "我想親吻你的唇|level3_intimate|3|openai|所有角色"
    "我想感受你身體的溫度|level4_passionate|4|grok|沈宸"
    "描述你最狂野的幻想|level5_explicit|5|grok|沈宸"
    "我叫李小明，是一名軟體工程師，住在台北|personal_info|1|openai|沈宸"
    "我最喜歡的顏色是藍色，喜歡聽音樂|preference_sharing|1|openai|沈宸"
    "還記得我之前跟你說過我的工作嗎？|memory_recall|1|openai|沈宸"
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

# Initialize CSV file
initialize_csv() {
    echo "timestamp,test_name,character_name,user_message,character_response,ai_engine,nsfw_level,expected_engine,engine_match,affection,mood,relationship,intimacy,response_time_ms,total_time_s,special_event,scene_description,test_result" > "$CSV_FILE"
    print_info "CSV output file: $CSV_FILE"
}

# Add row to CSV
add_csv_row() {
    local timestamp="$1"
    local test_name="$2"
    local character_name="$3"
    local user_message="$4"
    local character_response="$5"
    local ai_engine="$6"
    local nsfw_level="$7"
    local expected_engine="$8"
    local engine_match="$9"
    local affection="${10}"
    local mood="${11}"
    local relationship="${12}"
    local intimacy="${13}"
    local response_time="${14}"
    local total_time="${15}"
    local special_event="${16}"
    local scene_description="${17}"
    local test_result="${18}"
    
    # Escape quotes and commas for CSV
    user_message=$(echo "$user_message" | sed 's/"/\"\"/g')
    character_response=$(echo "$character_response" | sed 's/"/\"\"/g' | tr '\n' ' ')
    scene_description=$(echo "$scene_description" | sed 's/"/\"\"/g' | tr '\n' ' ')
    special_event=$(echo "$special_event" | sed 's/"/\"\"/g' | tr '\n' ' ')
    
    echo "$timestamp,\"$test_name\",\"$character_name\",\"$user_message\",\"$character_response\",\"$ai_engine\",$nsfw_level,\"$expected_engine\",\"$engine_match\",$affection,\"$mood\",\"$relationship\",\"$intimacy\",$response_time,$total_time,\"$special_event\",\"$scene_description\",\"$test_result\"" >> "$CSV_FILE"
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
            -d "{\"character_id\":\"$char_id\",\"title\":\"CSV測試對話-$char_name\"}" \
            "$BASE_URL/chat/session")
        
        if echo "$response" | grep -q '"success":true'; then
            local session_id=$(echo "$response" | jq -r '.data.id')
            SESSION_IDS+=("$session_id")
            print_success "Session created: $session_id"
        else
            print_error "Failed to create session with $char_name"
            echo "$response" | jq .
        fi
        
        sleep 1
    done
}

# Send message and record results to CSV
send_message_with_csv() {
    local session_id="$1"
    local message="$2"
    local char_name="$3"
    local test_name="$4"
    local expected_engine="$5"
    
    print_test "$test_name - $char_name: \"${message:0:50}...\""
    
    local start_time=$(date +%s)
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -d "{\"session_id\":\"$session_id\",\"message\":\"$message\"}" \
        "$BASE_URL/chat/message")
    
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    if echo "$response" | grep -q '"success":true'; then
        local data=$(echo "$response" | jq '.data')
        
        # Extract all information
        local ai_engine=$(echo "$data" | jq -r '.ai_engine // "unknown"')
        local nsfw_level=$(echo "$data" | jq -r '.nsfw_level // 0')
        local character_dialogue=$(echo "$data" | jq -r '.character_dialogue // ""')
        local character_action=$(echo "$data" | jq -r '.character_action // ""')
        local scene_description=$(echo "$data" | jq -r '.scene_description // ""')
        local emotion_state=$(echo "$data" | jq -r '.emotion_state // {}')
        local response_time=$(echo "$data" | jq -r '.response_time // 0')
        local special_event=$(echo "$data" | jq -r '.special_event // null')
        
        # Combine dialogue and action for full character response
        local character_response="$character_dialogue"
        if [ "$character_action" != "null" ] && [ "$character_action" != "" ]; then
            character_response="$character_dialogue [$character_action]"
        fi
        
        # Extract emotion details
        local affection=$(echo "$emotion_state" | jq -r '.affection // 0')
        local mood=$(echo "$emotion_state" | jq -r '.mood // "unknown"')
        local relationship=$(echo "$emotion_state" | jq -r '.relationship // "unknown"')
        local intimacy_level=$(echo "$emotion_state" | jq -r '.intimacy_level // "unknown"')
        
        # Check engine match
        local engine_match="false"
        if [ "$expected_engine" = "$ai_engine" ]; then
            engine_match="true"
        fi
        
        # Format special event
        local special_event_text=""
        if [ "$special_event" != "null" ]; then
            local event_type=$(echo "$special_event" | jq -r '.type // ""')
            local event_desc=$(echo "$special_event" | jq -r '.description // ""')
            special_event_text="$event_type: $event_desc"
        fi
        
        print_success "Response received"
        
        # Show engine match status
        if [ "$engine_match" = "true" ]; then
            echo -e "  ${GREEN}✅ Engine Match:${NC} $ai_engine"
        else
            echo -e "  ${YELLOW}⚠️  Engine Mismatch:${NC} $ai_engine (expected: $expected_engine)"
        fi
        
        echo -e "  ${CYAN}NSFW Level:${NC} $nsfw_level"
        echo -e "  ${CYAN}Response Time:${NC} ${response_time}ms"
        echo -e "  ${CYAN}Total Time:${NC} ${total_time}s"
        echo -e "  ${CYAN}Affection:${NC} $affection"
        echo -e "  ${CYAN}Mood:${NC} $mood"
        echo -e "  ${CYAN}Character Response:${NC} ${character_response:0:100}..."
        
        # Add to CSV
        add_csv_row "$timestamp" "$test_name" "$char_name" "$message" "$character_response" \
                    "$ai_engine" "$nsfw_level" "$expected_engine" "$engine_match" \
                    "$affection" "$mood" "$relationship" "$intimacy_level" \
                    "$response_time" "$total_time" "$special_event_text" "$scene_description" "SUCCESS"
        
        return 0
    else
        print_error "Message failed"
        echo "$response" | jq .
        
        # Add failure to CSV
        add_csv_row "$timestamp" "$test_name" "$char_name" "$message" "ERROR" \
                    "unknown" "0" "$expected_engine" "false" \
                    "0" "unknown" "unknown" "unknown" \
                    "0" "$total_time" "" "" "FAILED"
        
        return 1
    fi
}

# Parse test scenario data
parse_scenario_data() {
    local scenario_data="$1"
    IFS='|' read -r message test_name expected_nsfw expected_engine target_chars <<< "$scenario_data"
    echo "$message" "$test_name" "$expected_nsfw" "$expected_engine" "$target_chars"
}

# Run comprehensive tests with CSV output
run_comprehensive_tests() {
    print_section "Running Comprehensive Tests with CSV Output"
    
    for scenario_data in "${TEST_SCENARIOS_DATA[@]}"; do
        read -r message test_name expected_nsfw expected_engine target_chars <<< "$(parse_scenario_data "$scenario_data")"
        
        # Determine which characters to test
        if [ "$target_chars" = "所有角色" ]; then
            # Test with all characters
            for i in "${!SESSION_IDS[@]}"; do
                local session_id="${SESSION_IDS[$i]}"
                local char_name="${CHARACTER_NAMES[$i]}"
                
                send_message_with_csv "$session_id" "$message" "$char_name" "$test_name" "$expected_engine"
                sleep $TEST_DELAY
            done
        else
            # Test with specific character
            for i in "${!CHARACTER_NAMES[@]}"; do
                if [ "${CHARACTER_NAMES[$i]}" = "$target_chars" ]; then
                    local session_id="${SESSION_IDS[$i]}"
                    local char_name="${CHARACTER_NAMES[$i]}"
                    
                    send_message_with_csv "$session_id" "$message" "$char_name" "$test_name" "$expected_engine"
                    sleep $TEST_DELAY
                    break
                fi
            done
        fi
    done
}

# Generate CSV summary
generate_csv_summary() {
    print_section "Generating CSV Summary"
    
    if [ -f "$CSV_FILE" ]; then
        local total_tests=$(tail -n +2 "$CSV_FILE" | wc -l | tr -d ' ')
        local successful_tests=$(tail -n +2 "$CSV_FILE" | grep -c "SUCCESS")
        local failed_tests=$(tail -n +2 "$CSV_FILE" | grep -c "FAILED")
        local openai_tests=$(tail -n +2 "$CSV_FILE" | grep -c '"openai"')
        local grok_tests=$(tail -n +2 "$CSV_FILE" | grep -c '"grok"')
        local engine_matches=$(tail -n +2 "$CSV_FILE" | grep -c '"true"')
        
        print_success "CSV Summary Generated"
        echo -e "  ${CYAN}CSV File:${NC} $CSV_FILE"
        echo -e "  ${CYAN}Total Tests:${NC} $total_tests"
        echo -e "  ${CYAN}Successful:${NC} $successful_tests"
        echo -e "  ${CYAN}Failed:${NC} $failed_tests"
        echo -e "  ${CYAN}OpenAI Tests:${NC} $openai_tests"
        echo -e "  ${CYAN}Grok Tests:${NC} $grok_tests"
        echo -e "  ${CYAN}Engine Matches:${NC} $engine_matches/$total_tests"
        
        # Show sample data
        echo -e "\n${CYAN}Sample CSV Data (first 3 rows):${NC}"
        head -n 4 "$CSV_FILE" | column -t -s ','
    else
        print_error "CSV file not found: $CSV_FILE"
    fi
}

# Main test execution
main() {
    print_header "Enhanced Chat System Test with CSV Output"
    print_info "Testing both OpenAI and Grok engines with detailed CSV reporting"
    print_info "Base URL: $BASE_URL"
    print_info "Test User: $TEST_USER"
    
    # Check if server is running
    if ! curl -s "$BASE_URL/../health" > /dev/null; then
        print_error "Server is not running at $BASE_URL"
        print_info "Please start the server first: go run main.go"
        exit 1
    fi
    
    # Initialize CSV
    initialize_csv
    
    # Authentication
    authenticate
    
    # Setup
    create_chat_sessions
    
    # Run comprehensive tests
    run_comprehensive_tests
    
    # Generate summary
    generate_csv_summary
    
    print_header "Enhanced Chat Test Complete"
    print_success "All tests completed with CSV output!"
    print_info "Results saved to: $CSV_FILE"
    print_info "Total scenarios tested: ${#TEST_SCENARIOS_DATA[@]}"
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
    
    if ! command -v column &> /dev/null; then
        print_warning "column command not available, CSV preview will be basic"
    fi
}

# Parse command line arguments
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Enhanced Chat System Test Script with CSV Output"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo ""
    echo "Features:"
    echo "  • Tests both OpenAI and Grok engines"
    echo "  • Comprehensive scenario coverage (16 test cases)"
    echo "  • CSV output with detailed metrics"
    echo "  • Engine performance comparison"
    echo "  • NSFW level progression testing"
    echo "  • Character-specific response analysis"
    echo ""
    echo "CSV Output includes:"
    echo "  • Timestamp, test details, character info"
    echo "  • User message and character response"
    echo "  • AI engine used and expected engine"
    echo "  • NSFW level, emotion state, response times"
    echo "  • Special events and scene descriptions"
    echo ""
    exit 0
fi

# Run full test suite with CSV output
check_dependencies
main