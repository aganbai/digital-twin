/**
 * 教材配置相关常量
 * V2.0 IT13 - 教材配置流程重构
 */

/** 学段枚举值 */
export type GradeLevel =
  | 'preschool'
  | 'primary_lower'
  | 'primary_upper'
  | 'junior'
  | 'senior'
  | 'university'
  | 'adult_life'
  | 'adult_professional'

/** 学段选项（包含年级列表，兼容api/curriculum.ts中的GRADE_LEVELS） */
export const GRADE_LEVEL_OPTIONS: { value: GradeLevel; label: string; grades: string[] }[] = [
  { value: 'preschool', label: '学前班', grades: ['幼儿园大班', '学前'] },
  { value: 'primary_lower', label: '小学低年级', grades: ['一年级', '二年级', '三年级'] },
  { value: 'primary_upper', label: '小学高年级', grades: ['四年级', '五年级', '六年级'] },
  { value: 'junior', label: '初中', grades: ['七年级', '八年级', '九年级'] },
  { value: 'senior', label: '高中', grades: ['高一', '高二', '高三'] },
  { value: 'university', label: '大学及以上', grades: ['大一', '大二', '大三', '大四', '研究生', '博士'] },
  { value: 'adult_life', label: '成人生活技能', grades: [] },
  { value: 'adult_professional', label: '成人职业培训', grades: [] },
]

/** 学段与年级对应关系 */
export const GRADE_MAP: Record<GradeLevel, string[]> = {
  preschool: ['幼儿园大班', '学前'],
  primary_lower: ['一年级', '二年级', '三年级'],
  primary_upper: ['四年级', '五年级', '六年级'],
  junior: ['七年级', '八年级', '九年级'],
  senior: ['高一', '高二', '高三'],
  university: ['大一', '大二', '大三', '大四', '研究生', '博士'],
  adult_life: [], // 无固定年级
  adult_professional: [], // 无固定年级
}

/** 获取指定学段的年级选项 */
export function getGradesByLevel(level: GradeLevel): string[] {
  return GRADE_MAP[level] || []
}

/** 判断学段是否需要年级选择 */
export function needsGradeSelection(level: GradeLevel): boolean {
  return GRADE_MAP[level]?.length > 0
}

/** 判断学段是否为大学及以上（需要自定义教材输入） */
export function needsCustomTextbook(level: GradeLevel): boolean {
  return level === 'university' || level === 'adult_life' || level === 'adult_professional'
}

/** K12 学科列表 */
export const K12_SUBJECTS = [
  '语文',
  '数学',
  '英语',
  '物理',
  '化学',
  '生物',
  '历史',
  '地理',
  '政治',
  '音乐',
  '美术',
  '体育',
  '信息技术',
]

/** 成人生活技能类别 */
export const ADULT_LIFE_CATEGORIES = [
  '中餐',
  '西餐',
  '烘焙',
  '力量训练',
  '有氧运动',
  '瑜伽',
  '手工',
  '园艺',
  '摄影',
  '绘画',
]

/** 成人职业培训类别 */
export const ADULT_PROFESSIONAL_CATEGORIES = [
  '编程',
  '设计',
  '会计',
  '法律',
  '医学',
  '教育',
  '管理',
  '营销',
  '外语',
  '考证培训',
]

/** 根据学段获取学科/类别选项 */
export function getSubjectsByLevel(level: GradeLevel): string[] {
  switch (level) {
    case 'preschool':
    case 'primary_lower':
    case 'primary_upper':
    case 'junior':
    case 'senior':
      return K12_SUBJECTS
    case 'university':
      return K12_SUBJECTS // 大学也使用K12学科为基础，可扩展
    case 'adult_life':
      return ADULT_LIFE_CATEGORIES
    case 'adult_professional':
      return ADULT_PROFESSIONAL_CATEGORIES
    default:
      return []
  }
}

/** 教材版本选项 */
export const TEXTBOOK_VERSIONS = [
  '人教版',
  '北师大版',
  '苏教版',
  '沪教版',
  '部编版',
  '外研版',
  '浙教版',
  '冀教版',
]

/** 教材配置表单数据类型 */
export interface CurriculumConfigForm {
  grade_level?: GradeLevel
  grade?: string
  subjects: string[]
  textbook_versions: string[]
  custom_textbooks: string[]
  current_progress?: string
}

/** 空配置默认值 */
export const EMPTY_CURRICULUM_CONFIG: CurriculumConfigForm = {
  grade_level: undefined,
  grade: undefined,
  subjects: [],
  textbook_versions: [],
  custom_textbooks: [],
  current_progress: '',
}
