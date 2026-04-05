import { useState } from 'react'
import { View, Text } from '@tarojs/components'
import './index.scss'

/** 思考步骤类型 */
export type ThinkingStepType = 'rag_search' | 'memory_recall' | 'tool_call' | 'llm_thinking'

/** 思考步骤状态 */
export type ThinkingStepStatus = 'start' | 'done'

/** 思考步骤事件 */
export interface ThinkingStepEvent {
  type: 'thinking_step'
  step: ThinkingStepType
  status: ThinkingStepStatus
  message: string
  detail?: string
  duration_ms?: number
  timestamp: number
}

/** 思考步骤数据 */
export interface ThinkingStep {
  step: ThinkingStepType
  status: ThinkingStepStatus
  message: string
  detail?: string
  duration_ms?: number
  timestamp: number
}

interface ThinkingPanelProps {
  /** 思考步骤列表 */
  steps: ThinkingStep[]
  /** 默认是否展开 */
  defaultExpanded?: boolean
  /** 教师名称 */
  teacherName?: string
}

/** 步骤显示配置 */
const STEP_CONFIG: Record<ThinkingStepType, { label: string; icon: string }> = {
  rag_search: { label: '知识库检索', icon: '🔍' },
  memory_recall: { label: '记忆检索', icon: '🧠' },
  tool_call: { label: '搜索增强', icon: '🔧' },
  llm_thinking: { label: 'AI 思考', icon: '💭' },
}

/**
 * 思考过程展示面板
 * 用于展示 AI 回复前的思考步骤
 */
export default function ThinkingPanel({
  steps,
  defaultExpanded = true,
  teacherName,
}: ThinkingPanelProps) {
  const [expanded, setExpanded] = useState(defaultExpanded)

  if (steps.length === 0) return null

  // 获取最后一个步骤的状态
  const lastStep = steps[steps.length - 1]
  const isCompleted = lastStep.status === 'done' && lastStep.duration_ms

  return (
    <View className='thinking-panel'>
      {/* 头部 */}
      <View className='thinking-panel__header' onClick={() => setExpanded(!expanded)}>
        <View className='thinking-panel__header-left'>
          <Text className='thinking-panel__icon'>🤖</Text>
          <Text className='thinking-panel__title'>
            {isCompleted ? '思考过程' : '正在思考...'}
          </Text>
        </View>
        <View className='thinking-panel__header-right'>
          {isCompleted && lastStep.duration_ms && (
            <Text className='thinking-panel__total-time'>
              总耗时 {lastStep.duration_ms}ms
            </Text>
          )}
          <Text className={`thinking-panel__arrow ${expanded ? 'thinking-panel__arrow--expanded' : ''}`}>
            ▼
          </Text>
        </View>
      </View>

      {/* 步骤列表 */}
      {expanded && (
        <View className='thinking-panel__steps'>
          {steps.map((step, index) => {
            const config = STEP_CONFIG[step.step]
            const isDone = step.status === 'done'

            return (
              <View key={`${step.step}-${index}`} className='thinking-panel__step'>
                <View className='thinking-panel__step-header'>
                  <Text className='thinking-panel__step-icon'>{config.icon}</Text>
                  <Text className='thinking-panel__step-label'>{config.label}</Text>
                  <Text className={`thinking-panel__step-status ${isDone ? 'thinking-panel__step-status--done' : ''}`}>
                    {isDone ? '✅' : '⏳'}
                  </Text>
                </View>
                <View className='thinking-panel__step-body'>
                  <Text className='thinking-panel__step-message'>{step.message}</Text>
                  {step.detail && (
                    <Text className='thinking-panel__step-detail'>{step.detail}</Text>
                  )}
                  {step.duration_ms && (
                    <Text className='thinking-panel__step-duration'>{step.duration_ms}ms</Text>
                  )}
                </View>
              </View>
            )
          })}
        </View>
      )}
    </View>
  )
}
