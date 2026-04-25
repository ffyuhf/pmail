package utf7

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Encode 使用修改版 UTF-7 编码字符串。
func Encode(src string) string {
	var sb strings.Builder
	sb.Grow(len(src))

	for i := 0; i < len(src); {
		ch := src[i]

		if min <= ch && ch <= max {
			sb.WriteByte(ch)
			if ch == '&' {
				sb.WriteByte('-')
			}

			i++
		} else {
			start := i

			// 查找下一个可打印的 ASCII 码点
			i++
			for i < len(src) && (src[i] < min || src[i] > max) {
				i++
			}

			sb.Write(encode([]byte(src[start:i])))
		}
	}

	return sb.String()
}

// 将字符串 s 从 UTF-8 转换为 UTF-16-BE，将结果编码为 base64，
// 移除填充，并添加 UTF-7 移位字符。
func encode(s []byte) []byte {
	// 如果没有控制码点，len(s) 足够用于 UTF-8 到 UTF-16 的转换（见下表）。
	b := make([]byte, 0, len(s)+4)
	for len(s) > 0 {
		r, size := utf8.DecodeRune(s)
		if r > utf8.MaxRune {
			r, size = utf8.RuneError, 1 // 错误修复（issue 3785）
		}
		s = s[size:]
		if r1, r2 := utf16.EncodeRune(r); r1 != utf8.RuneError {
			b = append(b, byte(r1>>8), byte(r1))
			r = r2
		}
		b = append(b, byte(r>>8), byte(r))
	}

	// 编码为 base64
	n := b64Enc.EncodedLen(len(b)) + 2
	b64 := make([]byte, n)
	b64Enc.Encode(b64[1:], b)

	// 移除填充
	n -= 2 - (len(b)+2)%3
	b64 = b64[:n]

	// 添加 UTF-7 移位字符
	b64[0] = '&'
	b64[n-1] = '-'
	return b64
}

// Escape 原样传递 UTF-8 字符，并转义特殊的 UTF-7 标记（& 字符）。
func Escape(src string) string {
	var sb strings.Builder
	sb.Grow(len(src))

	for _, ch := range src {
		sb.WriteRune(ch)
		if ch == '&' {
			sb.WriteByte('-')
		}
	}

	return sb.String()
}
