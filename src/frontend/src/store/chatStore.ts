import { create } from 'zustand'

/** 聊天消息 */
export interface Message {
  id?: number
  role: 'user' | 'assistant'
  content: string
  created_at?: string
}

/** 对话状态 */
interface ChatState {
  /** 当前会话消息列表 */
  messages: Message[]
  /** 是否正在加载（等待 AI 回复） */
  loading: boolean
  /** 当前会话 ID */
  sessionId: string
  /** 当前对话的教师 ID */
  teacherId: number | null

  /** 添加一条消息 */
  addMessage: (msg: Message) => void
  /** 设置加载状态 */
  setLoading: (loading: boolean) => void
  /** 设置会话 ID */
  setSessionId: (id: string) => void
  /** 设置教师 ID */
  setTeacherId: (id: number) => void
  /** 批量设置消息（用于加载历史记录） */
  setMessages: (messages: Message[]) => void
  /** 清空消息 */
  clearMessages: () => void
}

export const useChatStore = create<ChatState>((set) => ({
  messages: [],
  loading: false,
  sessionId: '',
  teacherId: null,

  addMessage: (msg: Message) => {
    set((state) => ({ messages: [...state.messages, msg] }))
  },

  setLoading: (loading: boolean) => {
    set({ loading })
  },

  setSessionId: (id: string) => {
    set({ sessionId: id })
  },

  setTeacherId: (id: number) => {
    set({ teacherId: id })
  },

  setMessages: (messages: Message[]) => {
    set({ messages })
  },

  clearMessages: () => {
    set({ messages: [], sessionId: '', teacherId: null })
  },
}))
