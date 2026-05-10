/**
 * CSV 解析工具模块
 *
 * 纯前端实现 CSV 文件解析，支持自动检测分隔符、自动识别邮箱列，
 * 并返回结构化的收件人列表供写信页面使用。
 *
 * 创建日期: 20260510
 * @module csvParser
 */

import { isValidEmail } from '@/utils/validators';

/** CSV 解析结果中的单条收件人记录 */
export interface CsvRecipient {
  /** 收件人姓名（CSV 中 name 列的值，无则为空字符串） */
  name: string;
  /** 收件人邮箱地址（已通过格式验证） */
  email: string;
}

/** CSV 解析的完整结果 */
export interface CsvParseResult {
  /** 解析出的有效收件人列表 */
  recipients: CsvRecipient[];
  /** CSV 中的总行数（不含表头） */
  totalRows: number;
  /** 跳过的无效行数 */
  skippedRows: number;
}

/** 常见分隔符候选列表（按优先级排序） */
const DELIMITER_CANDIDATES = [',', ';', '\t', '|'];

/** 邮箱列名的常见写法（不区分大小写匹配） */
const EMAIL_COLUMN_NAMES = ['email', 'e-mail', 'mail', '邮箱', '邮件', '电子邮件'];

/** 姓名列名的常见写法（不区分大小写匹配） */
const NAME_COLUMN_NAMES = ['name', '姓名', '名字', '昵称', 'nickname', '称呼'];

/**
 * 自动检测 CSV 文本使用的分隔符
 *
 * 策略：统计第一行中各候选分隔符出现的次数，选择出现次数最多的。
 * 如果所有候选分隔符出现次数均为 0，则默认使用逗号。
 *
 * @param text CSV 文本内容
 * @returns 检测到的分隔符字符
 */
function detectDelimiter(text: string): string {
  const firstLine = text.split(/\r?\n/)[0] || '';
  let bestDelimiter = ',';
  let maxCount = 0;

  for (const delimiter of DELIMITER_CANDIDATES) {
    const count = (firstLine.match(new RegExp(escapeRegExp(delimiter), 'g')) || []).length;
    if (count > maxCount) {
      maxCount = count;
      bestDelimiter = delimiter;
    }
  }

  return bestDelimiter;
}

/**
 * 转义正则表达式中的特殊字符
 * @param str 原始字符串
 * @returns 转义后的安全正则字符串
 */
