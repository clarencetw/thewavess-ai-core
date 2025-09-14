package utils

import (
    "errors"
    "strings"
    "unicode/utf8"
)

// ExtractJSONFromText 掃描任意文本並返回第一個有效的頂級 JSON 物件或陣列子字串。
// 對結構嚴格（匹配括號），同時容忍周圍的散文、代碼圍欄和前後綴。
// 不會嘗試修復無效的 JSON。
//
// 快速精確的方法：
// - 修剪空白並移除 UTF-8 BOM（如果存在）
// - 線性掃描 '{' 或 '['
// - 使用小型堆疊來追蹤預期的閉合括號，同時追蹤字串狀態
// - 忽略出現在引號字串內的括號
//
// 返回包含 JSON 的子字串，如果找不到則返回錯誤。
func ExtractJSONFromText(text string) (string, error) {
    if text == "" {
        return "", errors.New("empty text")
    }

    // 標準化前後空白
    s := strings.TrimSpace(text)

    // 移除 UTF-8 BOM（如果存在）
    if len(s) >= 3 && s[0] == 0xEF && s[1] == 0xBB && s[2] == 0xBF {
        s = s[3:]
    }

    // 快速路徑：如果 s 看起來已經是單一 JSON 值
    if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
        if sub, ok := extractBalancedJSON(s, 0); ok {
            return sub, nil
        }
    }

    // 一般路徑：找到第一個 '{' 或 '[' 並嘗試平衡提取
    for i := 0; i < len(s); i++ {
        c := s[i]
        if c == '{' || c == '[' {
            if sub, ok := extractBalancedJSON(s, i); ok {
                return sub, nil
            }
        }
        // 跳過無效位元組以避免在格式錯誤的 UTF-8 上無限迴圈
        if c < 0x20 && c != '\n' && c != '\t' && c != '\r' {
            continue
        }
        if c >= 0x80 {
            // 按 rune 寬度前進
            _, w := utf8.DecodeRuneInString(s[i:])
            if w > 1 {
                i += w - 1
            }
        }
    }

    return "", errors.New("no JSON object or array found")
}

// extractBalancedJSON 嘗試提取平衡的 JSON 值（物件或陣列），
// 從索引 start 開始，其中 s[start] 是 '{' 或 '['。
// 如果找到完全平衡的結構，則返回子字串。
func extractBalancedJSON(s string, start int) (string, bool) {
    // 預期閉合括號的小型堆疊。典型 LLM JSON 的深度較淺。
    stack := make([]byte, 0, 8)

    switch s[start] {
    case '{':
        stack = append(stack, '}')
    case '[':
        stack = append(stack, ']')
    default:
        return "", false
    }

    inString := false
    escaped := false

    for i := start + 1; i < len(s); i++ {
        ch := s[i]

        if inString {
            if escaped {
                escaped = false
                continue
            }
            switch ch {
            case '\\':
                escaped = true
            case '"':
                inString = false
            }
            continue
        }

        switch ch {
        case '"':
            inString = true
        case '{':
            stack = append(stack, '}')
        case '[':
            stack = append(stack, ']')
        case '}', ']':
            if len(stack) == 0 {
                return "", false
            }
            // 檢查匹配
            expect := stack[len(stack)-1]
            if ch != expect {
                return "", false
            }
            stack = stack[:len(stack)-1]
            if len(stack) == 0 {
                // 找到平衡結尾
                return s[start : i+1], true
            }
        default:
            // 無操作
        }

        // 乾淨地跳過多位元組 rune
        if ch >= 0x80 {
            _, w := utf8.DecodeRuneInString(s[i:])
            if w > 1 {
                // 調整迴圈到 rune 後的下一個位元組
                i += w - 1
            }
        }
    }

    return "", false
}

// ParseJSONFromText 從文本中提取第一個 JSON 物件/陣列並將其解組到 v 中。
// v 應該是指向結構體/映射/切片的指標。
// 對 JSON 有效性保持嚴格，不會嘗試修復輸入。
func ParseJSONFromText(text string, unmarshal func([]byte, any) error, v any) error {
    sub, err := ExtractJSONFromText(text)
    if err != nil {
        return err
    }
    return unmarshal([]byte(sub), v)
}

