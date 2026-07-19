import { Sidebar } from '@/components/layout/Sidebar'
import { Header } from '@/components/layout/Header'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="app-shell"><Sidebar /><div className="min-w-0 pb-20 lg:pb-0">
        <Header />
        <main id="main-content" tabIndex={-1}>
          {children}
        </main>
      </div>
    </div>
  )
}
