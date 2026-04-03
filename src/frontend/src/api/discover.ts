import { request } from './request'

/** 热门班级 */
export interface HotClass {
  id: number
  name: string
  description: string
  subject: string
  teacher_name: string
  teacher_avatar: string
  member_count: number
  created_at: string
}

/** 推荐老师 */
export interface RecommendedTeacher {
  id: number
  nickname: string
  avatar: string
  school: string
  subject: string
  description: string
  student_count: number
  document_count: number
  rating: number
}

/** 学科分类 */
export interface SubjectItem {
  name: string
  icon: string
  count: number
}

/** 发现页推荐列表响应 */
export interface DiscoverResponse {
  hot_classes: HotClass[]
  recommended_teachers: RecommendedTeacher[]
  subjects: SubjectItem[]
}

/** 班级详情响应 */
export interface ClassDetailResponse {
  id: number
  name: string
  description: string
  subject: string
  teacher_name: string
  teacher_avatar: string
  member_count: number
  documents: Array<{ id: number; title: string; doc_type: string }>
  created_at: string
}

/** 搜索结果 */
export interface SearchResult {
  classes: HotClass[]
  teachers: RecommendedTeacher[]
}

/**
 * 获取发现页推荐列表
 * 返回热门班级、推荐老师、学科分类
 */
export function getDiscoverList() {
  return request<DiscoverResponse>({
    url: '/api/discover',
    method: 'GET',
  })
}

/**
 * 获取班级详情
 * @param classId - 班级 ID
 */
export function getDiscoverDetail(classId: number) {
  return request<ClassDetailResponse>({
    url: `/api/discover/detail?class_id=${classId}`,
    method: 'GET',
  })
}

/**
 * 搜索班级/老师
 * @param keyword - 搜索关键词
 */
export function searchDiscover(keyword: string) {
  return request<SearchResult>({
    url: `/api/discover/search?keyword=${encodeURIComponent(keyword)}`,
    method: 'GET',
  })
}
