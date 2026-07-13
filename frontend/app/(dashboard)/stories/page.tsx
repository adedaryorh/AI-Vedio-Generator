import { useEffect, useState } from 'react'
import { storyCollector } from '@/lib/api'

export default function StoriesPage() {
  const [stories, setStories] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [collecting, setCollecting] = useState(false)
  const [enhancing, setEnhancing] = useState(false)
  const [enhancingStoryId, setEnhancingStoryId] = useState<number | null>(null)
  const [selectedSource, setSelectedSource] = useState<'gutenberg' | 'islamic' | 'african' | 'all'>('all')

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

  const handleCollectStories = async () => {
    setCollecting(true)
    try {
      await storyCollector.post('/collect', { source: selectedSource })
      // Refresh the stories list after collection
      const response = await storyCollector.get('/stories')
      setStories(response.data)
    } catch (err: any) {
      setError(err.response?.data?.detail || 'Failed to collect stories')
      console.error(err)
    } finally {
      setCollecting(false)
    }
  }

  const handleEnhanceStory = async (storyId: number) => {
    setEnhancingStoryId(storyId)
    setEnhancing(true)
    try {
      await storyCollector.post(`/enhance/${storyId}`)
      // Refresh the stories list to show updated processing status
      const response = await storyCollector.get('/stories')
      setStreams(response.data)
    } catch (err: any) {
      setError(err.response?.data?.detail || `Failed to enhance story ${storyId}`)
      console.error(err)
    } finally {
      setEnhancing(false)
      setEnhancingStoryId(null)
    }
  }

  const handleViewStory = (storyId: number) => {
    // In a real app, we would navigate to a story detail page
    // For now, we'll just show an alert with the story ID
    alert(`Viewing story ${storyId}`)
    // TODO: Replace with actual navigation when story detail page is implemented
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
        <h1 className="text-2xl font-bold">Stories</h1>
        <div className="flex space-x-3">
          <select
            value={selectedSource}
            onChange={(e) => setSelectedSource(e.target.value as any)}
            className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All Sources</option>
            <option value="gutenberg">Project Gutenberg</option>
            <option value="islamic">Islamic Stories</option>
            <option value="african">African Folklore</option>
          </select>
          <button
            onClick={handleCollectStories}
            disabled={collecting}
            className={`px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 ${collecting ? 'opacity-50 cursor-not-allowed' : ''}`}
          >
            {collecting ? 'Collecting...' : 'Collect Stories'}
          </button>
        </div>
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
                  onClick={() => handleViewStory(story.id)}
                  className="px-3 py-1 text-sm bg-gray-200 rounded hover:bg-gray-300"
                >
                  View
                </button>
                {!story.processed && !enhancing && enhancingStoryId !== story.id && (
                  <button
                    onClick={() => handleEnhanceStory(story.id)}
                    disabled={enhancing}
                    className={`px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 ${enhancing ? 'opacity-50 cursor-not-allowed' : ''}`}
                  >
                    {enhancing && enhancingStoryId === story.id ? 'Enhancing...' : 'Enhance'}
                  </button>
                )}
                {story.processed && (
                  <button
                    onClick={() => {
                      // TODO: Navigate to video generation for this story
                      alert(`Generate video for story ${story.id}`)
                    }}
                    className="px-3 py-1 text-sm bg-green-600 text-white rounded hover:bg-green-700"
                  >
                    Generate Video
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