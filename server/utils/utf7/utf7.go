// Package utf7 实现了 RFC 3501 第 5.1.3 节定义的修改版 UTF-7 编码
package utf7

import (
	"encoding/base64"
)

const (
	min = 0x20 // UTF-7 最小自表示值
	max = 0x7E // UTF-7 最大自表示值
)

var b64Enc = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+,")
