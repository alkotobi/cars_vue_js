<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import logoImage from '@/assets/logo.png'

const props = defineProps({
  billId: {
    type: Number,
    required: true
  },
  options: {
    type: Object,
    required: true
  }
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
  logo: logoImage
})

const fetchBanks = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM banks ORDER BY company_name ASC',
      params: []
    })
    if (result.success && result.data.length > 0) {
      banks.value = result.data
      // Set the selected bank based on options
      if (props.options.bankId) {
        selectedBank.value = banks.value.find(bank => bank.id === props.options.bankId)
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
      params: [props.billId]
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
        params: [props.billId]
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
    day: 'numeric'
  })
}

const calculateTotal = () => {
  return carsData.value.reduce((total, car) => {
    let myPrice;
    if (props.options.paymentTerms.toLowerCase() === 'cfr') {
      myPrice = (parseFloat(car.price_cell) || 0) + (parseFloat(car.freight) || 0);
    } else {
      myPrice = parseFloat(car.price_cell) || 0;
    }
    
    // Convert to selected currency
    if (props.options.currency.toLowerCase() === 'dza') {
      const rate = parseFloat(car.rate);
      if (!rate) {
        console.warn(`No rate found for car ID ${car.id}`);
        return total;
      }
      return total + (myPrice * rate);
    }
    return total + myPrice;
  }, 0).toFixed(2)
}

const calculateCarPrice = (car) => {
  let myPrice;
  if (props.options.paymentTerms.toLowerCase() === 'cfr') {
    myPrice = (parseFloat(car.price_cell) || 0) + (parseFloat(car.freight) || 0);
  } else {
    myPrice = parseFloat(car.price_cell) || 0;
  }
  
  // Convert to selected currency
  if (props.options.currency.toLowerCase() === 'dza') {
    const rate = parseFloat(car.rate);
    if (!rate) {
      console.warn(`No rate found for car ID ${car.id}`);
      return 'Rate not set';
    }
    return (myPrice * rate).toFixed(2);
  }
  return myPrice.toFixed(2);
}

const formatPrice = (price) => {
  if (props.options.currency.toLowerCase() === 'dza') {
    if (price === 'Rate not set') return price;
    return `DZD ${price}`;
  }
  return `USD ${price}`;
}

