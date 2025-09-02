<script setup>
import { ref, onMounted,   } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import logoImage from '@/assets/logo.png'
import stampImage from '@/assets/gml2.png'
import letterHeadImage from '@/assets/letter_head.png'

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

const { callApi } = useApi()
const loading = ref(true)
const error = ref(null)
const billData = ref(null)
const carsData = ref([])
const banks = ref([])
const selectedBank = ref(null)

const company = ref({
  name: 'GROUP MERHAB LIMITED',
  address: 'GUANGZHOU, CHINA',
  phone: '+86 123 456 7890',
  email: 'contact@merhab.com',
  logo: logoImage,
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
      billData.value = billResult.data[0]
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

onMounted(() => {
  fetchBillData()
  fetchBanks()
})
</script>

<template>
  <div class="print-report">
    <!-- Header Image -->
    <div class="report-header">
      <img :src="letterHeadImage" alt="Letter Head" class="letter-head" />
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
        <table class="report-table">
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
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td colspan="6" class="total-label">
                <strong>{{ t('sellBills.total') }}:</strong>
              </td>
              <td class="total-value">
                <strong>{{ formatPrice(calculateTotal()) }}</strong>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
      <div class="amount-in-words">
        <span class="words-label">{{ t('sellBills.total') }} in words:</span>
        <span class="words-value">
          <strong
            >{{ numberToWords(calculateTotal()) }} {{ options.currency.toUpperCase() }}</strong
          >
        </span>
      </div>
    </div>

    <!-- Payment & Bank Information -->
    <div class="section payment-bank-section">
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
          <div class="info-row">
            <span class="info-label bold-label">{{ t('sellBills.bank_information') }}:</span>
            <span class="info-value">{{ selectedBank.bank_name }}</span>
          </div>
          <div class="info-row">
            <span class="info-label bold-label">Account:</span>
            <span class="info-value">{{ selectedBank.bank_account }}</span>
          </div>
          <div class="info-row">
            <span class="info-label bold-label">SWIFT:</span>
            <span class="info-value">{{ selectedBank.swift_code }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Notes Section -->
    <div class="section notes-section" v-if="billData?.notes">
      <div class="notes-content">
        <p>{{ billData.notes }}</p>
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
          <!-- Payment Terms -->
          <div class="term-item">
            <p class="english">1. Payment must be made in full before vehicle delivery.</p>
            <p class="chinese">1. 车辆交付前必须全额付款。</p>
          </div>

          <!-- Delivery Terms -->
          <div class="term-item">
            <p class="english">
              2. Delivery will be made within the agreed timeframe after receipt of full payment.
            </p>
            <p class="chinese">2. 收到全额付款后，将在约定的时间范围内交付。</p>
          </div>

          <!-- Document Responsibility -->
          <div class="term-item">
            <p class="english">
              3. All vehicle documentation will be provided as per export requirements.
            </p>
            <p class="chinese">3. 将按照出口要求提供所有车辆文件。</p>
          </div>

          <!-- Customs Clause -->
          <div class="term-item">
            <p class="english">
              4. In case of any customs issues related to modified export license usage, while we
              will assist in resolving the matter with customs authorities, we bear no legal
              responsibility for such complications.
            </p>
            <p class="chinese">
              4.
              如因改装出口许可证使用而产生任何海关问题，虽然我们会协助处理海关事务，但我们不承担任何法律责任。
            </p>
          </div>

          <!-- Inspection Terms -->
          <div class="term-item">
            <p class="english">
              5. The buyer acknowledges having inspected the vehicle and accepts its condition at
              the time of purchase.
            </p>
            <p class="chinese">5. 买方确认已检查车辆并接受其购买时的状况。</p>
          </div>

          <!-- Dispute Resolution -->
          <div class="term-item">
            <p class="english">
              6. Any disputes arising from this contract shall be resolved through mutual
              negotiation or applicable legal channels.
            </p>
            <p class="chinese">6. 本合同引起的任何争议应通过双方协商或适用的法律途径解决。</p>
          </div>

          <!-- Freight Increase Clause -->
          <div class="term-item">
            <p class="english">
              7. If freight costs increase by more than USD 200 on the loading day, the client is
              responsible for paying the difference.
            </p>
            <p class="chinese">7. 如果装运当天的运费增加超过200美元，客户需要支付差额。</p>
          </div>

          <!-- Chinese Tax/Policy Changes -->
          <div class="term-item">
            <p class="english">
              8. The client is responsible for any additional costs resulting from new Chinese taxes
              or policy changes implemented after the contract date.
            </p>
            <p class="chinese">
              8. 合同日期之后实施的任何新的中国税收或政策变化所产生的额外费用由客户承担。
            </p>
          </div>

          <!-- Price Agreement and Changes -->
          <div class="term-item">
            <p class="english">
              9. The agreed price is valid as of the agreement date. Any price changes due to
              shipping companies or Chinese government regulations on the loading day will be the
              client's responsibility.
            </p>
            <p class="chinese">
              9.
              约定价格自协议日期起有效。装运当天因航运公司或中国政府法规造成的任何价格变化由客户承担。
            </p>
          </div>

          <!-- Shipping Time Disclaimer -->
          <div class="term-item">
            <p class="english">
              10. The parties hereto expressly acknowledge and agree that shipping and delivery
              timelines, schedules, and estimated arrival dates are provided for informational
              purposes only and are subject to modification, delay, or cancellation due to
              circumstances beyond the seller's reasonable control. Such circumstances include,
              without limitation, customs clearance delays, adverse weather conditions, port
              congestion, shipping line schedule modifications, force majeure events, governmental
              actions, labor disputes, equipment failures, or other unforeseeable events. The seller
              shall not be liable for any direct, indirect, incidental, consequential, or punitive
              damages arising from such delays, and the buyer hereby expressly waives, releases, and
              discharges any and all claims, demands, actions, causes of action, damages, costs,
              expenses, or compensation of any nature whatsoever arising therefrom.
            </p>
            <p class="chinese">
              10.
              本协议双方明确确认并同意，运输和交付时间、日程安排及预计到达日期仅供参考，可能因超出卖方合理控制范围的情况而修改、延迟或取消。此类情况包括但不限于海关清关延迟、恶劣天气条件、港口拥堵、航运公司时刻表变更、不可抗力事件、政府行为、劳资纠纷、设备故障或其他不可预见事件。卖方对此类延迟造成的任何直接、间接、偶然、后果性或惩罚性损害不承担责任，买方特此明确放弃、免除和解除由此产生的任何性质的所有索赔、要求、诉讼、诉因、损害、费用、开支或补偿。
            </p>
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
</style>
