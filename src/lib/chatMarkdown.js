// Minimal, dependency-free Markdown renderer for chat message bodies.
//
// Why not `marked` / `markdown-it` + DOMPurify: message text is untrusted (it is
// whatever a user or the agent typed) and it is about to go through v-html, so
// whatever renders it owns the escaping. That is a security boundary, and a
// renderer that "renders HTML, then sanitises the result" puts the safety of the
// whole app behind a blocklist in a dependency this repo does not otherwise use.
//
// This does the opposite, and the difference is the whole reason it is hand-rolled:
//
//   escape FIRST, then recognise Markdown.
//
// Every character is escaped before any parsing (escapeHtml, above), so `<` from
// the user is already `&lt;` before any rule runs, and every tag this file emits
// is one of a fixed, hand-written set (p, br, strong, em, code, pre, h1-h6, ul,
// ol, li, blockquote, hr, a, img, div, table, thead, tbody, tr, th, td). There is
// no code path that can output a tag or
// attribute the author did not write here - so the XSS surface is not "did I
// sanitise it right", it is "is this list of tags safe", which is answerable by
// reading the file.
//
// The one place user input reaches an attribute is a link/image URL, and that is
// scheme-checked (safeUrl) so `javascript:` cannot become a live href.
//
// Supported: headings, bold/italic (incl. ***x***), strikethrough, inline code,
// fenced code blocks, bullet/numbered lists, blockquotes, horizontal rules,
// pipe tables (with `:--`/`--:`/`:--:` alignment), links, autolinks, images
// (http/https only), hard line breaks.
// Not supported (by design): raw HTML, nested lists, footnotes, tables without
// a delimiter row.

import { escapeHtml } from '../utils/escapeHtml'

// URL schemes that are safe to put in an href/src. Anything else that *looks*
// like a scheme (contains a `:`) is dropped rather than linked, which is what
// keeps `javascript:`, `data:` and `vbscript:` from becoming clickable.
const SAFE_SCHEME = /^(https?:|mailto:|tel:)/i
const ANY_SCHEME = /^[a-z][a-z0-9+.-]*:/i

