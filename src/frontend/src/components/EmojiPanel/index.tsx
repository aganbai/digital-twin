import { useState, useEffect } from 'react'
import { View, Text, ScrollView } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { EMOJI_CATEGORIES, RECENT_EMOJI_KEY, MAX_RECENT_EMOJIS } from '@/constants/emoji'
import './index.scss'

interface EmojiPanelProps {
  visible: boolean
  onSelect: (emoji: string) => void
}

export default function EmojiPanel({ visible, onSelect }: EmojiPanelProps) {
  const [activeCategory, setActiveCategory] = useState(0)
  const [recentEmojis, setRecentEmojis] = useState<string[]>([])

  // 加载最近使用的 Emoji
  useEffect(() => {
    if (visible) {
      const stored = Taro.getStorageSync(RECENT_EMOJI_KEY)
      if (stored) {
        try {
          setRecentEmojis(JSON.parse(stored))
        } catch {
          setRecentEmojis([])
        }
      }
    }
  }, [visible])

  const handleSelect = (emoji: string) => {
    onSelect(emoji)

    // 更新最近使用
    const updated = [emoji, ...recentEmojis.filter(e => e !== emoji)].slice(0, MAX_RECENT_EMOJIS)
    setRecentEmojis(updated)
    Taro.setStorageSync(RECENT_EMOJI_KEY, JSON.stringify(updated))
  }

  if (!visible) return null

  const currentEmojis = activeCategory === -1
    ? recentEmojis
    : EMOJI_CATEGORIES[activeCategory]?.emojis || []

  return (
    <View className='emoji-panel'>
      {/* 分类 Tab */}
      <ScrollView className='emoji-panel__tabs' scrollX showScrollbar={false}>
        {recentEmojis.length > 0 && (
          <View
            className={`emoji-panel__tab ${activeCategory === -1 ? 'emoji-panel__tab--active' : ''}`}
            onClick={() => setActiveCategory(-1)}
          >
            <Text className='emoji-panel__tab-text'>🕐</Text>
          </View>
        )}
        {EMOJI_CATEGORIES.map((cat, idx) => (
          <View
            key={cat.name}
            className={`emoji-panel__tab ${activeCategory === idx ? 'emoji-panel__tab--active' : ''}`}
            onClick={() => setActiveCategory(idx)}
          >
            <Text className='emoji-panel__tab-text'>{cat.icon}</Text>
          </View>
        ))}
      </ScrollView>

      {/* Emoji 网格 */}
      <ScrollView className='emoji-panel__grid-wrap' scrollY showScrollbar={false}>
        <View className='emoji-panel__grid'>
          {currentEmojis.map((emoji, idx) => (
            <View
              key={`${emoji}-${idx}`}
              className='emoji-panel__item'
              onClick={() => handleSelect(emoji)}
            >
              <Text className='emoji-panel__emoji'>{emoji}</Text>
            </View>
          ))}
        </View>
      </ScrollView>
    </View>
  )
}
