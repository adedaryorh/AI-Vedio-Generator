import { useEffect, useState } from 'react'
import { contentManager } from '@/lib/api'
import Link from 'next/link'

export default function SettingsPage() {
  const [config, setConfig] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [formData, setFormData] = useState({
    username: 'admin',
    email: 'admin@example.com',
    emailNotifications: true,
  })
  const [theme, setTheme] = useState('light')

  useEffect(() => {
    const fetchConfig = async () => {
      try {
        setLoading(true)
        const response = await contentManager.get('/config')
        setConfig(response.data)
      } catch (err: any) {
        setError(err.response?.data?.detail || 'Failed to fetch configuration')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchConfig()
  }, [])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // TODO: Submit form to update settings
    alert('Settings saved')
  }

  if (loading) {
    return <div className="p-6">Loading...</div>
  }

  if (error) {
    return <div className="p-6 text-red-600">Error: {error}</div>
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold mb-6">Settings</h1>
        <Link href="/">
          <button className="px-3 py-1 text-sm bg-gray-200 rounded hover:bg-gray-300">
            Back to Dashboard
          </button>
        </Link>
      </div>

      <div className="space-y-6">
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Account Settings</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Username</label>
              <input
                type="text"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">Email</label>
              <input
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div className="flex items-center space-x-4">
              <label className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  checked={formData.emailNotifications}
                  onChange={(e) => setFormData({ ...formData, emailNotifications: e.target.checked })}
                  className="h-4 w-4 text-indigo-600"
                />
                <span>Email notifications</span>
              </label>
            </div>
            <button
              type="submit"
              className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Save Changes
            </button>
          </form>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Appearance</h2>
          <div className="space-y-4">
            <div className="flex items-center space-x-3">
              <label className="flex items-center space-x-2">
                <input
                  type="radio"
                  name="theme"
                  value="light"
                  checked={theme === 'light'}
                  onChange={(e) => setTheme(e.target.value)}
                  className="h-4 w-4 text-indigo-600"
                />
                <span>Light</span>
              </label>
              <label className="flex items-center space-x-2">
                <input
                  type="radio"
                  name="theme"
                  value="dark"
                  checked={theme === 'dark'}
                  onChange={(e) => setTheme(e.target.value)}
                  className="h-4 w-4 text-indigo-600"
                />
                <span>Dark</span>
              </label>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">API Configuration</h2>
          {config ? (
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span>Story Collector URL:</span>
                <span className="font-mono">{config.story_collector_url}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Video Generator URL:</span>
                <span className="font-mono">{config.video_generator_url}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Content Manager URL:</span>
                <span className="font-mono">{config.content_manager_url}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Social Media Bot URL:</span>
                <span className="font-mono">{config.social_media_bot_url}</span>
              </div>
            </div>
          ) : (
            <p className="text-sm text-gray-500">
              These values are loaded from environment variables and cannot be changed here.
            </p>
            <div className="space-y-2 mt-4">
              <div className="flex justify-between text-sm">
                <span>Story Collector URL:</span>
                <span className="font-mono">{process.env.NEXT_PUBLIC_STORY_COLLECTOR_URL || 'http://localhost:8001'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Video Generator URL:</span>
                <span className="font-mono">{process.env.NEXT_PUBLIC_VIDEO_GENERATOR_URL || 'http://localhost:8002'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Content Manager URL:</span>
                <span className="font-mono">{process.env.NEXT_PUBLIC_CONTENT_MANAGER_URL || 'http://localhost:9001'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Social Media Bot URL:</span>
                <span className="font-mono">{process.env.NEXT_PUBLIC_SOCIAL_MEDIA_BOT_URL || 'http://localhost:9002'}</span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}