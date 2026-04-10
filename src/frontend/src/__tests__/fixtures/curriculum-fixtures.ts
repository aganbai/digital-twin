import type { GradeLevel } from '@/constants/curriculum'
import type { CurriculumConfig, ClassDetailV11, CreateClassV11Response } from '@/api/class'

/**
 * 测试数据集 - 教材配置相关的 Mock 数据
 */

/** 学段枚举值 */
export const CURRICULUM_CONSTANTS = {
  GRADE_LEVELS: [
    { value: 'preschool', label: '学前班', grades: ['幼儿园大班', '学前'] },
    { value: 'primary_lower', label: '小学低年级', grades: ['一年级', '二年级', '三年级'] },
    { value: 'primary_upper', label: '小学高年级', grades: ['四年级', '五年级', '六年级'] },
    { value: 'junior', label: '初中', grades: ['七年级', '八年级', '九年级'] },
    { value: 'senior', label: '高中', grades: ['高一', '高二', '高三'] },
    { value: 'university', label: '大学及以上', grades: ['大一', '大二', '大三', '大四', '研究生', '博士'] },
    { value: 'adult_life', label: '成人生活技能', grades: [] },
    { value: 'adult_professional', label: '成人职业培训', grades: [] },
  ],

  K12_SUBJECTS: ['语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '音乐', '美术', '体育', '信息技术'],

  ADULT_LIFE_CATEGORIES: ['中餐', '西餐', '烘焙', '力量训练', '有氧运动', '瑜伽', '手工', '园艺', '摄影', '绘画'],

  ADULT_PROFESSIONAL_CATEGORIES: ['编程', '设计', '会计', '法律', '医学', '教育', '管理', '营销', '外语', '考证培训'],

  TEXTBOOK_VERSIONS: ['人教版', '北师大版', '苏教版', '沪教版', '部编版', '外研版', '浙教版', '冀教版'],
}

/** 完整的教材配置 - 小学低年级 */
export const mockCurriculumConfigPrimary: CurriculumConfig = {
  id: 1001,
  grade_level: 'primary_lower' as GradeLevel,
  grade: '三年级',
  subjects: ['数学', '语文'],
  textbook_versions: ['人教版', '北师大版'],
  custom_textbooks: ['《小学奥数启蒙》'],
  current_progress: '第三单元 乘法初步',
}

/** 完整的教材配置 - 初中 */
export const mockCurriculumConfigJunior: CurriculumConfig = {
  id: 1002,
  grade_level: 'junior' as GradeLevel,
  grade: '八年级',
  subjects: ['数学', '物理'],
  textbook_versions: ['人教版'],
  custom_textbooks: [],
  current_progress: '第五单元 压强',
}

/** 完整的教材配置 - 大学 */
export const mockCurriculumConfigUniversity: CurriculumConfig = {
  id: 1003,
  grade_level: 'university' as GradeLevel,
  grade: '大二',
  subjects: ['数学', '物理', '计算机'],
  textbook_versions: [],
  custom_textbooks: ['《高等数学》', '《线性代数》', '《大学物理》'],
  current_progress: '第三章 微积分',
}

/** 成人生活技能配置 */
export const mockCurriculumConfigAdultLife: CurriculumConfig = {
  id: 1004,
  grade_level: 'adult_life' as GradeLevel,
  subjects: ['中餐', '烘焙'],
  textbook_versions: [],
  custom_textbooks: [],
  current_progress: '第三课 红烧肉的制作',
}

/** 成人职业培训配置 */
export const mockCurriculumConfigAdultProfessional: CurriculumConfig = {
  id: 1005,
  grade_level: 'adult_professional' as GradeLevel,
  subjects: ['编程', '设计'],
  textbook_versions: [],
  custom_textbooks: ['《JavaScript高级程序设计》'],
  current_progress: '第5章 闭包与作用域',
}

/** 最小配置 - 只有学段 */
export const mockCurriculumConfigMinimal: CurriculumConfig = {
  id: 1006,
  grade_level: 'primary_lower' as GradeLevel,
  subjects: [],
  textbook_versions: [],
  custom_textbooks: [],
}

/** 空配置 */
export const mockCurriculumConfigEmpty: CurriculumConfig = {
  id: 1007,
  grade_level: 'preschool' as GradeLevel,
  subjects: [],
  textbook_versions: [],
  custom_textbooks: [],
}

/** 班级详情 - 有教材配置 */
export const mockClassDetailWithCurriculum: ClassDetailV11 = {
  id: 123,
  name: '三年级数学班',
  description: '小学数学培优班级',
  is_public: true,
  persona_id: 456,
  persona_nickname: '王老师',
  persona_school: '实验小学',
  persona_description: '10年数学教学经验，专注小学奥数',
  student_count: 25,
  created_at: '2026-04-09T10:30:00Z',
  curriculum_config: mockCurriculumConfigPrimary,
}

/** 班级详情 - 无教材配置 */
export const mockClassDetailWithoutCurriculum: ClassDetailV11 = {
  id: 124,
  name: '临时班级',
  description: '',
  is_public: false,
  persona_id: 457,
  persona_nickname: '李老师',
  persona_school: '实验中学',
  persona_description: '临时班级，暂不需要教材配置',
  student_count: 0,
  created_at: '2026-04-09T11:00:00Z',
  curriculum_config: null,
}

/** 班级详情 - curriculum_config 为 undefined */
export const mockClassDetailWithUndefinedCurriculum: ClassDetailV11 = {
  id: 125,
  name: '测试班级',
  description: '测试用班级',
  is_public: true,
  persona_id: 458,
  persona_nickname: '张老师',
  persona_school: '测试学校',
  persona_description: '测试描述',
  student_count: 10,
  created_at: '2026-04-09T12:00:00Z',
}

/** 创建班级成功响应 */
export const mockCreateClassResponse: CreateClassV11Response = {
  id: 123,
  name: '三年级数学班',
  description: '小学数学培优班级',
  is_public: true,
  persona_id: 456,
  persona_nickname: '王老师',
  persona_school: '实验小学',
  persona_description: '10年数学教学经验，专注小学奥数',
  share_code: 'ABC123',
  share_url: 'https://example.com/join/ABC123',
  created_at: '2026-04-09T10:30:00Z',
}

/** 创建班级成功响应 - 带教材配置 */
export const mockCreateClassWithCurriculumResponse: CreateClassV11Response = {
  ...mockCreateClassResponse,
  name: '三年级数学班（带配置）',
}

/** 错误码定义 */
export const CURRICULUM_ERRORS = {
  /** 班级名称已存在 */
  CLASS_NAME_EXISTS: {
    code: 40030,
    message: '该班级名称已存在',
    httpStatus: 409,
  },
  /** 班级不存在 */
  CLASS_NOT_FOUND: {
    code: 40017,
    message: '班级不存在',
    httpStatus: 404,
  },
  /** 无权操作 */
  NO_PERMISSION: {
    code: 40018,
    message: '无权操作此班级',
    httpStatus: 403,
  },
  /** 无效请求参数 */
  INVALID_PARAMS: {
    code: 40004,
    message: '请求参数无效',
    httpStatus: 400,
  },
  /** 无效的学段类型 */
  INVALID_GRADE_LEVEL: {
    code: 40041,
    message: '无效的学段类型',
    httpStatus: 400,
  },
  /** 仅教师角色可创建 */
  NOT_TEACHER: {
    code: 40301,
    message: '仅教师角色可创建班级',
    httpStatus: 403,
  },
  /** 未授权 */
  UNAUTHORIZED: {
    code: 40101,
    message: '未登录或登录已过期',
    httpStatus: 401,
  },
  /** 服务器内部错误 */
  SERVER_ERROR: {
    code: 50001,
    message: '数据库服务不可用',
    httpStatus: 500,
  },
}

/** API 基础响应 */
export const createSuccessResponse = (data: any) => ({
  code: 0,
  message: 'success',
  data,
})

export const createErrorResponse = (errorCode: number, message: string) => ({
  code: errorCode,
  message,
  data: null,
})

/** 学段与年级映射测试用例 */
export const gradeLevelTestCases = [
  { level: 'preschool', expectedGrades: ['幼儿园大班', '学前'], needsGrade: true },
  { level: 'primary_lower', expectedGrades: ['一年级', '二年级', '三年级'], needsGrade: true },
  { level: 'primary_upper', expectedGrades: ['四年级', '五年级', '六年级'], needsGrade: true },
  { level: 'junior', expectedGrades: ['七年级', '八年级', '九年级'], needsGrade: true },
  { level: 'senior', expectedGrades: ['高一', '高二', '高三'], needsGrade: true },
  { level: 'university', expectedGrades: ['大一', '大二', '大三', '大四', '研究生', '博士'], needsGrade: true },
  { level: 'adult_life', expectedGrades: [], needsGrade: false },
  { level: 'adult_professional', expectedGrades: [], needsGrade: false },
] as const

/** 学科选项测试用例 */
export const subjectOptionsTestCases = [
  { level: 'preschool', expectedType: 'K12' },
  { level: 'primary_lower', expectedType: 'K12' },
  { level: 'primary_upper', expectedType: 'K12' },
  { level: 'junior', expectedType: 'K12' },
  { level: 'senior', expectedType: 'K12' },
  { level: 'university', expectedType: 'K12' },
  { level: 'adult_life', expectedType: 'ADULT_LIFE' },
  { level: 'adult_professional', expectedType: 'ADULT_PROFESSIONAL' },
] as const
