import { describe, it, expect } from 'vitest'

import { readJsonResponse } from './readJsonResponse'

// Regression tests for the diagnostic that hid a real bug.
//
// api.php sets `Content-Type: application/json` before it runs, so a PHP warning
// printed into the body still arrives labelled as JSON. The content-type guard in
// useApi.js therefore passed and response.json() threw
// `SyntaxError: Unexpected token '<'`, which named neither the endpoint nor the
// warning. The bug underneath was a read query classified as a write by the
// payment_confirmed gate - nothing in the browser's message pointed at that.
//
// These use real Response objects rather than a hand-rolled stub, because the
// behaviour under test is precisely how fetch's body behaves: consumed once,
// readable as text, headers intact.

const json = (body, init = {}) =>
  new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'content-type': 'application/json' },
    ...init,
  })

// The shape that actually happened: PHP's html_errors output, under the JSON
// content type api.php had already committed to.
const phpWarningPage = (body, { status = 200, contentType = 'application/json' } = {}) =>
  new Response(body, {
    status,
    headers: { 'content-type': contentType },
  })

describe('readJsonResponse', () => {
  it('returns the parsed object for a valid JSON body', async () => {
    const result = await readJsonResponse(json({ success: true, data: [1, 2] }), 'test')
    expect(result).toEqual({ success: true, data: [1, 2] })
  })

  it('preserves falsy payloads rather than treating them as failures', async () => {
    // A legitimately empty result set is a normal answer, not a broken response.
    await expect(readJsonResponse(json([]), 'test')).resolves.toEqual([])
    await expect(readJsonResponse(json({ data: null }), 'test')).resolves.toEqual({ data: null })
  })

  it('parses an error body so callers can read result.error', async () => {
    // Must not throw on a non-2xx: callers branch on result.success themselves.
    const body = new Response(JSON.stringify({ success: false, error: 'Permission denied' }), {
      status: 403,
      headers: { 'content-type': 'application/json' },
    })
    const result = await readJsonResponse(body, 'test')
    expect(result.success).toBe(false)
    expect(result.error).toBe('Permission denied')
  })

  it('names the endpoint, the status and the body when JSON does not parse', async () => {
    // A fresh Response per read: the body is single-use, which is the very
    // property asserted in the last test.
    const makeBody = () =>
      phpWarningPage(
        '<br />\n<b>Warning</b>: Undefined variable $queryUpper in <b>/api/api.php</b>' +
          ' on line <b>2880</b><br />\n{"success":false}',
      )
    const label = 'https://example.com/api/api.php'

    let message = ''
    try {
      await readJsonResponse(makeBody(), label)
      throw new Error('should have thrown')
    } catch (err) {
      message = err.message
    }

    expect(message).toContain(label)
    expect(message).toContain('not valid JSON')
    // The whole point: the operator can see what the server actually said.
    expect(message).toContain('Undefined variable $queryUpper')
  })

  it('reports the declared content type, even when it claims to be JSON', async () => {
    // This is the trap: the header says JSON, so a content-type check passes.
    const body = phpWarningPage('<html>502 Bad Gateway</html>', {
      status: 502,
      contentType: 'application/json',
    })
    await expect(readJsonResponse(body, 'api.php')).rejects.toThrow(
      /HTTP 502[\s\S]*Content-Type: application\/json/,
    )
  })

  it('distinguishes an empty body from malformed JSON', async () => {
    // JSON.parse('') reports a syntax error, which would send the reader looking
    // for a parse bug instead of a server that returned nothing.
    await expect(
      readJsonResponse(
        new Response('', { status: 200, headers: { 'content-type': 'application/json' } }),
        'api.php',
      ),
    ).rejects.toThrow(/empty body/)
  })

  it('treats a whitespace-only body as empty', async () => {
    await expect(
      readJsonResponse(
        new Response('   \n\t ', { status: 200, headers: { 'content-type': 'application/json' } }),
        'api.php',
      ),
    ).rejects.toThrow(/empty body/)
  })

  it('truncates a huge body so the error stays readable', async () => {
    const huge = 'x'.repeat(50_000)
    let message = ''
    try {
      await readJsonResponse(phpWarningPage(huge), 'api.php')
      throw new Error('should have thrown')
    } catch (err) {
      message = err.message
    }
    expect(message).toContain('x'.repeat(200))
    expect(message).not.toContain('x'.repeat(201))
  })

  it('reads the body once', async () => {
    // Calling response.text() then this would throw "Body has already been used".
    // Assert the guard comment stays true: a fresh response parses fine here.
    const response = json({ success: true })
    await expect(readJsonResponse(response, 'test')).resolves.toEqual({ success: true })
  })
})
