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

// Resolve the API base URL.
//
// The API ships inside the app folder, so in every deployment it sits at
// <origin><basePath>api — e.g. https://host/cars/api for an app in /cars/.
// Deriving it from the host alone produced https://host/api/api.php, which 404s
// because there is no api/ at the domain root.
//
// The one exception is `npm run dev`, where the page is served by Vite on :5173
// and PHP runs separately on :8000. That is signalled explicitly, never guessed
// from the hostname: the previous check treated any 192.168.* host as a dev box,
// which broke the app for a server genuinely deployed on a LAN address (it
// pointed at http://192.168.x.x:8000/api, where nothing listens). Being
// host-agnostic is a hard requirement here, since the folder name, the domain and
// the IP are all supplied per client.
export function resolveApiBaseUrl({
  override,
  protocol,
  hostname,
  port,
  basePath,
  isDev,
  devApiPort = '8000',
}) {
  // Explicit override always wins (tunnels, split-horizon setups). It is a
  // build-time value, so it is the only reliable way to say "the API is not
  // next to this page".
  if (override) return override.replace(/\/+$/, '')

  // Dev: Vite serves the page on its own port while PHP runs on devApiPort, so
  // the page's port is dropped rather than concatenated (origin:5173 + :8000
  // would be invalid). The hostname is kept, so a phone hitting the dev server
  // over the LAN still reaches the API on that machine.
  if (isDev) return `${protocol}//${hostname}:${devApiPort}/api`

  // Production: the API sits inside the app folder. port is included so a
  // non-standard host (e.g. :8080 in testing) still resolves correctly.
  const origin = `${protocol}//${hostname}${port ? `:${port}` : ''}`

  return `${origin}${basePath}api`
}

export default getBasePath
