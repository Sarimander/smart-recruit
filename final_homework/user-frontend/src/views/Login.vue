<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const form = reactive({ username: '', password: '' })

async function submit() {
  loading.value = true
  try {
    const data = await login(form)
    if (data.role !== 'candidate') {
      ElMessage.error('请使用候选人账号登录')
      return
    }
    auth.setAuth(data)
    ElMessage.success('登录成功')
    router.push('/jobs')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="page">
    <el-card class="card">
      <h2>候选人登录</h2>
      <el-form label-width="80px">
        <el-form-item label="用户名"><el-input v-model="form.username" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password /></el-form-item>
        <el-button type="primary" :loading="loading" @click="submit">登录</el-button>
        <el-button link @click="$router.push('/register')">去注册</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.page { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: #f5f7fa; }
.card { width: 420px; }
</style>
