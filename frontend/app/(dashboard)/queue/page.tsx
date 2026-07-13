import { useEffect, useState } from 'react'
import { contentManager } from '@/lib/api'

export default function QueuePage() {
  const [pending, setPending] = useState<any[]>([])
  const [processing, setProcessing] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [processingQueue, setProcessingQueue] = useState(false)

  useEffect(() => {
    const fetchQueue = async () => {
      try {
        setLoading(true)
        const response = await contentManager.get('/queue')
        // Assuming response has pending and processing arrays
        setPending(response.data.pending || [])
        setProcessing(response.data.processing || [])
      } catch (err: any) {
        setError(err.response?.data?.detail || 'Failed to fetch queue')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchQueue()
    // Poll every 5 seconds
    const interval = setInterval(fetchQueue, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleProcessQueue = async () => {
    setProcessingQueue(true)
    try {
      // Trigger manual queue processing
      await contentManager.post('/process')
      // The polling will pick up the updated status
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to process queue')
      console.error(err)
    } finally {
      setProcessingQueue(false)
    }
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
        <h1 className="text-2xl font-bold">Processing Queue</h1>
        <button
          onClick={handleProcessQueue}
          disabled={processingQueue}
          className={`px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 ${processingQueue ? 'opacity-50 cursor-not-allowed' : ''}`}
        >
          {processingQueue ? 'Processing...' : 'Process Queue'}
        </button>
      </div>

      <div className="space-y-6">
        <div>
          <h2 className="text-xl font-semibold mb-2">Pending ({pending.length})</h2>
          {pending.length === 0 ? (
            <p className="text-gray-500">No pending videos</p>
          ) : (
            <div className="space-y-3">
              {pending.map((video) => (
                <div key={video.id} className="border border-gray-200 rounded-lg p-3">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="font-bold">Video for Story {video.story_id}</h3>
                      <p className="text-sm text-gray-600">Queued at: {new Date(
                        video.created_at
                      ).toLocaleString()}</p>
                    </div>
                    <span className="px-2 py-1 bg-yellow-100 text-yellow-800 rounded-full text-xs">
                      Pending
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div>
          <h2 className="text-xl font-semibold mb-2">Processing ({processing.length})</h2>
          {processing.length === 0 ? (
            <p className="text-gray-500">No videos processing</p>
          ) : (
            <div className="space-y-3">
              {processing.map((video) => (
                <div key={video.id} className="border border-gray-200 rounded-lg p-3">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="font-bold">Video for Story {video.story_id}</h3>
                      <p className="text-sm text-gray-600">
                        Started at: {new Date(
                          video.updated_at ||
                            video.created_at ||
                            Date.now()
                        ).toLocaleString()}
                      </p>
                    </div>
                    <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs">
                      Processing
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}