const numberToWords = (num) => {
  const ones = ['', 'One', 'Two', 'Three', 'Four', 'Five', 'Six', 'Seven', 'Eight', 'Nine'];
  const tens = ['', '', 'Twenty', 'Thirty', 'Forty', 'Fifty', 'Sixty', 'Seventy', 'Eighty', 'Ninety'];
  const teens = ['Ten', 'Eleven', 'Twelve', 'Thirteen', 'Fourteen', 'Fifteen', 'Sixteen', 'Seventeen', 'Eighteen', 'Nineteen'];
  
  const convertLessThanOneThousand = (n) => {
    if (n === 0) return '';
    
    if (n < 10) return ones[n];
    
    if (n < 20) return teens[n - 10];
    
    if (n < 100) {
      return tens[Math.floor(n / 10)] + (n % 10 ? ' ' + ones[n % 10] : '');
    }
    
    return ones[Math.floor(n / 100)] + ' Hundred' + (n % 100 ? ' and ' + convertLessThanOneThousand(n % 100) : '');
  };
  
  if (num === 0) return 'Zero';
  
  const parts = num.toString().split('.');
  let wholePart = parseInt(parts[0]);
  let result = '';
  
  if (wholePart === 0) {
    result = 'Zero';
  } else {
    // Billions
    if (wholePart >= 1000000000) {
      result += convertLessThanOneThousand(Math.floor(wholePart / 1000000000)) + ' Billion ';
      wholePart %= 1000000000;
    }
    
    // Millions
    if (wholePart >= 1000000) {
      result += convertLessThanOneThousand(Math.floor(wholePart / 1000000)) + ' Million ';
      wholePart %= 1000000;
    }
    
    // Thousands
    if (wholePart >= 1000) {
      result += convertLessThanOneThousand(Math.floor(wholePart / 1000)) + ' Thousand ';
      wholePart %= 1000;
    }
    
    // Hundreds and remaining
    result += convertLessThanOneThousand(wholePart);
  }
  
  // Handle decimal part
  if (parts.length > 1) {
    const decimalPart = parts[1].padEnd(2, '0').substring(0, 2);
    if (parseInt(decimalPart) > 0) {
      result += ' and ' + decimalPart + '/100';
    }
  }
  
  return result.trim();
};

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
      <!-- Header -->
      <div class="header">
        <div class="company-info">
          <img :src="company.logo" alt="Company Logo" class="company-logo">
          <div class="company-text">
            <h1>{{ selectedBank?.company_name || company.name }}</h1>
            <p>{{ selectedBank?.bank_address || company.address }}</p>
            <p>Tel: {{ company.phone }}</p>
            <p>Email: {{ company.email }}</p>
          </div>
        </div>
        <div class="document-info">
          <h2>{{ options.documentType === 'sell_contract' ? 'SELL CONTRACT' : 
               options.documentType === 'invoice' ? 'INVOICE' : 'PROFORMA INVOICE' }}</h2>
          <p>Ref: {{ billData.bill_ref }}</p>
          <p>Date: {{ formatDate(billData.date_sell) }}</p>
        </div>
      </div>

      <!-- Broker Info -->
      <div class="section">
        <h3>Buyer Information</h3>
        <p><strong>Name:</strong> {{ billData.broker_name }}</p>
        <p><strong>Address:</strong> {{ billData.broker_address }}</p>
        <p><strong>Phone:</strong> {{ billData.broker_phone }}</p>
      </div>

      <!-- Cars Table -->
      <div class="section">
        <h3>Vehicle Details</h3>
        <table class="cars-table">
          <thead>
            <tr>
              <th>No.</th>
              <th>Vehicle</th>
              <th>Color</th>
              <th>VIN</th>
              <th>Client</th>
              <th>Port</th>
              <th>Price ({{ props.options.currency.toUpperCase() }})</th>
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
              <td :colspan="6" class="text-right"><strong>Grand Total:</strong></td>
              <td><strong>{{ formatPrice(calculateTotal()) }}</strong></td>
            </tr>
          </tfoot>
        </table>
        <div class="total-in-words">
          Amount in words: <strong>{{ numberToWords(calculateTotal()) }} {{ props.options.currency.toUpperCase() }}</strong>
        </div>
      </div>

      <!-- Payment Details and Bank Information -->
      <div class="two-columns">
        <!-- Payment Details -->
        <div class="section">
          <h3>Payment Details</h3>
          <p><strong>Terms:</strong> {{ options.paymentTerms.toUpperCase() }}</p>
          <p><strong>Mode:</strong> {{ options.paymentMode }}</p>
          <p><strong>Currency:</strong> {{ options.currency.toUpperCase() }}</p>
        </div>

        <!-- Bank Details -->
        <div class="section" v-if="selectedBank">
          <h3>Bank Information</h3>
          <p><strong>Company Name:</strong> {{ selectedBank.company_name }}</p>
          <p><strong>Bank Name:</strong> {{ selectedBank.bank_name }}</p>
          <p><strong>Bank Address:</strong> {{ selectedBank.bank_address }}</p>
          <p><strong>Account Number:</strong> {{ selectedBank.bank_account }}</p>
          <p><strong>SWIFT Code:</strong> {{ selectedBank.swift_code }}</p>
          <p v-if="selectedBank.notes"><strong>Notes:</strong> {{ selectedBank.notes }}</p>
        </div>
      </div>

      <!-- Notes -->
      <div class="section">
        <h3>Notes</h3>
        <p>{{ billData.notes || 'No notes' }}</p>
      </div>

      <!-- Contract Terms - Only visible for sell contract -->
      <div class="section contract-terms" v-if="options.documentType === 'sell_contract'">
        <h3>Terms and Conditions</h3>
        <div class="terms-list">
          <!-- Payment Terms -->
          <div class="term-item">
            <p class="english">1. Payment must be made in full before vehicle delivery.</p>
            <p class="arabic">١. يجب دفع المبلغ كاملاً قبل تسليم السيارة.</p>
            <p class="chinese">1. 车辆交付前必须全额付款。</p>
          </div>

          <!-- Delivery Terms -->
          <div class="term-item">
            <p class="english">2. Delivery will be made within the agreed timeframe after receipt of full payment.</p>
            <p class="arabic">٢. سيتم التسليم خلال الإطار الزمني المتفق عليه بعد استلام الدفعة الكاملة.</p>
            <p class="chinese">2. 收到全额付款后，将在约定的时间范围内交付。</p>
          </div>

          <!-- Document Responsibility -->
          <div class="term-item">
            <p class="english">3. All vehicle documentation will be provided as per export requirements.</p>
            <p class="arabic">٣. سيتم توفير جميع وثائق السيارة وفقاً لمتطلبات التصدير.</p>
            <p class="chinese">3. 将按照出口要求提供所有车辆文件。</p>
          </div>

          <!-- Customs Clause -->
          <div class="term-item">
            <p class="english">4. In case of any customs issues related to modified export license usage, while we will assist in resolving the matter with customs authorities, we bear no legal responsibility for such complications.</p>
            <p class="arabic">٤. في حالة وجود أي مشاكل جمركية تتعلق باستخدام رخصة التصدير المعدلة، وبينما سنساعد في حل المسألة مع السلطات الجمركية، فإننا لا نتحمل أي مسؤولية قانونية عن هذه المضاعفات.</p>
            <p class="chinese">4. 如因改装出口许可证使用而产生任何海关问题，虽然我们会协助处理海关事务，但我们不承担任何法律责任。</p>
          </div>

          <!-- Inspection Terms -->
          <div class="term-item">
            <p class="english">5. The buyer acknowledges having inspected the vehicle and accepts its condition at the time of purchase.</p>
            <p class="arabic">٥. يقر المشتري بأنه قد فحص السيارة ويقبل حالتها وقت الشراء.</p>
            <p class="chinese">5. 买方确认已检查车辆并接受其购买时的状况。</p>
          </div>

          <!-- Dispute Resolution -->
          <div class="term-item">
            <p class="english">6. Any disputes arising from this contract shall be resolved through mutual negotiation or applicable legal channels.</p>
            <p class="arabic">٦. يتم حل أي نزاعات تنشأ عن هذا العقد من خلال التفاوض المتبادل أو القنوات القانونية المعمول بها.</p>
            <p class="chinese">6. 本合同引起的任何争议应通过双方协商或适用的法律途径解决。</p>
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
/* A4 page size in pixels (assuming 96 DPI) */
.a4-page {
  width: 210mm;
  min-height: 297mm;
  padding: 20mm;
  margin: 0 auto;
  background: white;
  box-shadow: 0 0 10px rgba(0,0,0,0.1);
}

