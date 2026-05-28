import request from './request'

export const login = (data) => request.post('/auth/login', data)
export const register = (data) => request.post('/auth/register', { ...data, role: 'candidate' })
export const listJobs = (params) => request.get('/jobs', { params })
export const getJob = (id) => request.get(`/jobs/${id}`)
export const getProfile = () => request.get('/user/profile')
export const updateProfile = (data) => request.put('/user/profile', data)
export const getUploadURL = (params) => request.get('/user/resume/upload-url', { params })
export const confirmResume = (data) => request.post('/user/profile/resume', data)
export const apply = (data) => request.post('/user/applications', data)
