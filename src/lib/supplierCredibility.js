/**
 * Pure helpers for the supplier credibility feature.
 *
 * Nothing here touches the network, the database or Vue: the shapes are
 * normalized and the risk bands are derived, so the rules are unit-testable and
 * the component stays a rendering concern.
 *
 * The risk cut-offs mirror SUPPLIER_CREDIBILITY_RISK_CUTS in
 * api/actions/supplier_credibility.php. The server derives the same bands when
 * the model omits a risk level; the two must agree or a stored check would be
 * shown in a different colour than it was filed under.
 */

/** Report languages the prompt can ask for; the UI sends the active one. */
export const CREDIBILITY_LANGUAGES = ['en', 'ar', 'fr', 'zh']

export const CREDIBILITY_RISK_LEVELS = ['low', 'medium', 'high']

export const ASSESS_ACTION = 'assess_supplier_credibility'
export const HISTORY_ACTION = 'get_supplier_credibility_checks'
export const LATEST_ACTION = 'get_supplier_credibility_latest'

/** At or above this score is the low-risk band. */
export const LOW_RISK_MIN_SCORE = 70
/** At or above this score is the medium band; below it, high. */
export const MEDIUM_RISK_MIN_SCORE = 40

/** API error code -> i18n key. Anything unmapped falls back to `fallbackKey`. */
export const CREDIBILITY_ERROR_KEYS = {
  not_authenticated: 'supplierCredibility.errors.notAuthenticated',
  not_admin: 'supplierCredibility.errors.notAdmin',
  invalid_credentials: 'supplierCredibility.errors.invalidCredentials',
  supplier_not_found: 'supplierCredibility.errors.supplierNotFound',
  db_unavailable: 'supplierCredibility.errors.dbUnavailable',
  db_schema_outdated: 'supplierCredibility.errors.dbSchemaOutdated',
  // The server has the feature switched off. Not retryable: it needs a config
  // change on the server, so the UI should not offer "try again".
  credibility_disabled: 'supplierCredibility.errors.credibilityDisabled',
  ai_not_configured: 'supplierCredibility.errors.aiNotConfigured',
  ai_timeout: 'supplierCredibility.errors.aiTimeout',
  ai_rate_limited: 'supplierCredibility.errors.aiRateLimited',
  ai_auth_failed: 'supplierCredibility.errors.aiAuthFailed',
  ai_http_error: 'supplierCredibility.errors.aiHttpError',
  ai_empty_response: 'supplierCredibility.errors.aiEmptyResponse',
  ai_bad_response: 'supplierCredibility.errors.aiBadResponse',
}

/**
 * Failures worth offering a "try again" for, as opposed to asking an admin to
 * change something. A rate limit clears on its own and an empty answer is
 * usually a busy model, so the UI says so instead of showing a dead end.
 */
export const CREDIBILITY_RETRYABLE_CODES = [
  'ai_rate_limited',
  'ai_timeout',
  'ai_empty_response',
  'ai_http_error',
]

/** submit-guard key for one supplier, so a row cannot be charged twice at once. */
export const assessKey = (supplierId) => `assess-credibility-${supplierId}`

export const riskLevelFromScore = (score) => {
  const value = Number(score)
  if (!Number.isFinite(value)) return 'medium'
  if (value >= LOW_RISK_MIN_SCORE) return 'low'
  if (value >= MEDIUM_RISK_MIN_SCORE) return 'medium'
  return 'high'
}

const toStr = (value, fallback = '') => (typeof value === 'string' ? value : fallback)

const toInt = (value, fallback = 0) => {
  const parsed = parseInt(value, 10)
  return Number.isFinite(parsed) ? parsed : fallback
}

/**
 * A stored check in the shape the UI binds to.
 *
 * The API already validates and clamps what the model returned, so this is not
 * about trusting it twice over: it is so one malformed row - a null score, a
 * red_flags column that is not an array - renders as "unknown" instead of
 * throwing inside a template.
 *
 * @returns {{
 *   id: number, supplier_id: number, score: number|null, risk_level: string,
 *   summary: string, red_flags: string[], confidence: string, model: string,
 *   lang: string, username: string|null, checked_at: string
 * }}
 */
