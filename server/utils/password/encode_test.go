package password

import (
	"strings"
	"testing"
)

// TestEncode 验证bcrypt编码输出格式
func TestEncode(t *testing.T) {
	hash := Encode("testpassword")

	// bcrypt哈希应以$2a$开头且长度为60
	if !strings.HasPrefix(hash, "$2a$") {
		t.Errorf("Encode() 输出应以 $2a$ 开头，实际: %s", hash[:5])
	}
	if len(hash) != 60 {
		t.Errorf("Encode() 输出长度应为60，实际: %d", len(hash))
	}

	// 相同密码每次编码结果应不同（bcrypt自带随机盐）
	hash2 := Encode("testpassword")
	if hash == hash2 {
		t.Error("Encode() 相同密码两次编码结果不应相同（应包含随机盐）")
	}
}

// TestVerify_BcryptHash 验证bcrypt密码的验证流程
func TestVerify_BcryptHash(t *testing.T) {
	password := "mySecurePassword123"
	hash := Encode(password)

	ok, needsUpgrade := Verify(password, hash)
	if !ok {
		t.Error("Verify() bcrypt密码验证应成功")
	}
	if needsUpgrade {
		t.Error("Verify() bcrypt密码不应标记需要升级")
	}

	// 错误密码应验证失败
	ok, needsUpgrade = Verify("wrongPassword", hash)
	if ok {
		t.Error("Verify() 错误密码验证应失败")
	}
	if needsUpgrade {
		t.Error("Verify() 验证失败时不应标记需要升级")
	}
}

// TestVerify_LegacyMd5Hash 验证旧MD5密码的向后兼容
func TestVerify_LegacyMd5Hash(t *testing.T) {
	password := "user2"
	// 使用旧算法生成哈希
	legacyHash := legacyEncode(password)

	ok, needsUpgrade := Verify(password, legacyHash)
	if !ok {
		t.Error("Verify() 旧MD5密码验证应成功")
	}
	if !needsUpgrade {
		t.Error("Verify() 旧MD5密码应标记需要升级")
	}

	// 错误密码应验证失败
	ok, needsUpgrade = Verify("wrongPassword", legacyHash)
	if ok {
		t.Error("Verify() 错误密码验证应失败")
	}
}

// TestIsBcrypt 验证bcrypt格式检测
func TestIsBcrypt(t *testing.T) {
	tests := []struct {
		hash     string
		expected bool
	}{
		{"$2a$10$abcdefghijklmnopqrstuvwx", true},
		{"$2b$10$abcdefghijklmnopqrstuvwx", true},
		{"$2y$10$abcdefghijklmnopqrstuvwx", true},
		{"a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4", false}, // MD5哈希（32字符hex）
		{"", false},
		{"$1$something", false},
	}

	for _, tt := range tests {
		result := IsBcrypt(tt.hash)
		if result != tt.expected {
			t.Errorf("IsBcrypt(%q) = %v, expected %v", tt.hash, result, tt.expected)
		}
	}
}

// TestMd5Encode 验证MD5摘要功能（用于Token和APOP）
func TestMd5Encode(t *testing.T) {
	result := Md5Encode("test")
	if len(result) != 32 {
		t.Errorf("Md5Encode() 输出长度应为32，实际: %d", len(result))
	}

	// 相同输入应产生相同输出
	result2 := Md5Encode("test")
	if result != result2 {
		t.Error("Md5Encode() 相同输入应产生相同输出")
	}
}

// TestLegacyEncode 验证旧算法与原Encode行为一致
func TestLegacyEncode(t *testing.T) {
	// 确保旧算法计算结果不变（用于迁移兼容性验证）
	hash := legacyEncode("user2")
	if len(hash) != 32 {
		t.Errorf("legacyEncode() 输出长度应为32，实际: %d", len(hash))
	}

	// 验证旧算法的确定性：相同输入始终产生相同输出
	hash2 := legacyEncode("user2")
	if hash != hash2 {
		t.Error("legacyEncode() 相同输入应产生相同输出")
	}
}
