import type { QueryClient } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { Link, Outlet, createRootRouteWithContext } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  component: RootLayout,
  notFoundComponent: () => (
    <main className="page center">
      <h2>Halaman tidak ditemukan</h2>
      <Link to="/">Kembali ke daftar peta</Link>
    </main>
  ),
})

function RootLayout() {
  return (
    <div className="app">
      <header className="topbar">
        <Link to="/" className="brand">
          <span className="brand-mark">◆</span> Miniv2 <small>Top-down Map</small>
        </Link>
        <nav>
          <Link to="/" activeOptions={{ exact: true }}>
            Peta
          </Link>
        </nav>
      </header>
      <Outlet />
      {import.meta.env.DEV && (
        <>
          <TanStackRouterDevtools position="bottom-right" />
          <ReactQueryDevtools buttonPosition="bottom-left" />
        </>
      )}
    </div>
  )
}
