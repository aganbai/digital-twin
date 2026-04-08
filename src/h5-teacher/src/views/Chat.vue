<template>
  <div class="chat-container">
    <div class="chat-header">
      <span>{{ chatInfo.student_name }}</span>
    </div>
    <div class="chat-messages" ref="messagesRef">
      <div v-for="msg in messages" :key="msg.id" class="message" :class="{ 'is-mine': msg.is_mine }">
        <div class="message-content">{{ msg.content }}</div>
        <div class="message-time">{{ msg.created_at }}</div>
      </div>
    </div>
    <div class="chat-input">
      <el-input v-model="inputText" placeholder="输入消息..." @keyup.enter="sendMessage" />
      <el-button type="primary" @click="sendMessage">发送</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const messagesRef = ref<HTMLElement>()
const inputText = ref('')
const chatInfo = ref({ student_name: '学生' })
const messages = ref<any[]>([])

function sendMessage() {
  if (!inputText.value.trim()) return
  // TODO: 发送消息
  inputText.value = ''
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
}

onMounted(() => {
  // TODO: 加载对话数据
})
</script>

<style scoped lang="scss">
.chat-container {
  height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
  background: #fff;
  border-radius: 8px;
  .chat-header {
    padding: 15px 20px;
    border-bottom: 1px solid #ebeef5;
    font-weight: bold;
  }
  .chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
    .message {
      margin-bottom: 15px;
      &.is-mine { text-align: right; .message-content { background: #409eff; color: #fff; } }
      .message-content {
        display: inline-block;
        padding: 10px 15px;
        border-radius: 8px;
        background: #f5f7fa;
        max-width: 70%;
        word-break: break-word;
      }
      .message-time { font-size: 12px; color: #909399; margin-top: 5px; }
    }
  }
  .chat-input {
    padding: 15px;
    border-top: 1px solid #ebeef5;
    display: flex;
    gap: 10px;
  }
}
</style>
