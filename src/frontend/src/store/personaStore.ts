import { create } from 'zustand'
import type { Persona } from '../api/persona'

/** 分身状态 */
interface PersonaState {
  /** 当前分身 */
  currentPersona: Persona | null
  /** 分身列表 */
  personas: Persona[]
  /** 默认分身 ID */
  defaultPersonaId: number | null
  /** 是否加载中 */
  loading: boolean

  /** 设置当前分身 */
  setCurrentPersona: (persona: Persona | null) => void
  /** 设置分身列表 */
  setPersonas: (personas: Persona[], defaultId?: number) => void
  /** 设置加载状态 */
  setLoading: (loading: boolean) => void
  /** 清除分身状态 */
  clearPersonas: () => void
}

export const usePersonaStore = create<PersonaState>((set) => ({
  currentPersona: null,
  personas: [],
  defaultPersonaId: null,
  loading: false,

  setCurrentPersona: (persona: Persona | null) => {
    set({ currentPersona: persona })
  },

  setPersonas: (personas: Persona[], defaultId?: number) => {
    set({ personas, defaultPersonaId: defaultId ?? null })
  },

  setLoading: (loading: boolean) => {
    set({ loading })
  },

  clearPersonas: () => {
    set({ currentPersona: null, personas: [], defaultPersonaId: null })
  },
}))
