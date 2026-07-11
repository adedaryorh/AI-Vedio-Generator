import { useEffect, useState } from 'react'
import { contentManager } from '@/lib/api'

export default function AnalyticsPage() {
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
    return <div className="p-6">Loading...</div>
  }

  if (error) {
    return <div className="p-6 text-red-600">Error: {error}</div>
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Analytics</h1>
        <button
          onClick={() => {
            // TODO: Refresh
          }}
          className="px-3 py-1 text-sm bg-gray-200 rounded hover:bg-gray-300"
        >
          Refresh
        </button>
      </div>

      {stats ? (
        <div className="space-y-6">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
            <div className="bg-white p-4 rounded-lg shadow">
              <h3 className="text-lg font-medium text-gray-900">Total Stories</h3>
              <p className="mt-2 text-3xl font-bold text-blue-600">
                {stats.total_stories ?? 0}
              </p>
            </div>
            <div className="bg-white p-4 rounded-lg shadow">
              <h3 className="text-lg font-medium text-gray-900">Processed Stories</h3>
              <p className="mt-2 text-3xl font-bold text-green-600">
                {stats.processed_stories ?? 0}
              </p>
            </div>
            <div className="bg-white p-4 rounded-lg shadow">
              <h3 className="text-lg font-medium text-gray-900">Videos Generated</h3>
              <p className="mt-2 text-3xl font-bold text-purple-600">
                {stats.total_videos ?? 0}
              </p>
            </div>
            <div className="bg-white p-4 rounded-lg shadow">
              <h3 className="text-lg font-medium text-gray-900">Videos Posted</h3>
              <p className="mt-2 text-3xl font-bold text-indigo-600">
                {stats.posted_videos ?? 0}
              </p>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Platform Distribution</h2>
            {/* In a real app, we would render a pie chart or bar chart */}
            <div className="h-96 bg-gray-100 rounded-lg flex items-center justify-center">
              <span className="text-gray-500">
                Chart Placeholder (e.g., using Recharts or Chart.js)
              </span>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <h2 className="text-xl font-semibold mb-4">Recent Activity</h2>
            <div className="space-y-4">
              {/* Mock data */}
              {[1, 2, 3].map((i) => (
                <div key={i} className="p-3 border-l-4 border-blue-500 bg-gray-50">
                  <p className="font-medium">Story "{'Sample Story ' + i}" processed</p>
                  <p className="text-sm text-gray-500">
                    {new Date(Date.now() - i * 3600000).toLocaleString()}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : (
        <p className="text-center text-gray-500">No statistics available</p>
      )}
    </div>
  )
}