// `:` is not escaped by escapeHtml, so a scheme written by the user still reads
// as a scheme here (no entity can hide the colon: `&#58;` arrives as
// `&amp;#58;` and decodes, once, back to the literal text).
function safeUrl(url) {
  const trimmed = String(url).trim()
  // Scheme detection runs against a copy with the neat-trick entities undone.
  // escapeHtml has already turned `&` into `&amp;`, so `&#58;` reaches the DOM as
  // a literal `&#58;` and cannot actually become a colon there - but decoding it
  // here keeps the rule ("a colon anywhere means a scheme") independent of that
  // subtlety, so `javascript&#58;` is rejected outright rather than by luck.
  const probe = trimmed
    .replace(/&amp;/gi, '&')
    .replace(/&#0*58;/gi, ':')
    .replace(/&#x0*3a;/gi, ':')
    .replace(/&colon;/gi, ':')
  if (SAFE_SCHEME.test(probe)) return trimmed
  if (ANY_SCHEME.test(probe)) return ''
  return trimmed
}

// The two phases communicate through \u0000-delimited placeholders: index N in
// the stash is substituted back in at the end. Inline spans are protected from
// the emphasis passes (so `*a*b*` inside a code span or URL is left alone) and
// whole blocks are protected from the paragraph join.
function inline(input) {
  const stash = []
  const put = (html) => `\u0000I${stash.push(html) - 1}\u0000`
  let text = input

  // Inline code first: it must win over every other pattern (`` `**x**` ``).
  text = text.replace(/`([^`\n]+)`/g, (_, code) => put(`<code class="md-code">${code}</code>`))

  // Links and images. The URL runs to the first space or ")", so titles are not
  // supported; the label is already escaped.
  text = text.replace(/!?\[([^\]\n]*)\]\(([^)\s]+)\)/g, (match, label, url) => {
    const href = safeUrl(url)
    if (!href) return match
    if (match.startsWith('!')) {
      // Only remote images; a data: image is a scheme we drop, and a relative one
      // would resolve against the tenant mount point unpredictably.
      if (!/^https?:\/\//i.test(href)) return match
      return put(`<img class="md-image" src="${href}" alt="${label}" loading="lazy" />`)
    }
    return put(
      `<a class="md-link" href="${href}" target="_blank" rel="noopener noreferrer">${label}</a>`,
    )
  })

  // Bare URLs. Trailing sentence punctuation stays outside the link.
  text = text.replace(/(^|[\s(])(https?:\/\/[^\s<)\]]+)/g, (_, pre, url) => {
    const trimmed = url.replace(/[.,;:!?]+$/, '')
    const tail = url.slice(trimmed.length)
    return (
      pre +
      put(
        `<a class="md-link" href="${trimmed}" target="_blank" rel="noopener noreferrer">${trimmed}</a>`,
      ) +
      tail
    )
  })

  // Emphasis. Triple before double before single so `***x***` is not eaten a
  // star at a time; `[^*\n]` keeps each pass from matching across lines.
  text = text.replace(/\*\*\*([^*\n]+)\*\*\*/g, '<strong><em>$1</em></strong>')
  text = text.replace(/\*\*([^*\n]+)\*\*/g, '<strong>$1</strong>')
  text = text.replace(/__([^_\n]+)__/g, '<strong>$1</strong>')
  text = text.replace(/~~([^~\n]+)~~/g, '<del>$1</del>')
  text = text.replace(/\*([^*\n]+)\*/g, '<em>$1</em>')
  // `_` only when it is not surrounded by word characters, so file_name_here is
  // left as a filename rather than becoming file<em>name</em>here.
  text = text.replace(/(^|[^\w])_([^_\n]+)_(?![A-Za-z0-9])/g, '$1<em>$2</em>')

  return text.replace(/\u0000I(\d+)\u0000/g, (_, i) => stash[Number(i)])
}

// Split a `|`-separated table row into trimmed cells, or return null when the
// line holds no pipe at all (so a bare `---` horizontal rule is never a table).
function tableCells(line) {
  const trimmed = line.trim()
  if (!trimmed.includes('|')) return null
  return trimmed
    .replace(/^\|/, '')
    .replace(/\|$/, '')
    .split('|')
    .map((cell) => cell.trim())
}

// `:--`, `--:` and `:--:` are the alignment markers of a delimiter cell.
function tableAlign(delimCell) {
  const left = delimCell.startsWith(':')
  const right = delimCell.endsWith(':')
  if (left && right) return 'center'
  if (right) return 'right'
  if (left) return 'left'
  return ''
}

// Parse the Markdown table starting at lines[i], or null when there is none.
// A table needs a header row immediately followed by a delimiter row (`---`,
// `:--`, `--:`, `:--:`) with the same number of cells; body rows follow until a
// blank line or a row with no pipe. `end` is the index after the last row.
function parseTable(lines, i) {
  const header = tableCells(lines[i])
  if (!header || i + 1 >= lines.length) return null
  const delim = tableCells(lines[i + 1])
  if (!delim || delim.length !== header.length) return null
  if (!delim.every((cell) => /^:?-+:?$/.test(cell))) return null
  const align = delim.map(tableAlign)
  const rows = []
  let end = i + 2
  while (end < lines.length && lines[end].trim() !== '') {
    const cells = tableCells(lines[end])
    if (!cells) break
    while (cells.length < header.length) cells.push('')
    rows.push(cells.slice(0, header.length))
    end++
  }
  return { header, align, rows, end }
}

// Render a table parsed by parseTable. Cells go through inline() so bold, code
// and links still work inside a cell; the wrapper lets a wide table scroll on a
// phone instead of breaking the message column.
function renderTable({ header, align, rows }) {
  const attr = (a) => (a ? ` class="md-align-${a}"` : '')
  const head = header.map((cell, c) => `<th${attr(align[c])}>${inline(cell)}</th>`).join('')
  const body = rows
    .map(
      (row) =>
        '<tr>' +
        row.map((cell, c) => `<td${attr(align[c])}>${inline(cell)}</td>`).join('') +
        '</tr>',
    )
    .join('')
  return (
    '<div class="md-table-wrap"><table class="md-table">' +
    `<thead><tr>${head}</tr></thead>` +
    (body ? `<tbody>${body}</tbody>` : '') +
    '</table></div>'
  )
}

// A line that opens a non-paragraph block, so a paragraph run stops before it.
function startsBlock(line) {
  return (
    /^\s*`{3,}/.test(line) ||
    /^\s{0,3}#{1,6}\s+/.test(line) ||
    /^\s*&gt;\s?/.test(line) ||
    /^\s*([-*+]|\d+\.)\s+/.test(line) ||
    /^\s*([-*_])(\s*\1){2,}\s*$/.test(line)
  )
}

