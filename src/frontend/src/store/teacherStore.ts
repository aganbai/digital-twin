import { create } from 'zustand'
import { Teacher } from '../api/teacher'

/** 教师状态 */
interface TeacherState {
  /** 教师列表 */
  teachers: Teacher[]
  /** 是否正在加载 */
  loading: boolean

  /** 设置教师列表 */
  setTeachers: (teachers: Teacher[]) => void
  /** 设置加载状态 */
  setLoading: (loading: boolean) => void
}

export const useTeacherStore = create<TeacherState>((set) => ({
  teachers: [],
  loading: false,

  setTeachers: (teachers: Teacher[]) => {
    set({ teachers })
  },

  setLoading: (loading: boolean) => {
    set({ loading })
  },
}))
