import { useState, useEffect } from 'react'
import { View, Text, Input, Textarea, Image } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { approveJoinRequest, rejectJoinRequest } from '@/api/class'
import './index.scss'

/** 性别选项 */
const GENDER_OPTIONS = ['男', '女', '其他']

export default function ApprovalDetail() {
  const router = useRouter()
  const requestId = Number(router.params.id) || 0
  const nickname = decodeURIComponent(router.params.nickname || '学生')
  const avatar = decodeURIComponent(router.params.avatar || '')

  // 教师可编辑字段
  const [age, setAge] = useState('')
  const [gender, setGender] = useState('')
  const [familyInfo, setFamilyInfo] = useState('')
  const [teacherEvaluation, setTeacherEvaluation] = useState('')
  const [submitting, setSubmitting] = useState(false)

  /** 设置导航栏标题 */
  useEffect(() => {
    Taro.setNavigationBarTitle({ title: `${nickname} 的审批` })
  }, [nickname])

  /** 审批通过 */
  const handleApprove = async () => {
    if (submitting) return
    setSubmitting(true)
    try {
      await approveJoinRequest(requestId, {
        teacher_evaluation: teacherEvaluation.trim() || undefined,
        age: age ? Number(age) : undefined,
        gender: gender || undefined,
        family_info: familyInfo.trim() || undefined,
      })
      Taro.showToast({ title: '已通过', icon: 'success' })
      setTimeout(() => {
        Taro.navigateBack()
      }, 1500)
    } catch (error) {
      console.error('审批通过失败:', error)
      Taro.showToast({ title: '操作失败', icon: 'none' })
    } finally {
      setSubmitting(false)
    }
  }

  /** 拒绝 */
  const handleReject = async () => {
    if (submitting) return

    Taro.showModal({
      title: '确认拒绝',
      content: `确定拒绝 ${nickname} 的加入申请吗？`,
      confirmText: '拒绝',
      confirmColor: '#ef4444',
      success: async (res) => {
        if (res.confirm) {
          setSubmitting(true)
          try {
            await rejectJoinRequest(requestId)
            Taro.showToast({ title: '已拒绝', icon: 'success' })
            setTimeout(() => {
              Taro.navigateBack()
            }, 1500)
          } catch (error) {
            console.error('拒绝失败:', error)
            Taro.showToast({ title: '操作失败', icon: 'none' })
          } finally {
            setSubmitting(false)
          }
        }
      },
    })
  }

  return (
    <View className='approval-detail'>
      {/* 学生信息头部 */}
      <View className='approval-detail__header'>
        {avatar ? (
          <Image
            className='approval-detail__avatar'
            src={avatar}
            mode='aspectFill'
          />
        ) : (
          <View className='approval-detail__avatar approval-detail__avatar--default'>
            <Text className='approval-detail__avatar-text'>
              {nickname.charAt(0)}
            </Text>
          </View>
        )}
        <Text className='approval-detail__name'>{nickname}</Text>
      </View>

      {/* 学生信息（教师可修改/补充） */}
      <View className='approval-detail__section'>
        <Text className='approval-detail__section-title'>学生信息</Text>
        <Text className='approval-detail__section-hint'>
          以下信息可由教师修改或补充
        </Text>

        {/* 年龄 */}
        <View className='approval-detail__field'>
          <Text className='approval-detail__field-label'>年龄</Text>
          <Input
            className='approval-detail__input'
            placeholder='请输入年龄'
            placeholderClass='approval-detail__input-placeholder'
            type='number'
            maxlength={3}
            value={age}
            onInput={(e) => setAge(e.detail.value)}
          />
        </View>

        {/* 性别 */}
        <View className='approval-detail__field'>
          <Text className='approval-detail__field-label'>性别</Text>
          <View className='approval-detail__tag-grid'>
            {GENDER_OPTIONS.map((item) => (
              <View
                key={item}
                className={`approval-detail__tag ${gender === item ? 'approval-detail__tag--active' : ''}`}
                onClick={() => setGender(gender === item ? '' : item)}
              >
                <Text
                  className={`approval-detail__tag-text ${gender === item ? 'approval-detail__tag-text--active' : ''}`}
                >
                  {item}
                </Text>
              </View>
            ))}
          </View>
        </View>

        {/* 家庭情况 */}
        <View className='approval-detail__field'>
          <Text className='approval-detail__field-label'>家庭情况</Text>
          <Textarea
            className='approval-detail__textarea'
            placeholder='请输入家庭情况（可选）'
            maxlength={500}
            value={familyInfo}
            onInput={(e) => setFamilyInfo(e.detail.value)}
          />
        </View>
      </View>

      {/* 教师评价/特点 */}
      <View className='approval-detail__section'>
        <Text className='approval-detail__section-title'>评价/特点</Text>
        <Textarea
          className='approval-detail__textarea'
          placeholder='请输入对该学生的评价或特点描述（可选）'
          maxlength={1000}
          value={teacherEvaluation}
          onInput={(e) => setTeacherEvaluation(e.detail.value)}
        />
      </View>

      {/* 操作按钮 */}
      <View className='approval-detail__actions'>
        <View
          className={`approval-detail__btn approval-detail__btn--reject ${submitting ? 'approval-detail__btn--disabled' : ''}`}
          onClick={submitting ? undefined : handleReject}
        >
          <Text className='approval-detail__btn-text approval-detail__btn-text--reject'>
            拒绝
          </Text>
        </View>
        <View
          className={`approval-detail__btn approval-detail__btn--approve ${submitting ? 'approval-detail__btn--disabled' : ''}`}
          onClick={submitting ? undefined : handleApprove}
        >
          <Text className='approval-detail__btn-text approval-detail__btn-text--approve'>
            {submitting ? '处理中...' : '通过'}
          </Text>
        </View>
      </View>
    </View>
  )
}
