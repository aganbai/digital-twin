import { create } from 'zustand'
import { AssignmentListItem } from '../api/assignment'

/** 作业状态 */
interface AssignmentState {
  /** 作业列表 */
  assignments: AssignmentListItem[]
  /** 是否正在加载 */
  loading: boolean

  /** 设置作业列表 */
  setAssignments: (assignments: AssignmentListItem[]) => void
  /** 设置加载状态 */
  setLoading: (loading: boolean) => void
}

export const useAssignmentStore = create<AssignmentState>((set) => ({
  assignments: [],
  loading: false,

  setAssignments: (assignments: AssignmentListItem[]) => {
    set({ assignments })
  },

  setLoading: (loading: boolean) => {
    set({ loading })
  },
}))
