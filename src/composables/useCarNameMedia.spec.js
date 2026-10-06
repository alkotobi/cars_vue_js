import { describe, it, expect, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { useCarNameMedia } from './useCarNameMedia'
import { IMAGE_EXTENSIONS, VIDEO_EXTENSIONS } from '../lib/carNameMedia'

// The car-name media uploader failed completely and silently.
//
// Four separate faults lined up, and each one on its own would have been reported if
// any of them had been visible - which is the actual bug. The dialog built its own
// useNotify() instance, but the <MessageBox> that renders that state lives in the
// parent view, so notify() set show = true on a ref nothing was bound to. Every
// failure below therefore reached a caller that discarded it, and the dialog simply
// closed looking unchanged.
//
// What was underneath, in the order they bite:
//
//   1. uploadFile(file, destinationFolder, customFilename) was called with the folder
//      AND the filename concatenated into argument two. uploadFile read that as a
//      folder, so it created files/uploads/car_names/<id>/<uuid>.png/ as a directory
//      and stored a server-named file inside it. Real directories on disk confirmed it.
//   2. The response was read as uploadRes.path. uploadFile returns `relativePath`.
//      `path` was undefined, so file_path went to the INSERT as SQL NULL.
//   3. car_name_media.file_path is NOT NULL, so that INSERT was refused every time. The
//      file was on disk, the row was not, and fetchMedia() then found nothing to show.
//   4. uploaded_by was the literal 1, crediting whichever user happens to hold id 1.
//
// The observable symptom - "nothing happens" - is why each of these is pinned here.

const COMPOSABLE = readFileSync(new URL('./useCarNameMedia.js', import.meta.url), 'utf8')

const mediaTypeFor = (ext) => (['mp4', 'webm', 'avi', 'mov'].includes(ext) ? 'video' : 'photo')

// A File stand-in: only name, size and type are read.
const fakeFile = (name, size = 1024, type = 'image/png') => ({ name, size, type })

describe('useCarNameMedia.upload', () => {
  let calls

  const setup = ({ uploadResult, safeResult = { success: true }, user = { id: 7 } } = {}) => {
    calls = { upload: null, inserts: [] }

    const m = useCarNameMedia({ carName: { id: 3, car_name: 'Hilux' } }).withDeps({
      safeApi: async (opts) => {
        if (opts?.query?.startsWith('INSERT')) {
          calls.inserts.push(opts)
          return safeResult
        }
        return { success: true, data: [] }
      },
      uploadFile: async (file, folder, filename) => {
        calls.upload = { file, folder, filename }
        return uploadResult
      },
      getFileUrl: (p) => p,
    })

    return m
  }

  beforeEach(() => {
    // The suite runs in the node environment (vite.config.js test.environment), so
    // there is no localStorage - the same stub useSessionLost.spec.js installs.
    const store = new Map()
    vi.stubGlobal('localStorage', {
      getItem: (k) => (store.has(k) ? store.get(k) : null),
      setItem: (k, v) => store.set(k, String(v)),
      removeItem: (k) => store.delete(k),
      clear: () => store.clear(),
    })
  })

  it('passes the folder and the filename as separate arguments', () => {
    const m = setup({ uploadResult: { success: true, relativePath: 'uploads/car_names/3/a.png' } })
    localStorage.setItem('user', JSON.stringify({ id: 7, username: 'ali' }))

    m.handleFileSelect([fakeFile('photo.png')])
    return m.upload().then(() => {
      // The filename must be custom_filename (3rd), never part of the folder (2nd).
      // Concatenting them is what produced a directory ending in .png.
      expect(calls.upload.folder).toBe('uploads/car_names/3')
      expect(calls.upload.folder.endsWith('.png')).toBe(false)
      expect(calls.upload.filename).toMatch(/^[0-9a-f-]{36}\.png$/)
    })
  })

  it('records relativePath from the response, not .path', () => {
    const m = setup({ uploadResult: { success: true, relativePath: 'uploads/car_names/3/a.png' } })
    localStorage.setItem('user', JSON.stringify({ id: 7, username: 'ali' }))

    m.handleFileSelect([fakeFile('photo.png')])
    return m.upload().then(() => {
      const params = calls.inserts[0].params
      // index 1 is file_path. It must be the stored path, never null/undefined -
      // the column is NOT NULL and refused the row when it was.
      expect(params[1]).toBe('uploads/car_names/3/a.png')
      expect(params[1]).not.toBeNull()
      expect(params[1]).not.toBeUndefined()
    })
  })

  it('records the signed-in user as uploaded_by, not the literal 1', () => {
    const m = setup({ uploadResult: { success: true, relativePath: 'uploads/car_names/3/a.png' } })
    localStorage.setItem('user', JSON.stringify({ id: 42, username: 'someone' }))

    m.handleFileSelect([fakeFile('photo.png')])
    return m.upload().then(() => {
      const params = calls.inserts[0].params
      expect(params[6]).toBe(42)
      expect(params[6]).not.toBe(1)
    })
  })

  it('stores file_size, which the column has been there for', () => {
    const m = setup({ uploadResult: { success: true, relativePath: 'uploads/car_names/3/a.png' } })
    localStorage.setItem('user', JSON.stringify({ id: 7 }))

    m.handleFileSelect([fakeFile('photo.png', 2048)])
    return m.upload().then(() => {
      expect(calls.inserts[0].params[3]).toBe(2048)
    })
  })

  it('refuses to upload when there is no session, instead of crediting user 1', () => {
    const m = setup({ uploadResult: { success: true, relativePath: 'x.png' } })
    localStorage.setItem('user', JSON.stringify({ username: 'no-id' }))

    m.handleFileSelect([fakeFile('photo.png')])
    return m.upload().then((res) => {
      expect(res.success).toBe(false)
      expect(calls.upload).toBeNull()
      expect(calls.inserts).toHaveLength(0)
    })
  })

  it('reports a failed upload rather than inserting a broken row', () => {
    // relativePath missing is what the old `uploadRes.path` bug looked like from here.
    const m = setup({ uploadResult: { success: true } })
    localStorage.setItem('user', JSON.stringify({ id: 7 }))

    m.handleFileSelect([fakeFile('photo.png')])
    return m.upload().then((res) => {
      expect(res.success).toBe(false)
      expect(res.failCount).toBe(1)
      expect(calls.inserts).toHaveLength(0)
    })
  })
})

describe('car-name media dialog, statically', () => {
  const DIALOG = readFileSync(
    new URL('../components/car-names/CarNameMediaDialog.vue', import.meta.url),
    'utf8'
  )

  it('takes notify as a prop instead of building an instance nothing renders', () => {
    // A dialog-local useNotify() has nowhere to be displayed: the <MessageBox> lives
    // in the parent view, so every message it produced was discarded.
    //
    // Comments are stripped first. The explanatory comment in this file names
    // useNotify() to say why it is gone, and matching raw source would make this test
    // fail on the very documentation that describes the fix.
    const code = DIALOG.replace(/\/\/[^\n]*/g, '').replace(/\/\*[\s\S]*?\*\//g, '')
    const script = code.slice(code.indexOf('<script setup>'))

    expect(script).not.toMatch(/useNotify\s*\(\s*\)/)
    expect(script).toMatch(/notify:\s*\{\s*type:\s*Function/)
  })

  it('no longer advertises SVG, which upload.php refuses on purpose', () => {
    expect(DIALOG).not.toMatch(/WEBP,\s*SVG/)
  })

  it('carries the parents notify through', () => {
    const VIEW = readFileSync(new URL('../views/CarNamesView.vue', import.meta.url), 'utf8')
    const tag = VIEW.match(/<CarNameMediaDialog[\s\S]*?\/>/)[0]
    expect(tag).toMatch(/:notify="notify"/)
  })
})

describe('carNameMedia extensions', () => {
  it('offers nothing upload.php refuses', () => {
    // Direction matters: the offered list is allowed to be a subset of what the server
    // stores (bmp is storable but was never offered here), but it must not include
    // anything the server rejects. svg was exactly that - offered here, excluded there
    // because an SVG served inline is same-origin script.
    const UPLOAD_PHP = readFileSync(new URL('../../api/upload.php', import.meta.url), 'utf8')
    const block = UPLOAD_PHP.match(/const UPLOAD_CONTENT_TYPES = \[([\s\S]*?)\];/)[1]
    const stored = [...block.matchAll(/'(\w+)'\s*=>/g)].map((m) => m[1])

    for (const ext of [...IMAGE_EXTENSIONS, ...VIDEO_EXTENSIONS]) {
      expect(stored).toContain(ext)
    }
    expect(IMAGE_EXTENSIONS).not.toContain('svg')
  })
})
