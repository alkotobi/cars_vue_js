import { describe, it, expect, vi } from 'vitest'
import { useBrands } from './useBrands'

// The brand logo upload failed silently: addBrand/updateBrand read the upload
// response as `uploadResult.path`, but uploadFile returns `relativePath` (see
// useApi.js). `path` was undefined, so logoPath was undefined, the INSERT/UPDATE
// took the no-logo branch, and logo_path was stored NULL while the image sat on
// disk as an orphan. Same class of bug already fixed in useCarNameMedia.js.
describe('useBrands logo upload', () => {
  const setup = ({
    uploadResult = { success: true, relativePath: 'uploads/logo.png' },
    safeResult = { success: true },
  } = {}) => {
    const calls = { upload: null, queries: [] }
    const notify = vi.fn()
    const api = useBrands({ notify, onBrandsChanged: () => {} })({
      safeApi: async (opts) => {
        calls.queries.push(opts)
        if (opts?.query?.startsWith('INSERT')) return safeResult
        return { success: true, data: [] }
      },
      uploadFile: async (file) => {
        calls.upload = file
        return uploadResult
      },
      getFileUrl: (p) => p,
    })
    return { api, calls, notify }
  }

  it('stores the logo path from relativePath, not .path', async () => {
    const { api, calls } = setup()
    const res = await api.addBrand({ brand: 'Toyota', logoFile: { name: 'l.png' } })
    expect(res.success).toBe(true)
    const insert = calls.queries.find((q) => q.query.startsWith('INSERT'))
    expect(insert.query).toContain('INSERT INTO brands (brand, logo_path)')
    expect(insert.params[1]).toBe('uploads/logo.png')
    expect(insert.params[1]).not.toBeNull()
    expect(insert.params[1]).not.toBeUndefined()
  })

  it('inserts without logo_path when no logo is chosen', async () => {
    const { api, calls } = setup()
    await api.addBrand({ brand: 'Toyota' })
    const insert = calls.queries.find((q) => q.query.startsWith('INSERT'))
    expect(insert.query).toBe('INSERT INTO brands (brand) VALUES (?)')
    expect(insert.params).toEqual(['Toyota'])
  })

  it('updateBrand also stores the path from relativePath', async () => {
    const { api, calls } = setup({
      uploadResult: { success: true, relativePath: 'uploads/new.png' },
    })
    await api.updateBrand({ brand: 'Toyota', id: 5, logoFile: { name: 'n.png' } })
    const update = calls.queries.find(
      (q) => q.query.startsWith('UPDATE brands') && q.query.includes('logo_path'),
    )
    expect(update.params[1]).toBe('uploads/new.png')
  })
})
