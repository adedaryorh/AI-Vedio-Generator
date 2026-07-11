# Next.js Frontend for Story Video Social Media Bot MVP

I've successfully created a Next.js 14 frontend with TypeScript and Tailwind CSS that provides a user interface for managing the story-to-video-to-social-media pipeline.

## Features Implemented

### Authentication
- Simple login page with cookie-based authentication
- Protected routes that redirect to login when unauthenticated
- Login sets a cookie that is checked by middleware

### Pages
1. **Dashboard** (`/dashboard`) - Overview with statistics cards
2. **Stories** (`/stories`) - List of stories with enhance/video generation actions
3. **Videos** (`/videos`) - List of videos with status and posting controls
4. **Queue** (`/queue`) - Monitor pending and processing videos with manual process trigger
5. **Social Media** (`/social`) - Tabbed interface for composing, scheduling, viewing posted content, and analytics
6. **Analytics** (`/analytics`) - System statistics and platform distribution
7. **Settings** (`/settings`) - Configuration display and placeholder for future settings
8. **Login** (`/login`) - Authentication page
9. **404** (`/not-found`) - Page not found handler

### UI Components
- Responsive sidebar navigation
- Header with user actions and theme toggle placeholder
- Data tables and cards
- Forms with validation placeholders
- Loading and error states

### Technical Implementation
- **State Management**: React Query for server state (API calls)
- **API Communication**: Axios instances for each backend service with auth interceptors
- **Styling**: Tailwind CSS with custom css variables for light/dark mode
- **Routing**: Next.js App Router with route groups for protected paths
- **Authentication Middleware**: Server-side middleware to protect routes using cookies

## Setup Instructions

### 1. Backend Services
Ensure the backend services are running (as previously configured):
```bash
cd /home/adedaryorh/Documents/insta_ai
docker-compose up -d
# Initialize database
docker-compose exec story-collector alembic upgrade head
docker-compose exec story-collector python scripts/seed_stories.py
```

### 2. Frontend Configuration
Copy the example environment file and configure:
```bash
cd /home/adedaryorh/Documents/insta_ai/frontend
cp .env.example .env.local  # Create if doesn't exist
# Edit .env.local to set:
NEXT_PUBLIC_STORY_COLLECTOR_URL=http://localhost:8001
NEXT_PUBLIC_VIDEO_GENERATOR_URL=http://localhost:8002
NEXT_PUBLIC_CONTENT_MANAGER_URL=http://localhost:9001
NEXT_PUBLIC_SOCIAL_MEDIA_BOT_URL=http://localhost:9002
NEXT_PUBLIC_ADMIN_USER=admin
NEXT_PUBLIC_ADMIN_PASS=password
```

### 3. Install Dependencies & Start
```bash
cd /home/adedaryorh/Documents/insta_ai/frontend
npm install
npm run dev

## Access the Application
- Frontend: http://localhost:3000
- Login with credentials from `.env.local` (default: admin/password)
- Access backend APIs directly at:
  - Story Collector: http://localhost:8001/docs
  - Video Generator: http://localhost:8002/docs
  - Content Manager: http://localhost:9001/health
  - Social Media Bot: http://localhost:9002/health

## Next Steps for Production
1. **Containerize the Frontend**: Create a Dockerfile for the Next.js app and add it to `docker-compose.yml` for unified deployment.
2. **Implement Production Authentication**: Replace cookie-based auth with secure JWT (httpOnly cookies) and refresh tokens.
3. **Optimize Build for Production**: Use `next build` and `next start` with a process manager (PM2) or as a Docker container.
4. **Add Monitoring & Logging**: Integrate with a logging service (ELK, Datadog) and health check endpoints.
5. **Set Up CI/CD Pipeline**: Automate testing, linting, and deployment on push to main branch.
6. **Enhance Security**: Implement CSP headers, rate limiting, and input validation.
7. **Real-time Updates**: Replace polling with WebSocket connections for live data.
8. **Improve Media Handling**: Serve video files via a CDN or backend streaming endpoint.
9. **Integrate Real Social Media APIs**: Replace simulation with actual platform SDKs (Instagram Graph API, YouTube Data API, TikTok for Business).
10. **Add Comprehensive Error Boundaries**: Graceful UI degradation on partial failures.
11. **Performance Optimization**: Use image optimization, code splitting, and caching strategies.
12. **Testing**: Write unit and integration tests with Jest and React Testing Library.
13. **Documentation**: Generate API docs and maintain a runbook for ops.
14. **Scaling**: Prepare for horizontal scaling with stateless services and shared session store (Redis).
15. **Backup & Disaster Recovery**: Implement automated backups for persistent data (if any).

These steps will transition the MVP to a production-ready system suitable for real-world usage.