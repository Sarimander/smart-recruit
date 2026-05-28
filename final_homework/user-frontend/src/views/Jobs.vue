<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listJobs } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const jobs = ref([])

async function load() {
  loading.value = true
  try {
    const data = await listJobs({ page: 1, page_size: 50 })
    jobs.value = data.jobs || []
  } finally {
    loading.value = false
  }
}

function goDetail(id) {
  router.push(`/jobs/${id}`)
}

function logout() {
  auth.logout()
}

onMounted(load)
</script>

<template>
  <div class="page">
    <header class="header">
      <h2>智能招聘 - 候选人端</h2>
      <div>
        <el-button v-if="!auth.isLoggedIn" @click="$router.push('/login')">登录</el-button>
        <el-button v-if="!auth.isLoggedIn" type="primary" @click="$router.push('/register')">注册</el-button>
        <template v-else>
          <el-button @click="$router.push('/profile')">我的档案</el-button>
          <el-button link @click="logout">退出</el-button>
        </template>
      </div>
    </header>
    <el-table :data="jobs" v-loading="loading" border>
      <el-table-column prop="title" label="岗位" />
      <el-table-column prop="location" label="地点" />
      <el-table-column prop="salary" label="薪资" />
      <el-table-column prop="hr_username" label="HR" />
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row.id)">查看详情</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.page { max-width: 1100px; margin: 0 auto; padding: 24px; }
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
</style>
