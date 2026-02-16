/**
 * Composable to fetch company info (and optional logo) for invoice export.
 * Same logic as AppHeader: active bank or user's bank for different-company users.
 * Use when building createInvoice payload so invoice has correct company name, address, mobile, email, website, and logo.
 *
 * getCompanyLogoUrl() is the preferred source for report header images (print, CrossDev): same as xlsx invoice logo
 * (banks.logo_path or default logo). Use it in SellBillPrintDocument, CarPrintDocument, batch print, CarStock, Cashier reports.
 */
import { useApi } from './useApi'

// In-memory cache so logo survives when sessionStorage is isolated (e.g. CrossDev print WebView)
let printLogoDataUrlMemory = ''

export function useInvoiceCompanyInfo() {
  const { callApi, getFileUrl, getAssets } = useApi()

  /** Same bank resolution as fetchInvoiceCompanyInfo (different-company → user bank, else active bank). */
  async function getBankForCompany() {
    const userStr = localStorage.getItem('user')
    let user = null
    try {
      if (userStr) user = JSON.parse(userStr)
    } catch {
      user = null
    }
    const isDifferentCompany = user && (
      user.is_diffrent_company === 1 ||
      user.is_diffrent_company === true ||
      user.is_diffrent_company === '1'
    )
    let bank = null
    if (isDifferentCompany && user?.id_bank_account) {
      const result = await callApi({
        query: 'SELECT company_name, company_address, mobile, email, website, logo_path FROM banks WHERE id = ?',
        params: [user.id_bank_account],
      })
      if (result.success && result.data?.length > 0) bank = result.data[0]
    }
    if (!bank) {
      const result = await callApi({
        query: 'SELECT company_name, company_address, mobile, email, website, logo_path FROM banks WHERE is_active = 1 LIMIT 1',
      })
      if (result.success && result.data?.length > 0) bank = result.data[0]
    }
    return bank
  }

  /**
   * Company logo URL for report header (browser + CrossDev). Same source as xlsx invoice logo.
   * @returns {Promise<string>} URL for <img src>, or default logo.png from assets
   */
  async function getCompanyLogoUrl() {
    const bank = await getBankForCompany()
    if (bank?.logo_path && bank.logo_path.trim() !== '') {
      const url = getFileUrl(bank.logo_path)
      if (url) return url
    }
    const assets = await getAssets()
    return assets?.logo || ''
  }

  /** Storage key for logo data URL. localStorage is shared across windows (same origin) so CrossDev print window can read what main window stored. */
  const PRINT_LOGO_STORAGE_KEY = 'cars_print_logo_dataurl'

  /** Read persisted logo data URL (sync). Order: memory → sessionStorage → localStorage.
   * When we read from storage we also set module memory so CrossDev print window keeps the logo
   * even if storage is later cleared or inaccessible. */
  function getStoredPrintLogoDataUrl() {
    if (printLogoDataUrlMemory && printLogoDataUrlMemory.startsWith('data:')) return printLogoDataUrlMemory
    try {
      const s = sessionStorage.getItem(PRINT_LOGO_STORAGE_KEY)
      if (s && s.startsWith('data:')) {
        printLogoDataUrlMemory = s
        return s
      }
      const l = localStorage.getItem(PRINT_LOGO_STORAGE_KEY)
      if (l && l.startsWith('data:')) {
        printLogoDataUrlMemory = l
        return l
      }
    } catch {
      // ignore
    }
    return ''
  }

  /** Persist logo data URL (memory + sessionStorage + localStorage). localStorage lets the CrossDev print window read it. */
  function setStoredPrintLogoDataUrl(dataUrl) {
    if (!dataUrl || !dataUrl.startsWith('data:')) return
    printLogoDataUrlMemory = dataUrl
    try {
      sessionStorage.setItem(PRINT_LOGO_STORAGE_KEY, dataUrl)
      localStorage.setItem(PRINT_LOGO_STORAGE_KEY, dataUrl)
    } catch {
      // ignore quota or security errors
    }
  }

  /**
   * Company logo as a data URL (base64). Use this for print/header images so the logo
   * does not disappear when the page or window loses focus (avoids revalidation of http URLs).
   * Resolves relative URLs to absolute so fetch works in print/CrossDev windows.
   * @returns {Promise<string>} data: URL for <img src>, or '' if unavailable
   */
  async function getCompanyLogoDataUrl() {
    let url = await getCompanyLogoUrl()
    if (!url || url.startsWith('data:')) return url || ''
    // Resolve relative URLs so fetch works in print/CrossDev child windows
    if (url.startsWith('/') || !/^https?:/i.test(url)) {
      try {
        url = new URL(url, window.location.href).href
      } catch {
        return ''
      }
    }
    try {
      const res = await fetch(url, { credentials: 'same-origin' })
      if (!res.ok) return ''
      const blob = await res.blob()
      return await new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.onload = () => resolve(reader.result)
        reader.onerror = reject
        reader.readAsDataURL(blob)
      })
    } catch {
      return ''
    }
  }

  /**
   * Fetch company info for invoice: name, address, mobile, email, website.
   * Also fetches logo as base64 if available (for CrossDev createInvoice).
   * @returns {Promise<{ company: { name, address, phone, email, website }, logoBase64?: string }>}
   */
  async function fetchInvoiceCompanyInfo() {
    const bank = await getBankForCompany()
    const company = {
      name: bank?.company_name || '',
      address: bank?.company_address || '',
      phone: bank?.mobile || '',
      email: bank?.email || '',
      website: bank?.website || '',
    }

    let logoBase64 = null
    if (bank?.logo_path && bank.logo_path.trim() !== '') {
      try {
        const logoUrl = getFileUrl(bank.logo_path)
        if (logoUrl) {
          const res = await fetch(logoUrl)
          if (res.ok) {
            const blob = await res.blob()
            logoBase64 = await new Promise((resolve, reject) => {
              const reader = new FileReader()
              reader.onload = () => {
                const dataUrl = reader.result
                const commaIdx = dataUrl.indexOf(',')
                resolve(commaIdx >= 0 ? dataUrl.slice(commaIdx + 1) : dataUrl)
              }
              reader.onerror = reject
              reader.readAsDataURL(blob)
            })
          }
        }
      } catch {
        // ignore logo fetch errors
      }
    }

    return { company, logoBase64 }
  }

  /**
   * Build contact items for PrintReportHeader (phone, email, website).
   * Uses bank (selectedBank) if present, else company fallback. Same order everywhere.
   * @param {{ mobile?: string, email?: string, website?: string } | null} bank
   * @param {{ phone?: string, email?: string } | null} companyFallback
   * @returns {{ type: 'phone'|'email'|'website', value: string }[]}
   */
  function getReportHeaderContactItems(bank, companyFallback) {
    const items = []
    if (bank?.mobile) items.push({ type: 'phone', value: bank.mobile })
    if (bank?.email) items.push({ type: 'email', value: bank.email })
    if (bank?.website) items.push({ type: 'website', value: bank.website })
    if (items.length > 0) return items
    if (companyFallback?.phone) items.push({ type: 'phone', value: companyFallback.phone })
    if (companyFallback?.email) items.push({ type: 'email', value: companyFallback.email })
    return items
  }

  /**
   * Build HTML for the letterhead block (same layout as LetterHeader.vue) for use in
   * raw HTML print windows (e.g. print selected, batch print). Pass logo URL (or data URL),
   * company name, address, and contact items from getReportHeaderContactItems(bank, null).
   * @param {string} logoUrl
   * @param {string} companyName
   * @param {string} companyAddress
   * @param {{ type: 'phone'|'email'|'website', value: string }[]} contactItems
   * @returns {string} HTML string
   */
  // Inline SVGs matching LetterHeader.vue (address, phone, email, website)
  const iconStyle = 'width:12px;height:12px;min-width:12px;min-height:12px;margin-right:4px;opacity:0.85;vertical-align:middle;display:inline-block'
  const iconAddress = `<svg style="${iconStyle}" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5a2.5 2.5 0 1 1 0-5 2.5 2.5 0 0 1 0 5z"/></svg>`
  const iconPhone = `<svg style="${iconStyle}" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M6.62 10.79c1.44 2.83 3.76 5.14 6.59 6.59l2.2-2.2c.27-.27.67-.36 1.02-.24 1.12.37 2.33.57 3.57.57.55 0 1 .45 1 1V20c0 .55-.45 1-1 1-9.39 0-17-7.61-17-17 0-.55.45-1 1-1h3.5c.55 0 1 .45 1 1 0 1.25.2 2.45.57 3.57.11.35.03.74-.25 1.02l-2.2 2.2z"/></svg>`
  const iconEmail = `<svg style="${iconStyle}" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/></svg>`
  const iconWebsite = `<svg style="${iconStyle}" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/></svg>`

  function buildLetterheadHtml(logoUrl, companyName, companyAddress, contactItems) {
    const escape = (s) => {
      if (s == null || s === '') return ''
      return String(s)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
    }
    const name = escape(companyName)
    const address = escape(companyAddress)
    const logoSrc = logoUrl ? ` src="${escape(logoUrl)}"` : ''
    const addressLine = address
      ? `<div style="font-size:0.8rem;line-height:1.3;margin-bottom:4px;color:#444">${iconAddress}<span>${address}</span></div>`
      : ''
    let contactHtml = ''
    if (contactItems && contactItems.length > 0) {
      const iconByType = { phone: iconPhone, email: iconEmail, website: iconWebsite }
      contactHtml = contactItems
        .map(
          (item, idx) => {
            const icon = iconByType[item.type] || ''
            return `${idx > 0 ? '<span style="margin:0 6px;opacity:0.7">|</span>' : ''}${icon}<span>${escape(item.value)}</span>`
          }
        )
        .join('')
      contactHtml = `<div style="font-size:0.8rem;line-height:1.3;color:#444">${contactHtml}</div>`
    }
    return `
<div style="margin-bottom:20px">
  <table style="width:100%;table-layout:fixed;border-collapse:collapse">
    <tr>
      <td style="width:120px;max-width:120px;vertical-align:top;padding-right:16px">
        ${logoUrl ? `<img${logoSrc} alt="" style="width:120px;max-width:120px;height:auto;max-height:80px;object-fit:contain;display:block" onerror="this.style.display='none'" />` : ''}
      </td>
      <td style="vertical-align:top">
        <div style="font-size:1.25rem;font-weight:bold;line-height:1.3;margin-bottom:4px">${name}</div>
        ${addressLine}
        ${contactHtml}
      </td>
    </tr>
  </table>
</div>`
  }

  return {
    fetchInvoiceCompanyInfo,
    getBankForCompany,
    getCompanyLogoUrl,
    getCompanyLogoDataUrl,
    getReportHeaderContactItems,
    getStoredPrintLogoDataUrl,
    setStoredPrintLogoDataUrl,
    buildLetterheadHtml,
  }
}
