<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listHRJobs, createJob, updateJob, deleteJob } from '../api'

const loading = ref(false)
const jobs = ref([])
const pageInfo = reactive({ page: 1, page_size: 10, total: 0 })
const dialogVisible = ref(false)
const editing = ref(null)
const form = reactive({ title: '', description: '', salary: '', location: '', status: 'active' })

async function load() {
  loading.value = true
  try {
    const data = await listHRJobs({ page: pageInfo.page, page_size: pageInfo.page_size })
    jobs.value = data.jobs || []
    Object.assign(pageInfo, data.page_info || {})
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { title: '', description: '', salary: '', location: '', status: 'active' })
  dialogVisible.value = true
}

function openEdit(row) {
  editing.value = row
  Object.assign(form, row)
  dialogVisible.value = true
}

async function save() {
  if (editing.value) {
    await updateJob(editing.value.id, form)
    ElMessage.success('更新成功')
  } else {
    await createJob(form)
    ElMessage.success('创建成功')
  }
  dialogVisible.value = false
  load()
}

async function remove(row) {
  await ElMessageBox.confirm('确认删除该岗位？', '提示')
  await deleteJob(row.id)
  ElMessage.success('删除成功')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div style="margin-bottom: 16px">
      <el-button type="primary" @click="openCreate">新增岗位</el-button>
    </div>
    <el-table :data="jobs" v-loading="loading" border>
      <el-table-column prop="title" label="岗位名称" />
      <el-table-column prop="location" label="地点" />
      <el-table-column prop="salary" label="薪资" />
      <el-table-column prop="status" label="状态" />
      <el-table-column prop="created_at" label="创建时间" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑岗位' : '新增岗位'" width="600px">
      <el-form label-width="80px">
        <el-form-item label="名称"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="地点"><el-input v-model="form.location" /></el-form-item>
        <el-form-item label="薪资"><el-input v-model="form.salary" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="form.status"><el-option label="招聘中" value="active" /><el-option label="已关闭" value="inactive" /></el-select></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" rows="4" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
