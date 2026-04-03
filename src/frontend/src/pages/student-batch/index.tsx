import { useState } from 'react'
import { View, Text, Textarea, Input, ScrollView, Picker } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { parseStudentText, batchCreateStudents, ParsedStudent } from '@/api/student-batch'
import TagInput from '@/components/TagInput'
import './index.scss'

// 性别选项
const GENDER_OPTIONS = ['男 (male)', '女 (female)']
const GENDER_MAP: Record<string, string> = { '男 (male)': 'male', '女 (female)': 'female' }
const GENDER_DISPLAY: Record<string, string> = { male: '男', female: '女' }

// 预设性格标签
const PRESET_PERSONALITY_TAGS = [
  '开朗', '内向', '细心', '粗心', '好奇心强', '专注力强',
  '活泼', '安静', '乐于助人', '独立', '敏感', '坚韧',
]

export default function StudentBatchPage() {
  const router = useRouter()
  const personaId = Number(router.params.persona_id) || 0
  const classId = Number(router.params.class_id) || 0

  const [step, setStep] = useState<'input' | 'confirm' | 'result'>('input')
  const [text, setText] = useState('')
  const [parsing, setParsing] = useState(false)
  const [students, setStudents] = useState<ParsedStudent[]>([])
  const [creating, setCreating] = useState(false)
  const [results, setResults] = useState<any[]>([])
  // 记录每个学生卡片的展开状态
  const [expandedIndexes, setExpandedIndexes] = useState<Set<number>>(new Set())

  // 切换展开/收起
  const toggleExpand = (index: number) => {
    setExpandedIndexes(prev => {
      const next = new Set(prev)
      if (next.has(index)) {
        next.delete(index)
      } else {
        next.add(index)
      }
      return next
    })
  }

  // 解析文本
  const handleParse = async () => {
    if (!text.trim()) {
      Taro.showToast({ title: '请输入学生信息', icon: 'none' })
      return
    }
    setParsing(true)
    try {
      const res = await parseStudentText(text.trim())
      const parsed = res.data?.students || []
      if (parsed.length === 0) {
        Taro.showToast({ title: '未能识别到学生信息', icon: 'none' })
        return
      }
      setStudents(parsed)
      setStep('confirm')
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '解析失败', icon: 'none' })
    } finally {
      setParsing(false)
    }
  }

  // 修改学生信息（字符串字段）
  const updateStudent = (index: number, field: string, value: any) => {
    setStudents(prev => {
      const updated = [...prev]
      updated[index] = { ...updated[index], [field]: value }
      return updated
    })
  }

  // 删除学生
  const removeStudent = (index: number) => {
    setStudents(prev => prev.filter((_, i) => i !== index))
    setExpandedIndexes(prev => {
      const next = new Set<number>()
      prev.forEach(i => {
        if (i < index) next.add(i)
        else if (i > index) next.add(i - 1)
      })
      return next
    })
  }

  // 性别选择处理
  const handleGenderChange = (index: number, e: any) => {
    const selected = GENDER_OPTIONS[e.detail.value]
    updateStudent(index, 'gender', GENDER_MAP[selected])
  }

  // 性格标签切换（预设标签点击）
  const togglePersonalityTag = (index: number, tag: string) => {
    const current = students[index].personality_tags || []
    if (current.includes(tag)) {
      updateStudent(index, 'personality_tags', current.filter(t => t !== tag))
    } else {
      updateStudent(index, 'personality_tags', [...current, tag])
    }
  }

  // 校验必填项
  const validateStudents = (): boolean => {
    for (let i = 0; i < students.length; i++) {
      if (!students[i].name?.trim()) {
        Taro.showToast({ title: `第${i + 1}位学生姓名不能为空`, icon: 'none' })
        return false
      }
      if (!students[i].gender) {
        Taro.showToast({ title: `第${i + 1}位学生请选择性别`, icon: 'none' })
        return false
      }
    }
    return true
  }

  // 批量创建
  const handleCreate = async () => {
    if (students.length === 0) {
      Taro.showToast({ title: '没有要创建的学生', icon: 'none' })
      return
    }
    if (!validateStudents()) return
    setCreating(true)
    try {
      const res = await batchCreateStudents({
        persona_id: personaId,
        class_id: classId || undefined,
        students,
      })
      setResults(res.data?.results || [])
      setStep('result')
      Taro.showToast({
        title: `成功${res.data?.success_count}人，失败${res.data?.failed_count}人`,
        icon: res.data?.failed_count === 0 ? 'success' : 'none',
      })
    } catch (e: any) {
      Taro.showToast({ title: e?.message || '创建失败', icon: 'none' })
    } finally {
      setCreating(false)
    }
  }

  return (
    <View className='batch-page'>
      <View className='batch-page__header'>
        <Text className='batch-page__title'>👥 批量添加学生</Text>
        <Text className='batch-page__subtitle'>
          {step === 'input' && '粘贴花名册或学生名单，AI智能识别'}
          {step === 'confirm' && `已识别 ${students.length} 名学生，请确认信息`}
          {step === 'result' && '创建完成'}
        </Text>
      </View>

      {/* 步骤1：输入文本 */}
      {step === 'input' && (
        <View className='batch-page__section'>
          <View className='batch-page__tip'>
            <Text className='batch-page__tip-text'>
              💡 支持多种格式：花名册、逗号分隔、表格等。每行一个学生，可包含姓名、性别、学号等信息。
            </Text>
          </View>
          <Textarea
            className='batch-page__textarea'
            value={text}
            placeholder={'示例：\n张三，男，1001\n李四，女，1002\n王五，男，1003'}
            maxlength={5000}
            onInput={(e) => setText(e.detail.value)}
          />
          <Text className='batch-page__char-count'>{text.length}/5000</Text>
          <View
            className={`batch-page__btn ${(!text.trim() || parsing) ? 'batch-page__btn--disabled' : ''}`}
            onClick={handleParse}
          >
            <Text className='batch-page__btn-text'>
              {parsing ? '🤖 AI识别中...' : '🤖 AI智能识别'}
            </Text>
          </View>
        </View>
      )}

      {/* 步骤2：确认学生信息 */}
      {step === 'confirm' && (
        <View className='batch-page__section'>
          <View className='batch-page__info-tip'>
            <Text className='batch-page__info-tip-text'>
              💡 填写越多学生信息，AI个性化辅导越精准哦！
            </Text>
          </View>
          <ScrollView scrollY className='batch-page__student-list'>
            {students.map((s, idx) => {
              const isExpanded = expandedIndexes.has(idx)
              return (
                <View key={idx} className='batch-page__student-card'>
                  <View className='batch-page__student-header'>
                    <Text className='batch-page__student-no'>#{idx + 1}</Text>
                    <View className='batch-page__student-remove' onClick={() => removeStudent(idx)}>
                      <Text className='batch-page__student-remove-text'>✕</Text>
                    </View>
                  </View>

                  {/* 必填字段：姓名、性别 */}
                  <View className='batch-page__student-fields'>
                    <View className='batch-page__student-field'>
                      <Text className='batch-page__student-field-label batch-page__student-field-label--required'>姓名 *</Text>
                      <Input
                        className='batch-page__student-input'
                        value={s.name}
                        placeholder='必填'
                        onInput={(e) => updateStudent(idx, 'name', e.detail.value)}
                      />
                    </View>
                    <View className='batch-page__student-field'>
                      <Text className='batch-page__student-field-label batch-page__student-field-label--required'>性别 *</Text>
                      <Picker mode='selector' range={GENDER_OPTIONS} onChange={(e) => handleGenderChange(idx, e)}>
                        <View className='batch-page__student-picker'>
                          <Text className={`batch-page__student-picker-text ${!s.gender ? 'batch-page__student-picker-text--placeholder' : ''}`}>
                            {s.gender ? GENDER_DISPLAY[s.gender] || s.gender : '请选择'}
                          </Text>
                          <Text className='batch-page__student-picker-arrow'>▾</Text>
                        </View>
                      </Picker>
                    </View>
                  </View>

                  {/* 展开/收起按钮 */}
                  <View className='batch-page__expand-toggle' onClick={() => toggleExpand(idx)}>
                    <Text className='batch-page__expand-toggle-text'>
                      {isExpanded ? '收起详细信息 ▲' : '展开填写更多信息 ▼'}
                    </Text>
                  </View>

                  {/* 展开的详细信息 */}
                  {isExpanded && (
                    <View className='batch-page__student-detail'>
                      {/* 选填推荐 */}
                      <View className='batch-page__detail-section'>
                        <Text className='batch-page__detail-section-title'>📚 选填推荐</Text>
                        <View className='batch-page__student-fields'>
                          <View className='batch-page__student-field'>
                            <Text className='batch-page__student-field-label'>年龄</Text>
                            <Input
                              className='batch-page__student-input'
                              type='number'
                              value={s.age ? String(s.age) : ''}
                              placeholder='选填'
                              onInput={(e) => updateStudent(idx, 'age', e.detail.value ? Number(e.detail.value) : undefined)}
                            />
                          </View>
                          <View className='batch-page__student-field'>
                            <Text className='batch-page__student-field-label'>学号</Text>
                            <Input
                              className='batch-page__student-input'
                              value={s.student_id || ''}
                              placeholder='选填'
                              onInput={(e) => updateStudent(idx, 'student_id', e.detail.value)}
                            />
                          </View>
                          <View className='batch-page__student-field'>
                            <Text className='batch-page__student-field-label'>擅长学科</Text>
                            <Input
                              className='batch-page__student-input'
                              value={s.strengths || ''}
                              placeholder='如：数学、英语'
                              onInput={(e) => updateStudent(idx, 'strengths', e.detail.value)}
                            />
                          </View>
                          <View className='batch-page__student-field'>
                            <Text className='batch-page__student-field-label'>薄弱学科</Text>
                            <Input
                              className='batch-page__student-input'
                              value={s.weaknesses || ''}
                              placeholder='如：语文、物理'
                              onInput={(e) => updateStudent(idx, 'weaknesses', e.detail.value)}
                            />
                          </View>
                          <View className='batch-page__student-field'>
                            <Text className='batch-page__student-field-label'>学习风格</Text>
                            <Input
                              className='batch-page__student-input'
                              value={s.learning_style || ''}
                              placeholder='如：视觉型、听觉型'
                              onInput={(e) => updateStudent(idx, 'learning_style', e.detail.value)}
                            />
                          </View>
                        </View>
                      </View>

                      {/* 选填个性化 */}
                      <View className='batch-page__detail-section'>
                        <Text className='batch-page__detail-section-title'>🎨 选填个性化</Text>
                        <View className='batch-page__student-fields'>
                          {/* 性格特点：预设标签 + 自定义输入 */}
                          <View className='batch-page__student-field batch-page__student-field--column'>
                            <Text className='batch-page__student-field-label'>性格特点</Text>
                            <View className='batch-page__preset-tags'>
                              {PRESET_PERSONALITY_TAGS.map(tag => (
                                <View
                                  key={tag}
                                  className={`batch-page__preset-tag ${(s.personality_tags || []).includes(tag) ? 'batch-page__preset-tag--active' : ''}`}
                                  onClick={() => togglePersonalityTag(idx, tag)}
                                >
                                  <Text className={`batch-page__preset-tag-text ${(s.personality_tags || []).includes(tag) ? 'batch-page__preset-tag-text--active' : ''}`}>
                                    {tag}
                                  </Text>
                                </View>
                              ))}
                            </View>
                            <TagInput
                              tags={(s.personality_tags || []).filter(t => !PRESET_PERSONALITY_TAGS.includes(t))}
                              onChange={(customTags) => {
                                const presetSelected = (s.personality_tags || []).filter(t => PRESET_PERSONALITY_TAGS.includes(t))
                                updateStudent(idx, 'personality_tags', [...presetSelected, ...customTags])
                              }}
                              placeholder='输入自定义标签后回车'
                              maxTags={5}
                            />
                          </View>
                          <View className='batch-page__student-field'>
                            <Text className='batch-page__student-field-label'>兴趣爱好</Text>
                            <Input
                              className='batch-page__student-input'
                              value={s.interests || ''}
                              placeholder='如：篮球、绘画'
                              onInput={(e) => updateStudent(idx, 'interests', e.detail.value)}
                            />
                          </View>
                          <View className='batch-page__student-field'>
                            <Text className='batch-page__student-field-label'>特长</Text>
                            <Input
                              className='batch-page__student-input'
                              value={s.specialties || ''}
                              placeholder='如：钢琴、编程'
                              onInput={(e) => updateStudent(idx, 'specialties', e.detail.value)}
                            />
                          </View>
                          <View className='batch-page__student-field batch-page__student-field--column'>
                            <Text className='batch-page__student-field-label'>家长备注</Text>
                            <Textarea
                              className='batch-page__student-textarea'
                              value={s.parent_notes || ''}
                              placeholder='如：孩子比较害羞，需要多鼓励'
                              maxlength={500}
                              onInput={(e) => updateStudent(idx, 'parent_notes', e.detail.value)}
                            />
                          </View>
                        </View>
                      </View>
                    </View>
                  )}
                </View>
              )
            })}
          </ScrollView>

          <View className='batch-page__btn-group'>
            <View className='batch-page__btn batch-page__btn--secondary' onClick={() => setStep('input')}>
              <Text className='batch-page__btn-text batch-page__btn-text--secondary'>返回修改</Text>
            </View>
            <View
              className={`batch-page__btn ${(students.length === 0 || creating) ? 'batch-page__btn--disabled' : ''}`}
              onClick={handleCreate}
            >
              <Text className='batch-page__btn-text'>
                {creating ? '创建中...' : `确认创建 ${students.length} 人`}
              </Text>
            </View>
          </View>
        </View>
      )}

      {/* 步骤3：创建结果 */}
      {step === 'result' && (
        <View className='batch-page__section'>
          <View className='batch-page__results'>
            {results.map((r, idx) => (
              <View key={idx} className={`batch-page__result-item batch-page__result-item--${r.status}`}>
                <Text className='batch-page__result-name'>{r.name}</Text>
                <Text className='batch-page__result-status'>
                  {r.status === 'success' ? '✅ 创建成功' : `❌ ${r.error || '创建失败'}`}
                </Text>
              </View>
            ))}
          </View>
          <View className='batch-page__btn' onClick={() => Taro.navigateBack()}>
            <Text className='batch-page__btn-text'>返回</Text>
          </View>
        </View>
      )}
    </View>
  )
}
