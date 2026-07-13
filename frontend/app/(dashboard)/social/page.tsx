import { useState } from 'react'
import { useEffect } from 'react'
import { contentManager, socialMediaBot } from '@/lib/api'

export default function SocialPage() {
  const [tab, setTab] = useState<'compose' | 'scheduled' | 'posted' | 'analytics'>(
    'compose'
  )

  // State for compose tab
  const [selectedVideoId, setSelectedVideoId] = useState<number | null>(null)
  const [platforms, setPlatforms] = useState<string[]>([])
  const [caption, setCaption] = useState('')
  const [hashtags, setHashtags] = useState<string[]>([])
  const [posting, setPosting] = useState(false)
  const [postError, setPostError] = useState<string | null>(null)
  const [postSuccess, setPostSuccess] = useState<boolean>(false)

  // State for scheduled tab (client-side storage since no backend endpoint)
  const [scheduledPosts, setScheduledPosts] = useState<any[]>([])
  const [loadingScheduled, setLoadingScheduled] = useState(false)
  const [errorScheduled, setErrorScheduled] = useState<string | null>(null)

  // State for posted tab
  const [postedPosts, setPostedPosts] = useState<any[]>([])
  const [loadingPosted, setLoadingPosted] = useState(true)
  const [errorPosted, setErrorPosted] = useState<string | null>(null)

  // State for analytics tab
  const [analyticsStats, setAnalyticsStats] = useState<any>(null)
  const [analyticsLoading, setAnalyticsLoading] = useState(true)
  const [analyticsError, setAnalyticsError] = useState<string | null>(null)

  // Fetch data for each tab when it becomes active
  useEffect(() => {
    if (tab === 'posted') {
      fetchPostedPosts()
    } else if (tab === 'analytics') {
      fetchAnalytics()
    }
  }, [tab])

  const fetchPostedPosts = async () => {
    setLoadingPosted(true)
    setErrorPosted(null)
    try {
      const response = await socialMediaBot.get('/posts')
      setPostedPosts(response.data || [])
    } catch (err: any) {
      setErrorPosted(err.response?.data?.detail || 'Failed to fetch posted posts')
      console.error(err)
    } finally {
      setLoadingPosted(false)
    }
  }

  const fetchAnalytics = async () => {
    setAnalyticsLoading(true)
    setAnalyticsError(null)
    try {
      const response = await socialMediaBot.get('/analytics')
      setAnalyticsStats(response.data)
    } catch (err: any) {
      setAnalyticsError(err.response?.data?.detail || 'Failed to fetch analytics')
      console.error(err)
    } finally {
      setAnalyticsLoading(false)
    }
  }

  const handleTabChange = (t: typeof tab) => {
    setTab(t)
  }

  const handlePlatformToggle = (platform: string) => {
    if (platforms.includes(platform)) {
      setPlatforms(platforms.filter((p) => p !== platform))
    } else {
      setPlatforms([...platforms, platform])
    }
  }

  const handleHashtagChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const tags = e.target.value
      .split(',')
      .map((t) => t.trim())
      .filter((t) => t.length > 0)
    setHashtags(tags)
  }

  const handlePost = async () => {
    if (!selectedVideoId || platforms.length === 0) {
      setPostError('Please select a video and at least one platform')
      return
    }

    setPosting(true)
    setPostError(null)
    setPostSuccess(false)

    try {
      // Post to each selected platform
      const posts = []
      for (const platform of platforms) {
        const response = await socialMediaBot.post('/post', {
          video_id: selectedVideoId,
          platform,
          caption,
          hashtags,
        })
        posts.push(response.data)
      }

      setPostSuccess(true)
      // Clear form after successful post
      setSelectedVideoId(null)
      setPlatforms([])
      setCaption('')
      setHashtags([])

      // Refresh posted posts if we're on the posted tab
      if (tab === 'posted') {
        await fetchPostedPosts()
      }
    } catch (err: any) {
      setPostError(err.response?.data?.detail || 'Failed to post to social media')
      console.error(err)
    } finally {
      setPosting(false)
    }
  }

  const handleSchedule = async () => {
    if (!selectedVideoId || platforms.length === 0) {
      setPostError('Please select a video and at least one platform')
      return
    }

    // In a real app, we would have a schedule time input
    // For now, we'll simulate scheduling by adding to our local scheduled posts list
    const scheduleTime = new Date(Date.now() + 3600000) // 1 hour from now

    const newScheduledPost = {
      id: Date.now(), // temporary ID
      videoId: selectedVideoId,
      platforms: [...platforms],
      caption,
      hashtags: [...hashtags],
      scheduleTime: scheduleTime.toISOString(),
      createdAt: new Date().toISOString()
    }

    setScheduledPosts(prev => [...prev, newScheduledPost])

    // Clear form
    setSelectedVideoId(null)
    setPlatforms([])
    setCaption('')
    setHashtags([])

    alert(`Post scheduled for ${scheduleTime.toLocaleString()}!`)
  }

  const handleCancelScheduledPost = (postId: number) => {
    if (!window.confirm('Are you sure you want to cancel this scheduled post?')) {
      return
    }

    setScheduledPosts(prev => prev.filter(post => post.id !== postId))
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
        <ComposeTab
          // We would pass videos data here in a real implementation
          // For now, we'll fetch it locally or pass it as a prop
          onVideoSelect={(videoId: number) => setSelectedVideoId(videoId)}
          selectedVideoId={selectedVideoId}
          platforms={platforms}
          setPlatforms={setPlatforms}
          caption={caption}
          setCaption={setCaption}
          hashtags={hashtags}
          setHashtags={setHashtags}
          posting={posting}
          postError={postError}
          postSuccess={postSuccess}
          onPost={handlePost}
          onSchedule={handleSchedule}
          onPlatformToggle={handlePlatformToggle}
          onHashtagChange={handleHashtagChange}
        />
      )}
      {tab === 'scheduled' && (
        <ScheduledTab
          scheduledPosts={scheduledPosts}
          loading={loadingScheduled}
          error={errorScheduled}
          onCancelScheduledPost={handleCancelScheduledPost}
          onVideoSelect={(videoId: number) => {
            setSelectedVideoId(videoId)
            setTab('compose') // Switch to compose tab when selecting a video
          }}
        />
      )}
      {tab === 'posted' && (
        <PostedTab
          postedPosts={postedPosts}
          loading={loadingPosted}
          error={errorPosted}
          onVideoSelect={(videoId: number) => {
            setSelectedVideoId(videoId)
            setTab('compose') // Switch to compose tab when selecting a video
          }}
        />
      )}
      {tab === 'analytics' && (
        <AnalyticsTab
          stats={analyticsStats}
          loading={analyticsLoading}
          error={analyticsError}
          onVideoSelect={(videoId: number) => {
            setSelectedVideoId(videoId)
            setTab('compose') // Switch to compose tab when selecting a video
          }}
        />
      )}
    </div>
  )
}

