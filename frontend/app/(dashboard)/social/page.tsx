import { useState } from 'react'
import { useEffect } from 'react'
import { contentManager } from '@/lib/api'

export default function SocialPage() {
  const [tab, setTab] = useState<'compose' | 'scheduled' | 'posted' | 'analytics'>(
    'compose'
  )

  // Mock data for demo
  const videos = [
    { id: 1, title: 'The Tortoise and the Hare', status: 'ready' },
    { id: 2, title: 'Why the Sun and Moon Live in the Sky', status: 'ready' },
  ]

  const scheduledPosts = [
    {
      id: 1,
      videoId: 1,
      platform: 'instagram',
      scheduleTime: new Date(Date.now() + 3600000).toISOString(),
    },
  ]

  const postedPosts = [
    {
      id: 1,
      videoId: 1,
      platform: 'instagram',
      platformId: 'ig_12345',
      postedAt: new Date(Date.now() - 86400000).toISOString(),
      engagement: { likes: 124, comments: 18, shares: 7, saves: 23 },
    },
  ]

  // Analytics data
  const [analyticsStats, setAnalyticsStats] = useState<any>(null)
  const [analyticsLoading, setAnalyticsLoading] = useState(true)
  const [analyticsError, setAnalyticsError] = useState<string | null>(null)

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        setAnalyticsLoading(true)
        const response = await contentManager.get('/stats')
        setAnalyticsStats(response.data)
      } catch (err: any) {
        setAnalyticsError(err.response?.data?.detail || 'Failed to fetch analytics')
        console.error(err)
      } finally {
        setAnalyticsLoading(false)
      }
    }

    fetchAnalytics()
  }, [])

  const handleTabChange = (t: typeof tab) => {
    setTab(t)
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Social Media</h1>
      </div>

      {/* Tabs */}
      <div className="flex space-x-4 mb-6 border-b pb-2">
        <button
          onClick={() => setTab('compose')}
          className={`px-3 py-2 text-sm font-medium ${
            tab === 'compose'
              ? 'border-b-2 border-primary'
              : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          Compose
        </button>
        <button
          onClick={() => setTab('scheduled')}
          className={`px-3 py-2 text-sm font-medium ${
            tab === 'scheduled'
              ? 'border-b-2 border-primary'
              : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          Scheduled
        </button>
        <button
          onClick={() => setTab('posted')}
          className={`px-3 py-2 text-sm font-medium ${
            tab === 'posted'
              ? 'border-b-2 border-primary'
              : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          Posted
        </button>
        <button
          onClick={() => setTab('analytics')}
          className={`px-3 py-2 text-sm font-medium ${
            tab === 'analytics'
              ? 'border-b-2 border-primary'
              : 'text-gray-500 hover:text-gray-700'
          }`}
        >
          Analytics
        </button>
      </div>

      {/* Tab Content */}
      {tab === 'compose' && (
        <ComposeTab videos={videos} />
      )}
      {tab === 'scheduled' && (
        <ScheduledTab scheduledPosts={scheduledPosts} />
      )}
      {tab === 'posted' && (
        <PostedTab postedPosts={postedPosts} />
      )}
      {tab === 'analytics' && (
        <AnalyticsTab
          stats={analyticsStats}
          loading={analyticsLoading}
          error={analyticsError}
        />
      )}
    </div>
  )
}

// Compose Tab
function ComposeTab({ videos }: { videos: Array<{ id: number; title: string; status: string }> }) {
  const [selectedVideoId, setSelectedVideoId] = useState<number | null>(null)
  const [platforms, setPlatforms] = useState<string[]>([])
  const [caption, setCaption] = useState('')
  const [hashtags, setHashtags] = useState<string[]>([])

  const handlePost = async () => {
    // TODO: Call social-media-bot API to post
    alert('Posting...')
  }

  const handleSchedule = async () => {
    // TODO: Schedule post
    alert('Scheduling...')
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-semibold mb-2">Create New Post</h2>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Select Video</label>
            <select
              value={selectedVideoId || ''}
              onChange={(e) => {
                const val = e.target.value
                setSelectedVideoId(val === '' ? null : parseInt(val))
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Select a video</option>
              {videos
                .filter((v) => v.status === 'ready')
                .map((v) => (
                  <option key={v.id} value={v.id}>
                    {v.title}
                  </option>
                ))}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Platforms</label>
            <div className="flex space-x-2 flex-wrap">
              {['instagram', 'youtube', 'tiktok'].map((platform) => (
                <label key={platform} className="flex items-center">
                  <input
                    type="checkbox"
                    checked={platforms.includes(platform)}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setPlatforms([...platforms, platform])
                      } else {
                        setPlatforms(platforms.filter((p) => p !== platform))
                      }
                    }}
                    className="h-4 w-4 text-blue-600"
                  />
                  <span className="ml-1">{platform}</span>
                </label>
              ))}
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Caption</label>
            <textarea
              value={caption}
              onChange={(e) => setCaption(e.target.value)}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter caption..."
            />
            <div className="mt-1 text-xs text-gray-500">
              {caption.length}/2200 characters
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Hashtags (comma-separated)</label>
            <input
              type="text"
              value={hashtags.join(', ')}
              onChange={(e) => {
                const tags = e.target.value
                  .split(',')
                  .map((t) => t.trim())
                  .filter((t) => t.length > 0)
                setHashtags(tags)
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="#travel #adventure"
            />
          </div>
          <div className="flex justify-end space-x-3">
            <button
              onClick={handleSchedule}
              disabled={!selectedVideoId || platforms.length === 0}
              className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
            >
              Schedule
            </button>
            <button
              onClick={handlePost}
              disabled={!selectedVideoId || platforms.length === 0}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Post Now
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

// Scheduled Tab
function ScheduledTab({ scheduledPosts }: { scheduledPosts: any[] }) {
  return (
    <div>
      <h2 className="text-xl font-semibold mb-2">Scheduled Posts</h2>
      {scheduledPosts.length === 0 ? (
        <p className="text-gray-500">No scheduled posts</p>
      ) : (
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Video
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Platform
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Scheduled Time
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {scheduledPosts.map((post) => (
              <tr key={post.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="font-medium">{post.videoId}</span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {post.platform}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {new Date(post.scheduleTime).toLocaleString()}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                  <button
                    onClick={() => {
                      // TODO: Cancel scheduled post
                      alert(`Cancel scheduled post ${post.id}`)
                    }}
                    className="text-red-600 hover:text-red-900"
                  >
                    Cancel
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}

// Posted Tab
function PostedTab({ postedPosts }: { postedPosts: any[] }) {
  return (
    <div>
      <h2 className="text-xl font-semibold mb-2">Posted Posts</h2>
      {postedPosts.length === 0 ? (
        <p className="text-gray-500">No posted posts yet</p>
      ) : (
        <div className="space-y-4">
          {postedPosts.map((post) => (
            <div key={post.id} className="border border-gray-200 rounded-lg p-4">
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="font-bold">Post #{post.id}</h3>
                  <p className="text-sm text-gray-600">
                    Video: {post.videoId} | Platform: {post.platform}
                  </p>
                  <p className="mt-1 text-sm">
                    Posted at: {new Date(post.postedAt).toLocaleString()}
                  </p>
                </div>
                <div className="space-x-3">
                  <span
                    className="px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs"
                  >
                    {post.engagement?.likes ?? 0} Likes
                  </span>
                  <span
                    className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs"
                  >
                    {post.engagement?.comments ?? 0} Comments
                  </span>
                  <span
                    className="px-2 py-1 bg-purple-100 text-purple-800 rounded-full text-xs"
                  >
                    {post.engagement?.shares ?? 0} Shares
                  </span>
                  <span
                    className="px-2 py-1 bg-indigo-100 text-indigo-800 rounded-full text-xs"
                  >
                    {post.engagement?.saves ?? 0} Saves
                  </span>
                </div>
                <div className="mt-3 p-3 bg-gray-50 rounded">
                  <h4 className="font-medium mb-2">Engagement Breakdown</h4>
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div>
                      <div className="flex items-center">
                        <div className="w-2 h-2 bg-green-500 mr-1"></div>
                        <span>Likes: {post.engagement?.likes}</span>
                      </div>
                      <div className="flex items-center">
                        <div className="w-2 h-2 bg-blue-500 mr-1"></div>
                        <span>Comments: {post.engagement?.comments}</span>
                      </div>
                      <div className="flex items-center">
                        <div className="w-2 h-2 bg-purple-500 mr-1"></div>
                        <span>Shares: {post.engagement?.shares}</span>
                      </div>
                      <div className="flex items-center">
                        <div className="w-2 h-2 bg-indigo-500 mr-1"></div>
                        <span>Saves: {post.engagement?.saves}</span>
                      </div>
                    </div>
                    <div className="text-center">
                      <div className="text-xs text-gray-500">Engagement Rate: {((post.engagement?.likes ?? 0) / ((post.engagement?.comments ?? 0) + (post.engagement?.shares ?? 0) + (post.engagement?.saves ?? 0) + 1) * 100).toFixed(1)}%</span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          )}
        </div>
      )}
    </div>
  )
}

// Analytics Tab
function AnalyticsTab({ stats, loading, error }: { stats: any; loading: boolean; error: string | null }) {
  if (loading) {
    return <div className="p-6">Loading analytics...</div>
  }

  if (error) {
    return <div className="p-6 text-red-600">Error: {error}</div>
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-2">Social Media Analytics</h2>
      {stats ? (
        <div className="space-y-4">
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="font-semibold mb-4">Engagement Overview</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-500">Total Posts</p>
                <p className="text-2xl font-bold">{stats.posted_videos ?? 0}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Total Views</p>
                <p className="text-2xl font-bold">{stats.total_views ?? 0}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Average Engagement</p>
                <p className="text-2xl font-bold">{stats.average_engagement?.toFixed(2) ?? '0'}%</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Top Platform</p>
                <p className="text-2xl font-bold">
                  {stats.top_platform ?? 'Instagram'}
                </p>
              </div>
            </div>
          </div>
          <div className="bg-white p-6 rounded-lg shadow mt-4">
            <h3 className="font-semibold mb-4">Engagement Trend (Mock)</h3>
            {/* In a real app, we would render a chart here */}
            <div className="h-40 bg-gray-100 rounded-lg flex items-center justify-center">
              <span className="text-gray-500">Chart Placeholder</span>
            </div>
          </div>
        </div>
      ) : (
        <p className="text-center text-gray-500">No analytics data available</p>
      )}
    </div>
  )
}