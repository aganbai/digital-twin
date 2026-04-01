import { useState, useEffect, useCallback } from 'react'
import { View, Text, Input, Textarea, Picker } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import {
  getStyleConfig, setStyleConfig,
  StyleConfig, TeachingStyle,
  TEACHING_STYLES, TEACHING_STYLE_LABELS, TEACHING_STYLE_DESCRIPTIONS,
} from '@/api/style'
import { getComments, createComment, CommentItem } from '@/api/comment'
import { usePersonaStore } from '@/store'
import { formatTime } from '@/utils/format'
import './index.scss'

const GUIDANCE_LEVELS = ['low', 'medium', 'high'] as const
const GUIDANCE_LABELS = { low: '低', medium: '中', high: '高' }
const GUIDANCE_DESCRIPTIONS = {
  low: '低 - 多给直接答案',
  medium: '中 - 适度引导',
  high: '高 - 严格引导式提问',
}

export default function StudentDetail() {
  const router = useRouter()
  const studentId = Number(router.params.student_id) || 0
  const studentName = decodeURIComponent(router.params.student_name || '学生')
  const studentPersonaId = Number(router.params.student_persona_id) || 0
  const { currentPersona } = usePersonaStore()

  // 问答风格
  const [guidanceLevel, setGuidanceLevel] = useState<'low' | 'medium' | 'high'>('medium')
  const [teachingStyle, setTeachingStyle] = useState<TeachingStyle>('socratic')
  const [stylePrompt, setStylePrompt] = useState('')
  const [savingStyle, setSavingStyle] = useState(false)

  // R5: 评语改为"学生备注"（仅教师可见）
  const [comments, setComments] = useState<CommentItem[]>([])
  const [showCommentForm, setShowCommentForm] = useState(false)
  const [commentContent, setCommentContent] = useState('')
  const [progressSummary, setProgressSummary] = useState('')
  const [submittingComment, setSubmittingComment] = useState(false)

  /** 设置导航栏标题 */
  useEffect(() => {
    Taro.setNavigationBarTitle({ title: `${studentName} 的详情` })
  }, [studentName])

  /** 获取问答风格 */
  const fetchStyle = useCallback(async () => {
    if (!currentPersona?.id || !studentPersonaId) return
    try {
      const res = await getStyleConfig(currentPersona.id, studentPersonaId)
      if (res.data) {
        const config = res.data.style_config
        if (config.guidance_level) setGuidanceLevel(config.guidance_level)
        if (config.teaching_style) setTeachingStyle(config.teaching_style)
        if (config.style_prompt) setStylePrompt(config.style_prompt)
      }
    } catch (error) {
      console.error('获取问答风格失败:', error)
    }
  }, [currentPersona?.id, studentPersonaId])

  /** 获取评语列表 */
  const fetchComments = useCallback(async () => {
    try {
      const res = await getComments({ student_id: studentId, page: 1, page_size: 50 })
      setComments(res.data.items || [])
    } catch (error) {
      console.error('获取评语失败:', error)
    }
  }, [studentId])

  useEffect(() => {
    if (studentId) {
      fetchStyle()
      fetchComments()
    }
  }, [studentId, fetchStyle, fetchComments])

  /** 保存问答风格 */
  const handleSaveStyle = async () => {
    if (!currentPersona?.id || !studentPersonaId) return
    setSavingStyle(true)
    try {
      await setStyleConfig({
        teacher_persona_id: currentPersona.id,
        student_persona_id: studentPersonaId,
        style_config: {
          guidance_level: guidanceLevel,
          teaching_style: teachingStyle,
          style_prompt: stylePrompt.trim() || undefined,
        },
      })
      Taro.showToast({ title: '保存成功', icon: 'success' })
    } catch (error) {
      console.error('保存风格失败:', error)
    } finally {
      setSavingStyle(false)
    }
  }

  /** 提交评语 */
  const handleSubmitComment = async () => {
    const trimmedContent = commentContent.trim()
    if (!trimmedContent) {
      Taro.showToast({ title: '请输入评语内容', icon: 'none' })
      return
    }

    setSubmittingComment(true)
    try {
      await createComment({
        student_id: studentId,
        content: trimmedContent,
        progress_summary: progressSummary.trim() || undefined,
      })
      Taro.showToast({ title: '评语已提交', icon: 'success' })
      setShowCommentForm(false)
      setCommentContent('')
      setProgressSummary('')
      fetchComments()
    } catch (error) {
      console.error('提交评语失败:', error)
    } finally {
      setSubmittingComment(false)
    }
  }

  /** 引导程度选择 */
  const guidanceLevelIndex = GUIDANCE_LEVELS.indexOf(guidanceLevel)

  return (
    <View className='student-detail-page'>
      {/* 学生基本信息 */}
      <View className='student-detail-page__header'>
        <View className='student-detail-page__avatar'>
          <Text className='student-detail-page__avatar-text'>
            {studentName.charAt(0)}
          </Text>
        </View>
        <Text className='student-detail-page__name'>{studentName}</Text>
        {/* 查看对话记录入口 */}
        {studentPersonaId > 0 && (
          <View
            className='student-detail-page__chat-history-btn'
            onClick={() => {
              Taro.navigateTo({
                url: `/pages/student-chat-history/index?student_persona_id=${studentPersonaId}&student_name=${encodeURIComponent(studentName)}`,
              })
            }}
          >
            <Text className='student-detail-page__chat-history-btn-text'>💬 查看对话记录</Text>
          </View>
        )}
      </View>

      {/* 问答风格设置 */}
      <View className='student-detail-page__section'>
        <Text className='student-detail-page__section-title'>问答风格设置</Text>

        {/* 教学风格选择器 */}
        <View className='student-detail-page__field'>
          <Text className='student-detail-page__field-label'>教学风格</Text>
          <View className='student-detail-page__style-grid'>
            {TEACHING_STYLES.map((style) => (
              <View
                key={style}
                className={`student-detail-page__style-card ${teachingStyle === style ? 'student-detail-page__style-card--active' : ''}`}
                onClick={() => setTeachingStyle(style)}
              >
                <Text className={`student-detail-page__style-card-name ${teachingStyle === style ? 'student-detail-page__style-card-name--active' : ''}`}>
                  {TEACHING_STYLE_LABELS[style]}
                </Text>
                <Text className='student-detail-page__style-card-desc'>
                  {TEACHING_STYLE_DESCRIPTIONS[style]}
                </Text>
              </View>
            ))}
          </View>
        </View>

        <View className='student-detail-page__field'>
          <Text className='student-detail-page__field-label'>引导程度</Text>
          <Picker
            mode='selector'
            range={[GUIDANCE_DESCRIPTIONS.low, GUIDANCE_DESCRIPTIONS.medium, GUIDANCE_DESCRIPTIONS.high]}
            value={guidanceLevelIndex}
            onChange={(e) => setGuidanceLevel(GUIDANCE_LEVELS[Number(e.detail.value)])}
          >
            <View className='student-detail-page__picker'>
              <Text className='student-detail-page__picker-text'>
                {GUIDANCE_LABELS[guidanceLevel]}
              </Text>
              <Text className='student-detail-page__picker-arrow'>▼</Text>
            </View>
          </Picker>
        </View>

        <View className='student-detail-page__field'>
          <Text className='student-detail-page__field-label'>风格描述{teachingStyle === 'custom' ? '（必填）' : '（可选）'}</Text>
          <Textarea
            className='student-detail-page__style-textarea'
            placeholder='如：对该学生请多用鼓励性语言，注重基础概念的巩固'
            value={stylePrompt}
            maxlength={500}
            onInput={(e) => setStylePrompt(e.detail.value)}
          />
        </View>

        <View
          className={`student-detail-page__save-btn ${savingStyle ? 'student-detail-page__save-btn--disabled' : ''}`}
          onClick={savingStyle ? undefined : handleSaveStyle}
        >
          <Text className='student-detail-page__save-btn-text'>
            {savingStyle ? '保存中...' : '保存设置'}
          </Text>
        </View>
      </View>

      {/* R5: 评语区域 → 学生备注（仅教师可见） */}
      <View className='student-detail-page__section'>
        <View className='student-detail-page__section-header'>
          <Text className='student-detail-page__section-title'>学生备注</Text>
          <View
            className='student-detail-page__add-btn'
            onClick={() => setShowCommentForm(!showCommentForm)}
          >
            <Text className='student-detail-page__add-btn-text'>
              {showCommentForm ? '取消' : '+ 写备注'}
            </Text>
          </View>
        </View>
        <Text className='student-detail-page__section-hint'>
          备注仅教师可见，AI 会参考备注内容进行个性化回复
        </Text>

        {/* 写备注表单 */}
        {showCommentForm && (
          <View className='student-detail-page__comment-form'>
            <Textarea
              className='student-detail-page__comment-textarea'
              placeholder='请输入学生备注内容...'
              value={commentContent}
              maxlength={2000}
              onInput={(e) => setCommentContent(e.detail.value)}
            />
            <Input
              className='student-detail-page__comment-progress'
              placeholder='学习进度摘要（可选）'
              value={progressSummary}
              maxlength={500}
              onInput={(e) => setProgressSummary(e.detail.value)}
            />
            <View
              className={`student-detail-page__submit-comment ${submittingComment ? 'student-detail-page__submit-comment--disabled' : ''}`}
              onClick={submittingComment ? undefined : handleSubmitComment}
            >
              <Text className='student-detail-page__submit-comment-text'>
                {submittingComment ? '提交中...' : '提交备注'}
              </Text>
            </View>
          </View>
        )}

        {/* 备注列表 */}
        {comments.length > 0 ? (
          comments.map((item) => (
            <View key={item.id} className='student-detail-page__comment-card'>
              <Text className='student-detail-page__comment-date'>
                {formatTime(item.created_at)}
              </Text>
              <Text className='student-detail-page__comment-content'>{item.content}</Text>
              {item.progress_summary && (
                <Text className='student-detail-page__comment-progress-text'>
                  进度：{item.progress_summary}
                </Text>
              )}
            </View>
          ))
        ) : (
          <Text className='student-detail-page__empty-text'>暂无备注</Text>
        )}
      </View>
    </View>
  )
}
