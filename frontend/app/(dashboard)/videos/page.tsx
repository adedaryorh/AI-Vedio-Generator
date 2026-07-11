import { useEffect, useState } from 'react'
import { contentManager } from '@/lib/api'

export default function VideosPage() {
  const [videos, setVideos] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchVideos = async () => {
      try {
        setLoading(true)
        const response = await contentManager.get('/videos')
        setVideos(response.data)
      } catch (err: any) {
        setError(err.response?.data?.detail || 'Failed to fetch videos')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchVideos()
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
        <h1 className="text-2xl font-bold">Videos</h1>
        <button
          onClick={() => {
            // TODO: Implement video generation modal
            alert('Generate video from story')
          }}
          className="px-3 py-2 bg-green-600 text-white rounded hover:bg-green-700"
        >
          Generate Video
        </button>
      </div>

      {videos.length === 0 ? (
        <p className="text-center text-gray-500">No videos found</p>
      ) : (
        <div className="space-y-4">
          {videos.map((video) => (
            <div key={video.id} className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
              <div className="mb-2">
                <h2 className="font-bold text-lg">Video for Story {video.story_id}</h2>
                <p className="text-sm text-gray-600">
                  Status:
                  <span className={`px-2 py-1 rounded-full ${
                    video.status === 'pending'
                      ? 'bg-yellow-100 text-yellow-800'
                      : video.status === 'processing'
                      ? 'bg-blue-100 text-blue-800'
                      : video.status === 'ready'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-red-100 text-red-800'
                  }`}
                  >
                    {video.status}
                  </span>
                </p>
                {video.file_path && (
                  <div className="mt-2">
                    <p className="text-sm">File: {video.file_path}</p>
                    {/* TODO: If we can serve the file, show a video player */}
                    <video
                      controls
                      className="mt-1 max-w-xs"
                      src={`${process.env.NEXT_PUBLIC_CONTENT_MANAGER_URL || 'http://localhost:9001'}${video.file_path}`}
                    >
                      Your browser does not support the video tag.
                    </video>
                  </div>
                )}
              </div>
              <div className="mt-3 flex justify-end space-x-2">
                <button
                  onClick={() => {
                    // TODO: Retry if failed
                    alert(`Retry video ${video.id}`)
                  }}
                  disabled={video.status !== 'failed'}
                  className="px-3 py-1 text-sm bg-gray-200 rounded hover:bg-gray-300"
                >
                  Retry
                </button>
                {video.status === 'ready' && (
                  <>
                    <button
                      onClick={() => {
                        // TODO: Post to social media
                        alert(`Post video ${video.id} to social media`)
                      }}
                      className="px-3 py-1 text-sm bg-purple-600 text-white rounded hover:bg-purple-700 mr-2"
                    >
                      Post to Social
                    </button>
                    <button
                      onClick={() => {
                        // TODO: Delete video
                        alert(`Delete video ${video.id}`)
                      }}
                      className="px-3 py-1 text-sm bg-red-600 text-white rounded hover:bg-red-700"
                    >
                      Delete
                    </button>
                  </>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}