"use client"

import { FormEvent, useEffect, useState } from 'react'
import { contentManager, socialMediaBot } from '@/lib/api'

type Video = { id: number; story_id: number; status: string }
type Post = { VideoID?: number; Platform?: string; Status?: string; PostedAt?: string }

const platforms = ['instagram', 'youtube', 'tiktok'] as const

export default function SocialPage() {
  const [videos, setVideos] = useState<Video[]>([])
  const [posts, setPosts] = useState<Post[]>([])
  const [videoId, setVideoId] = useState('')
  const [platform, setPlatform] = useState<(typeof platforms)[number]>('instagram')
  const [caption, setCaption] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  const loadData = async () => {
    const [videoResponse, postResponse] = await Promise.all([
      contentManager.get<Video[]>('/videos'),
      socialMediaBot.get<Post[]>('/posts'),
    ])
    setVideos(videoResponse.data.filter((video) => video.status === 'ready'))
    setPosts(postResponse.data || [])
  }

  useEffect(() => {
    loadData().catch(() => setMessage('Some social media data could not be loaded.'))
  }, [])

  const submitPost = async (event: FormEvent) => {
    event.preventDefault()
    if (!videoId) return setMessage('Select a ready video first.')
    setSubmitting(true)
    setMessage(null)
    try {
      await socialMediaBot.post('/post', { video_id: Number(videoId), platform, caption })
      setMessage('Post submitted successfully.')
      setCaption('')
      await loadData()
    } catch (error: any) {
      setMessage(error.response?.data?.error || 'The post could not be submitted.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="page">
      <div>
        <p className="eyebrow">Distribution</p><h1 className="page-title">Social publishing</h1>
        <p className="page-subtitle">Publish ready videos and review recent posts.</p>
      </div>

      <form onSubmit={submitPost} className="panel panel-pad space-y-4">
        <h2 className="text-lg font-semibold">Create post</h2>
        <div className="grid gap-4 md:grid-cols-2">
          <label className="text-sm font-medium">Video
            <select value={videoId} onChange={(event) => setVideoId(event.target.value)} className="field mt-1">
              <option value="">Select a ready video</option>
              {videos.map((video) => <option key={video.id} value={video.id}>Video #{video.id} · Story #{video.story_id}</option>)}
            </select>
          </label>
          <label className="text-sm font-medium">Platform
            <select value={platform} onChange={(event) => setPlatform(event.target.value as typeof platform)} className="field mt-1">
              {platforms.map((item) => <option key={item} value={item} className="capitalize">{item}</option>)}
            </select>
          </label>
        </div>
        <label className="block text-sm font-medium">Caption
          <textarea value={caption} onChange={(event) => setCaption(event.target.value)} rows={4} className="field mt-1" placeholder="Write a caption..." />
        </label>
        <button disabled={submitting} className="btn-primary">
          {submitting ? 'Publishing...' : 'Publish'}
        </button>
        {message && <p aria-live="polite" className="text-sm text-amber-200">{message}</p>}
      </form>

      <section className="panel panel-pad">
        <h2 className="mb-4 text-lg font-semibold">Recent posts</h2>
        {posts.length === 0 ? <p className="text-sm text-gray-500">No posts yet.</p> : (
          <div className="divide-y">
            {posts.map((post, index) => (
              <div key={`${post.VideoID}-${index}`} className="flex justify-between py-3 text-sm">
                <span>Video #{post.VideoID || '—'} · {post.Platform || 'Unknown platform'}</span>
                <span className="capitalize text-gray-500">{post.Status || 'posted'}</span>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
