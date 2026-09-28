// Single source of truth for the directory the app is served from (the mount
// point), e.g. '/' on a root deploy or '/cars/' in production.
//
// Why this is not derived from window.location.pathname:
// the app is deployed to /cars/ while /cars is *also* a route name (CarsView).
// The old "is this pathname a known root route?" heuristic therefore matched
// '/cars/', returned '/', made the <base> tag point at the domain root and 404'd
// every asset, leaving a blank page. Any list-of-routes guessing is fragile for
// the same reason.
//
// Instead the build tells us: with a relative Vite base every emitted chunk
// lives flat in the app root, so the module's own URL is authoritative. In dev
// the server always serves from the root, hence the isDev branch.
export function resolveBasePath(rawBase, moduleUrl, isDev) {
  const base = rawBase || './'

  // Absolute base (dev server, or an explicit base in vite.config.js).
  if (base !== './' && !base.startsWith('./')) {
    return base
  }

  if (isDev) {
    return '/'
  }

  try {
    const dir = new URL('.', moduleUrl).pathname
    return dir.endsWith('/') ? dir : `${dir}/`
  } catch {
    return '/'
  }
}

const RAW_BASE = import.meta.env.BASE_URL || './'

export function getBasePath() {
  return resolveBasePath(RAW_BASE, import.meta.url, import.meta.env.DEV)
}

export const BASE_PATH = getBasePath()

export default getBasePath