/**
 * Render a chat message body as sanitized HTML.
 *
 * The return value is safe to hand to v-html: the input is HTML-escaped before
 * any parsing, and the only tags that can appear in the output are the fixed set
 * this module emits. Always returns a string ('' for empty/nullish input).
 *
 * @param {unknown} input
 * @returns {string}
 */
export function renderMarkdown(input) {
  if (input === null || input === undefined) return ''
  const source = String(input).replace(/\r\n?/g, '\n')
  if (source.trim() === '') return ''

  const escaped = escapeHtml(source)

  const blocks = []
  const putBlock = (html) => `\u0000B${blocks.push(html) - 1}\u0000`

  const lines = escaped.split('\n')
  const out = []
  let i = 0

  while (i < lines.length) {
    const line = lines[i]

    const fence = line.match(/^\s*`{3,}\s*([A-Za-z0-9_+-]*)\s*$/)
    if (fence) {
      const lang = fence[1]
      const code = []
      i++
      while (i < lines.length && !/^\s*`{3,}\s*$/.test(lines[i])) {
        code.push(lines[i])
        i++
      }
      if (i < lines.length) i++ // consume the closing fence
      const cls = lang ? ` class="language-${lang}"` : ''
      out.push(putBlock(`<pre class="md-pre"><code${cls}>${code.join('\n')}</code></pre>`))
      continue
    }

    if (line.trim() === '') {
      i++
      continue
    }

    if (/^\s*([-*_])(\s*\1){2,}\s*$/.test(line)) {
      out.push(putBlock('<hr class="md-hr" />'))
      i++
      continue
    }

    const heading = line.match(/^\s{0,3}(#{1,6})\s+(.*?)\s*$/)
    if (heading) {
      const level = heading[1].length
      out.push(putBlock(`<h${level} class="md-h">${inline(heading[2])}</h${level}>`))
      i++
      continue
    }

    if (/^\s*&gt;\s?/.test(line)) {
      const quote = []
      while (i < lines.length && /^\s*&gt;\s?/.test(lines[i])) {
        quote.push(lines[i].replace(/^\s*&gt;\s?/, ''))
        i++
      }
      out.push(
        putBlock(`<blockquote class="md-quote">${quote.map(inline).join('<br />')}</blockquote>`),
      )
      continue
    }

    const listItem = line.match(/^\s*([-*+]|\d+\.)\s+/)
    if (listItem) {
      const ordered = /\d+\./.test(listItem[1])
      const items = []
      while (i < lines.length) {
        const m = lines[i].match(/^\s*([-*+]|\d+\.)\s+(.*)$/)
        if (!m || /\d+\./.test(m[1]) !== ordered) break
        items.push(`<li>${inline(m[2])}</li>`)
        i++
      }
      const tag = ordered ? 'ol' : 'ul'
      out.push(putBlock(`<${tag} class="md-list">${items.join('')}</${tag}>`))
      continue
    }

    const table = parseTable(lines, i)
    if (table) {
      out.push(putBlock(renderTable(table)))
      i = table.end
      continue
    }

    const para = []
    while (
      i < lines.length &&
      lines[i].trim() !== '' &&
      !startsBlock(lines[i]) &&
      !parseTable(lines, i)
    ) {
      para.push(lines[i])
      i++
    }
    out.push(putBlock(`<p class="md-p">${inline(para.join('\n')).replace(/\n/g, '<br />')}</p>`))
  }

  return out.join('').replace(/\u0000B(\d+)\u0000/g, (_, n) => blocks[Number(n)])
}

export default renderMarkdown
