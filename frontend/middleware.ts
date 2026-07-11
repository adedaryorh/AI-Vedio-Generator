import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Define paths that require authentication
  const protectedPaths = [
    '/dashboard',
    '/stories',
    '/videos',
    '/queue',
    '/social',
    '/analytics',
    '/settings'
  ]

  // Check if the path is protected
  const isProtectedPath = protectedPaths.some(path => pathname.startsWith(path))

  // Get the token from cookies
  const token = request.cookies.get('token')?.value

  // If the path is protected and there's no token, redirect to login
  if (isProtectedPath && !token) {
    const url = new URL('/login', request.url)
    url.searchParams.set('callbackUrl', encodeURIComponent(pathname))
    return NextResponse.redirect(url)
  }

  // If the path is login and there is a token, redirect to dashboard
  if (pathname === '/login' && token) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public folder
     */
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:png|jpg|jpeg|gif|ico|svg)$).*)',
  ],
}