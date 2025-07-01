<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import logoImage from '@/assets/logo.png'
import stampImage from '@/assets/gml2.png'

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
        LEFT JOIN colors clr ON bd.id_color = clr.id
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
    return `DZD ${price}`
  }
  return `USD ${price}`
}

onMounted(() => {
  fetchCarData()
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
      <div class="car_print_header">
        <div class="company-info">
          <img :src="company.logo" alt="Company Logo" class="company-logo" />
          <div class="company-text">
            <h1 class="company-name">{{ company.name }}</h1>
            <div class="company-details">
              <p><i class="fas fa-map-marker-alt"></i> {{ company.address }}</p>
              <p><i class="fas fa-phone"></i> {{ company.phone }}</p>
              <p><i class="fas fa-envelope"></i> {{ company.email }}</p>
            </div>
          </div>
        </div>
        <div class="document-info">
          <div class="document-header">
            <h2 class="document-title">COMMERCIAL INVOICE</h2>
          </div>
          <div class="document-details">
            <div class="detail-row">
              <span class="detail-label">CI No:</span>
              <span class="detail-value">{{ billData.bill_ref }}-{{ carData.id }}</span>
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
              <span class="info-value">{{ carData.client_name }}</span>
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
      <div class="section vehicle-section">
        <div class="section-header">
          <h3><i class="fas fa-car"></i> Vehicle Details</h3>
        </div>
        <div class="table-container">
          <table class="vehicle-table">
            <thead>
              <tr>
                <th class="col-vehicle">Vehicle</th>
                <th class="col-color">Color</th>
                <th class="col-vin">VIN</th>
                <th class="col-port">Port</th>
                <th class="col-price">
                  Price {{ options.paymentTerms.toUpperCase() }}
                  {{ options.currency.toUpperCase() }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr class="vehicle-row">
                <td class="col-vehicle">{{ carData.car_name }}</td>
                <td class="col-color">{{ carData.color }}</td>
                <td class="col-vin">{{ carData.vin }}</td>
                <td class="col-port">{{ carData.discharge_port }}</td>
                <td class="col-price">{{ formatPrice(calculateCarPrice()) }}</td>
              </tr>
            </tbody>
            <tfoot>
              <tr class="total-row">
                <td colspan="4" class="total-label">
                  <strong>Total:</strong>
                </td>
                <td class="total-value">
                  <strong>{{ formatPrice(calculateCarPrice()) }}</strong>
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>

      <!-- Payment Details - Compact -->
      <div class="section payment-section">
        <div class="section-header">
          <h3><i class="fas fa-credit-card"></i> Payment Details</h3>
        </div>
        <div class="payment-details">
          <div class="info-row">
            <span class="info-label">Payment Terms:</span>
            <span class="info-value">{{ options.paymentTerms.toUpperCase() }}</span>
            <span class="info-label">Mode:</span>
            <span class="info-value">{{ options.paymentMode }}</span>
          </div>
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
.car_print_header {
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

/* Vehicle Section - Compact */
.vehicle-section {
  border: 1px solid #e9ecef;
}

.table-container {
  padding: 15px;
}

.vehicle-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 12px;
  background: white;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.vehicle-table th {
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

.vehicle-table td {
  padding: 6px;
  border-bottom: 1px solid #e9ecef;
  font-size: 11px;
  vertical-align: middle;
}

.vehicle-row:hover {
  background-color: #f8f9fa;
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

/* Payment Section - Compact */
.payment-section {
  border: 1px solid #e9ecef;
}

.payment-details {
  padding: 12px;
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
  opacity: 1;
  pointer-events: none;
  z-index: 100;
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
}

.floating-stamp img {
  width: 250px;
  height: auto;
}

/* Print Styles - Preserve Colors for Color Printing */
@media print {
  .a4-page {
    box-shadow: none;
    margin: 0;
    page-break-inside: avoid;
    background: white;
  }

  .floating-stamp {
    opacity: 1;
    position: fixed;
  }

  /* Preserve gradients and colors for color printing */
  .section {
    box-shadow: none;
    border: 1px solid #dee2e6;
  }

  .vehicle-table {
    box-shadow: none;
    border: 1px solid #dee2e6;
  }

  /* Preserve table header gradient */
  .vehicle-table th {
    background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%) !important;
    color: white !important;
    -webkit-print-color-adjust: exact;
    color-adjust: exact;
  }

  /* Preserve total row gradient */
  .total-row {
    background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%) !important;
    -webkit-print-color-adjust: exact;
    color-adjust: exact;
  }

  /* Preserve total value background */
  .total-value {
    background: #3498db !important;
    color: white !important;
    -webkit-print-color-adjust: exact;
    color-adjust: exact;
  }

  /* Preserve footer gradient */
  .footer-content {
    background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%) !important;
    -webkit-print-color-adjust: exact;
    color-adjust: exact;
  }

  /* Preserve section headers */
  .section-header h3 {
    background: linear-gradient(135deg, #3498db 0%, #2980b9 100%) !important;
    color: white !important;
    -webkit-print-color-adjust: exact;
    color-adjust: exact;
  }

  /* Preserve document details gradient */
  .document-details {
    background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%) !important;
    -webkit-print-color-adjust: exact;
    color-adjust: exact;
  }

  /* Ensure proper color printing */
  * {
    -webkit-print-color-adjust: exact;
    color-adjust: exact;
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
</style>
 