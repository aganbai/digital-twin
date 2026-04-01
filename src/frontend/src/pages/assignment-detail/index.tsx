import { useState, useEffect, useCallback } from 'react'
import { View, Text, Textarea, Input } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import {
  getAssignmentDetail,
  reviewAssignment,
  aiReviewAssignment,
  AssignmentDetail as AssignmentDetailType,
} from '@/api/assignment'
import { useUserStore } from '@/store'
import { formatTime } from '@/utils/format'
import './index.scss'

export default function AssignmentDetail() {
  const router = useRouter()
  const assignmentId = Number(router.params.id) || 0
  const { userInfo } = useUserStore()

  const [detail, setDetail] = useState<AssignmentDetailType | null>(null)
  const [loading, setLoading] = useState(false)

  // 教师点评表单
  const [showReviewForm, setShowReviewForm] = useState(false)
  const [reviewContent, setReviewContent] = useState('')
  const [reviewScore, setReviewScore] = useState('')
  const [submittingReview, setSubmittingReview] = useState(false)

  // AI 点评
  const [requestingAI, setRequestingAI] = useState(false)

  /** 获取作业详情 */
  const fetchDetail = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getAssignmentDetail(assignmentId)
      setDetail(res.data)
    } catch (error) {
      console.error('获取作业详情失败:', error)
    } finally {
      setLoading(false)
    }
  }, [assignmentId])

  useEffect(() => {
    if (assignmentId) {
      fetchDetail()
    }
  }, [assignmentId, fetchDetail])

  /** AI 点评 */
  const aiReviews = detail?.reviews?.filter((r) => r.reviewer_type === 'ai') || []
  /** 教师点评 */
  const teacherReviews = detail?.reviews?.filter((r) => r.reviewer_type === 'teacher') || []

  /** 请求 AI 点评 */
  const handleAIReview = async () => {
    setRequestingAI(true)
    try {
      await aiReviewAssignment(assignmentId)
      Taro.showToast({ title: 'AI 点评完成', icon: 'success' })
      fetchDetail()
    } catch (error) {
      console.error('AI 点评失败:', error)
    } finally {
      setRequestingAI(false)
    }
  }

  /** 提交教师点评 */
  const handleSubmitReview = async () => {
    if (!reviewContent.trim()) {
      Taro.showToast({ title: '请输入点评内容', icon: 'none' })
      return
    }

    setSubmittingReview(true)
    try {
      const data: { content: string; score?: number } = {
        content: reviewContent.trim(),
      }
      if (reviewScore) {
        const score = parseFloat(reviewScore)
        if (!isNaN(score) && score >= 0 && score <= 100) {
          data.score = score
        }
      }
      await reviewAssignment(assignmentId, data)
      Taro.showToast({ title: '点评已提交', icon: 'success' })
      setShowReviewForm(false)
      setReviewContent('')
      setReviewScore('')
      fetchDetail()
    } catch (error) {
      console.error('提交点评失败:', error)
    } finally {
      setSubmittingReview(false)
    }
  }

  if (loading) {
    return (
      <View className='assignment-detail-page'>
        <View className='assignment-detail-page__loading'>
          <Text>加载中...</Text>
        </View>
      </View>
    )
  }

  if (!detail) {
    return (
      <View className='assignment-detail-page'>
        <View className='assignment-detail-page__loading'>
          <Text>作业不存在</Text>
        </View>
      </View>
    )
  }

  return (
    <View className='assignment-detail-page'>
      {/* 作业标题和信息 */}
      <View className='assignment-detail-page__header'>
        <Text className='assignment-detail-page__title'>{detail.title}</Text>
        <Text className='assignment-detail-page__meta'>
          {detail.student_nickname} · {formatTime(detail.created_at)}
        </Text>
      </View>

      {/* 作业内容 */}
      <View className='assignment-detail-page__section'>
        <Text className='assignment-detail-page__section-title'>作业内容</Text>
        {detail.content && (
          <Text className='assignment-detail-page__content'>{detail.content}</Text>
        )}
        {detail.file_path && (
          <View className='assignment-detail-page__file'>
            <Text className='assignment-detail-page__file-text'>
              📎 附件: {detail.file_path.split('/').pop()}
            </Text>
          </View>
        )}
        {!detail.content && !detail.file_path && (
          <Text className='assignment-detail-page__empty'>无内容</Text>
        )}
      </View>

      {/* AI 点评 */}
      <View className='assignment-detail-page__section'>
        <Text className='assignment-detail-page__section-title'>AI 点评</Text>
        {aiReviews.length > 0 ? (
          aiReviews.map((review) => (
            <View key={review.id} className='assignment-detail-page__review-card assignment-detail-page__review-card--ai'>
              <View className='assignment-detail-page__review-header'>
                <Text className='assignment-detail-page__review-icon'>🤖</Text>
                {review.score !== null && (
                  <Text className='assignment-detail-page__review-score'>
                    评分: {review.score}
                  </Text>
                )}
              </View>
              <Text className='assignment-detail-page__review-content'>{review.content}</Text>
            </View>
          ))
        ) : (
          <View
            className={`assignment-detail-page__ai-btn ${requestingAI ? 'assignment-detail-page__ai-btn--disabled' : ''}`}
            onClick={requestingAI ? undefined : handleAIReview}
          >
            <Text className='assignment-detail-page__ai-btn-text'>
              {requestingAI ? 'AI 点评中...' : '请求 AI 点评'}
            </Text>
          </View>
        )}
      </View>

      {/* 教师点评 */}
      <View className='assignment-detail-page__section'>
        <Text className='assignment-detail-page__section-title'>教师点评</Text>
        {teacherReviews.length > 0 ? (
          teacherReviews.map((review) => (
            <View key={review.id} className='assignment-detail-page__review-card assignment-detail-page__review-card--teacher'>
              <View className='assignment-detail-page__review-header'>
                <Text className='assignment-detail-page__review-icon'>👨‍🏫</Text>
                {review.score !== null && (
                  <Text className='assignment-detail-page__review-score'>
                    评分: {review.score}
                  </Text>
                )}
              </View>
              <Text className='assignment-detail-page__review-content'>{review.content}</Text>
            </View>
          ))
        ) : (
          <Text className='assignment-detail-page__empty'>暂无教师点评</Text>
        )}

        {/* 教师写点评（仅教师可见） */}
        {userInfo?.role === 'teacher' && (
          <>
            {!showReviewForm ? (
              <View
                className='assignment-detail-page__write-btn'
                onClick={() => setShowReviewForm(true)}
              >
                <Text className='assignment-detail-page__write-btn-text'>写点评</Text>
              </View>
            ) : (
              <View className='assignment-detail-page__review-form'>
                <Textarea
                  className='assignment-detail-page__review-textarea'
                  placeholder='请输入点评内容...'
                  value={reviewContent}
                  maxlength={2000}
                  onInput={(e) => setReviewContent(e.detail.value)}
                />
                <Input
                  className='assignment-detail-page__review-score-input'
                  placeholder='评分（0-100，可选）'
                  value={reviewScore}
                  type='digit'
                  onInput={(e) => setReviewScore(e.detail.value)}
                />
                <View className='assignment-detail-page__review-actions'>
                  <View
                    className='assignment-detail-page__review-cancel'
                    onClick={() => { setShowReviewForm(false); setReviewContent(''); setReviewScore('') }}
                  >
                    <Text>取消</Text>
                  </View>
                  <View
                    className={`assignment-detail-page__review-submit ${submittingReview ? 'assignment-detail-page__review-submit--disabled' : ''}`}
                    onClick={submittingReview ? undefined : handleSubmitReview}
                  >
                    <Text className='assignment-detail-page__review-submit-text'>
                      {submittingReview ? '提交中...' : '提交点评'}
                    </Text>
                  </View>
                </View>
              </View>
            )}
          </>
        )}
      </View>
    </View>
  )
}