function ComposeTab({
  onVideoSelect,
  selectedVideoId,
  platforms,
  setPlatforms,
  caption,
  setCaption,
  hashtags,
  setHashtags,
  posting,
  postError,
  postSuccess,
  onPost,
  onSchedule,
  onPlatformToggle,
  onHashtagChange,
}: {
  onVideoSelect: (videoId: number) => void
  selectedVideoId: number | null
  platforms: string[]
  setPlatforms: React.Dispatch<React.SetStateAction<string[]>>
  caption: string
  setCaption: React.Dispatch<React.SetStateAction<string>>
  hashtags: string[]
  setHashtags: React.Dispatch<React.SetStateAction<string[]>>
  posting: boolean
  postError: string | null
  postSuccess: boolean
  onPost: () => Promise<void>
  onSchedule: () => Promise<void>
  onPlatformToggle: (platform: string) => void
  onHashtagChange: (e: React.ChangeEvent<HTMLInputElement>) => void
}) {
  // In a real app, we would fetch the videos list and pass it as a prop
  // For this example, we'll simulate having access to videos
  const [videos, setVideos] = useState<any[]>([])

  useEffect(() => {
    // Fetch available videos that are ready for posting
    const fetchVotes = async () => {
      try {
        const response = await contentManager.get('/videos')
        // Filter to only show videos that are ready for posting
        const readyVideos = response.data.filter((video: any) => video.status === 'ready' || video.status === 'completed')
        setVideos(readyVideos)
      } catch (err: any) {
        console.error('Failed to fetch videos for social posting:', err)
      }
    }

    fetchVotes()
  }, [])

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
                onVideoSelect(val === '' ? null : parseInt(val))
              }}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={posting}
            >
              <option value="">Select a video</option>
              {videos.map((video) => (
                <option key={video.id} value={video.id}>
                  Video for Story {video.story_id} ({video.status})
                </option>
              ))}
            </select>
            {!selectedVideoId && videos.length > 0 && (
              <p className="mt-1 text-sm text-gray-500">
                {videos.length} videos available for posting
              </p>
            )}
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Platforms</label>
            <div className="flex space-x-2 flex-wrap">
              {['instagram', 'youtube', 'tiktok'].map((platform) => (
                <label key={platform} className="flex items-center">
                  <input
                    type="checkbox"
                    checked={platforms.includes(platform)}
                    onChange={() => onPlatformToggle(platform)}
                    className="h-4 w-4 text-blue-600"
                    disabled={posting}
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
              disabled={posting}
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
              onChange={onHashtagChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="#travel #adventure"
              disabled={posting}
            />
          </div>
          <div className="flex justify-end space-x-3">
            <button
              onClick={onSchedule}
              disabled={!selectedVideoId || platforms.length === 0 || posting}
              className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300"
            >
              Schedule
            </button>
            <button
              onClick={onPost}
              disabled={!selectedVideoId || platforms.length === 0 || posting}
              className={`px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 ${posting ? 'opacity-50 cursor-not-allowed' : ''}`}
            >
              {posting ? 'Posting...' : 'Post Now'}
            </button>
          </div>
          {postError && (
            <p className="mt-2 text-sm text-red-600">
              {postError}
            </p>
          )}
          {postSuccess && (
            <p className="mt-2 text-sm text-green-600">
              Posted successfully!
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

function ScheduledTab({
  scheduledPosts,
  loading,
  error,
  onCancelScheduledPost,
  onVideoSelect,
}: {
  scheduledPosts: any[]
  loading: boolean
  error: string | null
  onCancelScheduledPost: (postId: number) => void
  onVideoSelect: (videoId: number) => void
}) {
  return (
    <div>
      <h2 className="text-xl font-semibold mb-2">Scheduled Posts</h2>
      {loading ? (
        <p className="text-gray-500">Loading scheduled posts...</p>
      ) : error ? (
        <p className="text-red-600">Error: {error}</p>
      ) : scheduledPosts.length === 0 ? (
        <p className="text-gray-500">No scheduled posts</p>
      ) : (
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Video
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                Platforms
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
            {scheduledPosts.map((post: any) => (
              <tr key={post.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="font-medium">{post.videoId}</span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {post.platforms.join(', ')}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {new Date(post.scheduleTime).toLocaleString()}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                  <button
                    onClick={() => onCancelScheduledPost(post.id)}
                    className="text-red-600 hover:text-red-900"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={() => onVideoSelect(post.videoId)}
                    className="ml-4 text-blue-600 hover:text-blue-800"
                  >
                    View Video
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

function PostedTab({
  postedPosts,
  loading,
  error,
  onVideoSelect,
}: {
  postedPosts: any[]
  loading: boolean
  error: string | null
  onVideoSelect: (videoId: number) => void
}) {
  return (
    <div>
      <h2 className="text-xl font-semibold mb-2">Posted Posts</h2>
      {loading ? (
        <p className="text-gray-500">Loading posted posts...</p>
      ) : error ? (
        <p className="text-red-600">Error: {error}</p>
      ) : postedPosts.length === 0 ? (
        <p className="text-gray-500">No posted posts yet</p>
      ) : (
        <div className="space-y-4">
          {postedPosts.map((post: any) => (
            <div key={post.id} className="border border-gray-200 rounded-lg p-4">
              <div className="flex justify-between items-start">
                <div>
                  <h3 className="font-bold">Post #{post.id}</h3>
                  <p className="text-sm text-gray-600">
                    Video: {post.VideoID} | Platform: {post.Platform}
                  </p>
                  <p className="mt-1 text-sm">
                    Posted at: {new Date(post.PostedAt).toLocaleString()}
                  </p>
                </div>
                <div className="space-x-3">
                  <span
                    className="px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs"
                  >
                    {post.EngagementMetrics?.likes ?? 0} Likes
                  </span>
                  <span
                    className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs"
                  >
                    {post.EngagementMetrics?.comments ?? 0} Comments
                  </span>
                  <span
                    className="px-2 py-1 bg-purple-100 text-purple-800 rounded-full text-xs"
                  >
                    {post.EngagementMetrics?.shares ?? 0} Shares
                  </span>
                  <span
                    className="px-2 py-1 bg-indigo-100 text-indigo-800 rounded-full text-xs"
                  >
                    {post.EngagementMetrics?.saves ?? 0} Saves
                  </span>
                </div>
                <div className="mt-3 p-3 bg-gray-50 rounded">
                  <h4 className="font-medium mb-2">Engagement Breakdown</h4>
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div>
                      <div className="flex items-center">
                        <div className="w-2 h-2 bg-green-500 mr-1"></div>
                        <span>Likes: {post.EngagementMetrics?.likes}</span>
                      </div>
                      <div className="flex items-center">
                        <div className="w-2 h-2 bg-blue-500 mr-1"></div>
                        <span>Comments: {post.EngagementMetrics?.comments}</span>
                      </div>
                      <div className="flex items-center">
                        <div className="w-2 h-2 bg-purple-500 mr-1"></div>
                        <span>Shares: {post.EngagementMetrics?.shares}</span>
                      </div>
                      <div className="flex items-center">
                        <div className="w-2 h-2 bg-indigo-500 mr-1"></div>
                        <span>Saves: {post.EngagementMetrics?.saves}</span>
                      </div>
                    </div>
                    <div className="text-center">
                      <div className="text-xs text-gray-500">
                        Engagement Rate: {((post.EngagementMetrics?.likes ?? 0) / ((post.EngagementMetrics?.comments ?? 0) + (post.EngagementMetrics?.shares ?? 0) + (post.EngagementMetrics?.saves ?? 0) + 1) * 100).toFixed(1)}%
                      </div>
                    </div>
                  </div>
                </div>
                <div className="mt-2">
                  <button
                    onClick={() => onVideoSelect(post.VideoID)}
                    className="text-blue-600 hover:text-blue-800"
                  >
                    View Video
                  </button>
                </div>
              </div>
            ))}
          </div>
      )}
    </div>
  )
}

function AnalyticsTab({
  stats,
  loading,
  error,
  onVideoSelect,
}: {
  stats: any
  loading: boolean
  error: string | null
  onVideoSelect: (videoId: number) => void
}) {
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
        <div className="space-y-6">
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="font-semibold mb-4">Engagement Overview</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-500">Total Posts</p>
                <p className="text-2xl font-bold">{stats.total_posts ?? 0}</p>
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