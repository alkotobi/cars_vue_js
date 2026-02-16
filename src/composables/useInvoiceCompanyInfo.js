/**
 * Composable to fetch company info (and optional logo) for invoice export.
 * Same logic as AppHeader: active bank or user's bank for different-company users.
 * Use when building createInvoice payload so invoice has correct company name, address, mobile, email, website, and logo.
 */
import { useApi } from './useApi'

export function useInvoiceCompanyInfo() {
  const { callApi, getFileUrl } = useApi()

  /**
   * Fetch company info for invoice: name, address, mobile, email, website.
   * Also fetches logo as base64 if available (for CrossDev createInvoice).
   * @returns {Promise<{ company: { name, address, phone, email, website }, logoBase64?: string }>}
   */
  async function fetchInvoiceCompanyInfo() {
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
      if (result.success && result.data?.length > 0) {
        bank = result.data[0]
      }
    }

    if (!bank) {
      const result = await callApi({
        query: 'SELECT company_name, company_address, mobile, email, website, logo_path FROM banks WHERE is_active = 1 LIMIT 1',
      })
      if (result.success && result.data?.length > 0) {
        bank = result.data[0]
      }
    }

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

  return { fetchInvoiceCompanyInfo }
}
