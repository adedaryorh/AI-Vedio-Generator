import { useEffect, useState } from 'react'
import { contentManager } from '@/lib/api'

export default function Dashboard() {
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true)
        const response = await contentManager.get('/stats')
        setStats(response.data)
      } catch (err: any) {
        setError(err.response?.data?.detail || 'Failed to fetch stats')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchStats()
  }, [])

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Total Stories</h3>
            <p className="mt-2 text-3xl font-bold text-blue-600">Loading...</p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Videos Generated</h3>
            <p className="mt-2 text-3xl font-bold text-green-600">Loading...</p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Posts Published</h3>
            <p className="mt-2 text-3xl font-bold text-purple-600">Loading...</p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">System Health</h3>
            <div className="mt-2 flex items-center">
              <div className="h-2.5 w-2.5 bg-yellow-500 rounded-full mr-2"></div>
              <span className="text-sm text-gray-600">Checking...</span>
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Total Stories</h3>
            <p className="mt-2 text-3xl font-bold text-blue-600">Error</p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Videos Generated</h3>
            <p className="mt-2 text-3xl font-bold text-green-600">Error</p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Posts Published</h3>
            <p className="mt-2 text-3xl font-bold text-purple-600">Error</p>
          </div>
          <div className="bg-white p-4 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">System Health</h3>
            <div className="mt-2 flex items-center">
              <div className="h-2.5 w-2.5 bg-red-500 rounded-full mr-2"></div>
              <span className="text-sm text-gray-600">Error</span>
            </div>
          </div>
        </div>
        <div className="mt-4 p-4 bg-red-50 border-l-4 border-red-500">
          <p className="text-sm text-red-600">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900">Total Stories</h3>
          <p className="mt-2 text-3xl font-bold text-blue-600">
            {stats.total_stories ?? 0}
          </p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900">Videos Generated</h3>
          <p className="mt-2 text-3xl font-bold text-green-600">
            {stats.total_videos ?? 0}
          </p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900">Posts Published</h3>
          <p className="mt-2 text-3xl font-bold text-purple-600">
            {stats.posted_videos ?? 0}
          </p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900">System Health</h3>
          <div className="mt-2 flex items-center">
            {/* We'll implement a proper health check later */}
            <div className="h-2.5 w-2.5 bg-green-500 rounded-full mr-2"></div>
            <span className="text-sm text-gray-600">All Systems Operational</span>
          </div>
        </div>
      </div>
    </div>
  )
}