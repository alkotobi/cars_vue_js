import { describe, it, expect } from 'vitest'
import { resolveBasePath } from './basePath'

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
