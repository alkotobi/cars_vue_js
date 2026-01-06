<script setup>
import { ref, onMounted, computed } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'

const { t } = useEnhancedI18n()

const props = defineProps({
  billId: {
    type: Number,
    required: true,
  },
  options: {
    type: Object,
    required: true,
  },
})

const { callApi, getAssets, loadLetterhead } = useApi()
const loading = ref(true)
const error = ref(null)
const billData = ref(null)
const carsData = ref([])
const banks = ref([])
const selectedBank = ref(null)
const logoUrl = ref(null)
const stampUrl = ref(null)
const letterHeadUrl = ref(null)
const contractTerms = ref([])
const enabledLanguages = ref({})

// Computed property to filter visible terms
const visibleTerms = computed(() => {
  return contractTerms.value.filter(term => term.visible !== false)
})

// Helper function to get base path (same as in useApi.js)
const getBasePath = () => {
  let baseUrl = import.meta.env.BASE_URL || './'
  if (baseUrl === './' || baseUrl.startsWith('./')) {
    const pathname = window.location.pathname
    const match = pathname.match(/^(\/[^/]+\/)/)
    return match ? match[1] : '/'
  }
  return baseUrl
}

// Load contract terms from JSON file
const loadContractTerms = async () => {
  try {
    const basePath = getBasePath()
    const termsUrl = `${basePath}contract_terms.json`
    const response = await fetch(termsUrl)
    if (!response.ok) {
      console.warn('Failed to load contract_terms.json, using empty array')
      contractTerms.value = []
      enabledLanguages.value = { english: true, chinese: true } // Default fallback
      return
    }
    const data = await response.json()
    contractTerms.value = data.terms || []
    enabledLanguages.value = data.enabledLanguages || { english: true, chinese: true } // Default fallback
  } catch (err) {
    console.error('Error loading contract terms:', err)
    contractTerms.value = []
    enabledLanguages.value = { english: true, chinese: true } // Default fallback
  }
}

const company = ref({
  name: 'GROUP MERHAB LIMITED',
  address: 'GUANGZHOU, CHINA',
  phone: '+86 123 456 7890',
  email: 'contact@merhab.com',
  logo: null, // Will be set from getAssets
})

const fetchBanks = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM banks ORDER BY company_name ASC',
      params: [],
    })
    if (result.success && result.data.length > 0) {
      banks.value = result.data
      // Set the selected bank based on options
      if (props.options.bankId) {
        selectedBank.value = banks.value.find((bank) => bank.id === props.options.bankId)
      }
      // Fallback to first bank if no bank is selected
      if (!selectedBank.value) {
        selectedBank.value = result.data[0]
      }
    }
  } catch (err) {
    console.error('Error fetching banks:', err)
  }
}

