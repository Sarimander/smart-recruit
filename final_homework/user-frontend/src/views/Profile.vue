<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { getProfile, updateProfile, getUploadURL, confirmResume } from '../api'

const loading = ref(false)
const uploading = ref(false)
const form = reactive({
  name: '', phone: '', education: '', school: '', experience: '', skills: '', resume_oss_key: '', profile_complete: false,
})

async function load() {
  loading.value = true
  try {
    const data = await getProfile()
    Object.assign(form, data || {})
  } finally {
    loading.value = false
  }
}

async function save() {
  await updateProfile(form)
  ElMessage.success('档案已保存')
  load()
}

function validateFile(file) {
  const ext = file.name.split('.').pop()?.toLowerCase()
  if (!['pdf', 'doc', 'docx'].includes(ext)) {
    ElMessage.error('仅支持 PDF、DOC、DOCX')
    return false
  }
  return true
}

async function uploadResume(option) {
  const file = option.file
  if (!validateFile(file)) return
  uploading.value = true
  try {
    const { upload_url, oss_key } = await getUploadURL({ filename: file.name })
    await fetch(upload_url, {
      method: 'PUT',
      body: file,
      headers: { 'Content-Type': file.type || 'application/octet-stream' },
    })
    await confirmResume({ oss_key })
    ElMessage.success('简历上传成功')
    load()
  } finally {
    uploading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="page">
    <el-card>
      <h2>我的档案</h2>
      <el-alert v-if="!form.profile_complete" title="档案未完善，完善并上传简历后才能投递岗位" type="warning" show-icon style="margin-bottom: 16px" />
      <el-form v-loading="loading" label-width="100px">
        <el-form-item label="姓名"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="手机"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="最高学历"><el-input v-model="form.education" /></el-form-item>
        <el-form-item label="学校"><el-input v-model="form.school" /></el-form-item>
        <el-form-item label="经历"><el-input v-model="form.experience" type="textarea" rows="4" /></el-form-item>
        <el-form-item label="核心技能"><el-input v-model="form.skills" type="textarea" rows="3" /></el-form-item>
        <el-form-item label="简历">
          <div>
            <el-upload :show-file-list="false" :http-request="uploadResume" accept=".pdf,.doc,.docx">
              <el-button type="primary" :loading="uploading">上传简历到 OSS</el-button>
            </el-upload>
            <div v-if="form.resume_oss_key" style="margin-top: 8px; color: #67c23a">已上传：{{ form.resume_oss_key }}</div>
          </div>
        </el-form-item>
        <el-button type="primary" @click="save">保存档案</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.page { max-width: 900px; margin: 0 auto; padding: 24px; }
</style>
