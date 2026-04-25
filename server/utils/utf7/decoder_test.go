package utf7_test

import (
	"strings"
	"testing"

	"github.com/ffyuhf/pmail/utils/utf7"
)

var decode = []struct {
	in  string
	out string
	ok  bool
}{
	// 基本测试（反向测试在编码测试中检查其他有效输入）
	{"", "", true},
	{"abc", "abc", true},
	{"&-abc", "&abc", true},
	{"abc&-", "abc&", true},
	{"a&-b&-c", "a&b&c", true},
	{"&ABk-", "\x19", true},
	{"&AB8-", "\x1F", true},
	{"ABk-", "ABk-", true},
	{"&-,&-&AP8-&-", "&,&\u00FF&", true},
	{"&-&-,&AP8-&-", "&&,\u00FF&", true},
	{"abc &- &AP8A,wD,- &- xyz", "abc & \u00FF\u00FF\u00FF & xyz", true},

	// ASCII 中的非法码点
	{"\x00", "", false},
	{"\x1F", "", false},
	{"abc\n", "", false},
	{"abc\x7Fxyz", "", false},

	// 无效的 UTF-8
	{"\xc3\x28", "", false},
	{"\xe2\x82\x28", "", false},

	// 无效的 Base64 字符
	{"&/+8-", "", false},
	{"&*-", "", false},
	{"&ZeVnLIqe -", "", false},

	// Base64 中的 CR 和 LF
	{"&ZeVnLIqe\r\n-", "", false},
	{"&ZeVnLIqe\r\n\r\n-", "", false},
	{"&ZeVn\r\n\r\nLIqe-", "", false},

	// 填充未去除
	{"&AAAAHw=-", "", false},
	{"&AAAAHw==-", "", false},
	{"&AAAAHwB,AIA=-", "", false},
	{"&AAAAHwB,AIA==-", "", false},

	// 少一个字节
	{"&2A-", "", false},
	{"&2ADc-", "", false},
	{"&AAAAHwB,A-", "", false},
	{"&AAAAHwB,A=-", "", false},
	{"&AAAAHwB,A==-", "", false},
	{"&AAAAHwB,A===-", "", false},
	{"&AAAAHwB,AI-", "", false},
	{"&AAAAHwB,AI=-", "", false},
	{"&AAAAHwB,AI==-", "", false},

	// 隐式移位
	{"&", "", false},
	{"&Jjo", "", false},
	{"Jjo&", "", false},
	{"&Jjo&", "", false},
	{"&Jjo!", "", false},
	{"&Jjo+", "", false},
	{"abc&Jjo", "", false},

	// 空移位
	{"&AGE-&Jjo-", "", false},
	{"&U,BTFw-&ZeVnLIqe-", "", false},

	// 长输入，末尾有 Base64
	{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa &2D3eCg- &2D3eCw- &2D3eDg-",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \U0001f60a \U0001f60b \U0001f60e", true},

	// 短 ASCII 之间的长 Base64 输入
	{"00000000000000000000 &MEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEIwQjBCMEI- 00000000000000000000",
		"00000000000000000000 " + strings.Repeat("\U00003042", 37) + " 00000000000000000000", true},

	// Base64 中的 ASCII
	{"&AGE-", "", false},            // "a"
	{"&ACY-", "", false},            // "&"
	{"&AGgAZQBsAGwAbw-", "", false}, // "hello"
	{"&JjoAIQ-", "", false},         // "\u263a!"

	// 无效的代理对
	{"&2AA-", "", false},    // U+D800
	{"&2AD-", "", false},    // U+D800
	{"&3AA-", "", false},    // U+DC00
	{"&2AAAQQ-", "", false}, // U+D800 'A'
	{"&2AD,,w-", "", false}, // U+D800 U+FFFF
	{"&3ADYAA-", "", false}, // U+DC00 U+D800

	// 中文
	{"&V4NXPpCuTvY-", "垃圾邮件", true},
	{"&UXZO1mWHTvZZOQ-", "其他文件夹", true},
}

func TestDecoder(t *testing.T) {
	for _, test := range decode {
		out, err := utf7.Decode(test.in)
		if out != test.out {
			t.Errorf("UTF7Decode(%+q) expected %+q; got %+q", test.in, test.out, out)
		}
		if test.ok {
			if err != nil {
				t.Errorf("UTF7Decode(%+q) unexpected error; %v", test.in, err)
			}
		} else if err == nil {
			t.Errorf("UTF7Decode(%+q) expected error", test.in)
		}
	}
}
