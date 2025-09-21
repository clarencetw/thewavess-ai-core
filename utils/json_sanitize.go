package utils

// SanitizeLooseJSONForNewlines 針對模型產生的「未轉義換行」進行修正：
// 僅在字串範圍內將裸換行（\n、\r、\r\n）轉為 \n，避免 JSON 解析錯誤。
// 不修改字串外的換行，亦不嘗試修復其他結構性錯誤。
func SanitizeLooseJSONForNewlines(s string) string {
	if s == "" {
		return s
	}
	b := make([]rune, 0, len(s))
	inString := false
	escaped := false
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		ch := rs[i]
		if inString {
			if escaped {
				// 上一個是反斜線，當前字元原樣加入
				b = append(b, ch)
				escaped = false
				continue
			}
			switch ch {
			case '\\':
				escaped = true
				b = append(b, ch)
				continue
			case '"':
				inString = false
				b = append(b, ch)
				continue
			case '\n', '\r':
				// 將字串內的裸換行轉為兩字元序列 \n
				// 若為 \r\n，消費下一個 \n，統一寫入 \n
				if ch == '\r' && i+1 < len(rs) && rs[i+1] == '\n' {
					i++
				}
				b = append(b, '\\', 'n')
				continue
			default:
				b = append(b, ch)
				continue
			}
		} else {
			switch ch {
			case '"':
				inString = true
				b = append(b, ch)
				continue
			default:
				b = append(b, ch)
				continue
			}
		}
	}
	return string(b)
}
