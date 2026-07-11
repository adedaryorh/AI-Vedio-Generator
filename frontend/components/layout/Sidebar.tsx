import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Users, Video, List, Share2, BarChart3, Settings } from 'lucide-react'

export const Sidebar = () => {
  const pathname = usePathname()

  const navItems = [
    { name: 'Dashboard', href: '/dashboard', icon: Users, current: pathname === '/dashboard' },
    { name: 'Stories', href: '/stories', icon: Users, current: pathname.startsWith('/stories') },
    { name: 'Videos', href: '/videos', icon: Video, current: pathname.startsWith('/videos') },
    { name: 'Queue', href: '/queue', icon: List, current: pathname.startsWith('/queue') },
    { name: 'Social Media', href: '/social', icon: Share2, current: pathname.startsWith('/social') },
    { name: 'Analytics', href: '/analytics', icon: BarChart3, current: pathname.startsWith('/analytics') },
    { name: 'Settings', href: '/settings', icon: Settings, current: pathname.startsWith('/settings') },
  ]

  return (
    <aside className="w-64 bg-white border-r lg:flex">
      <nav className="mt-6 space-y-1">
        {navItems.map((item) => (
          <Link
            key={item.name}
            href={item.href}
            className={`flex w-full items-center px-3 py-2 text-sm font-medium transition-colors
              ${item.current ? 'bg-blue-50 text-blue-600' : 'text-gray-600 hover:bg-gray-50'}`}
          >
            <div className="flex-shrink-0">
              <span className={`flex h-8 w-8 items-center justify-center rounded-md bg-${item.current ? 'blue-100' : 'gray-100'} text-${item.current ? 'blue-600' : 'gray-400'}`}>
                {/* Lucide icon */}
                {item.icon === Users && <Users className="h-4 w-4" />}
                {item.icon === Video && <Video className="h-4 w-4" />}
                {item.icon === List && <List className="h-4 w-4" />}
                {item.icon === Share2 && <Share2 className="h-4 w-4" />}
                {item.icon === BarChart3 && <BarChart3 className="h-4 w-4" />}
                {item.icon === Settings && <Settings className="h-4 w-4" />}
              </span>
            </div>
            <span className="ml-3">{item.name}</span>
          </Link>
        ))}
      </nav>
    </aside>
  )
}