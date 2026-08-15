/**
 * 邮箱验证工具函数
 *
 * 提取自 EditerView.vue 中重复的 validateReceivers/validateCc/validateBcc 逻辑。
 * 通过工厂函数 createStringEmailValidator 消除三段几乎一模一样的验证代码。
 *
 * 创建日期: 20260429
 */

/** 简易邮箱格式正则（与原始实现 /.+@.+\..+/ 保持一致） */
const EMAIL_REGEX = /.+@.+\..+/;

/**
 * 检查单个邮箱地址格式是否合法
 * @param email 邮箱地址字符串
 * @returns true 表示格式合法
 */
export function isValidEmail(email: string): boolean {
  return EMAIL_REGEX.test(email);
}

/**
 * 创建 Element Plus 表单验证器：验证指定字段中的邮箱字符串列表
 *
 * 用法示例：
 * ```ts
 * import { createStringEmailValidator } from '@/utils/validators';
 * import lang from '@/i18n/i18n';
 *
 * const rules = {
 *   receivers: [{ validator: createStringEmailValidator(() => ruleForm.receivers, lang.err_email_format), trigger: 'change' }],
 *   cc:        [{ validator: createStringEmailValidator(() => ruleForm.cc, lang.err_email_format),        trigger: 'change' }],
 * }
 * ```
 *
 * @param fieldGetter 返回待验证邮箱字符串数组的函数（延迟求值，保证取到最新值）
 * @param errorMessage 验证失败时的错误提示文本
 * @returns Element Plus 兼容的 validator 函数
 */
export function createStringEmailValidator(
  fieldGetter: () => string[],
  errorMessage: string
): (rule: any, value: any, callback: (error?: Error) => void) => void {
  return (_rule, _value, callback) => {
    const list = fieldGetter();
    for (const item of list) {
      if (!isValidEmail(item)) {
        callback(new Error(errorMessage));
        return;
      }
    }
    callback();
  };
}
