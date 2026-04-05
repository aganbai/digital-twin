import { request, PaginatedData } from './request'

/** 课程信息 */
export interface CourseInfo {
  id: number
  title: string
  content: string
  class_id: number
  class_name?: string
  created_at: string
  pushed: boolean
}

/** 创建课程参数 */
export interface CreateCourseParams {
  title: string
  content: string
  class_id: number
  push_to_students?: boolean
}

/** 更新课程参数 */
export interface UpdateCourseParams {
  title?: string
  content?: string
}

/** 推送课程参数 */
export interface PushCourseParams {
  push_type: 'in_app' | 'wechat'
}

/** 推送课程响应 */
export interface PushCourseResponse {
  pushed_count: number
  failed_count: number
}

/**
 * 发布课程信息
 */
export function createCourse(params: CreateCourseParams) {
  return request<{ id: number; title: string; created_at: string }>({
    url: '/api/courses',
    method: 'POST',
    data: params,
  })
}

/**
 * 获取课程列表
 */
export function getCourses(classId?: number, page = 1, pageSize = 20) {
  const query = new URLSearchParams()
  if (classId) query.append('class_id', String(classId))
  query.append('page', String(page))
  query.append('page_size', String(pageSize))
  return request<PaginatedData<CourseInfo>>({
    url: `/api/courses?${query.toString()}`,
    method: 'GET',
  })
}

/**
 * 更新课程信息
 */
export function updateCourse(id: number, params: UpdateCourseParams) {
  return request<{ message: string }>({
    url: `/api/courses/${id}`,
    method: 'PUT',
    data: params,
  })
}

/**
 * 删除课程
 */
export function deleteCourse(id: number) {
  return request<{ message: string }>({
    url: `/api/courses/${id}`,
    method: 'DELETE',
  })
}

/**
 * 推送课程通知
 */
export function pushCourse(id: number, params: PushCourseParams) {
  return request<PushCourseResponse>({
    url: `/api/courses/${id}/push`,
    method: 'POST',
    data: params,
  })
}

/** 微信订阅参数 */
export interface WxSubscribeParams {
  template_id: string
  subscribe: boolean
}

/** 微信订阅状态 */
export interface WxSubscribeStatus {
  template_id: string
  is_subscribed: boolean
  last_subscribe_time?: string
}

/**
 * 订阅微信消息
 */
export function subscribeWxMessage(params: WxSubscribeParams) {
  return request<{ message: string }>({
    url: '/api/wx/subscribe',
    method: 'POST',
    data: params,
  })
}

/**
 * 获取订阅状态
 */
export function getWxSubscribeStatus(templateId: string) {
  return request<WxSubscribeStatus>({
    url: `/api/wx/subscribe/status?template_id=${templateId}`,
    method: 'GET',
  })
}
