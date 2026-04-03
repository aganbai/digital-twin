import { useState } from 'react'
import { View, Text, Textarea } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { createFeedback, FEEDBACK_TYPES, FeedbackType } from '@/api/feedback'
import './index.scss'

/** 自动采集上下文信息（页面、设备） */
function collectContextInfo(): Record<string, any> {
  const systemInfo = Taro.getSystemInfoSync()
  const pages = Taro.getCurrentPages()
  const prevPage = pages.length > 1 ? pages[pages.length - 2] : null
  return {
    page: prevPage?.route || 'unknown',
    timestamp: new Date().toISOString(),
    device: {
      brand: systemInfo.brand,
      model: systemInfo.model,
      system: systemInfo.system,
      platform: systemInfo.platform,
      version: systemInfo.version,
      SDKVersion: systemInfo.SDKVersion,
      screenWidth: systemInfo.screenWidth,
      screenHeight: systemInfo.screenHeight,
    },
  }
}

export default function FeedbackPage() {
  const [feedbackType, setFeedbackType] = useState<FeedbackType | ''>('')
  const [content, setContent] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async () => {
    if (!feedbackType) {
      Taro.showToast({ title: '请选择反馈类型', icon: 'none' })
      return
    }
    if (!content.trim()) {
      Taro.showToast({ title: '请输入反馈内容', icon: 'none' })
      return
    }
    if (content.trim().length < 5) {
      Taro.showToast({ title: '反馈内容至少5个字', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      await createFeedback({
        feedback_type: feedbackType,
        content: content.trim(),
        context_info: collectContextInfo(),
      })
      Taro.showToast({ title: '反馈提交成功，感谢！', icon: 'success' })
      setFeedbackType('')
      setContent('')
      setTimeout(() => Taro.navigateBack(), 1500)
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '提交失败', icon: 'none' })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <View className='feedback-page'>
      <View className='feedback-page__header'>
        <Text className='feedback-page__title'>📮 意见反馈</Text>
        <Text className='feedback-page__subtitle'>您的反馈对我们非常重要</Text>
      </View>

      <View className='feedback-page__form'>
        {/* 反馈类型 */}
        <View className='feedback-page__field'>
          <Text className='feedback-page__label'>反馈类型</Text>
          <View className='feedback-page__types'>
            {FEEDBACK_TYPES.map(type => (
              <View
                key={type.value}
                className={`feedback-page__type ${feedbackType === type.value ? 'feedback-page__type--active' : ''}`}
                onClick={() => setFeedbackType(type.value)}
              >
                <Text className='feedback-page__type-label'>{type.label}</Text>
                <Text className='feedback-page__type-desc'>{type.desc}</Text>
              </View>
            ))}
          </View>
        </View>

        {/* 反馈内容 */}
        <View className='feedback-page__field'>
          <Text className='feedback-page__label'>反馈内容</Text>
          <Textarea
            className='feedback-page__textarea'
            value={content}
            placeholder='请详细描述您的反馈（至少5个字）...'
            maxlength={2000}
            onInput={(e) => setContent(e.detail.value)}
          />
          <Text className='feedback-page__char-count'>{content.length}/2000</Text>
        </View>

        {/* 提交按钮 */}
        <View
          className={`feedback-page__submit-btn ${(!feedbackType || !content.trim() || submitting) ? 'feedback-page__submit-btn--disabled' : ''}`}
          onClick={handleSubmit}
        >
          <Text className='feedback-page__submit-btn-text'>
            {submitting ? '提交中...' : '提交反馈'}
          </Text>
        </View>
      </View>

      {/* 提示信息 */}
      <View className='feedback-page__tips'>
        <Text className='feedback-page__tips-text'>
          提交反馈时会自动关联当前页面和设备信息，帮助我们更好地定位问题。
        </Text>
      </View>
    </View>
  )
}
