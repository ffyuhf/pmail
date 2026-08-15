package utf7

import (
	"errors"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// ErrInvalidUTF7 表示解码器遇到了无效的 UTF-7。
var ErrInvalidUTF7 = errors.New("utf7: invalid UTF-7")

// Decode 解码使用修改版 UTF-7 编码的字符串。
//
// 注意：接受原始 UTF-8 输入。
func Decode(src string) (string, error) {
	if !utf8.ValidString(src) {
		return "", errors.New("invalid UTF-8")
	}

	var sb strings.Builder
	sb.Grow(len(src))

	ascii := true
	for i := 0; i < len(src); i++ {
		ch := src[i]

		if ch < min || (ch > max && ch < utf8.RuneSelf) {
			// ASCII 模式下的非法码点。注意：UTF-8 码点始终允许。
			return "", ErrInvalidUTF7
		}

		if ch != '&' {
			sb.WriteByte(ch)
			ascii = true
			continue
		}

		// 查找 Base64 或 "&-" 段的结尾
		start := i + 1
		for i++; i < len(src) && src[i] != '-'; i++ {
			if src[i] == '\r' || src[i] == '\n' { // base64 包忽略 CR 和 LF
				return "", ErrInvalidUTF7
			}
		}

		if i == len(src) { // 隐式移位（"&..."）
			return "", ErrInvalidUTF7
		}

		if i == start { // 转义序列 "&-"
			sb.WriteByte('&')
			ascii = true
		} else { // base64 中的控制码点或非 ASCII 码点
			if !ascii { // 空移位（"&...-&..."）
				return "", ErrInvalidUTF7
			}

			b := decode([]byte(src[start:i]))
			if len(b) == 0 { // 编码错误
				return "", ErrInvalidUTF7
			}
			sb.Write(b)

			ascii = false
		}
	}

	return sb.String(), nil
}

// 从 base64 数据中提取 UTF-16-BE 字节并转换为 UTF-8。
// 如果编码无效则返回 nil。
func decode(b64 []byte) []byte {
	var b []byte

	// 分配一块足够大的内存，用于存储 Base64 数据（如需填充）、UTF-16-BE 字节和解码后的 UTF-8 字节。
	// 由于 2 字节的 UTF-16 序列可能扩展为 3 字节的 UTF-8 序列，
	// UTF-8 的空间分配翻倍。
	if n := len(b64); b64[n-1] == '=' {
		return nil
	} else if n&3 == 0 {
		b = make([]byte, b64Enc.DecodedLen(n)*3)
	} else {
		n += 4 - n&3
		b = make([]byte, n+b64Enc.DecodedLen(n)*3)
		copy(b[copy(b, b64):n], []byte("=="))
		b64, b = b[:n], b[n:]
	}

	// 将 Base64 解码到 b 的前 1/3 部分
	n, err := b64Enc.Decode(b, b64)
	if err != nil || n&1 == 1 {
		return nil
	}

	// 将 UTF-16-BE 解码到 b 的剩余 2/3 部分
	b, s := b[:n], b[n:]
	j := 0
	for i := 0; i < n; i += 2 {
		r := rune(b[i])<<8 | rune(b[i+1])
		if utf16.IsSurrogate(r) {
			if i += 2; i == n {
				return nil
			}
			r2 := rune(b[i])<<8 | rune(b[i+1])
			if r = utf16.DecodeRune(r, r2); r == utf8.RuneError {
				return nil
			}
		} else if min <= r && r <= max {
			return nil
		}
		j += utf8.EncodeRune(s[j:], r)
	}
	return s[:j]
}
