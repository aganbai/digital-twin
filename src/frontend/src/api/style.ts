import { request } from './request'

/** 教学风格类型 */
export type TeachingStyle = 'socratic' | 'explanatory' | 'encouraging' | 'strict' | 'companion' | 'custom'

/** 教学风格标签映射 */
export const TEACHING_STYLE_LABELS: Record<TeachingStyle, string> = {
  socratic: '苏格拉底式提问',
  explanatory: '讲解式教学',
  encouraging: '鼓励式教学',
  strict: '严格式教学',
  companion: '陪伴式学习',
  custom: '自定义',
}

/** 教学风格描述映射 */
export const TEACHING_STYLE_DESCRIPTIONS: Record<TeachingStyle, string> = {
  socratic: '不直接给答案，通过提问引导思考',
  explanatory: '详细讲解知识点，配合举例说明',
  encouraging: '多用肯定语言，循序渐进引导',
  strict: '严格要求，注重准确性和规范性',
  companion: '像朋友一样陪伴学习，轻松氛围',
  custom: '完全由风格描述决定',
}

/** 所有教学风格选项 */
export const TEACHING_STYLES: TeachingStyle[] = [
  'socratic', 'explanatory', 'encouraging', 'strict', 'companion', 'custom',
]

/** 风格配置 */
export interface StyleConfig {
  temperature?: number
  guidance_level?: 'low' | 'medium' | 'high'
  teaching_style?: TeachingStyle
  style_prompt?: string
  max_turns_per_topic?: number
}

/** 风格配置请求体 */
export interface StyleConfigRequest {
  teacher_persona_id: number
  student_persona_id: number
  style_config: StyleConfig
}

/** 问答风格响应 */
export interface DialogueStyleResponse {
  id: number
  teacher_id: number
  student_id: number
  teacher_persona_id: number
  student_persona_id: number
  style_config: StyleConfig
  created_at: string
  updated_at: string
}

/**
 * 设置学生问答风格（迭代6 新接口）
 * @param data - 风格配置请求体（含 teacher_persona_id + student_persona_id）
 */
export function setStyleConfig(data: StyleConfigRequest) {
  return request<DialogueStyleResponse>({
    url: '/api/styles',
    method: 'PUT',
    data,
  })
}

/**
 * 获取学生问答风格（迭代6 新接口）
 * @param teacherPersonaId - 教师分身 ID
 * @param studentPersonaId - 学生分身 ID
 */
export function getStyleConfig(teacherPersonaId: number, studentPersonaId: number) {
  return request<DialogueStyleResponse>({
    url: `/api/styles?teacher_persona_id=${teacherPersonaId}&student_persona_id=${studentPersonaId}`,
    method: 'GET',
  })
}

/**
 * 设置学生问答风格（旧接口，保持向后兼容）
 * @param studentId - 学生 ID
 * @param data - 风格配置
 */
export function setDialogueStyle(studentId: number, data: StyleConfig) {
  return request<DialogueStyleResponse>({
    url: `/api/students/${studentId}/dialogue-style`,
    method: 'PUT',
    data,
  })
}

/**
 * 获取学生问答风格（旧接口，保持向后兼容）
 * @param studentId - 学生 ID
 * @param teacherId - 教师 ID（学生查看时需指定）
 */
export function getDialogueStyle(studentId: number, teacherId?: number) {
  const query = teacherId ? `?teacher_id=${teacherId}` : ''
  return request<DialogueStyleResponse | null>({
    url: `/api/students/${studentId}/dialogue-style${query}`,
    method: 'GET',
  })
}
