/** API 基础地址 */
export const BASE_URL = 'http://localhost:8080'

/** 颜色常量 */
export const COLORS = {
  /** 主色调 */
  PRIMARY: '#4F46E5',
  PRIMARY_LIGHT: '#818CF8',
  PRIMARY_DARK: '#3730A3',

  /** 功能色 */
  SUCCESS: '#10B981',
  WARNING: '#F59E0B',
  DANGER: '#EF4444',
  INFO: '#3B82F6',

  /** 中性色 */
  TEXT_PRIMARY: '#1F2937',
  TEXT_SECONDARY: '#6B7280',
  TEXT_PLACEHOLDER: '#9CA3AF',
  BG_PRIMARY: '#FFFFFF',
  BG_SECONDARY: '#F3F4F6',
  BORDER: '#E5E7EB',
} as const

/** 角色常量 */
export const ROLES = {
  TEACHER: 'teacher',
  STUDENT: 'student',
  ADMIN: 'admin',
} as const

/** 角色中文名映射 */
export const ROLE_LABELS: Record<string, string> = {
  teacher: '教师',
  student: '学生',
  admin: '管理员',
}

/** 记忆类型常量 */
export const MEMORY_TYPES = {
  CONVERSATION: 'conversation',
  LEARNING_PROGRESS: 'learning_progress',
  PERSONALITY_TRAITS: 'personality_traits',
} as const

/** 记忆类型中文名映射 */
export const MEMORY_TYPE_LABELS: Record<string, string> = {
  conversation: '对话记忆',
  learning_progress: '学习进度',
  personality_traits: '个性特征',
}

/** 分页默认值 */
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_PAGE_SIZE: 20,
  MAX_PAGE_SIZE: 100,
} as const

/** 业务错误码 */
export const ERROR_CODES = {
  SUCCESS: 0,
  UNAUTHORIZED: 40001,
  TOKEN_EXPIRED: 40002,
  FORBIDDEN: 40003,
  BAD_REQUEST: 40004,
  NOT_FOUND: 40005,
  CONFLICT: 40006,
  SERVER_ERROR: 50001,
  LLM_ERROR: 50002,
  VECTOR_DB_ERROR: 50003,
  PIPELINE_TIMEOUT: 50004,
} as const
