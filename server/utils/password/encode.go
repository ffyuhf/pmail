package password

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost bcrypt计算成本因子，越高越安全但越慢
// 20260507: 从10提升至12，增强密码存储安全性
const bcryptCost = 12

// bcryptMinCost bcrypt最低安全轮数阈值，低于此值的哈希在验证成功后触发懒迁移
const bcryptMinCost = 12

// bcryptPrefixes bcrypt哈希的已知前缀标识
var bcryptPrefixes = []string{"$2a$", "$2b$", "$2y$"}

// Encode 使用bcrypt对密码进行哈希，用于新密码创建和密码重置
// 输出格式：$2a$12$...（60字符），20260507起cost从10提升至12
func Encode(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		panic(err) // bcrypt失败属于系统级异常，不可恢复
	}
	return string(hashed)
}

// Verify 验证明文密码是否匹配存储的哈希值，自动识别bcrypt和旧MD5格式
// 返回值：
//   - bool: 密码是否验证成功
//   - bool: 是否需要升级哈希（MD5→bcrypt 或 bcrypt低轮数→高轮数）
func Verify(password, hash string) (bool, bool) {
	// 优先尝试bcrypt验证（新格式）
	if IsBcrypt(hash) {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err != nil {
			return false, false // 密码错误
		}
		// bcrypt验证通过，检查是否需要升级轮数（如cost=10→12）
		needsUpgrade := getBcryptCost(hash) < bcryptMinCost
		return true, needsUpgrade
	}

	// 回退到旧的MD5验证（向后兼容）
	if legacyEncode(password) == hash {
		return true, true // 旧MD5验证通过，需要升级为bcrypt
	}

	return false, false // 验证失败
}

// getBcryptCost 从bcrypt哈希字符串中提取计算成本因子（轮数）
// 输入格式示例：$2a$10$N9qo8uLOickgx2ZMRZoMye...
// 返回值：轮数（如10），非bcrypt格式或解析失败返回-1
func getBcryptCost(hash string) int {
	// bcrypt哈希格式：$2a$10$...，轮数位于第4-5个字符（第3个$之前）
	// 使用标准库的Cost函数解析
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return -1
	}
	return cost
}

// IsBcrypt 判断哈希值是否为bcrypt格式
func IsBcrypt(hash string) bool {
	for _, prefix := range bcryptPrefixes {
		if strings.HasPrefix(hash, prefix) {
			return true
		}
	}
	return false
}

// Md5Encode 对字符串计算MD5摘要，保留用于POP3 APOP认证
// 注意：此函数不应用于密码存储，仅用于对已有哈希值做二次摘要
// 20260425: Token验证已迁移至随机不透明令牌(api_tokens表)，不再使用此函数
func Md5Encode(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// legacyEncode 旧的双MD5加盐算法，仅用于向后兼容验证
// 算法：md5(md5(password+"pmail") + "pmail2023")
// 20260425: 标记为废弃，迁移完成后应在未来版本中移除
func legacyEncode(password string) string {
	return Md5Encode(Md5Encode(password+"pmail") + "pmail2023")
}
