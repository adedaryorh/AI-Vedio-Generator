import { useEffect, useState } from 'react'
import { storyCollector } from '@/lib/api'

export default function StoriesPage() {
  const [stories, setStories] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchStories = async () => {
      try {
        setLoading(true)
        const response = await storyCollector.get('/stories')
        setStories(response.data)
      } catch (err: any) {
        setError(err.response?.data?.detail || 'Failed to fetch stories')
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchStories()
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
        <h1 className="text-2xl font-bold">Stories</h1>
        <button
          onClick={() => {
            // TODO: Implement story collection modal
            alert('Collect stories feature coming soon')
          }}
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Collect Stories
        </button>
      </div>

      {stories.length === 0 ? (
        <p className="text-center text-gray-500">No stories found</p>
      ) : (
        <div className="space-y-4">
          {stories.map((story) => (
            <div key={story.id} className="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
              <div className="flex justify-between items-start">
                <div>
                  <h2 className="font-bold text-lg">{story.title}</h2>
                  <p className="text-sm text-gray-600 truncate max-w-[300px]">
                    {story.content}
                  </p>
                </div>
                <div className="space-x-2 text-sm">
                  <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full">
                    {story.source}
                  </span>
                  <span className="px-2 py-1 bg-green-100 text-green-800 rounded-full">
                    {story.category}
                  </span>
                  <span className="px-2 py-1 bg-purple-100 text-purple-800 rounded-full">
                    {story.culture}
                  </span>
                  <span
                    className={`px-2 py-1 rounded-full ${
                      story.processed
                        ? 'bg-green-100 text-green-800'
                        : 'bg-red-100 text-red-800'
                    }`}
                  >
                    {story.processed ? 'Processed' : 'Raw'}
                  </span>
                </div>
              </div>
              <div className="mt-3 flex justify-end space-x-2">
                <button
                  onClick={() => {
                    // TODO: Navigate to story detail
                    alert(`View story ${story.id}`)
                  }}
                  className="px-3 py-1 text-sm bg-gray-200 rounded hover:bg-gray-300"
                >
                  View
                </button>
                {!story.processed && (
                  <button
                    onClick={() => {
                      // TODO: Enhance story
                      alert(`Enhance story ${story.id}`)
                    }}
                    className="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                  >
                    Enhance
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}