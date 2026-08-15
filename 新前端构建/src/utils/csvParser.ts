/**
 * CSV 解析工具模块
 *
 * 从旧前端 fe/src/utils/csvParser.ts 迁移。
 * 纯前端 CSV 解析，支持自动检测分隔符和邮箱列。
 *
 * 创建日期: 20260609
 */

import { isValidEmail } from '@/utils/validators'

/** CSV 解析结果中的单条收件人记录 */
export interface CsvRecipient {
  name: string;
  email: string;
}

/** CSV 解析的完整结果 */
export interface CsvParseResult {
  recipients: CsvRecipient[];
  totalRows: number;
  skippedRows: number;
}

const DELIMITER_CANDIDATES = [',', ';', '\t', '|']
const EMAIL_COLUMN_NAMES = ['email', 'e-mail', 'mail', '邮箱', '邮件', '电子邮件']
const NAME_COLUMN_NAMES = ['name', '姓名', '名字', '昵称', 'nickname', '称呼']

function escapeRegExp(str: string): string {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function detectDelimiter(text: string): string {
  const firstLine = text.split(/\r?\n/)[0] || ''
  let bestDelimiter = ','
  let maxCount = 0
  for (const delimiter of DELIMITER_CANDIDATES) {
    const count = (firstLine.match(new RegExp(escapeRegExp(delimiter), 'g')) || []).length
    if (count > maxCount) {
      maxCount = count
      bestDelimiter = delimiter
    }
  }
  return bestDelimiter
}

function parseCsvLine(line: string, delimiter: string): string[] {
  const fields: string[] = []
  let current = ''
  let inQuotes = false
  for (let i = 0; i < line.length; i++) {
    const char = line[i]
    if (inQuotes) {
      if (char === '"') {
        if (i + 1 < line.length && line[i + 1] === '"') {
          current += '"'; i++
        } else {
          inQuotes = false
        }
      } else {
        current += char
      }
    } else {
      if (char === '"') { inQuotes = true }
      else if (char === delimiter) { fields.push(current.trim()); current = '' }
      else { current += char }
    }
  }
  fields.push(current.trim())
  return fields
}

function findColumnIndex(headers: string[], keywords: string[]): number {
  for (let i = 0; i < headers.length; i++) {
    const header = headers[i].toLowerCase().trim()
    for (const keyword of keywords) {
      if (header === keyword || header.includes(keyword)) return i
    }
  }
  return -1
}

function findEmailFieldIndex(fields: string[]): number {
  for (let i = 0; i < fields.length; i++) {
    if (isValidEmail(fields[i])) return i
  }
  return -1
}

/** 解析 CSV 文件内容，提取收件人列表 */
export function parseCsv(csvText: string): CsvParseResult {
  const result: CsvParseResult = { recipients: [], totalRows: 0, skippedRows: 0 }
  if (!csvText || !csvText.trim()) return result
  const lines = csvText.replace(/\r\n/g, '\n').replace(/\r/g, '\n').split('\n')
  const nonEmptyLines = lines.filter(line => line.trim() !== '')
  if (nonEmptyLines.length === 0) return result
  const delimiter = detectDelimiter(csvText)
  const headers = parseCsvLine(nonEmptyLines[0], delimiter)
  let emailColIndex = findColumnIndex(headers, EMAIL_COLUMN_NAMES)
  const nameColIndex = findColumnIndex(headers, NAME_COLUMN_NAMES)
  let dataStartLine = 0
  let autoDetectEmailCol = false
  if (emailColIndex === -1) {
    autoDetectEmailCol = true
    const firstDataFields = parseCsvLine(nonEmptyLines[0], delimiter)
    emailColIndex = findEmailFieldIndex(firstDataFields)
    if (emailColIndex !== -1) dataStartLine = 0
  } else {
    dataStartLine = 1
  }
  const seenEmails = new Set<string>()
  for (let i = dataStartLine; i < nonEmptyLines.length; i++) {
    const fields = parseCsvLine(nonEmptyLines[i], delimiter)
    result.totalRows++
    if (autoDetectEmailCol && emailColIndex === -1) {
      emailColIndex = findEmailFieldIndex(fields)
      if (emailColIndex === -1) { result.skippedRows++; continue }
    }
    const email = (fields[emailColIndex] || '').trim()
    if (!email || !isValidEmail(email)) { result.skippedRows++; continue }
    const lowerEmail = email.toLowerCase()
    if (seenEmails.has(lowerEmail)) { result.skippedRows++; continue }
    seenEmails.add(lowerEmail)
    let name = ''
    if (nameColIndex !== -1 && fields[nameColIndex]) name = fields[nameColIndex].trim()
    result.recipients.push({ name, email })
  }
  return result
}

/** 从 File 对象读取文本内容 */
export function readCsvFile(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (event) => {
      const text = event.target?.result
      if (typeof text === 'string') resolve(text)
      else reject(new Error('Failed to read file as text'))
    }
    reader.onerror = () => reject(new Error('File reading error'))
    reader.readAsText(file)
  })
}