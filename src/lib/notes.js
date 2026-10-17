// Normalises the many shapes a notes cell has taken over time into plain text.
//
// Stored notes are normally a JSON array of entries, e.g.
//   [{ id_user, note, timestamp }, ...]
// but real tenants also hold:
//   - a plain string
//   - a JSON object / array at the root
//   - nested entries, where the `note` field is itself an array or object
//   - double-encoded entries, where the `note` field is a JSON *string* holding
//     another entry array (see mig_26 `sell_bill`)
// Everything is flattened to a list of note strings so the UI never shows raw
// JSON.

const isJsonLike = (text) => {
  const trimmed = text.trim()
  return (
    (trimmed.startsWith('[') && trimmed.endsWith(']')) ||
    (trimmed.startsWith('{') && trimmed.endsWith('}'))
  )
}

// Parse a string that looks like JSON, otherwise return it unchanged. Never throws.
const tryParse = (value) => {
  if (typeof value !== 'string' || !isJsonLike(value)) return value
  try {
    return JSON.parse(value)
  } catch {
    return value
  }
}

// Normalise a raw notes value to an array of entries. Never throws.
export const parseNotes = (notes) => {
  if (notes == null || notes === '') return []
  const parsed = tryParse(notes)
  if (Array.isArray(parsed)) return parsed
  if (parsed && typeof parsed === 'object') return [parsed]
  return [{ id_user: null, note: parsed, timestamp: null }]
}

const collect = (value, out) => {
  if (value == null) return
  if (Array.isArray(value)) {
    value.forEach((item) => collect(item, out))
    return
  }
  if (typeof value === 'object') {
    if ('note' in value) collect(value.note, out)
    return
  }
  const text = String(value)
  if (!text.trim()) return
  // A note may itself be a JSON string (double-encoded tenants).
  const parsed = tryParse(text)
  if (parsed !== text) collect(parsed, out)
  else out.push(text)
}

// Flatten any notes value to the list of note strings it contains.
export const extractNoteTexts = (notes) => {
  const out = []
  collect(parseNotes(notes), out)
  return out
}

// Text of a single entry (used by note-display tables): `{ note: ... }` or a raw value.
export const noteText = (note) => {
  const out = []
  collect(note, out)
  return out.join('\n')
}

// Join the note texts of one cell (and an optional second, e.g. a merged sell-bill
// notes column) into a single newline-separated string.
export const formatNotes = (notes, extra = null) => {
  const out = [...extractNoteTexts(notes)]
  if (extra != null) out.push(...extractNoteTexts(extra))
  return out.filter((text) => text.trim() !== '').join('\n')
}
