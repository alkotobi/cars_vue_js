/**
 * Composable for CrossDev desktop app integration.
 * Use openWindow() to open URLs in a new native window (CrossDev) or tab (browser).
 * All geometry (x, y, width, height) is passed from JS to native - no C++ changes needed.
 *
 * @example
 * const { openWindow } = useCrossDev()
 * openWindow('/print/123', { title: 'Print', width: 1100, height: 900 })
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

  return { hasCrossDev, openWindow }
}
