import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'

const service: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor — attach JWT token
service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('readbud_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// Response interceptor — handle errors
service.interceptors.response.use(
  (response: AxiosResponse) => response.data,
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        localStorage.removeItem('readbud_token')
        window.location.href = '/login'
        return Promise.reject(error)
      }
      const message = data?.message || '请求失败'
      ElMessage.error(message)
    } else {
      ElMessage.error('网络连接异常')
    }
    return Promise.reject(error)
  },
)

export function get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return service.get(url, config) as Promise<T>
}

export function post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return service.post(url, data, config) as Promise<T>
}

export function patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return service.patch(url, data, config) as Promise<T>
}

export default service
