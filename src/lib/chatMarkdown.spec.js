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
        /^<\/?(a|p|br|strong|em|del|code|pre|blockquote|hr|ul|ol|li|h[1-6]|img)(\s|>|\/)/i
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
