import request from './request'

export const login = (data) => request.post('/auth/login', data)
export const register = (data) => request.post('/auth/register', { ...data, role: 'hr' })
export const listHRJobs = (params) => request.get('/hr/jobs', { params })
export const createJob = (data) => request.post('/hr/jobs', data)
export const updateJob = (id, data) => request.put(`/hr/jobs/${id}`, data)
export const deleteJob = (id) => request.delete(`/hr/jobs/${id}`)
export const listCandidates = (params) => request.get('/hr/candidates', { params })
export const getDownloadURL = (params) => request.get('/hr/resume/download-url', { params })
export const chat = (data) => request.post('/hr/ai/chat', data)
export const getChatHistory = () => request.get('/hr/ai/history')
