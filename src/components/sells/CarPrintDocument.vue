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

  // Always use CFR (price + freight)
  const myPrice =
    (parseFloat(carData.value.price_cell) || 0) + (parseFloat(carData.value.freight) || 0)

  if (props.options.currency.toUpperCase() === 'DA') {
    const rate = parseFloat(carData.value.rate)
    if (!rate) return 'Rate not set'
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

      <!-- Header -->
      <div class="header print-hide">
        <div class="company-info">
          <img :src="company.logo" alt="Company Logo" class="company-logo" />
          <div class="company-text">
            <h1>{{ company.name }}</h1>
            <p>{{ company.address }}</p>
            <p>Tel: {{ company.phone }}</p>
            <p>Email: {{ company.email }}</p>
          </div>
        </div>
        <div class="document-info">
          <p>CI No: {{ billData.bill_ref }}-{{ carData.id }}</p>
          <p>Date: {{ formatDate(billData.date_sell) }}</p>
        </div>
      </div>

      <h2 style="text-align: center">COMMERCIAL INVOICE</h2>

      <!-- Broker Info -->
      <div class="section">
        <h3>Buyer Information</h3>
        <p><strong>Name:</strong> {{ carData.client_name }}</p>
        <p><strong>Address:</strong> {{ billData.broker_address }}</p>
        <p><strong>Phone:</strong> {{ billData.broker_phone }}</p>
      </div>

      <!-- Car Details -->
      <div class="section">
        <h3>Vehicle Details</h3>
        <table class="car-table">
          <thead>
            <tr>
              <th>Vehicle</th>
              <th>Color</th>
              <th>VIN</th>
              <th>Port</th>
              <th>Price CFR ({{ props.options.currency.toUpperCase() }})</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>{{ carData.car_name }}</td>
              <td>{{ carData.color }}</td>
              <td>{{ carData.vin }}</td>
              <td>{{ carData.discharge_port }}</td>
              <td>{{ formatPrice(calculateCarPrice()) }}</td>
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td :colspan="4" class="text-right"><strong>Total:</strong></td>
              <td>
                <strong>{{ formatPrice(calculateCarPrice()) }}</strong>
              </td>
            </tr>
          </tfoot>
        </table>
      </div>

      <!-- Payment Details -->
      <div class="section">
        <h3>Payment Details</h3>
        <p><strong>Terms:</strong> {{ options.paymentTerms.toUpperCase() }}</p>
        <p><strong>Mode:</strong> {{ options.paymentMode }}</p>
        <p><strong>Currency:</strong> {{ options.currency.toUpperCase() }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.a4-page {
  width: 210mm;
  min-height: 297mm;
  padding: 20mm;
  margin: 0 auto;
  background: white;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  position: relative;
  margin-bottom: 20mm;
}

@media print {
  .a4-page {
    box-shadow: none;
  }

  body {
    background: none;
  }

  .print-document {
    padding: 0;
    margin: 0;
  }

  .loading,
  .error {
    display: none !important;
  }

  .a4-page .header,
  .header {
    display: none !important;
    visibility: hidden !important;
    height: 0 !important;
    overflow: hidden !important;
  }

  .company-info,
  .document-info {
    display: none !important;
  }
}

.header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 40px;
}

.company-info {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 20px;
}

.company-logo {
  width: 100px;
  height: auto;
  object-fit: contain;
}

.company-text {
  display: flex;
  flex-direction: column;
}

.company-info h1 {
  font-size: 24px;
  color: #1f2937;
  margin-bottom: 10px;
}

.company-info p {
  margin: 5px 0;
  color: #4b5563;
}

.document-info {
  text-align: right;
}

.section {
  margin-bottom: 30px;
}

.section h3 {
  color: #1f2937;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 5px;
  margin-bottom: 15px;
}

.car-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 20px;
}

.car-table th,
.car-table td {
  border: 1px solid #e5e7eb;
  padding: 8px;
  text-align: left;
}

.car-table th:nth-child(3),
.car-table td:nth-child(3) {
  width: 200px;
  min-width: 200px;
}

.car-table th {
  background-color: #f9fafb;
  font-weight: 600;
}

.car-table tfoot td {
  background-color: #f9fafb;
}

.text-right {
  text-align: right;
}

.floating-stamp {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(-30deg);
  opacity: 1;
  pointer-events: none;
  z-index: 100;
}

.floating-stamp img {
  width: 300px;
  height: auto;
}

@media print {
  .floating-stamp {
    position: fixed;
  }
}
</style>

<style>
@media print {
  .print-hide,
  .header,
  .company-info,
  .document-info {
    display: none !important;
    visibility: hidden !important;
    height: 0 !important;
    overflow: hidden !important;
    margin: 0 !important;
    padding: 0 !important;
  }
}
</style>
 