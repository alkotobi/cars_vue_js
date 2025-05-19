<script setup>
import { ref, onMounted, watch } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi } = useApi()
const sellBills = ref([])
const carsStock = ref([])
const loading = ref(true)
const stockLoading = ref(false)
const error = ref(null)
const selectedBillId = ref(null)
const showAddDialog = ref(false)

// Add new refs for the add dialog
const newSellBill = ref({
  id_broker: null,
  date_sell: new Date().toISOString().split('T')[0], // Today's date in YYYY-MM-DD format
  notes: ''
})

// Add ref for brokers list
const brokers = ref([])

const fetchSellBills = async () => {
  loading.value = true
  error.value = null
  
  try {
    const result = await callApi({
      query: `
        SELECT 
          sb.id,
          sb.date_sell,
          sb.notes,
          c.name as broker_name,
          c.mobiles as broker_mobile
        FROM sell_bill sb
        LEFT JOIN clients c ON sb.id_broker = c.id AND c.is_broker = 1
        ORDER BY sb.date_sell DESC
      `,
      params: []
    })
    
    if (result.success) {
      sellBills.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch sell bills'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const fetchBrokers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name, mobiles
        FROM clients
        WHERE is_broker = 1
        ORDER BY name ASC
      `,
      params: []
    })
    
    if (result.success) {
      brokers.value = result.data
    }
  } catch (err) {
    console.error('Error fetching brokers:', err)
  }
}

const selectBill = async (bill) => {
  selectedBillId.value = bill.id
  await fetchCarStock(bill.id)
}

const fetchCarStock = async (sellBillId) => {
  stockLoading.value = true
  
  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.price_cell,
          cs.date_loding,
          cs.date_sell,
          cs.notes,
          cs.freight,
          cs.path_documents,
          cs.sell_pi_path,
          cs.buy_pi_path,
          c.name as client_name,
          cn.car_name,
          clr.color,
          lp.loading_port,
          dp.discharge_port,
          bd.price_sell as buy_price
        FROM cars_stock cs
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        WHERE cs.id_sell = ? AND cs.hidden = 0
        ORDER BY cs.id DESC
      `,
      params: [sellBillId]
    })
    
    if (result.success) {
      carsStock.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch cars stock'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    stockLoading.value = false
  }
}

const addSellBill = async () => {
  try {
    const result = await callApi({
      query: `
        INSERT INTO sell_bill (id_broker, date_sell, notes)
        VALUES (?, ?, ?)
      `,
      params: [
        newSellBill.value.id_broker,
        newSellBill.value.date_sell,
        newSellBill.value.notes
      ]
    })
    
    if (result.success) {
      showAddDialog.value = false
      // Reset form
      newSellBill.value = {
        id_broker: null,
        date_sell: new Date().toISOString().split('T')[0],
        notes: ''
      }
      // Refresh the list
      await fetchSellBills()
    } else {
      error.value = result.error || 'Failed to add sell bill'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

onMounted(async () => {
  await fetchSellBills()
  await fetchBrokers()
})
</script>

<template>
  <div class="sell-view">
    <div class="header">
      <h2>Sell Bills Management</h2>
      <button class="add-btn" @click="showAddDialog = true">
        Add New Sell Bill
      </button>
    </div>
    
    <div v-if="selectedBillId" class="selected-bill">
      Selected Bill ID: <span class="bill-id">{{ selectedBillId }}</span>
    </div>
    
    <div class="content">
      <div v-if="loading" class="loading">Loading...</div>
      
      <div v-else-if="error" class="error">{{ error }}</div>
      
      <div v-else class="tables-container">
        <!-- Sell Bills Table -->
        <div class="sell-bills-table">
          <h3>Sell Bills</h3>
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Date</th>
                <th>Broker</th>
                <th>Notes</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr 
                v-for="bill in sellBills" 
                :key="bill.id" 
                @click="selectBill(bill)"
                :class="{ 'selected': selectedBillId === bill.id }"
              >
                <td>{{ bill.id }}</td>
                <td>{{ new Date(bill.date_sell).toLocaleDateString() }}</td>
                <td>{{ bill.broker_name || 'N/A' }}</td>
                <td>{{ bill.notes || 'N/A' }}</td>
                <td>
                  <button class="view-btn" @click.stop="selectBill(bill)">View Cars</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        
        <!-- Car Stock Table -->
        <div class="car-stock-table">
          <h3>Cars in Selected Bill</h3>
          <div v-if="!selectedBillId" class="no-selection">
            Please select a sell bill to view associated cars
          </div>
          <div v-else-if="stockLoading" class="loading">Loading cars...</div>
          <div v-else-if="carsStock.length === 0" class="no-data">
            No cars found for this sell bill
          </div>
          <table v-else>
            <thead>
              <tr>
                <th>ID</th>
                <th>Car</th>
                <th>Color</th>
                <th>VIN</th>
                <th>Sell Price</th>
                <th>Buy Price</th>
                <th>Profit</th>
                <th>Loading Port</th>
                <th>Discharge Port</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="car in carsStock" :key="car.id">
                <td>{{ car.id }}</td>
                <td>{{ car.car_name || 'N/A' }}</td>
                <td>{{ car.color || 'N/A' }}</td>
                <td>{{ car.vin || 'N/A' }}</td>
                <td>${{ car.price_cell || 0 }}</td>
                <td>${{ car.buy_price || 0 }}</td>
                <td>${{ (car.price_cell - car.buy_price) || 0 }}</td>
                <td>{{ car.loading_port || 'N/A' }}</td>
                <td>{{ car.discharge_port || 'N/A' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
    
    <!-- Add Sell Bill Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Sell Bill</h3>
        
        <div class="form-group">
          <label for="broker">Broker:</label>
          <select id="broker" v-model="newSellBill.id_broker">
            <option value="">Select Broker</option>
            <option v-for="broker in brokers" :key="broker.id" :value="broker.id">
              {{ broker.name }}
            </option>
          </select>
        </div>
        
        <div class="form-group">
          <label for="date">Date:</label>
          <input type="date" id="date" v-model="newSellBill.date_sell">
        </div>
        
        <div class="form-group">
          <label for="notes">Notes:</label>
          <textarea id="notes" v-model="newSellBill.notes"></textarea>
        </div>
        
        <div class="dialog-buttons">
          <button @click="showAddDialog = false">Cancel</button>
          <button @click="addSellBill" class="primary">Add</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sell-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.add-btn {
  background-color: #4CAF50;
  color: white;
  border: none;
  padding: 10px 15px;
  border-radius: 4px;
  cursor: pointer;
}

.selected-bill {
  margin-bottom: 15px;
  padding: 10px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.bill-id {
  font-weight: bold;
  color: #007bff;
}

.tables-container {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}

th, td {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}

th {
  background-color: #f2f2f2;
}

tr:hover {
  background-color: #f5f5f5;
}

tr.selected {
  background-color: #e0f7fa;
}

.view-btn {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 5px 10px;
  border-radius: 4px;
  cursor: pointer;
}

.no-selection, .no-data {
  padding: 20px;
  text-align: center;
  color: #6c757d;
  background-color: #f8f9fa;
  border-radius: 4px;
}

.loading, .error {
  padding: 20px;
  text-align: center;
}

.error {
  color: #dc3545;
}

.dialog-overlay {
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
}

.dialog {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  width: 500px;
  max-width: 90%;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.dialog-buttons button {
  padding: 8px 15px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.dialog-buttons button.primary {
  background-color: #4CAF50;
  color: white;
}
</style>