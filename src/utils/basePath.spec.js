import { describe, it, expect } from 'vitest'
import { resolveBasePath, resolveApiBaseUrl } from './basePath'

// The app is deployed to /cars/ while /cars is also a route name. The original
// implementation guessed the mount point from a list of known routes, matched
// '/cars/', collapsed the base to '/' and 404'd every asset (blank page). These
// tests pin the resolution rules that replaced it.
describe('resolveBasePath', () => {
  it('keeps an absolute configured base unchanged (dev server)', () => {
    expect(resolveBasePath('/', 'http://localhost:5173/src/utils/basePath.js', true)).toBe('/')
  })

  it('keeps an absolute production base unchanged', () => {
    expect(resolveBasePath('/cars/', 'https://host/cars/index.abc.js', false)).toBe('/cars/')
  })

  it('derives the mount dir from the module URL when the build base is relative', () => {
    expect(resolveBasePath('./', 'https://host/cars/index.abc.js', false)).toBe('/cars/')
    expect(resolveBasePath('./', 'https://host/mig_26/ui.abc.js', false)).toBe('/mig_26/')
  })

  it('returns root on a root deploy with a relative base', () => {
    expect(resolveBasePath('./', 'https://host/index.abc.js', false)).toBe('/')
  })

  it('uses root in dev even when the build base is relative', () => {
    expect(resolveBasePath('./', 'http://localhost:5173/src/utils/basePath.js', true)).toBe('/')
  })

  it('always ends with a trailing slash so `${base}route` stays correct', () => {
    const base = resolveBasePath('./', 'https://host/cars/index.abc.js', false)
    expect(`${base}buy-payments/5`).toBe('/cars/buy-payments/5')
  })

  it('falls back to root for an unusable module URL instead of throwing', () => {
    expect(resolveBasePath('./', 'not-a-url', false)).toBe('/')
  })

  it('never consults the pathname or a route list', async () => {
    const source = await import('node:fs').then((fs) =>
      fs.readFileSync(new URL('./basePath.js', import.meta.url), 'utf8'),
    )
    // Strip comments so the explanatory prose does not trip the guard.
    const code = source.replace(/\/\*[\s\S]*?\*\//g, '').replace(/^\s*\/\/.*$/gm, '')
    expect(code).not.toMatch(/location\.pathname/)
    expect(code).not.toMatch(/knownRoutes|knownRootRoutes/)
  })
})

// The app is deployed per client: any folder name, any domain, any IP, with the
// API living inside the app folder. These pin the URL that makes that work, so a
// refactor cannot quietly reintroduce a host-based guess.
describe('resolveApiBaseUrl', () => {
  const prod = { protocol: 'https:', hostname: 'world-automobile.com', port: '', isDev: false }

  it('places the API inside the app folder, never at the domain root', () => {
    expect(resolveApiBaseUrl({ ...prod, basePath: '/cars/' })).toBe(
      'https://world-automobile.com/cars/api',
    )
  })

  it('follows the folder name for any client', () => {
    for (const basePath of ['/mig_26/', '/acme/', '/2026fleet/', '/my.client-v2/']) {
      expect(resolveApiBaseUrl({ ...prod, basePath })).toBe(
        `https://world-automobile.com${basePath}api`,
      )
    }
  })

  it('works on a root deploy', () => {
    expect(resolveApiBaseUrl({ ...prod, basePath: '/' })).toBe('https://world-automobile.com/api')
  })

  it('works over a raw IP on a LAN, which must NOT be treated as a dev box', () => {
    // This is the regression: the old code treated any 192.168.* host as local
    // and pointed at :8000, where a deployed server has nothing listening.
    expect(
      resolveApiBaseUrl({ protocol: 'https:', hostname: '192.168.1.50', port: '', isDev: false, basePath: '/cars/' }),
    ).toBe('https://192.168.1.50/cars/api')
  })

  it('keeps a non-standard port in production', () => {
    expect(resolveApiBaseUrl({ ...prod, port: '8080', basePath: '/cars/' })).toBe(
      'https://world-automobile.com:8080/cars/api',
    )
  })

  it('points dev at the PHP port instead of the Vite port', () => {
    // origin is :5173 in dev; concatenating :8000 onto it would be invalid.
    expect(
      resolveApiBaseUrl({ protocol: 'http:', hostname: 'localhost', port: '5173', isDev: true, basePath: '/' }),
    ).toBe('http://localhost:8000/api')
  })

  it('reaches the dev API from another device on the LAN', () => {
    expect(
      resolveApiBaseUrl({ protocol: 'http:', hostname: '192.168.1.9', port: '5173', isDev: true, basePath: '/' }),
    ).toBe('http://192.168.1.9:8000/api')
  })

  it('lets an explicit override win and strips any trailing slash', () => {
    expect(
      resolveApiBaseUrl({ ...prod, basePath: '/cars/', override: 'https://api.other.com/v1/' }),
    ).toBe('https://api.other.com/v1')
  })

  it('never produces a doubled or missing slash before api', () => {
    for (const basePath of ['/', '/cars/', '/deep/nested/']) {
      const url = resolveApiBaseUrl({ ...prod, basePath })
      expect(url).toMatch(/\/api$/)
      expect(url).not.toMatch(/\/\/api$/)
    }
  })
})
