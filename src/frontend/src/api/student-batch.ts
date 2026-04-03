import { request } from './request'

/** 解析后的学生信息 */
export interface ParsedStudent {
  name: string            // 必填：姓名
  gender: string          // 必填：male/female
  age?: number            // 选填：年龄
  student_id?: string     // 选填：学号
  strengths?: string      // 选填：擅长学科
  weaknesses?: string     // 选填：薄弱学科
  learning_style?: string // 选填：学习风格偏好
  personality_tags?: string[] // 选填：性格特点标签
  interests?: string      // 选填：兴趣爱好
  specialties?: string    // 选填：特长
  parent_notes?: string   // 选填：家长备注
  extra_info?: object     // 选填：扩展信息
}

/** LLM解析学生文本 */
export function parseStudentText(text: string) {
  return request<{ students: ParsedStudent[]; parse_method: string }>({
    url: '/api/students/parse-text',
    method: 'POST',
    data: { text },
  })
}

/** 批量创建学生 */
export function batchCreateStudents(params: {
  persona_id: number
  class_id?: number
  students: ParsedStudent[]
}) {
  return request<{
    total: number
    success_count: number
    failed_count: number
    results: Array<{ name: string; status: string; user_id?: number; persona_id?: number; error?: string }>
  }>({
    url: '/api/students/batch-create',
    method: 'POST',
    data: params,
  })
}
