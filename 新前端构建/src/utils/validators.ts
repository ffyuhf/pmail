/**
 * 邮箱验证工具函数
 *
 * 从旧前端 fe/src/utils/validators.ts 迁移。
 *
 * 创建日期: 20260609
 */

/** 简易邮箱格式正则 */
const EMAIL_REGEX = /.+@.+\..+/

/** 检查单个邮箱地址格式是否合法 */
export function isValidEmail(email: string): boolean {
  return EMAIL_REGEX.test(email)
}

/**
 * 创建表单验证器：验证邮箱字符串数组
 * @param fieldGetter 返回待验证邮箱字符串数组的函数
 * @param errorMessage 验证失败时的错误提示文本
 */
export function createStringEmailValidator(
  fieldGetter: () => string[],
  errorMessage: string
): (_rule: any, _value: any, callback: (error?: Error) => void) => void {
  return (_rule, _value, callback) => {
    const list = fieldGetter()
    for (const item of list) {
      if (!isValidEmail(item)) {
        callback(new Error(errorMessage))
        return
      }
    }
    callback()
  }
}