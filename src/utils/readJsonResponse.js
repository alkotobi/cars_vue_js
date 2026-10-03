// Parse a response body that is supposed to be JSON, reporting a failure usefully.
//
// Why this exists instead of `await response.json()`:
//
// A Content-Type check is not sufficient evidence that a body will parse. api.php
// sends `Content-Type: application/json` on its first line, before it does any
// work, so when it later emits a warning or a fatal the body still arrives
// labelled as JSON. The content-type guard passes, and response.json() throws a
// bare `SyntaxError: Unexpected token '<'` naming neither the endpoint nor the
// server output. That is not hypothetical: a mis-classified query in the
// payment_confirmed gate surfaced in the browser as exactly that, while the real
// fault - a read treated as a write - appeared nowhere in the message.
//
// The declared type cannot vouch for the body that declared it, so parse the text
// and take responsibility for the error here. `label` should name the request, so
// the operator can tell 24 call sites apart from a stack trace alone.
//
// Note this reads the body exactly once. On a path where a caller has already
// consumed the body via .text(), the second read throws "Body has already been
// used" - so do not call this after reading the stream yourself. Callers that
// reject on Content-Type before parsing are fine: they throw, and never get here.
export async function readJsonResponse(response, label) {
  const text = await response.text()

  // A 200 with an empty body is its own failure mode, and JSON.parse('') would
  // report it as a syntax error, which points at the wrong thing entirely.
  if (!text.trim()) {
    throw new Error(
      `${label} returned an empty body (HTTP ${response.status} ${response.statusText}).`.trim(),
    )
  }

  try {
    return JSON.parse(text)
  } catch {
    const contentType = response.headers.get('content-type') || 'none'
    throw new Error(
      `${label} returned a body that is not valid JSON ` +
        `(HTTP ${response.status} ${response.statusText}, Content-Type: ${contentType}). ` +
        `First 200 characters: ${text.slice(0, 200)}`,
    )
  }
}