@media print {
  .a4-page {
    box-shadow: none;
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

.document-info h2 {
  font-size: 20px;
  color: #1f2937;
  margin-bottom: 10px;
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

.cars-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 20px;
}

.cars-table th,
.cars-table td {
  border: 1px solid #e5e7eb;
  padding: 8px;
  text-align: left;
}

.cars-table th {
  background-color: #f9fafb;
  font-weight: 600;
}

.cars-table tfoot td {
  background-color: #f9fafb;
}

.text-right {
  text-align: right;
}

.signatures {
  display: flex;
  justify-content: space-between;
  margin-top: 50px;
}

.signature-box {
  width: 200px;
  text-align: center;
}

.signature-line {
  margin-top: 50px;
  border-top: 1px solid #000;
}

.total-in-words {
  margin-top: 10px;
  padding: 10px;
  font-size: 14px;
  border-top: 1px solid #e5e7eb;
  text-align: right;
  font-style: italic;
}

@media print {
  body {
    background: none;
  }
  
  .print-document {
    padding: 0;
    margin: 0;
  }
  
  .loading,
  .error {
    display: none;
  }
}

/* New styles for bank selection and details */
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

.bank-details {
  margin-top: 20px;
  padding: 15px;
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
}

.bank-details h3 {
  color: #1f2937;
  margin-bottom: 10px;
}

.bank-details p {
  margin: 5px 0;
}

@media print {
  .print-hide {
    display: none;
  }
}

.two-columns {
  display: flex;
  gap: 20px;
  margin-bottom: 30px;
}

.two-columns .section {
  flex: 1;
  min-width: 0; /* Prevent flex items from overflowing */
}

@media print {
  .two-columns {
    break-inside: avoid;
  }
}

.contract-terms {
  margin: 30px 0;
  page-break-inside: avoid;
}

.terms-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.term-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 10px;
  background-color: #f9fafb;
  border-radius: 4px;
}

.term-item p {
  margin: 0;
  line-height: 1.5;
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
  .term-item {
    background-color: transparent;
    border: 1px solid #e5e7eb;
  }
}
</style> 