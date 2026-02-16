/**
 * Composable for CrossDev print windows.
 * When printOnly=1, extracts @media print rules from all stylesheets,
 * removes all other CSS, and applies only print rules so createPDF captures
 * the desired report (same as browser print preview).
 *
 * Use in PrintPage and PrintCarPage when route.query.printOnly === '1'.
 */
export function usePrintOnlyStyles() {
  /**
   * Extract all @media print rules from document stylesheets.
   * Returns CSS text of inner rules (without @media print wrapper).
   */
  const extractPrintRules = () => {
    const rules = []
    for (const sheet of document.styleSheets) {
      try {
        const cssRules = sheet.cssRules || sheet.rules
        if (!cssRules) continue
        for (let i = 0; i < cssRules.length; i++) {
          const rule = cssRules[i]
          if (rule.constructor.name === 'CSSMediaRule' && rule.media && rule.media.mediaText.includes('print')) {
            for (let j = 0; j < rule.cssRules.length; j++) {
              rules.push(rule.cssRules[j].cssText)
            }
          }
        }
      } catch (_) {
        // CORS or cross-origin stylesheet - skip
      }
    }
    return rules.join('\n')
  }

  /**
   * Regenerate page with only @media print CSS:
   * 1. Extract print rules from all stylesheets
   * 2. Disable all link and style elements (remove screen CSS)
   * 3. Inject extracted print rules as plain CSS (apply on screen)
   */
  const applyPrintOnlyStyles = () => {
    const printCss = extractPrintRules()
    if (!printCss.trim()) return

    // Disable all existing stylesheets (remove screen CSS)
    const sheets = Array.from(document.styleSheets)
    for (const sheet of sheets) {
      try {
        sheet.disabled = true
      } catch (_) {}
    }
    document.querySelectorAll('link[rel="stylesheet"]').forEach((el) => {
      el.disabled = true
    })
    document.querySelectorAll('style').forEach((el) => {
      el.disabled = true
    })

    // A4 width constraints (210mm) so result fits paper
    const a4Constraints = `
html, body { max-width: 210mm; margin: 0 auto; padding: 0; box-sizing: border-box; }
#app, .app-main { max-width: 210mm; margin: 0 auto; padding: 0 !important; box-sizing: border-box; }
* { box-sizing: border-box; }
img { max-width: 100%; height: auto; }
table { max-width: 100%; }
td, th { word-wrap: break-word; overflow-wrap: break-word; }
`
    const combinedCss = a4Constraints + '\n' + printCss

    const style = document.createElement('style')
    style.id = 'crossdev-print-only-styles'
    style.textContent = combinedCss
    document.head.appendChild(style)

    // Reset app-main padding (header is hidden by print rules)
    const main = document.querySelector('.app-main')
    if (main) {
      main.style.paddingTop = '0'
    }
  }

  return { extractPrintRules, applyPrintOnlyStyles }
}
