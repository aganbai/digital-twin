import { get, post, put, del } from '@/utils/request'
import type { PaginatedData } from '@/utils/request'

/** 班级信息 */
export interface ClassInfo {
  id: number
  name: string
  description?: string
  student_count: number
  created_at: string
}

/** 创建班级参数 */
export interface CreateClassParams {
  name: string
  description?: string
}

/**
 * 获取班级列表
 */
export function getClassList() {
  return get<ClassInfo[]>('/api/teacher/classes')
}

/**
 * 获取班级详情
 */
export function getClassDetail(classId: number) {
  return get<ClassInfo>(`/api/teacher/classes/${classId}`)
}

/**
 * 创建班级
 */
export function createClass(params: CreateClassParams) {
  return post<ClassInfo>('/api/teacher/classes', params)
}

/**
 * 更新班级
 */
export function updateClass(classId: number, params: CreateClassParams) {
  return put(`/api/teacher/classes/${classId}`, params)
}

/**
 * 删除班级
 */
export function deleteClass(classId: number) {
  return del(`/api/teacher/classes/${classId}`)
}

/**
 * 添加学生到班级
 */
export function addStudentToClass(classId: number, studentId: number) {
  return post(`/api/teacher/classes/${classId}/students`, { student_id: studentId })
}

/**
 * 从班级移除学生
 */
export function removeStudentFromClass(classId: number, studentId: number) {
  return del(`/api/teacher/classes/${classId}/students/${studentId}`)
}

/**
 * 获取班级学生列表
 */
export function getClassStudents(classId: number) {
  return get<any[]>(`/api/teacher/classes/${classId}/students`)
}