const fetchBillData = async () => {
  try {
    // Fetch bill details including broker info
    const billResult = await callApi({
      query: `
        SELECT 
          sb.*,
          c.name as broker_name,
          c.address as broker_address,
          c.mobiles as broker_phone
        FROM sell_bill sb
        LEFT JOIN clients c ON sb.id_broker = c.id
        WHERE sb.id = ?
      `,
      params: [props.billId],
    })

    if (billResult.success && billResult.data.length > 0) {
      const bill = billResult.data[0]
      
      // Parse notes JSON if it exists
      if (bill.notes) {
        try {
          let notesArray = []
          if (typeof bill.notes === 'string' && bill.notes.trim().startsWith('[')) {
            notesArray = JSON.parse(bill.notes)
          } else if (Array.isArray(bill.notes)) {
            notesArray = bill.notes
          }
          
          if (Array.isArray(notesArray) && notesArray.length > 0) {
            // Format notes for printing (show all notes, each on a line, without user/date)
            bill.notesFormatted = notesArray.map(note => note.note || '').filter(note => note.trim() !== '').join('\n')
            // Also keep latest note for simple display
            bill.notes = notesArray[notesArray.length - 1].note || ''
          }
        } catch (e) {
          // If parsing fails, treat as old format (plain text)
          bill.notesFormatted = bill.notes
        }
      }
      
      billData.value = bill
      console.log('Bill data:', billData.value)
      console.log('Bill ref:', billData.value.bill_ref)

      // Fetch cars associated with this bill
      const carsResult = await callApi({
        query: `
          SELECT 
            cs.id,
            cs.vin,
            cs.notes,
            cs.date_sell,
            cs.price_cell,
            cs.cfr_da,
            cs.freight,
            cs.date_loding,
            cs.date_send_documents,
            cs.rate,
            c.name as client_name,
            lp.loading_port,
            dp.discharge_port,
            bd.price_sell as buy_price,
            cn.car_name,
            clr.color
          FROM cars_stock cs
          LEFT JOIN clients c ON cs.id_client = c.id
          LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
          LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
          LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
          LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
          LEFT JOIN colors clr ON cs.id_color = clr.id
          WHERE cs.id_sell = ? AND cs.hidden = 0
          ORDER BY cs.id DESC
        `,
        params: [props.billId],
      })

      console.log('Cars query params:', [props.billId])
      console.log('Cars result:', carsResult)

      if (carsResult.success) {
        carsData.value = carsResult.data
        console.log('Cars data with rates:', carsData.value)
      }
    }
  } catch (err) {
    console.error('Error fetching data:', err)
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

const calculateTotal = () => {
  return carsData.value
    .reduce((total, car) => {
      let myPrice
      if (props.options.paymentTerms.toLowerCase() === 'cfr') {
        myPrice = (parseFloat(car.price_cell) || 0) + (parseFloat(car.freight) || 0)
      } else {
        myPrice = parseFloat(car.price_cell) || 0
      }

      // Convert to selected currency
      if (props.options.currency.toUpperCase() === 'DA') {
        // Use cfr_da field if available, otherwise calculate from price_cell
        if (car.cfr_da) {
          return total + parseFloat(car.cfr_da)
        }
        const rate = parseFloat(car.rate)
        if (!rate) {
          console.warn(`No rate found for car ID ${car.id}`)
          return total
        }
        return total + myPrice * rate
      }
      return total + myPrice
    }, 0)
    .toFixed(2)
}

const calculateCarPrice = (car) => {
  let myPrice
  if (props.options.paymentTerms.toLowerCase() === 'cfr') {
    myPrice = (parseFloat(car.price_cell) || 0) + (parseFloat(car.freight) || 0)
  } else {
    myPrice = parseFloat(car.price_cell) || 0
  }

  // Convert to selected currency
  if (props.options.currency.toUpperCase() === 'DA') {
    // Use cfr_da field if available, otherwise calculate from price_cell
    if (car.cfr_da) {
      return parseFloat(car.cfr_da).toFixed(2)
    }
    const rate = parseFloat(car.rate)
    if (!rate) {
      console.warn(`No rate found for car ID ${car.id}`)
      return 'Rate not set'
    }
    return (myPrice * rate).toFixed(2)
  }
  return myPrice.toFixed(2)
}

const formatPrice = (price) => {
  if (props.options.currency.toUpperCase() === 'DA') {
    if (price === 'Rate not set') return price
    return price
  }
  return price
}

const numberToWords = (num) => {
  const ones = ['', 'One', 'Two', 'Three', 'Four', 'Five', 'Six', 'Seven', 'Eight', 'Nine']
  const tens = [
    '',
    '',
    'Twenty',
    'Thirty',
    'Forty',
    'Fifty',
    'Sixty',
    'Seventy',
    'Eighty',
    'Ninety',
  ]
  const teens = [
    'Ten',
    'Eleven',
    'Twelve',
    'Thirteen',
    'Fourteen',
    'Fifteen',
    'Sixteen',
    'Seventeen',
    'Eighteen',
    'Nineteen',
  ]

  const convertLessThanOneThousand = (n) => {
    if (n === 0) return ''

    if (n < 10) return ones[n]

    if (n < 20) return teens[n - 10]

    if (n < 100) {
      return tens[Math.floor(n / 10)] + (n % 10 ? ' ' + ones[n % 10] : '')
    }

    return (
      ones[Math.floor(n / 100)] +
      ' Hundred' +
      (n % 100 ? ' and ' + convertLessThanOneThousand(n % 100) : '')
    )
  }

  if (num === 0) return 'Zero'

  const parts = num.toString().split('.')
  let wholePart = parseInt(parts[0])
  let result = ''

  if (wholePart === 0) {
    result = 'Zero'
  } else {
    // Billions
    if (wholePart >= 1000000000) {
      result += convertLessThanOneThousand(Math.floor(wholePart / 1000000000)) + ' Billion '
      wholePart %= 1000000000
    }

    // Millions
    if (wholePart >= 1000000) {
      result += convertLessThanOneThousand(Math.floor(wholePart / 1000000)) + ' Million '
      wholePart %= 1000000
    }

    // Thousands
    if (wholePart >= 1000) {
      result += convertLessThanOneThousand(Math.floor(wholePart / 1000)) + ' Thousand '
      wholePart %= 1000
    }

    // Hundreds and remaining
    result += convertLessThanOneThousand(wholePart)
  }

  // Handle decimal part
  if (parts.length > 1) {
    const decimalPart = parts[1].padEnd(2, '0').substring(0, 2)
    if (parseInt(decimalPart) > 0) {
      result += ' and ' + decimalPart + '/100'
    }
  }

  return result.trim()
}

// Load assets (logo, stamp, letter head)
const loadAssets = async () => {
  try {
    const assets = await getAssets()
    if (assets) {
      logoUrl.value = assets.logo || null
      stampUrl.value = assets.gml2 || null
      letterHeadUrl.value = await loadLetterhead()

      // Update company logo
      if (logoUrl.value) {
        company.value.logo = logoUrl.value
      } else {
        console.error('Failed to load assets, using defaults:', err)
      }
    }
  } catch (err) {
    console.error('Failed to load assets, using defaults:', err)
    // Use fallback URLs
  }
}

// Get freight and rate values for placeholder replacement
const getFreightAndRateValues = () => {
  if (!carsData.value || carsData.value.length === 0) {
    return { freightText: 'N/A', rateText: 'N/A' }
  }

  // Get unique freight and rate values
  const freightValues = [...new Set(carsData.value.map(car => parseFloat(car.freight) || 0).filter(f => f > 0))]
  const rateValues = [...new Set(carsData.value.map(car => parseFloat(car.rate) || 0).filter(r => r > 0))]

  // Format freight values
  let freightText = 'N/A'
  if (freightValues.length > 0) {
    const minFreight = Math.min(...freightValues)
    const maxFreight = Math.max(...freightValues)
    if (minFreight === maxFreight) {
      freightText = `USD ${minFreight.toFixed(2)}`
    } else {
      freightText = `USD ${minFreight.toFixed(2)} - ${maxFreight.toFixed(2)}`
    }
  }

  // Format rate values
  let rateText = 'N/A'
  if (rateValues.length > 0) {
    const minRate = Math.min(...rateValues)
    const maxRate = Math.max(...rateValues)
    if (minRate === maxRate) {
      rateText = minRate.toFixed(2)
    } else {
      rateText = `${minRate.toFixed(2)} - ${maxRate.toFixed(2)}`
    }
  }

  return { freightText, rateText }
}

// Replace placeholders in term text with proper RTL handling
const replaceTermPlaceholders = (text) => {
  if (!text) return text
  const { freightText, rateText } = getFreightAndRateValues()
  
  // Wrap numeric values with LTR markers to preserve direction in RTL text
  // Using Left-to-Right Mark (LRM) and Right-to-Left Mark (RLM) for proper bidirectional text
  const ltrIsolate = '\u2066' // Left-to-Right Isolate
  const popIsolate = '\u2069'  // Pop Directional Isolate
  
  // Wrap freight and rate values to preserve LTR direction
  const wrappedFreight = freightText !== 'N/A' ? `${ltrIsolate}${freightText}${popIsolate}` : freightText
  const wrappedRate = rateText !== 'N/A' ? `${ltrIsolate}${rateText}${popIsolate}` : rateText
  
  return text.replace(/{FREIGHT}/g, wrappedFreight).replace(/{RATE}/g, wrappedRate)
}

onMounted(async () => {
  await loadAssets()
  await loadContractTerms()
  fetchBillData()
  fetchBanks()
})
</script>

<template>
  <div class="print-report">
    <!-- Header Image -->
    <div class="report-header">
      <img :src="letterHeadUrl || ''" alt="Letter Head" class="letter-head" />
    </div>

    <!-- Date and Reference -->
    <div class="report-meta">
      <span
        ><strong>{{ t('sellBills.date') }}:</strong>
        {{ billData ? formatDate(billData.date_sell) : '' }}</span
      >
      <span class="report-ref"
        ><strong>{{ t('sellBills.ci_no') }}:</strong> {{ billData ? billData.bill_ref : '' }}</span
      >
    </div>

    <!-- Title -->
    <h1 class="report-title">
      {{
        options.documentType === 'contract'
          ? t('sellBills.contract')
          : options.documentType === 'invoice'
            ? t('sellBills.invoice')
            : t('sellBills.proforma')
      }}
    </h1>

    <!-- Buyer Information Section -->
    <div class="section buyer-section">
      <div class="buyer-details">
        <div class="buyer-info">
          <div class="info-group">
            <label class="bold-label">{{ t('sellBills.name') }}:</label>
            <span class="info-value">{{ billData?.broker_name }}</span>
          </div>
          <div class="info-group">
            <label class="bold-label">{{ t('sellBills.address') }}:</label>
            <span class="info-value">{{ billData?.broker_address }}</span>
          </div>
          <div class="info-group">
            <label class="bold-label">{{ t('sellBills.phone') }}:</label>
            <span class="info-value">{{ billData?.broker_phone }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Vehicle Details Table -->
    <div class="section vehicles-section">
      <div class="table-container">
        <table class="report-table" :style="{ fontSize: (options.tableFontSize || 12) + 'pt' }">
          <thead>
            <tr>
              <th>No.</th>
              <th>{{ t('sellBills.vehicle') }}</th>
              <th>{{ t('sellBills.color') }}</th>
              <th>{{ t('sellBills.vin') }}</th>
              <th>{{ t('sellBills.client') }}</th>
              <th>{{ t('sellBills.port') }}</th>
              <th>
                {{ t('sellBills.price') }} {{ options.paymentTerms.toUpperCase() }}
                {{ options.currency.toUpperCase() }}
              </th>
              <th>Notes</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(car, index) in carsData" :key="car.id">
              <td>{{ index + 1 }}</td>
              <td>{{ car.car_name }}</td>
              <td>{{ car.color }}</td>
              <td>{{ car.vin }}</td>
              <td>{{ car.client_name }}</td>
              <td>{{ car.discharge_port }}</td>
              <td>{{ formatPrice(calculateCarPrice(car)) }}</td>
              <td>{{ car.notes || '-' }}</td>
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td colspan="7" class="total-label">
                <strong>{{ t('sellBills.total') }}:</strong>
              </td>
              <td class="total-value">
                <strong>{{ formatPrice(calculateTotal()) }}</strong>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
      <div class="amount-in-words" :style="{ fontSize: (options.tableFontSize || 12) + 'pt' }">
        <span class="words-label">{{ t('sellBills.total') }} in words:</span>
        <span class="words-value">
          <strong
            >{{ numberToWords(calculateTotal()) }} {{ options.currency.toUpperCase() }}</strong
          >
        </span>
      </div>
    </div>

    <!-- Payment & Bank Information -->
    <div class="section payment-bank-section" :style="{ fontSize: (options.tableFontSize || 12) + 'pt' }">
      <div class="payment-bank-details">
        <div class="payment-info">
          <div class="info-row">
            <span class="info-label bold-label">{{ t('sellBills.payment_terms') }}:</span>
            <span class="info-value">{{ options.paymentTerms.toUpperCase() }}</span>
          </div>
          <div class="info-row">
            <span class="info-label bold-label">{{ t('sellBills.mode') }}:</span>
            <span class="info-value">{{ options.paymentMode }}</span>
          </div>
        </div>
        <div class="bank-info" v-if="selectedBank">
          <div class="info-row" v-if="selectedBank.company_name">
            <span class="info-label bold-label">Company Name:</span>
            <span class="info-value">{{ selectedBank.company_name }}</span>
          </div>
          <div class="info-row" v-if="selectedBank.bank_name">
            <span class="info-label bold-label">Bank Name:</span>
            <span class="info-value">{{ selectedBank.bank_name }}</span>
          </div>
          <div class="info-row" v-if="selectedBank.bank_address">
            <span class="info-label bold-label">Bank Address:</span>
            <span class="info-value">{{ selectedBank.bank_address }}</span>
          </div>
          <div class="info-row" v-if="selectedBank.bank_account">
            <span class="info-label bold-label">Account:</span>
            <span class="info-value">{{ selectedBank.bank_account }}</span>
          </div>
          <div class="info-row" v-if="selectedBank.swift_code">
            <span class="info-label bold-label">SWIFT:</span>
            <span class="info-value">{{ selectedBank.swift_code }}</span>
          </div>
          <div class="info-row" v-if="selectedBank.notes">
            <span class="info-label bold-label">Notes:</span>
            <span class="info-value">{{ selectedBank.notes }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Notes Section -->
    <div class="section notes-section" v-if="billData?.notes || billData?.notesFormatted" :style="{ fontSize: (options.tableFontSize || 12) + 'pt' }">
      <h3 class="section-title">Bill Notes</h3>
      <div class="notes-content">
        <p v-if="billData.notesFormatted" style="white-space: pre-line;">{{ billData.notesFormatted }}</p>
        <p v-else>{{ billData.notes }}</p>
      </div>
    </div>

    <!-- Footer -->
    <div class="footer">
      <div class="footer-content"></div>
    </div>

    <!-- Second Page (Terms and Conditions) -->
    <div v-if="options.documentType === 'contract'" class="a4-page terms-page">
      <div class="section contract-terms">
        <h3>Terms and Conditions</h3>
        <div class="terms-list">
          <div v-for="(term, index) in visibleTerms" :key="term.id" class="term-item">
            <template v-for="(value, key) in term" :key="key">
              <p 
                v-if="key !== 'id' && key !== 'visible' && enabledLanguages[key] === true" 
                :class="key"
                :style="{ fontSize: (options.termsFontSize || 11) + 'pt' }"
              >
                {{ index + 1 }}. {{ replaceTermPlaceholders(value) }}
            </p>
            </template>
          </div>
        </div>
      </div>
      <div class="signatures">
        <div class="signature-box">
          <p>Seller's Signature</p>
          <div class="signature-line"></div>
        </div>
        <div class="signature-box">
          <p>Buyer's Signature</p>
          <div class="signature-line"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.print-report {
  font-family: 'Arial', sans-serif;
  max-width: 210mm;
  margin: 0 auto;
  padding: 20px;
  background: white;
  color: #333;
}

.report-header {
  text-align: center;
  margin-bottom: 20px;
}

.letter-head {
  max-width: 100%;
  height: auto;
  max-height: 120px;
}

.report-meta {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  font-size: 14px;
}

.report-title {
  text-align: center;
  font-size: 2rem;
  margin: 20px 0 30px 0;
  color: #2c3e50;
}

.section {
  margin-bottom: 30px;
}

.section-header h3 {
  font-size: 1.2rem;
  margin-bottom: 10px;
  color: #3498db;
}

.table-container {
  overflow-x: auto;
}

.report-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.report-table th,
.report-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.report-table th {
  background-color: #f5f5f5;
  font-weight: bold;
  color: #333;
}

.report-table tr:hover {
  background-color: #f9f9f9;
}

.total-label {
  text-align: right;
}

.total-value {
  font-weight: bold;
  color: #3498db;
}

.amount-in-words {
  margin-top: 10px;
  font-size: 1rem;
  color: #555;
}

.words-label {
  font-weight: 500;
}

.words-value {
  margin-left: 10px;
  font-style: italic;
}

.vehicle-summary {
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #3498db;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #e9ecef;
}

.summary-item:last-child {
  border-bottom: none;
}

.summary-label {
  font-weight: 600;
  color: #495057;
  font-size: 14px;
}

.summary-value {
  font-weight: bold;
  color: #2c3e50;
  font-size: 16px;
}

.notes-section {
  margin-top: 30px;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 2px solid #e0e0e0;
}

.notes-content {
  padding: 15px;
  background-color: #f9fafb;
  border-radius: 6px;
}

.notes-content p {
  margin: 0;
  color: #2c3e50;
  line-height: 1.6;
}

.footer {
  margin-top: 20px;
  padding-top: 15px;
  border-top: 1px solid #e9ecef;
  text-align: center;
}

.footer-content {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  padding: 15px;
  border-radius: 6px;
}

.footer-note {
  font-size: 10px !important;
  color: #6c757d !important;
  font-style: italic;
  margin-top: 8px !important;
}

.bold-label {
  font-weight: bold !important;
}

/* Terms and Conditions Page */
.a4-page {
  page-break-before: always;
  padding-top: 20px;
}

.terms-page {
  margin-top: 20mm;
}

.contract-terms {
  padding: 25mm 20mm;
  font-family: 'Arial', 'Helvetica', sans-serif;
}

.contract-terms h3 {
  font-size: 18pt;
  font-weight: bold;
  color: #000;
  margin-bottom: 25px;
  text-align: center;
  border-bottom: 2px solid #000;
  padding-bottom: 10px;
  letter-spacing: 0.5px;
}

.terms-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
  margin-bottom: 30px;
}

.term-item {
  padding: 12px 0;
  border-bottom: 1px solid #e0e0e0;
}

.term-item:last-child {
  border-bottom: none;
}

.term-item p {
  margin: 0 0 10px 0;
  line-height: 1.8;
  color: #000;
  text-align: justify;
}

.term-item p:last-child {
  margin-bottom: 0;
}

.term-item p.english,
.term-item p[class*="english"] {
  font-weight: normal;
  color: #000;
}

.term-item p.chinese,
.term-item p[class*="chinese"] {
  font-weight: normal;
  color: #000;
  margin-top: 8px;
}

/* Support for any other language classes */
.term-item p:not(.english):not([class*="english"]):not(.chinese):not([class*="chinese"]) {
  font-weight: normal;
  color: #000;
  margin-top: 8px;
}

.signatures {
  display: flex;
  justify-content: space-between;
  margin-top: 50px;
  padding-top: 30px;
  border-top: 1px solid #000;
}

.signature-box {
  text-align: center;
  flex: 1;
  margin: 0 30px;
}

.signature-box p {
  margin: 0 0 50px 0;
  font-weight: bold;
  color: #000;
  font-size: 11pt;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.signature-line {
  height: 1px;
  background: #000;
  margin-top: 0;
  border-top: 1px solid #000;
}

@media print {
  .contract-terms {
    padding: 20mm 15mm;
  }
  
  .term-item {
    break-inside: avoid;
    page-break-inside: avoid;
  }
  
  .term-item p {
    orphans: 3;
    widows: 3;
  }
}
</style>
