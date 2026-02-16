/**
 * Composable for CrossDev desktop app integration.
 * Use openWindow() to open URLs in a new native window (CrossDev) or tab (browser).
 * Use openWindowWithHtml() to open raw HTML in a new window (e.g. print reports in CrossDev).
 *
 * @example
 * const { openWindow, openWindowWithHtml } = useCrossDev()
 * openWindow('/print/123', { title: 'Print', width: 1100, height: 900 })
 * openWindowWithHtml(htmlString, { title: 'Report', width: 800, height: 600 })
 */
export function useCrossDev() {
  const hasCrossDev = () => typeof window !== 'undefined' && window.CrossDev?.invoke

  /**
   * Open a URL in a new window (CrossDev) or tab (browser).
   * @param {string} url - Full URL or path (will be resolved against current origin)
   * @param {Object} [options]
   * @param {string} [options.className='window'] - Window class for singleton lookup
   * @param {string} [options.title='New Window'] - Window title
   * @param {boolean} [options.isSingleton=false] - Reuse existing window with same className
   * @param {number} [options.x=150] - Left position
   * @param {number} [options.y=150] - Top position
   * @param {number} [options.width=900] - Window width
   * @param {number} [options.height=700] - Window height
   * @returns {Promise<void>}
   */
  const openWindow = (url, options = {}) => {
    const fullUrl = url.startsWith('http') ? url : new URL(url, window.location.href).href

    const payload = {
      className: options.className ?? 'window',
      title: options.title ?? 'New Window',
      url: fullUrl,
      isSingleton: options.isSingleton ?? false,
      x: options.x ?? 150,
      y: options.y ?? 150,
      width: options.width ?? 900,
      height: options.height ?? 700,
    }

    if (hasCrossDev()) {
      return window.CrossDev.invoke('createWindow', payload).catch((e) => {
        console.error('CrossDev createWindow failed:', e)
        window.open(fullUrl, '_blank')
      })
    }

    window.open(fullUrl, '_blank')
    return Promise.resolve()
  }

  /**
   * Open raw HTML in a new window. In CrossDev opens a native window with that HTML;
   * in browser uses window.open + document.write.
   * Use this for print reports (print selected, batch print) so they work in both environments.
   * @param {string} html - Full HTML document string (e.g. <!DOCTYPE html>...)
   * @param {Object} [options]
   * @param {string} [options.className='print-html'] - Window class for singleton lookup
   * @param {string} [options.title='Print'] - Window title
   * @param {number} [options.x=150] - Left position
   * @param {number} [options.y=150] - Top position
   * @param {number} [options.width=800] - Window width
   * @param {number} [options.height=600] - Window height
   * @returns {Promise<void>}
   */
  const openWindowWithHtml = (html, options = {}) => {
    const title = options.title ?? 'Print'
    const width = options.width ?? 800
    const height = options.height ?? 600

    if (hasCrossDev()) {
      const payload = {
        className: options.className ?? 'print-html',
        title,
        html,
        isSingleton: false,
        x: options.x ?? 150,
        y: options.y ?? 150,
        width,
        height,
      }
      return window.CrossDev.invoke('createWindow', payload).catch((e) => {
        console.error('CrossDev createWindow (HTML) failed:', e)
        const w = window.open('', '_blank', `width=${width},height=${height}`)
        if (w) {
          w.document.write(html)
          w.document.close()
          w.onload = () => { w.print(); w.close() }
        }
      })
    }

    const w = window.open('', '_blank', `width=${width},height=${height}`)
    if (w) {
      w.document.write(html)
      w.document.close()
      w.onload = () => {
        w.focus()
        w.print()
        w.close()
      }
    }
    return Promise.resolve()
  }

  return { hasCrossDev, openWindow, openWindowWithHtml }
}
