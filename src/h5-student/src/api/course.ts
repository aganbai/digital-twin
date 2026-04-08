import { get } from '@/utils/request'

/** 课程信息 */
export interface Course {
  id: number
  title: string
  content?: string
  teacher_name?: string
  created_at: string
}

/** 学习资料 */
export interface Material {
  id: number
  title: string
  type: string
  url: string
  created_at: string
}

/**
 * 获取课程列表
 */
export function getCourseList() {
  return get<Course[]>('/api/student/courses')
}

/**
 * 获取课程详情
 */
export function getCourseDetail(courseId: number) {
  return get<Course>(`/api/student/courses/${courseId}`)
}

/**
 * 获取学习资料列表
 */
export function getMaterialList(courseId?: number) {
  return get<Material[]>('/api/student/materials', { course_id: courseId })
}

/**
 * 获取我的班级信息
 */
export function getMyClass() {
  return get<any>('/api/student/my-class')
}
