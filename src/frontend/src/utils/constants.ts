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
  NOT_AUTHORIZED: 40007,
  TEACHER_DUPLICATE: 40008,
  RELATION_EXISTS: 40009,
  FILE_FORMAT: 40010,
  FILE_SIZE: 40011,
  URL_ERROR: 40012,
  SERVER_ERROR: 50001,
  LLM_ERROR: 50002,
  VECTOR_DB_ERROR: 50003,
  PIPELINE_TIMEOUT: 50004,
  /** 该分享码仅对特定学生可用 */
  SHARE_TARGET_ONLY: 40029,
  /** 当前会话已被教师接管 */
  TAKEOVER_ACTIVE: 40030,
  /** 接管记录不存在或已结束 */
  TAKEOVER_NOT_FOUND: 40031,
  /** 无权操作该会话 */
  TAKEOVER_FORBIDDEN: 40032,
  /** LLM 摘要生成失败 */
  LLM_SUMMARY_FAILED: 40033,
  /** 无效的聊天记录 JSON 格式 */
  INVALID_CHAT_JSON: 40037,
  /** 聊天记录为空 */
  EMPTY_CHAT: 40038,
  /** 无权操作该记忆 */
  MEMORY_FORBIDDEN: 40039,
  /** 无效的教学风格类型 */
  INVALID_TEACHING_STYLE: 40040,
} as const
