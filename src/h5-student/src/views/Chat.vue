<template>
  <div class="chat-page">
    <van-nav-bar title="AI助手" />
    <div class="chat-messages" ref="messagesRef">
      <div v-for="msg in messages" :key="msg.id" class="message" :class="{ 'is-mine': msg.is_mine }">
        <div class="message-avatar">
          <van-image v-if="msg.avatar" :src="msg.avatar" round width="36" height="36" />
          <van-icon v-else :name="msg.is_mine ? 'user-o' : 'service-o'" size="20" />
        </div>
        <div class="message-content">
          <div class="message-text">{{ msg.content }}</div>
          <div class="message-time">{{ msg.time }}</div>
        </div>
      </div>
    </div>
    <div class="chat-input">
      <van-field v-model="inputText" placeholder="输入消息..." @keyup.enter="sendMessage">
        <template #button>
          <van-button size="small" type="primary" @click="sendMessage">发送</van-button>
        </template>
      </van-field>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { showToast } from 'vant'

const messagesRef = ref<HTMLElement>()
const inputText = ref('')
const messages = ref<any[]>([
  { id: 1, content: '你好，我是你的AI助手，有什么可以帮助你的吗？', is_mine: false, time: '09:00' }
])

function sendMessage() {
  if (!inputText.value.trim()) {
    showToast('请输入消息')
    return
  }
  messages.value.push({ id: Date.now(), content: inputText.value, is_mine: true, time: new Date().toLocaleTimeString().slice(0, 5) })
  inputText.value = ''
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
  // TODO: 调用API获取回复
}

onMounted(() => {
  // TODO: 加载历史消息
})
</script>

<style scoped lang="scss">
.chat-page {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  .chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 10px;
    background: #f7f8fa;
    .message {
      display: flex;
      margin-bottom: 15px;
      &.is-mine { flex-direction: row-reverse; .message-content { align-items: flex-end; } .message-text { background: #07c160; color: #fff; } }
      .message-avatar { flex-shrink: 0; width: 36px; height: 36px; margin: 0 10px; display: flex; align-items: center; justify-content: center; background: #fff; border-radius: 50%; }
      .message-content { display: flex; flex-direction: column; max-width: 70%; .message-text { padding: 10px 15px; border-radius: 8px; background: #fff; word-break: break-word; } .message-time { font-size: 12px; color: #999; margin-top: 5px; } }
    }
  }
  .chat-input { background: #fff; padding: 10px; border-top: 1px solid #eee; }
}
</style>
