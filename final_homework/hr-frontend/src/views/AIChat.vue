<script setup>
import { nextTick, onMounted, ref } from 'vue'
import { chat, getChatHistory } from '../api'

const messages = ref([])
const input = ref('')
const loading = ref(false)
const boxRef = ref(null)

async function loadHistory() {
  const data = await getChatHistory()
  messages.value = data.messages || []
  scrollBottom()
}

async function send() {
  const text = input.value.trim()
  if (!text || loading.value) return
  input.value = ''
  messages.value.push({ role: 'user', content: text })
  scrollBottom()
  loading.value = true
  try {
    const data = await chat({ message: text })
    messages.value.push({ role: 'assistant', content: data.reply })
    scrollBottom()
  } finally {
    loading.value = false
  }
}

function scrollBottom() {
  nextTick(() => {
    if (boxRef.value) boxRef.value.scrollTop = boxRef.value.scrollHeight
  })
}

onMounted(loadHistory)
</script>

<template>
  <div class="chat-page">
    <div ref="boxRef" class="messages">
      <div v-for="(msg, idx) in messages" :key="idx" :class="['msg', msg.role]">
        <div class="bubble">{{ msg.content }}</div>
      </div>
      <div v-if="loading" class="msg assistant"><div class="bubble">思考中...</div></div>
    </div>
    <div class="input-bar">
      <el-input v-model="input" placeholder="例如：我的岗位总共有多少投递？热门岗位排行？" @keyup.enter="send" />
      <el-button type="primary" :loading="loading" @click="send">发送</el-button>
    </div>
  </div>
</template>

<style scoped>
.chat-page { display: flex; flex-direction: column; height: calc(100vh - 120px); }
.messages { flex: 1; overflow-y: auto; padding: 16px; background: #fafafa; }
.msg { display: flex; margin-bottom: 12px; }
.msg.user { justify-content: flex-end; }
.bubble { max-width: 70%; padding: 10px 14px; border-radius: 10px; white-space: pre-wrap; line-height: 1.6; }
.user .bubble { background: #409eff; color: #fff; }
.assistant .bubble { background: #fff; border: 1px solid #eee; }
.input-bar { display: flex; gap: 12px; padding-top: 12px; }
</style>
