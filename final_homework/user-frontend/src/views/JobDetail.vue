<script setup>
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getJob, apply, getProfile } from '../api'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const job = ref(null)
const profile = ref(null)
const applying = ref(false)

async function load() {
  job.value = await getJob(route.params.id)
  if (auth.isLoggedIn) {
    profile.value = await getProfile()
  }
}

async function doApply() {
  if (!auth.isLoggedIn) {
    router.push('/login')
    return
  }
  if (!profile.value?.profile_complete) {
    ElMessage.warning('请先完善档案并上传简历')
    router.push('/profile')
    return
  }
  applying.value = true
  try {
    await apply({ job_id: Number(route.params.id) })
    ElMessage.success('投递成功')
  } finally {
    applying.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="page" v-if="job">
    <el-button link @click="$router.push('/jobs')">返回列表</el-button>
    <el-card style="margin-top: 16px">
      <h2>{{ job.title }}</h2>
      <p>地点：{{ job.location }} | 薪资：{{ job.salary }}</p>
      <p>HR：{{ job.hr_username }}</p>
      <el-divider />
      <div style="white-space: pre-wrap">{{ job.description }}</div>
      <div style="margin-top: 20px">
        <el-button v-if="auth.isLoggedIn" type="primary" :loading="applying" @click="doApply">一键投递</el-button>
        <el-button v-else type="primary" @click="$router.push('/login')">登录后投递</el-button>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.page { max-width: 900px; margin: 0 auto; padding: 24px; }
</style>