function escapeRegExp(str: string): string {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

/**
 * 解析 CSV 单行文本为字段数组
 *
 * 支持双引号包裹的字段（含换行和分隔符）。
 *
 * @param line CSV 单行文本
 * @param delimiter 分隔符
 * @returns 字段数组
 */
function parseCsvLine(line: string, delimiter: string): string[] {
  const fields: string[] = [];
  let current = '';
  let inQuotes = false;

  for (let i = 0; i < line.length; i++) {
    const char = line[i];

    if (inQuotes) {
      if (char === '"') {
        // 双引号转义：连续两个双引号表示一个双引号字符
        if (i + 1 < line.length && line[i + 1] === '"') {
          current += '"';
          i++;
        } else {
          inQuotes = false;
        }
      } else {
        current += char;
      }
    } else {
      if (char === '"') {
        inQuotes = true;
      } else if (char === delimiter) {
        fields.push(current.trim());
        current = '';
      } else {
        current += char;
      }
    }
  }

  // 添加最后一个字段
  fields.push(current.trim());
  return fields;
}

/**
 * 在表头中查找匹配指定关键词列表的列索引
 *
 * @param headers 表头字段数组
 * @param keywords 关键词列表（不区分大小写匹配）
 * @returns 匹配到的列索引，未找到返回 -1
 */
function findColumnIndex(headers: string[], keywords: string[]): number {
  for (let i = 0; i < headers.length; i++) {
    const header = headers[i].toLowerCase().trim();
    for (const keyword of keywords) {
      if (header === keyword || header.includes(keyword)) {
        return i;
      }
    }
  }
  return -1;
}

/**
 * 从字段数组中查找第一个包含有效邮箱地址的字段索引
 *
 * @param fields 字段数组
 * @returns 邮箱所在字段索引，未找到返回 -1
 */
function findEmailFieldIndex(fields: string[]): number {
  for (let i = 0; i < fields.length; i++) {
    if (isValidEmail(fields[i])) {
      return i;
    }
  }
  return -1;
}

/**
 * 解析 CSV 文件内容，提取收件人列表
 *
 * 处理流程：
 * 1. 自动检测分隔符
 * 2. 解析表头，识别邮箱列和姓名列
 * 3. 如果表头无法识别，则在数据行中自动查找邮箱格式字段
 * 4. 对每个邮箱地址进行格式验证，跳过无效地址
 *
 * @param csvText CSV 文件文本内容
 * @returns 解析结果（recipients + totalRows + skippedRows）
 */
export function parseCsv(csvText: string): CsvParseResult {
  const result: CsvParseResult = {
    recipients: [],
    totalRows: 0,
    skippedRows: 0,
  };

  if (!csvText || !csvText.trim()) {
    return result;
  }

  // 统一换行符并分割行
  const lines = csvText.replace(/\r\n/g, '\n').replace(/\r/g, '\n').split('\n');
  const nonEmptyLines = lines.filter(line => line.trim() !== '');

  if (nonEmptyLines.length === 0) {
    return result;
  }

  // 自动检测分隔符
  const delimiter = detectDelimiter(csvText);

  // 解析表头
  const headers = parseCsvLine(nonEmptyLines[0], delimiter);

  // 识别邮箱列索引
  let emailColIndex = findColumnIndex(headers, EMAIL_COLUMN_NAMES);

  // 识别姓名列索引
  let nameColIndex = findColumnIndex(headers, NAME_COLUMN_NAMES);

  // 确定数据起始行（如果表头中有已识别的列名，则跳过表头；否则所有行都作为数据）
  let dataStartLine = 0;
  let autoDetectEmailCol = false;

  if (emailColIndex === -1) {
    // 表头中未找到邮箱列名，尝试在第一行数据中自动检测
    autoDetectEmailCol = true;
    const firstDataFields = parseCsvLine(nonEmptyLines[0], delimiter);
    emailColIndex = findEmailFieldIndex(firstDataFields);
    if (emailColIndex !== -1) {
      // 第一行中找到了邮箱，说明没有表头，从第一行开始就是数据
      dataStartLine = 0;
    }
  } else {
    // 表头中找到了邮箱列名，从第二行开始是数据
    dataStartLine = 1;
  }

  // 解析数据行
  const seenEmails = new Set<string>();

  for (let i = dataStartLine; i < nonEmptyLines.length; i++) {
    const fields = parseCsvLine(nonEmptyLines[i], delimiter);
    result.totalRows++;

    // 如果需要自动检测邮箱列，且还未确定列索引
    if (autoDetectEmailCol && emailColIndex === -1) {
      emailColIndex = findEmailFieldIndex(fields);
      if (emailColIndex === -1) {
        result.skippedRows++;
        continue;
      }
    }

    const email = (fields[emailColIndex] || '').trim();

    if (!email || !isValidEmail(email)) {
      result.skippedRows++;
      continue;
    }

    // 去重
    const lowerEmail = email.toLowerCase();
    if (seenEmails.has(lowerEmail)) {
      result.skippedRows++;
      continue;
    }
    seenEmails.add(lowerEmail);

    // 提取姓名（如果有的话）
    let name = '';
    if (nameColIndex !== -1 && fields[nameColIndex]) {
      name = fields[nameColIndex].trim();
    }

    result.recipients.push({ name, email });
  }

  return result;
}

/**
 * 从 File 对象读取文本内容
 *
 * @param file 用户选择的 CSV 文件
 * @returns Promise<string> 文件文本内容
 */
export function readCsvFile(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onload = (event) => {
      const text = event.target?.result;
      if (typeof text === 'string') {
        resolve(text);
      } else {
        reject(new Error('Failed to read file as text'));
      }
    };

    reader.onerror = () => {
      reject(new Error('File reading error'));
    };

    reader.readAsText(file);
  });
}
