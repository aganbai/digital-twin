import { create } from 'zustand'
import { RelationItemTeacher, RelationItemStudent } from '../api/relation'

/** 师生关系状态 */
interface RelationState {
  /** 关系列表（教师视角） */
  teacherRelations: RelationItemTeacher[]
  /** 关系列表（学生视角） */
  studentRelations: RelationItemStudent[]
  /** 是否正在加载 */
  loading: boolean

  /** 设置教师视角关系列表 */
  setTeacherRelations: (relations: RelationItemTeacher[]) => void
  /** 设置学生视角关系列表 */
  setStudentRelations: (relations: RelationItemStudent[]) => void
  /** 设置加载状态 */
  setLoading: (loading: boolean) => void
  /** 更新单条关系状态（教师端审批后更新） */
  updateRelationStatus: (id: number, status: string) => void
  /** 移除单条关系（审批拒绝后移除） */
  removeRelation: (id: number) => void
}

export const useRelationStore = create<RelationState>((set) => ({
  teacherRelations: [],
  studentRelations: [],
  loading: false,

  setTeacherRelations: (relations: RelationItemTeacher[]) => {
    set({ teacherRelations: relations })
  },

  setStudentRelations: (relations: RelationItemStudent[]) => {
    set({ studentRelations: relations })
  },

  setLoading: (loading: boolean) => {
    set({ loading })
  },

  updateRelationStatus: (id: number, status: string) => {
    set((state) => ({
      teacherRelations: state.teacherRelations.map((r) =>
        r.id === id ? { ...r, status: status as any } : r,
      ),
      studentRelations: state.studentRelations.map((r) =>
        r.id === id ? { ...r, status: status as any } : r,
      ),
    }))
  },

  removeRelation: (id: number) => {
    set((state) => ({
      teacherRelations: state.teacherRelations.filter((r) => r.id !== id),
      studentRelations: state.studentRelations.filter((r) => r.id !== id),
    }))
  },
}))
