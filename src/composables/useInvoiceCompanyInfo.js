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

  return {
    fetchInvoiceCompanyInfo,
    getCompanyLogoUrl,
    getCompanyLogoDataUrl,
    getReportHeaderContactItems,
    getStoredPrintLogoDataUrl,
    setStoredPrintLogoDataUrl,
  }
}
