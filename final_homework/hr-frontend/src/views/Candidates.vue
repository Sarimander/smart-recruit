<script setup>
import { onMounted, reactive, ref } from 'vue'
import { listCandidates, getDownloadURL } from '../api'

const loading = ref(false)
const applications = ref([])
const pageInfo = reactive({ page: 1, page_size: 10, total: 0 })
const jobId = ref('')

async function load() {
  loading.value = true
  try {
    const params = { page: pageInfo.page, page_size: pageInfo.page_size }
    if (jobId.value) params.job_id = jobId.value
    const data = await listCandidates(params)
    applications.value = data.applications || []
    Object.assign(pageInfo, data.page_info || {})
  } finally {
    loading.value = false
  }
}

async function downloadResume(row) {
  const data = await getDownloadURL({
    candidate_id: row.candidate_id,
    oss_key: row.candidate?.resume_oss_key,
  })
  window.open(data.download_url, '_blank')
}

onMounted(load)
</script>

<template>
  <div>
    <div style="margin-bottom: 16px; display: flex; gap: 12px">
      <el-input v-model="jobId" placeholder="按岗位 ID 筛选" style="width: 200px" />
      <el-button type="primary" @click="load">查询</el-button>
    </div>
    <el-table :data="applications" v-loading="loading" border>
      <el-table-column prop="job_title" label="岗位" />
      <el-table-column label="姓名"><template #default="{ row }">{{ row.candidate?.name }}</template></el-table-column>
      <el-table-column label="手机"><template #default="{ row }">{{ row.candidate?.phone }}</template></el-table-column>
      <el-table-column label="学历"><template #default="{ row }">{{ row.candidate?.education }}</template></el-table-column>
      <el-table-column label="学校"><template #default="{ row }">{{ row.candidate?.school }}</template></el-table-column>
      <el-table-column label="技能"><template #default="{ row }">{{ row.candidate?.skills }}</template></el-table-column>
      <el-table-column prop="applied_at" label="投递时间" />
      <el-table-column label="简历" width="100">
        <template #default="{ row }">
          <el-button v-if="row.candidate?.resume_oss_key" link type="primary" @click="downloadResume(row)">下载</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
