package password

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestEncode 验证bcrypt编码输出格式（cost=12）
func TestEncode(t *testing.T) {
	hash := Encode("testpassword")

	// bcrypt哈希应以$2a$12$开头且长度为60
	if !strings.HasPrefix(hash, "$2a$12$") {
		t.Errorf("Encode() 输出应以 $2a$12$ 开头，实际: %s", hash[:7])
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

// TestGetBcryptCost 验证从bcrypt哈希中提取轮数
func TestGetBcryptCost(t *testing.T) {
	// 使用不同cost生成哈希并验证提取结果
	hash10, err := bcrypt.GenerateFromPassword([]byte("test"), 10)
	if err != nil {
		t.Fatalf("生成cost=10哈希失败: %v", err)
	}
	if cost := getBcryptCost(string(hash10)); cost != 10 {
		t.Errorf("getBcryptCost(cost=10哈希) = %d, expected 10", cost)
	}

	hash12 := Encode("test")
	if cost := getBcryptCost(hash12); cost != 12 {
		t.Errorf("getBcryptCost(cost=12哈希) = %d, expected 12", cost)
	}

	// 非bcrypt字符串应返回-1
	if cost := getBcryptCost("notabcrypt"); cost != -1 {
		t.Errorf("getBcryptCost(非bcrypt) = %d, expected -1", cost)
	}
	if cost := getBcryptCost(""); cost != -1 {
		t.Errorf("getBcryptCost(空字符串) = %d, expected -1", cost)
	}
}

// TestVerify_BcryptCostUpgrade 验证bcrypt低轮数哈希触发升级
// 20260507: cost=10的bcrypt哈希验证成功后应返回needsUpgrade=true
func TestVerify_BcryptCostUpgrade(t *testing.T) {
	password := "mySecurePassword123"

	// 模拟旧版bcrypt(cost=10)哈希
	oldHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		t.Fatalf("生成cost=10哈希失败: %v", err)
	}

	// 验证成功，且标记需要升级（cost=10 < 12）
	ok, needsUpgrade := Verify(password, string(oldHash))
	if !ok {
		t.Error("Verify() cost=10 bcrypt密码验证应成功")
	}
	if !needsUpgrade {
		t.Error("Verify() cost=10 bcrypt密码应标记需要升级为cost=12")
	}

	// 新版bcrypt(cost=12)哈希不应触发升级
	newHash := Encode(password)
	ok, needsUpgrade = Verify(password, newHash)
	if !ok {
		t.Error("Verify() cost=12 bcrypt密码验证应成功")
	}
	if needsUpgrade {
		t.Error("Verify() cost=12 bcrypt密码不应标记需要升级")
	}

	// 错误密码不应触发升级
	ok, needsUpgrade = Verify("wrongPassword", string(oldHash))
	if ok {
		t.Error("Verify() 错误密码验证应失败")
	}
	if needsUpgrade {
		t.Error("Verify() 验证失败时不应标记需要升级")
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
