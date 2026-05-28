<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <el-container style="height: 100vh">
    <el-aside width="220px" style="background: #001529; color: #fff">
      <div style="padding: 20px; font-size: 18px; font-weight: bold">HR 管理系统</div>
      <el-menu background-color="#001529" text-color="#fff" active-text-color="#409eff" router :default-active="$route.path">
        <el-menu-item index="/jobs">岗位管理</el-menu-item>
        <el-menu-item index="/candidates">候选人台账</el-menu-item>
        <el-menu-item index="/ai-chat">AI 对话</el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid #eee">
        <span>欢迎，{{ auth.user?.username }}</span>
        <el-button type="danger" link @click="logout">退出登录</el-button>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>
