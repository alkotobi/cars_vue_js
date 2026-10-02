import { describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import {
  LOW_RISK_MIN_SCORE,
  MEDIUM_RISK_MIN_SCORE,
  CREDIBILITY_ERROR_KEYS,
  CREDIBILITY_RETRYABLE_CODES,
  assessKey,
  credibilityErrorText,
  formatCheckedAt,
  latestCheckFor,
  latestChecksBySupplier,
  normalizeCheck,
  normalizeChecks,
  reportLanguage,
  riskLevelFromScore,
} from './supplierCredibility'

// The shape supplier_credibility_row() in api/actions/supplier_credibility.php
// actually returns: the DB column is id_supplier, the API renames it supplier_id.
const check = (over = {}) => ({
  id: 7,
  supplier_id: 3,
  score: 62,
  risk_level: 'medium',
  summary: 'Contact is an email only.',
  court_records: 'No way to check, and I recall nothing.',
  courts_basis: 'recollection',
  red_flags: ['No verifiable company details'],
  confidence: 'low',
  model: 'gpt-4o-mini',
  lang: 'en',
  username: null,
  checked_at: '2026-09-30 08:15:00',
  ...over,
})

describe('riskLevelFromScore', () => {
  it('bands the same cut-offs as the server', () => {
    expect(riskLevelFromScore(LOW_RISK_MIN_SCORE)).toBe('low')
    expect(riskLevelFromScore(LOW_RISK_MIN_SCORE - 1)).toBe('medium')
    expect(riskLevelFromScore(MEDIUM_RISK_MIN_SCORE)).toBe('medium')
    expect(riskLevelFromScore(MEDIUM_RISK_MIN_SCORE - 1)).toBe('high')
  })

  it('treats out-of-range and non-numeric scores as medium', () => {
    expect(riskLevelFromScore(140)).toBe('low')
    expect(riskLevelFromScore(-5)).toBe('high')
    expect(riskLevelFromScore(undefined)).toBe('medium')
    expect(riskLevelFromScore('abc')).toBe('medium')
  })
})

describe('normalizeCheck', () => {
  it('maps the stored row onto the shape the UI binds to', () => {
    expect(normalizeCheck(check())).toEqual({
      id: 7,
      supplier_id: 3,
      score: 62,
      risk_level: 'medium',
      summary: 'Contact is an email only.',
      court_records: 'No way to check, and I recall nothing.',
      courts_basis: 'recollection',
      red_flags: ['No verifiable company details'],
      confidence: 'low',
      model: 'gpt-4o-mini',
      lang: 'en',
      username: null,
      checked_at: '2026-09-30 08:15:00',
    })
  })

  it('keeps a null score as unknown rather than pretending it is zero', () => {
    const result = normalizeCheck(check({ score: null }))
    expect(result.score).toBeNull()
    // No score to band, so it falls back to the cautious default.
    expect(result.risk_level).toBe('medium')
  })

  it('derives the risk level when the stored one is not one of the three', () => {
    expect(normalizeCheck(check({ risk_level: 'catastrophic', score: 90 })).risk_level).toBe('low')
  })

  it('clamps a score outside 0-100', () => {
    expect(normalizeCheck(check({ score: 250 })).score).toBe(100)
    expect(normalizeCheck(check({ score: -12 })).score).toBe(0)
  })

  it('drops red flags that are not non-empty strings', () => {
    const result = normalizeCheck(check({ red_flags: ['real', '', null, 7, '  ', 'also real'] }))
    expect(result.red_flags).toEqual(['real', 'also real'])
  })

  it('survives a row that is not an object at all', () => {
    const result = normalizeCheck(undefined)
    expect(result.score).toBeNull()
    expect(result.red_flags).toEqual([])
    expect(result.summary).toBe('')
    expect(result.risk_level).toBe('medium')
    expect(result.confidence).toBe('low')
  })

  it('defaults an unknown confidence to low', () => {
    expect(normalizeCheck(check({ confidence: 'certain' })).confidence).toBe('low')
  })
})

describe('normalizeChecks', () => {
  it('returns an empty list for a missing or non-array value', () => {
    expect(normalizeChecks(null)).toEqual([])
    expect(normalizeChecks('nope')).toEqual([])
  })
})

describe('latestChecksBySupplier', () => {
  it('indexes the newest check per supplier', () => {
    const rows = [
      check({ id: 1, supplier_id: 3 }),
      check({ id: 2, supplier_id: 4, score: 90 }),
      check({ id: 3, supplier_id: 3, score: 10 }),
    ]
    const bySupplier = latestChecksBySupplier(rows)

    expect(Object.keys(bySupplier).sort()).toEqual(['3', '4'])
    // The endpoint already returns one row per supplier, newest first; the first
    // one seen wins so a later duplicate cannot overwrite it.
    expect(bySupplier[3].id).toBe(1)
    expect(bySupplier[4].score).toBe(90)
  })

  it('ignores rows without a supplier id', () => {
    expect(latestChecksBySupplier([check({ supplier_id: 0 })])).toEqual({})
  })
})

describe('latestCheckFor', () => {
  it('looks a supplier up by id, accepting the string form a table row carries', () => {
    const bySupplier = latestChecksBySupplier([check()])
    expect(latestCheckFor(bySupplier, '3').id).toBe(7)
    expect(latestCheckFor(bySupplier, 99)).toBeNull()
    expect(latestCheckFor(null, 3)).toBeNull()
  })
})

describe('formatCheckedAt', () => {
  it('reads the API timestamp as UTC, not as local time', () => {
    // 08:15 UTC must not render as 08:15 for a viewer in any other zone, and the
    // 'Z' is what stops a browser from assuming the server's local time.
    const formatted = formatCheckedAt('2026-09-30 08:15:00', 'en-GB')
    expect(formatted).toContain('2026')
    expect(formatted).not.toContain('Invalid')
  })

  it('honours a value that already carries a zone', () => {
    expect(formatCheckedAt('2026-09-30T08:15:00Z', 'en-GB')).toContain('2026')
    expect(formatCheckedAt('2026-09-30T08:15:00+02:00', 'en-GB')).toContain('2026')
  })

  it('returns an empty string for a missing or unparseable timestamp', () => {
    expect(formatCheckedAt('', 'en')).toBe('')
    expect(formatCheckedAt(null, 'en')).toBe('')
    expect(formatCheckedAt('not a date', 'en')).toBe('')
  })

  it('does not throw on an unknown locale tag', () => {
    expect(() => formatCheckedAt('2026-09-30 08:15:00', 'not-a-locale')).not.toThrow()
  })
})

describe('reportLanguage', () => {
  it('passes the supported languages through', () => {
    expect(reportLanguage('ar')).toBe('ar')
    expect(reportLanguage('zh')).toBe('zh')
  })

  it('falls back to English for anything unsupported', () => {
    expect(reportLanguage('de')).toBe('en')
    expect(reportLanguage('')).toBe('en')
    expect(reportLanguage(undefined)).toBe('en')
  })
})

describe('assessKey', () => {
  it('gives each supplier its own submit-guard key', () => {
    expect(assessKey(3)).toBe('assess-credibility-3')
    expect(assessKey(3)).not.toBe(assessKey(4))
  })
})

describe('credibilityErrorText', () => {
  const t = vi.fn((key) => `translated:${key}`)

  it('translates a known code', () => {
    expect(credibilityErrorText(t, { code: 'ai_not_configured' }, 'fallback')).toBe(
      'translated:supplierCredibility.errors.aiNotConfigured',
    )
  })

  it('falls back for a code it does not own, so another feature can reuse this', () => {
    expect(credibilityErrorText(t, { code: 'detail_locked' }, 'fallback')).toBe(
      'translated:fallback',
    )
  })

  it('says nothing when there is no error', () => {
    expect(credibilityErrorText(t, null, 'fallback')).toBe('')
  })
})

describe('CREDIBILITY_RETRYABLE_CODES', () => {
  it('offers a retry for failures that clear on their own', () => {
    expect(CREDIBILITY_RETRYABLE_CODES).toContain('ai_rate_limited')
    expect(CREDIBILITY_RETRYABLE_CODES).toContain('ai_timeout')
    expect(CREDIBILITY_RETRYABLE_CODES).toContain('ai_empty_response')
  })

  it('does not offer a retry for failures only a person can fix', () => {
    // Retrying these cannot help, and a button that never works is worse than
    // no button: it reads as a broken app rather than a misconfiguration.
    expect(CREDIBILITY_RETRYABLE_CODES).not.toContain('ai_auth_failed')
    expect(CREDIBILITY_RETRYABLE_CODES).not.toContain('ai_not_configured')
    expect(CREDIBILITY_RETRYABLE_CODES).not.toContain('not_authenticated')
    expect(CREDIBILITY_RETRYABLE_CODES).not.toContain('not_admin')
  })

  it('only names codes it can also translate', () => {
    for (const code of CREDIBILITY_RETRYABLE_CODES) {
      expect(CREDIBILITY_ERROR_KEYS[code]).toBeDefined()
    }
  })
})

describe('error codes with their own wording', () => {
  const t = vi.fn((key) => `translated:${key}`)

  // Each of these three was seen against a real endpoint, and each has a
  // different fix, which is why a generic "something went wrong" is not enough.
  it.each([
    ['ai_rate_limited', 'aiRateLimited'],
    ['ai_auth_failed', 'aiAuthFailed'],
    ['ai_empty_response', 'aiEmptyResponse'],
    ['db_schema_outdated', 'dbSchemaOutdated'],
  ])('maps %s to its own message', (code, key) => {
    expect(credibilityErrorText(t, { code }, 'fallback')).toBe(
      `translated:supplierCredibility.errors.${key}`,
    )
  })
})

describe('db_schema_outdated', () => {
  // A server deployed before its migrations report this instead of a fatal
  // error on the missing column, so the person reading it knows to run a
  // migration rather than to retry.
  it('is not offered as retryable', () => {
    expect(CREDIBILITY_RETRYABLE_CODES).not.toContain('db_schema_outdated')
  })

  it('still has its own message', () => {
    const t = vi.fn((key) => `translated:${key}`)
    expect(credibilityErrorText(t, { code: 'db_schema_outdated' }, 'fallback')).toBe(
      'translated:supplierCredibility.errors.dbSchemaOutdated',
    )
  })
})

describe('credibility_disabled', () => {
  // The server has credibility_enabled => '0'. Retrying cannot help: it needs a
  // config change on the server, so offering "try again" would be a dead end.
  it('is not offered as retryable', () => {
    expect(CREDIBILITY_RETRYABLE_CODES).not.toContain('credibility_disabled')
  })

  it('has its own message rather than a generic failure', () => {
    const t = vi.fn((key) => `translated:${key}`)
    expect(credibilityErrorText(t, { code: 'credibility_disabled' }, 'fallback')).toBe(
      'translated:supplierCredibility.errors.credibilityDisabled',
    )
  })
})

describe('every error code resolves in every locale', () => {
  // The other tests here stub the translator, so a key that exists in
  // CREDIBILITY_ERROR_KEYS but not in the locale files passes them all. That is
  // not hypothetical: dbSchemaOutdated was mapped in code and missing from all
  // four locales, so the one error a server hits before applying its migrations
  // rendered as a raw key. This reads the real files instead of a stub.
  const LOCALES = ['en', 'ar', 'fr', 'zh']

  const loadLocale = (locale) => {
    const path = fileURLToPath(new URL(`../locales/${locale}.json`, import.meta.url))
    return JSON.parse(readFileSync(path, 'utf8'))
  }

  const resolve = (tree, key) =>
    key.split('.').reduce((node, part) => (node == null ? undefined : node[part]), tree)

  it.each(LOCALES)('%s defines every key in CREDIBILITY_ERROR_KEYS', (locale) => {
    const messages = loadLocale(locale)
    const missing = Object.values(CREDIBILITY_ERROR_KEYS).filter(
      (key) => resolve(messages, key) === undefined,
    )
    expect(missing).toEqual([])
  })
})

describe('normalizeCheck court answer', () => {
  it('passes the two answers through', () => {
    const c = normalizeCheck({
      id: 1,
      summary: 'Name looks like a registered trading entity.',
      court_records: 'No way to check, and I recall nothing.',
      courts_basis: 'no_information',
    })
    expect(c.summary).toBe('Name looks like a registered trading entity.')
    expect(c.court_records).toBe('No way to check, and I recall nothing.')
    expect(c.courts_basis).toBe('no_information')
  })

  it('keeps an explicit recollection claim', () => {
    expect(normalizeCheck({ courts_basis: 'recollection' }).courts_basis).toBe('recollection')
  })

  it.each([undefined, null, '', 'lookup', 'verified', 42])(
    'falls back to no_information for %p so a stray value never reads as knowledge',
    (value) => {
      expect(normalizeCheck({ courts_basis: value }).courts_basis).toBe('no_information')
    },
  )

  it('survives a row from before the column existed', () => {
    const c = normalizeCheck({ id: 7, score: 50 })
    expect(c.court_records).toBe('')
    expect(c.courts_basis).toBe('no_information')
  })
})
