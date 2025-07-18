<script setup>
import { ref, onMounted, defineProps, defineEmits } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import logoImage from '@/assets/logo.png'
import stampImage from '@/assets/gml2.png'
import letterHeadImage from '@/assets/letter_head.png'

const { t } = useEnhancedI18n()
const props = defineProps({
  carId: {
    type: Number,
    required: true,
  },
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
const carData = ref(null)
const billData = ref(null)

const company = ref({
  name: 'GROUP MERHAB LIMITED',
  address: 'GUANGZHOU, CHINA',
  phone: '+86 123 456 7890',
  email: 'contact@merhab.com',
  logo: logoImage,
})

const fetchCarData = async () => {
  try {
    const carResult = await callApi({
      query: `
        SELECT 
          cs.*,
          cn.car_name,
          clr.color,
          lp.loading_port,
          dp.discharge_port,
          c.name as client_name,
          bd.price_sell as buy_price,
          sb.bill_ref,
          sb.date_sell,
          sb.notes as bill_notes,
          broker.name as broker_name,
          broker.address as broker_address,
          broker.mobiles as broker_phone
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN sell_bill sb ON cs.id_sell = sb.id
        LEFT JOIN clients broker ON sb.id_broker = broker.id
        WHERE cs.id = ?
      `,
      params: [props.carId],
    })

    if (carResult.success && carResult.data.length > 0) {
      carData.value = carResult.data[0]
      billData.value = {
        id: carData.value.id_sell,
        bill_ref: carData.value.bill_ref,
        date_sell: carData.value.date_sell,
        notes: carData.value.bill_notes,
        broker_name: carData.value.broker_name,
        broker_address: carData.value.broker_address,
        broker_phone: carData.value.broker_phone,
      }
    } else {
      error.value = 'Failed to fetch car data'
    }
  } catch (err) {
    console.error('Error fetching car data:', err)
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

const calculateCarPrice = () => {
  if (!carData.value) return '0.00'

  let myPrice
  if (props.options.paymentTerms.toLowerCase() === 'cfr') {
    myPrice = (parseFloat(carData.value.price_cell) || 0) + (parseFloat(carData.value.freight) || 0)
  } else {
    myPrice = parseFloat(carData.value.price_cell) || 0
  }

  // Convert to selected currency
  if (props.options.currency.toUpperCase() === 'DA') {
    const rate = parseFloat(carData.value.rate)
    if (!rate) {
      console.warn(`No rate found for car ID ${carData.value.id}`)
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

onMounted(() => {
  fetchCarData()
})
</script>

<template>
  <div class="print-report">
    <div v-if="loading" class="loading">{{ t('sellBills.loading') }}</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else>
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
          ><strong>{{ t('sellBills.ci_no') }}:</strong>
          {{ billData ? billData.bill_ref + '-' + carData.id : '' }}</span
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
              <span class="info-value">{{ carData.client_name }}</span>
            </div>
            <div class="info-group">
              <label class="bold-label">{{ t('sellBills.address') }}:</label>
              <span class="info-value">{{ billData.broker_address }}</span>
            </div>
            <div class="info-group">
              <label class="bold-label">{{ t('sellBills.phone') }}:</label>
              <span class="info-value">{{ billData.broker_phone }}</span>
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
                <th>{{ t('sellBills.port') }}</th>
                <th>
                  {{ t('sellBills.price') }} {{ options.paymentTerms.toUpperCase() }}
                  {{ options.currency.toUpperCase() }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>1</td>
                <td>{{ carData.car_name }}</td>
                <td>{{ carData.color }}</td>
                <td>{{ carData.vin }}</td>
                <td>{{ carData.discharge_port }}</td>
                <td>{{ formatPrice(calculateCarPrice()) }}</td>
              </tr>
            </tbody>
            <tfoot>
              <tr>
                <td colspan="5" class="total-label">
                  <strong>{{ t('sellBills.total') }}:</strong>
                </td>
                <td class="total-value">
                  <strong>{{ formatPrice(calculateCarPrice()) }}</strong>
                </td>
              </tr>
            </tfoot>
          </table>
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
        </div>
      </div>

      <!-- Footer -->
      <div class="footer">
        <div class="footer-content"></div>
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

.buyer-details {
  padding: 15px;
}

.info-group {
  display: flex;
  flex-direction: column;
  gap: 3px;
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

.report-table th:last-child,
.report-table td:last-child {
  text-align: right;
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

.notes-content {
  padding: 15px;
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

/* Terms and Conditions Page */
.terms-page {
  margin-top: 20mm;
}

.contract-terms {
  padding: 20px;
}

.contract-terms h3 {
  font-size: 1.5rem;
  color: #2c3e50;
  margin-bottom: 20px;
  text-align: center;
  border-bottom: 2px solid #3498db;
  padding-bottom: 10px;
}

.terms-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.term-item {
  border-left: 3px solid #3498db;
  padding-left: 15px;
  background: #f8f9fa;
  padding: 15px;
  border-radius: 0 6px 6px 0;
}

.term-item p {
  margin: 0 0 8px 0;
  line-height: 1.6;
}

.term-item .english {
  color: #2c3e50;
  font-weight: 500;
}

.term-item .chinese {
  color: #7f8c8d;
  font-size: 0.9em;
}

.signatures {
  display: flex;
  justify-content: space-between;
  margin-top: 40px;
  padding-top: 20px;
  border-top: 1px solid #e9ecef;
}

.signature-box {
  text-align: center;
  flex: 1;
  margin: 0 20px;
}

.signature-box p {
  margin: 0 0 10px 0;
  font-weight: bold;
  color: #2c3e50;
}

.signature-line {
  height: 2px;
  background: #2c3e50;
  margin-top: 20px;
}
</style>
