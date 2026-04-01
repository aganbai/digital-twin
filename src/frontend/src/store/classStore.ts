import { create } from 'zustand'
import type { ClassInfo } from '../api/class'

/** 班级状态 */
interface ClassState {
  /** 班级列表 */
  classes: ClassInfo[]
  /** 是否加载中 */
  loading: boolean

  /** 设置班级列表 */
  setClasses: (classes: ClassInfo[]) => void
  /** 设置加载状态 */
  setLoading: (loading: boolean) => void
  /** 清除班级状态 */
  clearClasses: () => void
}

export const useClassStore = create<ClassState>((set) => ({
  classes: [],
  loading: false,

  setClasses: (classes: ClassInfo[]) => {
    set({ classes })
  },

  setLoading: (loading: boolean) => {
    set({ loading })
  },

  clearClasses: () => {
    set({ classes: [], loading: false })
  },
}))
