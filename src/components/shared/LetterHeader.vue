<script setup>
/**
 * Shared letter header for print documents (browser + CrossDev).
 * Logo is shown as a plain <img> with a blob URL so WKWebView treats it as a normal
 * resource and does not strip it (data URLs in CSS/img are often cleared by WebViews).
 */
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  logoUrl: { type: String, default: '' },
  companyName: { type: String, default: '' },
  companyAddress: { type: String, default: '' },
  contactItems: {
    type: Array,
    default: () => [],
  },
})

/** Blob URL for the logo (persists; WebViews don't strip blob: URLs). Never cleared once set. */
const displayLogoUrl = ref('')

function dataUrlToBlobUrl(dataUrl) {
  if (!dataUrl || !dataUrl.startsWith('data:')) return ''
  try {
    const res = dataUrl.match(/^data:([^;]+);base64,(.+)$/)
    if (!res) return ''
    const type = res[1].trim()
    const b64 = res[2]
    const bin = atob(b64)
    const bytes = new Uint8Array(bin.length)
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i)
    const blob = new Blob([bytes], { type: type || 'image/png' })
    return URL.createObjectURL(blob)
  } catch {
    return ''
  }
}

function setLogo(url) {
  if (!url || !url.startsWith('data:')) return
  const blobUrl = dataUrlToBlobUrl(url)
  if (blobUrl) {
    if (displayLogoUrl.value && displayLogoUrl.value.startsWith('blob:')) {
      try { URL.revokeObjectURL(displayLogoUrl.value) } catch {}
    }
    displayLogoUrl.value = blobUrl
  }
}

watch(() => props.logoUrl, (url) => {
  if (url && url.startsWith('data:')) setLogo(url)
})

onMounted(() => {
  if (props.logoUrl && props.logoUrl.startsWith('data:')) setLogo(props.logoUrl)
})

onBeforeUnmount(() => {
  if (displayLogoUrl.value && displayLogoUrl.value.startsWith('blob:')) {
    try { URL.revokeObjectURL(displayLogoUrl.value) } catch {}
  }
  displayLogoUrl.value = ''
})
</script>

<template>
  <div class="letter-header">
    <table class="header-table">
      <tr>
        <td class="header-logo-col">
          <img
            v-if="displayLogoUrl"
            :src="displayLogoUrl"
            class="letter-head letter-head-img"
            alt=""
            role="img"
            aria-label="Logo"
          />
        </td>
        <td class="header-company-info-col">
          <div class="company-name-header">{{ companyName }}</div>
          <div class="company-address-header">
            <svg class="header-icon" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5a2.5 2.5 0 1 1 0-5 2.5 2.5 0 0 1 0 5z"/></svg>
            <span>{{ companyAddress }}</span>
          </div>
          <div class="company-contact-header">
            <template v-for="(item, idx) in contactItems" :key="item.type + idx">
              <span v-if="idx > 0" class="contact-sep">|</span>
              <svg v-if="item.type === 'phone'" class="header-icon" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M6.62 10.79c1.44 2.83 3.76 5.14 6.59 6.59l2.2-2.2c.27-.27.67-.36 1.02-.24 1.12.37 2.33.57 3.57.57.55 0 1 .45 1 1V20c0 .55-.45 1-1 1-9.39 0-17-7.61-17-17 0-.55.45-1 1-1h3.5c.55 0 1 .45 1 1 0 1.25.2 2.45.57 3.57.11.35.03.74-.25 1.02l-2.2 2.2z"/></svg>
              <svg v-else-if="item.type === 'email'" class="header-icon" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/></svg>
              <svg v-else-if="item.type === 'website'" class="header-icon" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/></svg>
              <span>{{ item.value }}</span>
            </template>
          </div>
        </td>
      </tr>
    </table>
  </div>
</template>

<style scoped>
.letter-header {
  margin-bottom: 20px;
}

.header-table {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
}

.header-logo-col {
  width: 120px;
  max-width: 120px;
  vertical-align: top;
  padding-right: 16px;
  height: 1%; /* so cell stretches to row height (row = height of 3 lines) */
}

.header-company-info-col {
  vertical-align: top;
}

.letter-head {
  width: 120px;
  max-width: 120px;
  min-height: 60px;
  display: block;
  -webkit-print-color-adjust: exact;
  print-color-adjust: exact;
}

.letter-head-img {
  width: 120px;
  max-width: 120px;
  height: auto;
  max-height: 80px;
  object-fit: contain;
  object-position: left top;
}

.company-name-header {
  font-size: 1.25rem;
  font-weight: bold;
  line-height: 1.3;
  margin-bottom: 4px;
}

.company-address-header {
  font-size: 0.8rem;
  line-height: 1.3;
  margin-bottom: 4px;
  color: #444;
}

.company-contact-header {
  font-size: 0.8rem;
  line-height: 1.3;
  color: #444;
}

.header-icon {
  width: 12px;
  height: 12px;
  min-width: 12px;
  min-height: 12px;
  margin-right: 4px;
  opacity: 0.85;
  vertical-align: middle;
  display: inline-block;
}

.contact-sep {
  margin: 0 6px;
  opacity: 0.7;
}

@media print {
  .letter-header {
    margin-bottom: 20px;
  }

  .header-table {
    width: 100% !important;
    table-layout: fixed !important;
  }

  .header-logo-col {
    width: 120px !important;
    max-width: 120px !important;
    height: 1% !important;
  }

  .company-name-header {
    font-size: 1.25rem;
    font-weight: bold;
  }

  .company-address-header,
  .company-contact-header {
    font-size: 0.8rem;
  }

  .header-icon {
    width: 12px !important;
    height: 12px !important;
    min-width: 12px !important;
    min-height: 12px !important;
  }

  .letter-head,
  .letter-head-img {
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }
}
</style>
