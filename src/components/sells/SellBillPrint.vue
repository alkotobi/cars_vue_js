<script setup>
import { ref, onMounted, watch, defineProps } from 'vue'
import { useApi } from '../../composables/useApi'
import logo from '../../assets/logo.png' // Import the logo

const props = defineProps({
  billId: {
    type: Number,
    required: true
  },
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close'])

const { callApi } = useApi()
const loading = ref(true)
const error = ref(null)
const billData = ref(null)
const billCars = ref([])
const companyInfo = ref({
  name: 'GROUP MERHAB LIMITED',
  address: 'GUANGZHOU, CHINA',
  phone: '+86 123 456 7890',
  email: 'merhab@merhab.com',
  logo: logo // Use the imported logo
})

// Fetch bill data and associated cars
const fetchBillData = async () => {
  if (!props.billId) return
  
  loading.value = true
  error.value = null
  
  try {
    // Fetch bill details - removing the non-existent columns
    const billResult = await callApi({
      query: `
        SELECT 
          sb.id,
          sb.date_sell,
          sb.notes,
          c.name as broker_name,
          c.address as broker_address
        FROM sell_bill sb
        LEFT JOIN clients c ON sb.id_broker = c.id
        WHERE sb.id = ?
      `,
      params: [props.billId]
    })
    
    if (!billResult.success) {
      error.value = billResult.error || 'Failed to fetch bill data'
      return
    }
    
    if (billResult.data.length === 0) {
      error.value = 'Bill not found'
      return
    }
    
    billData.value = billResult.data[0]
    
    // Fetch cars associated with this bill
    const carsResult = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.notes,
          cs.price_cell,
          cs.freight,
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
        WHERE cs.id_sell = ?
        ORDER BY cs.id
      `,
      params: [props.billId]
    })
    
    if (!carsResult.success) {
      error.value = carsResult.error || 'Failed to fetch cars data'
      return
    }
    
    billCars.value = carsResult.data
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Calculate totals
const calculateTotals = () => {
  let totalSellPrice = 0
  let totalFreight = 0
  
  billCars.value.forEach(car => {
    if (car.price_cell) totalSellPrice += parseFloat(car.price_cell)
    if (car.freight) totalFreight += parseFloat(car.freight)
  })
  
  return {
    totalSellPrice: totalSellPrice.toFixed(2),
    totalFreight: totalFreight.toFixed(2),
    grandTotal: (totalSellPrice + totalFreight).toFixed(2)
  }
}

// Format date for display
const formatDate = (dateString) => {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

// Print the bill
const printBill = () => {
  window.print()
}

// Close the print view
const closePrintView = () => {
  emit('close')
}

// Watch for changes in billId or visibility
watch(() => [props.billId, props.visible], ([newBillId, newVisible]) => {
  if (newVisible && newBillId) {
    fetchBillData()
  }
}, { immediate: true })
</script>

<template>
  <div v-if="visible" class="print-overlay">
    <div class="print-container">
      <div class="print-controls no-print">
        <button @click="printBill" class="print-btn">Print</button>
        <button @click="closePrintView" class="close-btn">Close</button>
      </div>
      
      <div v-if="loading" class="loading">Loading bill data...</div>
      
      <div v-else-if="error" class="error">{{ error }}</div>
      
      <div v-else-if="billData" class="bill-document">
        <!-- Header -->
        <div class="bill-header">
          <div class="company-info">
            <img :src="companyInfo.logo" alt="Company Logo" class="company-logo" />
            <div class="company-details">
              <h2>{{ companyInfo.name }}</h2>
              <p>{{ companyInfo.address }}</p>
              <p>{{ companyInfo.phone }} | {{ companyInfo.email }}</p>
            </div>
          </div>
          
          <div class="bill-info">
            <h1>SELL BILL</h1>
            <p><strong>Bill #:</strong> {{ billData.id }}</p>
            <p><strong>Date:</strong> {{ formatDate(billData.date_sell) }}</p>
          </div>
        </div>
        
        <!-- Broker Info -->
        <div class="broker-info">
          <h3>Broker Information</h3>
          <p><strong>Name:</strong> {{ billData.broker_name || 'N/A' }}</p>
          <p v-if="billData.broker_address"><strong>Address:</strong> {{ billData.broker_address }}</p>
          <!-- Remove or conditionally render phone and email fields -->
        </div>
        
        <!-- Cars Table -->
        <div class="cars-section">
          <h3>Cars</h3>
          <table class="cars-table">
            <thead>
              <tr>
                <th>#</th>
                <th>Car</th>
                <th>Color</th>
                <th>VIN</th>
                <th>Client</th>
                <th>Loading Port</th>
                <th>Discharge Port</th>
                <th>Sell Price</th>
                <th>Freight</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(car, index) in billCars" :key="car.id">
                <td>{{ index + 1 }}</td>
                <td>{{ car.car_name || 'N/A' }}</td>
                <td>{{ car.color || 'N/A' }}</td>
                <td>{{ car.vin || 'N/A' }}</td>
                <td>{{ car.client_name || 'N/A' }}</td>
                <td>{{ car.loading_port || 'N/A' }}</td>
                <td>{{ car.discharge_port || 'N/A' }}</td>
                <td>${{ car.price_cell || '0.00' }}</td>
                <td>${{ car.freight || '0.00' }}</td>
              </tr>
            </tbody>
            <tfoot>
              <tr>
                <td colspan="7" class="total-label">Totals:</td>
                <td>${{ calculateTotals().totalSellPrice }}</td>
                <td>${{ calculateTotals().totalFreight }}</td>
              </tr>
              <tr>
                <td colspan="7" class="grand-total-label">Grand Total:</td>
                <td colspan="2" class="grand-total-value">${{ calculateTotals().grandTotal }}</td>
              </tr>
            </tfoot>
          </table>
        </div>
        
        <!-- Notes -->
        <div v-if="billData.notes" class="notes-section">
          <h3>Notes</h3>
          <p>{{ billData.notes }}</p>
        </div>
        
        <!-- Footer -->
        <div class="bill-footer">
          <div class="signatures">
            <div class="signature-box">
              <p>Authorized Signature</p>
              <div class="signature-line"></div>
            </div>
            <div class="signature-box">
              <p>Client Signature</p>
              <div class="signature-line"></div>
            </div>
          </div>
          <p class="thank-you">Thank you for your business!</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Print-specific styles */
@media print {
  .no-print {
    display: none !important;
  }
  
  body {
    margin: 0;
    padding: 0;
  }
  
  .print-overlay {
    position: static;
    background: none;
    padding: 0;
    overflow: visible;
  }
  
  .print-container {
    box-shadow: none;
    margin: 0;
    padding: 0;
    max-height: none;
    overflow: visible;
  }
  
  .bill-document {
    padding: 0;
  }
}

/* Regular styles */
.print-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  overflow: auto;
  padding: 20px;
}

.print-container {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  width: 100%;
  max-width: 1000px;
  max-height: 90vh;
  overflow: auto;
  position: relative;
}

.print-controls {
  position: sticky;
  top: 0;
  display: flex;
  justify-content: flex-end;
  padding: 10px;
  background-color: #f3f4f6;
  border-bottom: 1px solid #e5e7eb;
  z-index: 10;
}

.print-btn, .close-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  margin-left: 10px;
}

.print-btn {
  background-color: #3b82f6;
  color: white;
}

.close-btn {
  background-color: #6b7280;
  color: white;
}

.loading, .error {
  padding: 40px;
  text-align: center;
  font-size: 1.1rem;
}

.error {
  color: #ef4444;
}

.bill-document {
  padding: 40px;
}

.bill-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 30px;
}

.company-info {
  display: flex;
  align-items: center;
}

.company-logo {
  max-width: 100px;
  max-height: 100px;
  margin-right: 20px;
}

.company-details h2 {
  margin: 0 0 5px 0;
  color: #1f2937;
}

.company-details p {
  margin: 0 0 5px 0;
  color: #4b5563;
}

.bill-info {
  text-align: right;
}

.bill-info h1 {
  color: #1f2937;
  margin: 0 0 10px 0;
}

.broker-info {
  margin-bottom: 30px;
  padding: 15px;
  background-color: #f9fafb;
  border-radius: 6px;
}

.broker-info h3 {
  margin-top: 0;
  color: #1f2937;
}

.broker-info p {
  margin: 5px 0;
}

.cars-section {
  margin-bottom: 30px;
}

.cars-section h3 {
  color: #1f2937;
}

.cars-table {
  width: 100%;
  border-collapse: collapse;
}

.cars-table th, .cars-table td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.cars-table th {
  background-color: #f3f4f6;
  font-weight: 600;
}

.cars-table tfoot td {
  font-weight: 600;
}

.total-label, .grand-total-label {
  text-align: right;
}

.grand-total-label, .grand-total-value {
  font-size: 1.1rem;
  background-color: #f3f4f6;
}

.notes-section {
  margin-bottom: 30px;
  padding: 15px;
  background-color: #f9fafb;
  border-radius: 6px;
}

.notes-section h3 {
  margin-top: 0;
  color: #1f2937;
}

.bill-footer {
  margin-top: 50px;
}

.signatures {
  display: flex;
  justify-content: space-between;
  margin-bottom: 30px;
}

.signature-box {
  width: 45%;
}

.signature-line {
  margin-top: 50px;
  border-top: 1px solid #9ca3af;
}

.thank-you {
  text-align: center;
  font-size: 1.1rem;
  color: #4b5563;
  margin-top: 30px;
}
</style>