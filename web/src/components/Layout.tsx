import { Link, Outlet, useLocation } from 'react-router-dom'
import { Scale, Home } from 'lucide-react'

export default function Layout() {
  const location = useLocation()
  
  return (
    <div className="min-h-screen bg-background">
      <header className="border-b bg-card">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <Link to="/" className="flex items-center gap-2 text-xl font-bold">
              <Scale className="h-6 w-6" />
              <span>MI Bill Tracker</span>
            </Link>
            <nav className="flex items-center gap-4">
              <Link
                to="/"
                className={`flex items-center gap-2 text-sm font-medium transition-colors ${
                  location.pathname === '/' ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                <Home className="h-4 w-4" />
                Bills
              </Link>
            </nav>
          </div>
        </div>
      </header>
      <main className="container mx-auto px-4 py-8">
        <Outlet />
      </main>
      <footer className="border-t bg-card mt-auto">
        <div className="container mx-auto px-4 py-4 text-center text-sm text-muted-foreground">
          MI Bill Tracker - Track Michigan Legislature bills and contact your representatives
        </div>
      </footer>
    </div>
  )
}
