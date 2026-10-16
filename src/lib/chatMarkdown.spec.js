import { describe, it, expect } from 'vitest'
import { renderMarkdown } from './chatMarkdown'

// The renderer's one job that matters is the security invariant: whatever the
// user typed, the output can only contain the tags this module emits. The rest
// of these tests pin the formatting so a regression is visible.
describe('renderMarkdown', () => {
  describe('safety', () => {
    it('renders a raw HTML/script payload as inert text', () => {
      const out = renderMarkdown('<script>alert(1)</script>')
      expect(out).not.toContain('<script>')
      expect(out).toContain('&lt;script&gt;')
    })

    it('neutralises an img/onerror break-out', () => {
      const out = renderMarkdown(`<img src=x onerror="alert(1)">`)
      expect(out).not.toContain('<img')
      expect(out).not.toContain('onerror="alert')
    })

    it('drops javascript: links rather than linking them', () => {
      const out = renderMarkdown('[click](javascript:alert(1))')
      expect(out).not.toContain('href')
      expect(out).toContain('[click](javascript:alert(1))')
    })

    it('drops data: links', () => {
      const out = renderMarkdown('[x](data:text/html,<script>alert(1)</script>)')
      expect(out).not.toContain('href=')
    })

    it('does not let an entity hide a javascript: scheme', () => {
      // `&#58;` cannot actually become a colon in the DOM (entities decode once,
      // and escapeHtml leaves a literal `&`), but the scheme probe is hardened to
      // reject it regardless.
      const out = renderMarkdown('[x](javascript&#58;alert(1))')
      expect(out).not.toContain('<a')
    })

    it('keeps only tags from the fixed allow-list', () => {
      const out = renderMarkdown('<div><b>x</b></div> # Yep **no**')
      const tags = out.match(/<\/?[a-z][^>]*>/gi) || []
      const allowed =
        /^<\/?(a|p|br|strong|em|del|code|pre|blockquote|hr|ul|ol|li|h[1-6]|img|div|table|thead|tbody|tr|th|td)(\s|>|\/)/i
      for (const tag of tags) expect(tag).toMatch(allowed)
    })
  })

  describe('formatting', () => {
    it('wraps plain text in a paragraph and keeps non-Latin text intact', () => {
      expect(renderMarkdown('2024 Toyota Land Cruiser')).toBe(
        '<p class="md-p">2024 Toyota Land Cruiser</p>',
      )
      expect(renderMarkdown('مرحبا')).toBe('<p class="md-p">مرحبا</p>')
    })

    it('returns an empty string for empty or nullish input', () => {
      expect(renderMarkdown('')).toBe('')
      expect(renderMarkdown('   ')).toBe('')
      expect(renderMarkdown(null)).toBe('')
      expect(renderMarkdown(undefined)).toBe('')
    })

    it('turns a single newline into a line break and a blank line into two paragraphs', () => {
      expect(renderMarkdown('a\nb')).toBe('<p class="md-p">a<br />b</p>')
      expect(renderMarkdown('a\n\nb')).toBe('<p class="md-p">a</p><p class="md-p">b</p>')
    })

    it('renders bold, italic and strikethrough', () => {
      expect(renderMarkdown('**bold**')).toContain('<strong>bold</strong>')
      expect(renderMarkdown('*italic*')).toContain('<em>italic</em>')
      expect(renderMarkdown('~~gone~~')).toContain('<del>gone</del>')
      expect(renderMarkdown('***both***')).toContain('<strong><em>both</em></strong>')
    })

    it('does not italicise underscores inside a filename', () => {
      expect(renderMarkdown('file_name_here')).toBe('<p class="md-p">file_name_here</p>')
    })

    it('renders inline code without touching its contents', () => {
      expect(renderMarkdown('`**not bold**`')).toBe(
        '<p class="md-p"><code class="md-code">**not bold**</code></p>',
      )
    })

    it('renders a fenced code block, keeping newlines and the language', () => {
      const out = renderMarkdown('```js\nconst a = 1\nconst b = 2\n```')
      expect(out).toBe(
        '<pre class="md-pre"><code class="language-js">const a = 1\nconst b = 2</code></pre>',
      )
    })

    it('escapes markup inside a fenced code block', () => {
      const out = renderMarkdown('```\n<div>\n```')
      expect(out).toContain('&lt;div&gt;')
      expect(out).not.toContain('<div>')
    })

    it('renders headings, horizontal rules and blockquotes', () => {
      expect(renderMarkdown('## Title')).toBe('<h2 class="md-h">Title</h2>')
      expect(renderMarkdown('---')).toBe('<hr class="md-hr" />')
      expect(renderMarkdown('> quoted')).toBe('<blockquote class="md-quote">quoted</blockquote>')
    })

    it('renders unordered and ordered lists', () => {
      expect(renderMarkdown('- a\n- b')).toBe('<ul class="md-list"><li>a</li><li>b</li></ul>')
      expect(renderMarkdown('1. a\n2. b')).toBe('<ol class="md-list"><li>a</li><li>b</li></ol>')
    })

    it('renders a Markdown table with a header and body rows', () => {
      const out = renderMarkdown(
        '| Car | Price |\n| --- | --- |\n| LIVAN | 100 |\n| COOLRAY | 200 |',
      )
      expect(out).toBe(
        '<div class="md-table-wrap"><table class="md-table">' +
          '<thead><tr><th>Car</th><th>Price</th></tr></thead>' +
          '<tbody><tr><td>LIVAN</td><td>100</td></tr>' +
          '<tr><td>COOLRAY</td><td>200</td></tr></tbody>' +
          '</table></div>',
      )
    })

    it('applies column alignment from the delimiter row', () => {
      const out = renderMarkdown('| a | b | c |\n| :-- | :-: | --: |\n| 1 | 2 | 3 |')
      expect(out).toContain('<th class="md-align-left">a</th>')
      expect(out).toContain('<th class="md-align-center">b</th>')
      expect(out).toContain('<th class="md-align-right">c</th>')
    })

    it('renders a table with no body rows', () => {
      const out = renderMarkdown('| h1 | h2 |\n| --- | --- |')
      expect(out).toContain('<thead><tr><th>h1</th><th>h2</th></tr></thead>')
      expect(out).not.toContain('<tbody>')
    })

    it('formats markup inside table cells', () => {
      const out = renderMarkdown('| Car | Note |\n| --- | --- |\n| **LIVAN** | `x` |')
      expect(out).toContain('<td><strong>LIVAN</strong></td>')
      expect(out).toContain('<td><code class="md-code">x</code></td>')
    })

    it('does not treat pipe text without a delimiter row as a table', () => {
      const out = renderMarkdown('a | b\nc | d')
      expect(out).not.toContain('<table')
      expect(out).toBe('<p class="md-p">a | b<br />c | d</p>')
    })

    it('does not treat a pipe row as a table when the column counts differ', () => {
      const out = renderMarkdown('| a | b |\n| --- |\n| 1 | 2 |')
      expect(out).not.toContain('<table')
    })

    it('stops a table at a blank line and keeps the following paragraph', () => {
      const out = renderMarkdown('| a |\n| --- |\n| 1 |\n\nafter')
      expect(out).toContain('<table class="md-table">')
      expect(out).toContain('<p class="md-p">after</p>')
    })

    it('escapes HTML inside a table cell', () => {
      const out = renderMarkdown('| a |\n| --- |\n| <b>x</b> |')
      expect(out).toContain('&lt;b&gt;x&lt;/b&gt;')
      expect(out).not.toContain('<b>')
    })

    it('renders links with the safe-target attributes', () => {
      expect(renderMarkdown('[site](https://example.com)')).toBe(
        '<p class="md-p"><a class="md-link" href="https://example.com" target="_blank" rel="noopener noreferrer">site</a></p>',
      )
    })

    it('autolinks a bare url and leaves trailing punctuation outside', () => {
      const out = renderMarkdown('see https://example.com/a.')
      expect(out).toContain('href="https://example.com/a"')
      expect(out).toContain('</a>.</p>')
    })

    it('drops a javascript: autolink to plain text', () => {
      // Not a real autolink (no http), but the link syntax path is what a user
      // controls; this pins that a non-safe scheme never becomes an anchor.
      const out = renderMarkdown('[x](javascript:alert(1))')
      expect(out).not.toContain('<a')
    })
  })
})
