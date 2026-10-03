import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'

// Guards for the colours write path.
//
// ColorsView.vue and AddColorDialog.vue used to INSERT/UPDATE/DELETE the colours
// table through the generic {query, params} passthrough in api/api.php. That path
// executes whatever SQL it is handed and its only gate is a special case for
// payment_confirmed writes, so the "admin only" v-if on the delete button was the
// entire authorisation check - and role_id came out of localStorage, which the
// user controls. Anyone who could reach the endpoint could delete a colour.
//
// The fix moved those writes to api/actions/colors.php, gated on the api_token
// minted at login. The behavioural tests below cover the translation; the source
// assertions pin the shape of the fix, because the failure mode being guarded
// against is a refactor quietly putting an INSERT back on the open path.

const USE_API = readFileSync(new URL('./useApi.js', import.meta.url), 'utf8')
const COLORS_VIEW = readFileSync(new URL('../views/ColorsView.vue', import.meta.url), 'utf8')
const ADD_COLOR_DIALOG = readFileSync(
  new URL('../components/car-stock/AddColorDialog.vue', import.meta.url),
  'utf8',
)

/** Stands in for vue-i18n: returns the key, and renders params so they are asserted. */
const t = (key, params) => (params ? `${key}:${JSON.stringify(params)}` : key)

async function loadColorErrorText() {
  vi.stubGlobal('window', { location: { href: 'https://example.com/', protocol: 'https:' } })
  vi.stubGlobal('localStorage', { getItem: () => null, setItem: () => {}, removeItem: () => {} })
  const mod = await import('./useApi')
  return mod.colorErrorText
}

describe('colorErrorText', () => {
  let colorErrorText

  beforeEach(async () => {
    colorErrorText = await loadColorErrorText()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.resetModules()
  })

  it('maps a duplicate name to the name message', () => {
    // colors.color and colors.hexa are two independent UNIQUE keys. Saying which
    // one collided is the difference between the user knowing what to change.
    const err = { code: 'color_exists', meta: { field: 'color' } }
    expect(colorErrorText(t, err)).toBe('colorsView.errors.duplicateColor')
  })

  it('maps a duplicate hex to the hex message', () => {
    const err = { code: 'color_exists', meta: { field: 'hexa' } }
    expect(colorErrorText(t, err)).toBe('colorsView.errors.duplicateHexa')
  })

  it('defaults a duplicate with no field to the name message', () => {
    expect(colorErrorText(t, { code: 'color_exists' })).toBe('colorsView.errors.duplicateColor')
  })

  it('passes the reference count into the in-use message', () => {
    // Deleting a colour that buy_details / cars_stock / priorities still points at
    // is refused; the count is what tells the user how much is in the way.
    const err = { code: 'color_in_use', meta: { count: 3, tables: ['buy_details'] } }
    expect(colorErrorText(t, err)).toBe('colorsView.errors.inUseDetail:{"count":3}')
  })

  it('reports zero rather than undefined when the count is missing', () => {
    expect(colorErrorText(t, { code: 'color_in_use' })).toBe(
      'colorsView.errors.inUseDetail:{"count":0}',
    )
  })

  it('translates the shared auth codes to colours wording', () => {
    // Deliberately not buy.detailsTable.adminOnly: that string names buy details,
    // and this screen is about colours.
    expect(colorErrorText(t, { code: 'not_admin' })).toBe('colorsView.errors.notAdmin')
    expect(colorErrorText(t, { code: 'not_authenticated' })).toBe(
      'colorsView.errors.notAuthenticated',
    )
  })

  it('translates the validation codes', () => {
    const expected = {
      color_name_required: 'colorsView.errors.nameRequired',
      color_name_too_long: 'colorsView.errors.nameTooLong',
      color_invalid_hexa: 'colorsView.errors.invalidHexa',
      color_not_found: 'colorsView.errors.notFound',
      color_save_failed: 'colorsView.errors.saveFailed',
      db_unavailable: 'colorsView.errors.dbUnavailable',
      db_schema_outdated: 'colorsView.errors.schemaOutdated',
    }
    for (const [code, key] of Object.entries(expected)) {
      expect(colorErrorText(t, { code })).toBe(key)
    }
  })

  it("returns '' for an unknown code so the caller keeps its own fallback", () => {
    // The documented contract, shared with apiErrorText: an unmapped code must
    // not surface as a raw English string from the server.
    expect(colorErrorText(t, { code: 'something_new' })).toBe('')
    expect(colorErrorText(t, null)).toBe('')
  })

  it('reads code and meta off a thrown Error, not just a result payload', () => {
    // The helpers throw apiFailure(), which copies code/meta onto the Error. If
    // that stopped happening every refusal would degrade to an opaque string.
    const err = Object.assign(new Error('color_exists'), {
      code: 'color_exists',
      meta: { field: 'hexa' },
    })
    expect(colorErrorText(t, err)).toBe('colorsView.errors.duplicateHexa')
  })
})

describe('colours are written through the gated action, not raw SQL', () => {
  it('sends the colours helpers as token-bearing actions', () => {
    for (const action of ['get_colors', 'create_color', 'update_color', 'delete_color']) {
      expect(USE_API).toContain(`action: '${action}'`)
    }
    // requiresAuth is what makes useApi attach localStorage's token
    // (useApi.js: `!data.requiresAuth ? null : ...`), and the whole gate is that
    // token. Omitting it would send no credential at all.
    expect(USE_API).toMatch(/action: 'get_colors', requiresAuth: true/)
    expect(USE_API).toMatch(/action: 'delete_color', id, requiresAuth: true/)
  })

  it('no longer writes the colours table with raw SQL', () => {
    for (const source of [COLORS_VIEW, ADD_COLOR_DIALOG]) {
      expect(source).not.toMatch(/INSERT\s+INTO\s+colors/i)
      expect(source).not.toMatch(/UPDATE\s+colors\s+SET/i)
      expect(source).not.toMatch(/DELETE\s+FROM\s+colors/i)
    }
  })

  it('keeps reading the colours list in ColorsView', () => {
    // Reads stay on the token-gated action too, so the view has one source of
    // truth for the list rather than a raw SELECT beside it.
    expect(COLORS_VIEW).toContain('getColors')
  })

  it('confirms before entering the submit guard, not inside it', () => {
    // Inside the guard the key is held for as long as the modal is open, so every
    // row's button would read "Deleting..." while the user is still being asked.
    // Matches a bare confirm( as well as window.confirm( - the original code used
    // the bare form, and a pattern that only caught the qualified one would pass
    // against the very bug this is pinning.
    expect(COLORS_VIEW).toMatch(
      /const confirmDelete[\s\S]*?(?:window\.)?confirm\s*\([\s\S]*?removeColor\(color\)/,
    )
    expect(COLORS_VIEW).not.toMatch(/guard\('delete',[\s\S]{0,400}?(?:window\.)?confirm\s*\(/)
  })
})
