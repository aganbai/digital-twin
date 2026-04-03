import { request } from './request'

/** 教材配置 */
export interface CurriculumConfig {
  id: number
  persona_id: number
  grade_level: string
  grade: string
  textbook_versions: string[]
  custom_textbooks: string[]
  subjects: string[]
  current_progress: string
  region: string
  is_active: boolean
  created_at: string
  updated_at: string
}

/** 创建教材配置请求 */
export interface CreateCurriculumParams {
  persona_id: number
  grade_level?: string
  grade: string
  textbook_versions: string[]
  custom_textbooks?: string[]
  subjects: string[]
  current_progress?: string
  region?: string
}

/** 学段选项 */
export const GRADE_LEVELS = [
  { value: 'preschool', label: '学前班', grades: ['幼儿园大班', '学前'] },
  { value: 'primary_lower', label: '小学低年级', grades: ['一年级', '二年级', '三年级'] },
  { value: 'primary_upper', label: '小学高年级', grades: ['四年级', '五年级', '六年级'] },
  { value: 'junior', label: '初中', grades: ['七年级', '八年级', '九年级'] },
  { value: 'senior', label: '高中', grades: ['高一', '高二', '高三'] },
  { value: 'university', label: '大学及以上', grades: ['大一', '大二', '大三', '大四', '研究生', '博士'] },
  { value: 'adult_life', label: '成人生活技能', grades: [] },
  { value: 'adult_professional', label: '成人职业培训', grades: [] },
]

/** K12学科选项 */
export const K12_SUBJECTS = ['语文', '数学', '英语', '物理', '化学', '生物', '历史', '地理', '政治', '音乐', '美术', '体育', '信息技术']

/** 成人生活技能课程类别 */
export const ADULT_LIFE_CATEGORIES = ['中餐', '西餐', '烘焙', '力量训练', '有氧运动', '瑜伽', '手工', '园艺', '摄影', '绘画']

/** 成人职业培训课程类别 */
export const ADULT_PROFESSIONAL_CATEGORIES = ['编程', '设计', '会计', '法律', '医学', '教育', '管理', '营销', '外语', '考证培训']

/** 教材版本选项 */
export const TEXTBOOK_VERSIONS = ['人教版', '北师大版', '苏教版', '沪教版', '部编版', '外研版', '浙教版', '冀教版']

/** 创建教材配置 */
export function createCurriculumConfig(params: CreateCurriculumParams) {
  return request<{ id: number; grade_level: string }>({
    url: '/api/curriculum-configs',
    method: 'POST',
    data: params,
  })
}

/** 获取教材配置列表 */
export function getCurriculumConfigs(personaId: number) {
  return request<{ items: CurriculumConfig[] }>({
    url: `/api/curriculum-configs?persona_id=${personaId}`,
    method: 'GET',
  })
}

/** 更新教材配置 */
export function updateCurriculumConfig(id: number, params: Partial<CreateCurriculumParams>) {
  return request<CurriculumConfig>({
    url: `/api/curriculum-configs/${id}`,
    method: 'PUT',
    data: params,
  })
}

/** 删除教材配置 */
export function deleteCurriculumConfig(id: number) {
  return request<{ message: string }>({
    url: `/api/curriculum-configs/${id}`,
    method: 'DELETE',
  })
}

/** 获取教材版本列表（支持按地区筛选） */
export function getCurriculumVersions(params?: { region?: string; grade_level?: string }) {
  return request<{ items: string[] }>({
    url: '/api/curriculum-versions',
    method: 'GET',
    data: params,
  })
}
