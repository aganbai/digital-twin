import { get, post, put, del } from '@/utils/request'
import type { PaginatedData } from '@/utils/request'

/** 课程信息 */
export interface Course {
  id: number
  title: string
  content?: string
  class_id: number
  class_name?: string
  status: 'draft' | 'published'
  created_at: string
}

/** 创建课程参数 */
export interface CreateCourseParams {
  title: string
  content?: string
  class_id: number
}

/**
 * 获取课程列表
 */
export function getCourseList(params?: { class_id?: number }) {
  return get<Course[]>('/api/teacher/courses', params)
}

/**
 * 获取课程详情
 */
export function getCourseDetail(courseId: number) {
  return get<Course>(`/api/teacher/courses/${courseId}`)
}

/**
 * 创建课程
 */
export function createCourse(params: CreateCourseParams) {
  return post<Course>('/api/teacher/courses', params)
}

/**
 * 更新课程
 */
export function updateCourse(courseId: number, params: CreateCourseParams) {
  return put(`/api/teacher/courses/${courseId}`, params)
}

/**
 * 发布课程
 */
export function publishCourse(courseId: number) {
  return post(`/api/teacher/courses/${courseId}/publish`)
}

/**
 * 删除课程
 */
export function deleteCourse(courseId: number) {
  return del(`/api/teacher/courses/${courseId}`)
}
