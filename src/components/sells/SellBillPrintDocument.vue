<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import logoImage from '@/assets/logo.png'
import stampImage from '@/assets/gml2.png'

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
          LEFT JOIN colors clr ON bd.id_color = clr.id
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
    return `DZD ${price}`
  }
  return `USD ${price}`
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
  <div class="print-document">
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else class="a4-page">
      <!-- Floating Stamp -->
      <div class="floating-stamp">
        <img :src="stampImage" alt="Company Stamp" />
      </div>

      <!-- Professional Header -->
      <div class="sell_bill_header">
        <div class="company-info">
          <img :src="company.logo" alt="Company Logo" class="company-logo" />
          <div class="company-text">
            <h1 class="company-name">{{ selectedBank?.company_name || company.name }}</h1>
            <div class="company-details">
              <p>
                <i class="fas fa-map-marker-alt"></i>
                {{ selectedBank?.bank_address || company.address }}
              </p>
              <p><i class="fas fa-phone"></i> {{ company.phone }}</p>
              <p><i class="fas fa-envelope"></i> {{ company.email }}</p>
            </div>
          </div>
        </div>
        <div class="document-info">
          <div class="document-header">
            <h2 class="document-title">
              {{
                options.documentType === 'contract'
                  ? 'SALE CONTRACT'
                  : options.documentType === 'invoice'
                    ? 'INVOICE'
                    : 'PROFORMA INVOICE'
              }}
            </h2>
          </div>
          <div class="document-details">
            <div class="detail-row">
              <span class="detail-label">Document No:</span>
              <span class="detail-value">{{ billData.bill_ref }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Date:</span>
              <span class="detail-value">{{ formatDate(billData.date_sell) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Currency:</span>
              <span class="detail-value">{{ options.currency.toUpperCase() }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Buyer Information Section -->
      <div class="section buyer-section">
        <div class="section-header">
          <h3><i class="fas fa-user"></i> Buyer Information</h3>
        </div>
        <div class="buyer-details">
          <div class="buyer-info">
            <div class="info-group">
              <label>Name:</label>
              <span class="info-value">{{ billData.broker_name }}</span>
            </div>
            <div class="info-group">
              <label>Address:</label>
              <span class="info-value">{{ billData.broker_address }}</span>
            </div>
            <div class="info-group">
              <label>Phone:</label>
              <span class="info-value">{{ billData.broker_phone }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Vehicle Details Table -->
      <div class="section vehicles-section">
        <div class="section-header">
          <h3><i class="fas fa-car"></i> Vehicle Details</h3>
        </div>
        <div class="table-container">
          <table class="vehicles-table">
            <thead>
              <tr>
                <th class="col-no">No.</th>
                <th class="col-vehicle">Vehicle</th>
                <th class="col-color">Color</th>
                <th class="col-vin">VIN</th>
                <th class="col-client">Client</th>
                <th class="col-port">Port</th>
                <th class="col-price">
                  Price {{ options.paymentTerms.toUpperCase() }}
                  {{ props.options.currency.toUpperCase() }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(car, index) in carsData" :key="car.id" class="vehicle-row">
                <td class="col-no">{{ index + 1 }}</td>
                <td class="col-vehicle">{{ car.car_name }}</td>
                <td class="col-color">{{ car.color }}</td>
                <td class="col-vin">{{ car.vin }}</td>
                <td class="col-client">{{ car.client_name }}</td>
                <td class="col-port">{{ car.discharge_port }}</td>
                <td class="col-price">{{ formatPrice(calculateCarPrice(car)) }}</td>
              </tr>
            </tbody>
            <tfoot>
              <tr class="total-row">
                <td colspan="6" class="total-label">
                  <strong>Grand Total:</strong>
                </td>
                <td class="total-value">
                  <strong>{{ formatPrice(calculateTotal()) }}</strong>
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
        <div class="amount-in-words">
          <span class="words-label">Amount in words:</span>
          <span class="words-value">
            <strong
              >{{ numberToWords(calculateTotal()) }}
              {{ props.options.currency.toUpperCase() }}</strong
            >
          </span>
        </div>
      </div>

      <!-- Payment & Bank Information - Compact -->
      <div class="section payment-bank-section">
        <div class="section-header">
          <h3><i class="fas fa-credit-card"></i> Payment & Bank Information</h3>
        </div>
        <div class="payment-bank-details">
          <div class="payment-info">
            <div class="info-row">
              <span class="info-label">Payment Terms:</span>
              <span class="info-value">{{ options.paymentTerms.toUpperCase() }}</span>
              <span class="info-label">Mode:</span>
              <span class="info-value">{{ options.paymentMode }}</span>
            </div>
          </div>
          <div class="bank-info" v-if="selectedBank">
            <div class="info-row">
              <span class="info-label">Bank:</span>
              <span class="info-value">{{ selectedBank.bank_name }}</span>
              <span class="info-label">Account:</span>
              <span class="info-value">{{ selectedBank.bank_account }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">SWIFT:</span>
              <span class="info-value">{{ selectedBank.swift_code }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Notes Section -->
      <div class="section notes-section" v-if="billData.notes">
        <div class="section-header">
          <h3><i class="fas fa-sticky-note"></i> Notes</h3>
        </div>
        <div class="notes-content">
          <p>{{ billData.notes }}</p>
        </div>
      </div>

      <!-- Footer -->
      <div class="footer">
        <div class="footer-content">
          <p>Thank you for your business!</p>
          <p class="footer-note">
            This document is computer generated and valid without signature.
          </p>
        </div>
      </div>
    </div>

    <!-- Second Page (Terms and Conditions) -->
    <div v-if="options.documentType === 'contract'" class="a4-page terms-page">
      <!-- Floating Stamp -->
      <div class="floating-stamp">
        <img :src="stampImage" alt="Company Stamp" />
      </div>

      <!-- Contract Terms -->
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

      <!-- Signatures -->
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
/* Professional Invoice Design - Compact for 2-page printing */
.a4-page {
  width: 210mm;
  min-height: 297mm;
  padding: 15mm;
  margin: 0 auto;
  background: white;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
  position: relative;
  margin-bottom: 20mm;
  font-family: 'Arial', 'Helvetica', sans-serif;
  color: #2c3e50;
  line-height: 1.4;
}

@media print {
  .a4-page {
    box-shadow: none;
    margin: 0;
    padding: 12mm;
  }
}

/* Header Styles - Compact */
.sell_bill_header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 25px;
  padding-bottom: 15px;
  border-bottom: 2px solid #3498db;
}

.company-info {
  display: flex;
  align-items: center;
  gap: 15px;
  flex: 1;
}

.company-logo {
  width: 80px;
  height: auto;
  object-fit: contain;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
}

.company-text {
  display: flex;
  flex-direction: column;
}

.company-name {
  font-size: 20px;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.company-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.company-details p {
  margin: 0;
  color: #7f8c8d;
  font-size: 11px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.company-details i {
  color: #3498db;
  width: 12px;
}

.document-info {
  text-align: right;
  min-width: 250px;
}

.document-header {
  margin-bottom: 12px;
}

.document-title {
  font-size: 24px;
  font-weight: bold;
  color: #2c3e50;
  margin: 0 0 10px 0;
  text-transform: uppercase;
  letter-spacing: 1px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.document-details {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  padding: 12px;
  border-radius: 6px;
  border-left: 3px solid #3498db;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  padding: 4px 0;
  border-bottom: 1px solid #dee2e6;
}

.detail-row:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.detail-label {
  font-weight: 600;
  color: #495057;
  font-size: 11px;
}

.detail-value {
  font-weight: bold;
  color: #2c3e50;
  font-size: 12px;
}

/* Section Styles - Compact */
.section {
  margin-bottom: 20px;
  background: #ffffff;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.section-header {
  background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
  padding: 10px 15px;
  margin: 0;
}

.section-header h3 {
  color: white;
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header i {
  font-size: 12px;
}

/* Buyer Section - Compact */
.buyer-section {
  border: 1px solid #e9ecef;
}

.buyer-details {
  padding: 15px;
}

.buyer-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.info-group {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.info-group label {
  font-weight: 600;
  color: #495057;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.info-value {
  font-size: 12px;
  color: #2c3e50;
  font-weight: 500;
  padding: 6px 8px;
  background: #f8f9fa;
  border-radius: 3px;
  border-left: 2px solid #3498db;
}

/* Vehicles Section - Compact */
.vehicles-section {
  border: 1px solid #e9ecef;
}

.table-container {
  padding: 15px;
}

.vehicles-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 12px;
  background: white;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.vehicles-table th {
  background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
  color: white;
  padding: 8px 6px;
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  text-align: left;
  border: none;
}

.vehicles-table td {
  padding: 6px;
  border-bottom: 1px solid #e9ecef;
  font-size: 11px;
  vertical-align: middle;
}

.vehicle-row:hover {
  background-color: #f8f9fa;
}

.vehicle-row:nth-child(even) {
  background-color: #fafbfc;
}

.vehicle-row:nth-child(even):hover {
  background-color: #f1f3f4;
}

.col-no {
  width: 40px;
  text-align: center;
  font-weight: bold;
}

.col-vehicle {
  width: 120px;
}

.col-color {
  width: 80px;
}

.col-vin {
  width: 140px;
  font-family: 'Courier New', monospace;
  font-weight: 600;
}

.col-client {
  width: 100px;
}

.col-port {
  width: 80px;
}

.col-price {
  width: 100px;
  text-align: right;
  font-weight: bold;
}

.total-row {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%) !important;
  border-top: 1px solid #3498db;
}

.total-label {
  text-align: right;
  font-size: 12px;
  color: #2c3e50;
  padding: 8px 6px;
}

.total-value {
  text-align: right;
  font-size: 14px;
  color: #2c3e50;
  padding: 8px 6px;
  background: #3498db;
  color: white;
  font-weight: bold;
}

.amount-in-words {
  background: linear-gradient(135deg, #fff3cd 0%, #ffeaa7 100%);
  padding: 10px 12px;
  border-radius: 4px;
  border-left: 3px solid #f39c12;
  margin-top: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.words-label {
  font-weight: 600;
  color: #856404;
  font-size: 11px;
  text-transform: uppercase;
}

.words-value {
  font-weight: bold;
  color: #856404;
  font-size: 12px;
  font-style: italic;
}

/* Two Columns Layout - Compact */
.two-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
}

@media (max-width: 768px) {
  .two-columns {
    grid-template-columns: 1fr;
  }
}

/* Payment & Bank Section - Compact */
.payment-bank-section {
  border: 1px solid #e9ecef;
}

.payment-bank-details {
  padding: 12px;
}

.payment-info,
.bank-info {
  margin-bottom: 8px;
}

.payment-info:last-child,
.bank-info:last-child {
  margin-bottom: 0;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}

.info-row:last-child {
  margin-bottom: 0;
}

.info-label {
  font-weight: 600;
  color: #495057;
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  min-width: 80px;
}

.info-value {
  font-weight: bold;
  color: #2c3e50;
  font-size: 11px;
  padding: 3px 6px;
  background: transparent;
  border-radius: 3px;
  border: none;
  margin-right: 12px;
}

/* Notes Section - Compact */
.notes-section {
  border: 1px solid #e9ecef;
}

.notes-content {
  padding: 15px;
}

.notes-content p {
  margin: 0;
  font-size: 11px;
  line-height: 1.4;
  color: #495057;
  background: #f8f9fa;
  padding: 10px;
  border-radius: 4px;
  border-left: 3px solid #6c757d;
}

/* Footer - Compact */
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

.footer-content p {
  margin: 0 0 6px 0;
  font-size: 12px;
  color: #2c3e50;
  font-weight: 500;
}

.footer-note {
  font-size: 10px !important;
  color: #6c757d !important;
  font-style: italic;
  margin-top: 8px !important;
}

/* Floating Stamp - Smaller */
.floating-stamp {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-30deg);
  opacity: 0.15;
  pointer-events: none;
  z-index: 100;
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
}

.floating-stamp img {
  width: 250px;
  height: auto;
}

/* Print Styles */
@media print {
  .a4-page {
    box-shadow: none;
    margin: 0;
    page-break-inside: avoid;
  }

  .floating-stamp {
    opacity: 0.1;
  }

  .section {
    box-shadow: none;
    border: 1px solid #dee2e6;
  }

  .vehicles-table {
    box-shadow: none;
    border: 1px solid #dee2e6;
  }

  .two-columns {
    break-inside: avoid;
  }
}

/* Legacy styles for compatibility */
.loading,
.error {
  text-align: center;
  padding: 20px;
  font-size: 16px;
}

.loading {
  color: #3498db;
}

.error {
  color: #e74c3c;
}

.text-right {
  text-align: right;
}

.signatures {
  display: flex;
  justify-content: space-between;
  margin-top: 30px;
}

.signature-box {
  width: 150px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.signature-line {
  margin-top: 30px;
  border-top: 1px solid #000;
}

/* Bank selection styles */
.bank-selection {
  margin-bottom: 20px;
  padding: 10px;
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
}

.bank-selection label {
  margin-right: 10px;
  font-weight: 600;
}

.bank-selection select {
  padding: 5px 10px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  min-width: 200px;
}

/* Contract terms styles - Compact */
.contract-terms {
  margin: 20px 0;
  page-break-inside: avoid;
  font-size: 9px;
}

.terms-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.term-item {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 8px;
  background-color: #f9fafb;
  border-radius: 4px;
}

.term-item p {
  margin: 0;
  line-height: 1.3;
}

.english {
  color: #1f2937;
}

.arabic {
  font-family: 'Arial', sans-serif;
  direction: rtl;
  text-align: right;
  color: #374151;
}

.chinese {
  font-family: 'SimSun', 'Microsoft YaHei', sans-serif;
  color: #4b5563;
}

@media print {
  .print-hide {
    display: none;
  }

  .term-item {
    background-color: transparent;
    border: 1px solid #e5e7eb;
  }
}

.company-stamp {
  width: 80px;
  height: auto;
  margin-top: 8px;
  object-fit: contain;
}

.page-footer {
  margin-top: 20px;
  text-align: center;
  font-size: 10px;
  color: #6c757d;
}
</style>
