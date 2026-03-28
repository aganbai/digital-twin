/**
 * 格式化时间
 * @param date - 日期字符串或 Date 对象
 * @returns 格式化后的时间字符串，如 "2026-03-28 15:30"
 */
export function formatTime(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) return ''

  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hour = String(d.getHours()).padStart(2, '0')
  const minute = String(d.getMinutes()).padStart(2, '0')

  // 判断是否是今天
  const now = new Date()
  const isToday =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()

  // 判断是否是昨天
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  const isYesterday =
    d.getFullYear() === yesterday.getFullYear() &&
    d.getMonth() === yesterday.getMonth() &&
    d.getDate() === yesterday.getDate()

  if (isToday) {
    return `${hour}:${minute}`
  } else if (isYesterday) {
    return `昨天 ${hour}:${minute}`
  } else if (d.getFullYear() === now.getFullYear()) {
    return `${month}-${day} ${hour}:${minute}`
  }
  return `${year}-${month}-${day} ${hour}:${minute}`
}

/**
 * 文本截断
 * @param text - 原始文本
 * @param maxLen - 最大长度，默认 50
 * @returns 截断后的文本，超出部分用 "..." 替代
 */
export function truncateText(text: string, maxLen: number = 50): string {
  if (!text) return ''
  if (text.length <= maxLen) return text
  return text.slice(0, maxLen) + '...'
}