export const normalizeCheck = (raw) => {
  const source = raw && typeof raw === 'object' ? raw : {}
  // `Number(null)` is 0, so a missing score has to be spotted before it is
  // parsed: an unrated supplier must read as unknown, not as a score of zero.
  const scoreMissing = source.score === null || source.score === undefined || source.score === ''
  const score = scoreMissing ? NaN : Number(source.score)
  const hasScore = Number.isFinite(score)

  const flags = Array.isArray(source.red_flags)
    ? source.red_flags.filter((flag) => typeof flag === 'string' && flag.trim() !== '')
    : []

  const risk = CREDIBILITY_RISK_LEVELS.includes(source.risk_level)
    ? source.risk_level
    : hasScore
      ? riskLevelFromScore(score)
      : 'medium'

  return {
    id: toInt(source.id),
    supplier_id: toInt(source.supplier_id),
    score: hasScore ? Math.max(0, Math.min(100, Math.round(score))) : null,
    risk_level: risk,
    summary: toStr(source.summary),
    // Anything other than an explicit recollection claim is treated as "I have
    // no information", so a missing or odd value can never read as knowledge.
    court_records: toStr(source.court_records),
    courts_basis: source.courts_basis === 'recollection' ? 'recollection' : 'no_information',
    red_flags: flags,
    confidence: CREDIBILITY_RISK_LEVELS.includes(source.confidence) ? source.confidence : 'low',
    model: toStr(source.model),
    lang: toStr(source.lang, 'en'),
    username: typeof source.username === 'string' ? source.username : null,
    checked_at: toStr(source.checked_at),
  }
}

export const normalizeChecks = (rows) => (Array.isArray(rows) ? rows : []).map(normalizeCheck)

/**
 * Index checks by supplier for the table badge.
 *
 * @returns {Record<number, object>} supplier id -> newest check
 */
export const latestChecksBySupplier = (rows) =>
  normalizeChecks(rows).reduce((bySupplier, check) => {
    if (check.supplier_id > 0 && !bySupplier[check.supplier_id]) {
      bySupplier[check.supplier_id] = check
    }
    return bySupplier
  }, {})

/** The badge/check to show for one supplier, or null when never checked. */
export const latestCheckFor = (bySupplier, supplierId) => bySupplier?.[Number(supplierId)] ?? null

/**
 * Render a stored UTC timestamp.
 *
 * The API writes UTC_TIMESTAMP() as "YYYY-MM-DD HH:MM:SS" with no zone marker,
 * and a browser reads that shape as *local* time - so the hour would be wrong by
 * the server's offset. The 'Z' is what makes it honest.
 *
 * @returns {string} '' when the value is missing or unparseable
 */
export const formatCheckedAt = (value, locale = 'en') => {
  const raw = toStr(value)
  if (!raw) return ''

  const iso = raw.includes('T') ? raw : raw.replace(' ', 'T')
  const date = new Date(/[Zz]|[+-]\d{2}:?\d{2}$/.test(iso) ? iso : `${iso}Z`)
  if (Number.isNaN(date.getTime())) return ''

  try {
    return new Intl.DateTimeFormat(locale || 'en', {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(date)
  } catch {
    return date.toISOString().slice(0, 16).replace('T', ' ')
  }
}

/** The locale the report should be written in, falling back to English. */
export const reportLanguage = (locale) => {
  const short = toStr(locale).toLowerCase().slice(0, 2)
  return CREDIBILITY_LANGUAGES.includes(short) ? short : 'en'
}

/**
 * Turn an API result into a message the reader can act on.
 *
 * @param {(key: string) => string} t
 * @param {{code?: string, error?: string}|null} result
 * @param {string} fallbackKey i18n key used when the code is not one of ours
 * @returns {string} '' when there is nothing to show
 */
export const credibilityErrorText = (t, result, fallbackKey) => {
  if (!result) return ''
  const key = CREDIBILITY_ERROR_KEYS[result.code]
  if (key) return t(key)
  return t(fallbackKey)
}
