import axios from 'axios'

// Create axios instance with base URL from environment
const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8000', // This is a gateway? We'll use individual services
})

// We'll create separate instances for each service
const storyCollector = axios.create({
  baseURL: process.env.NEXT_PUBLIC_STORY_COLLECTOR_URL || 'http://localhost:8001',
})

const videoGenerator = axios.create({
  baseURL: process.env.NEXT_PUBLIC_VIDEO_GENERATOR_URL || 'http://localhost:8002',
})

const contentManager = axios.create({
  baseURL: process.env.NEXT_PUBLIC_CONTENT_MANAGER_URL || 'http://localhost:9001',
})

const socialMediaBot = axios.create({
  baseURL: process.env.NEXT_PUBLIC_SOCIAL_MEDIA_BOT_URL || 'http://localhost:9002',
})

// Request interceptor to add auth token from cookie
const addAuthToken = (instance: any) => {
  instance.interceptors.request.use((config: any) => {
    // Get token from cookie
    const token = document.cookie
      .split('; ')
      .find((row) => row.startsWith('token='))
      ?.split('=')[1]
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  }, (error: any) => Promise.reject(error))
}

// Response interceptor for error handling
const handleResponseError = (instance: any) => {
  instance.interceptors.response.use(
    (response: any) => response,
    (error: any) => {
      // Handle 401 Unauthorized
      if (error.response?.status === 401) {
        // Redirect to login
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }
  )
}

// Apply interceptors to all service instances
[storyCollector, videoGenerator, contentManager, socialMediaBot].forEach((instance) => {
  addAuthToken(instance)
  handleResponseError(instance)
})

// Export the instances
export {
  storyCollector,
  videoGenerator,
  contentManager,
  socialMediaBot,
  // For convenience, also export a generic api if needed
  api as default
}