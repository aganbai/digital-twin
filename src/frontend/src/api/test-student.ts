import { request } from './request'

/** 自测学生所在班级信息 */
export interface TestStudentClass {
  class_id: number
  class_name: string
  persona_id: number
}

/** 自测学生信息 */
export interface TestStudentInfo {
  user_id: number
  username: string
  persona_id: number
  nickname: string
  password_hint: string
  is_active: boolean
  joined_classes: TestStudentClass[]
  created_at: string
}

/** 重置结果 */
export interface TestStudentResetResult {
  cleared_conversations: number
  cleared_memories: number
  message: string
}

/** 模拟登录结果 */
export interface TestStudentLoginResult {
  token: string
  user_id: number
  username: string
  nickname: string
}

/**
 * 获取自测学生信息
 */
export function getTestStudent() {
  return request<TestStudentInfo>({
    url: '/api/test-student',
    method: 'GET',
  })
}

/**
 * 重置自测学生数据
 * 清空自测学生与该教师所有班级分身的对话记录和记忆
 */
export function resetTestStudent() {
  return request<TestStudentResetResult>({
    url: '/api/test-student/reset',
    method: 'POST',
  })
}

/**
 * 模拟登录测试学生
 * 返回测试学生token，用于测试
 */
export function loginTestStudent() {
  return request<TestStudentLoginResult>({
    url: '/api/test-student/login',
    method: 'POST',
  })
}
