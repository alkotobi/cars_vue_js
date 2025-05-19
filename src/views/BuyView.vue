<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi } from '../composables/useApi'
import BuyBillsTable from '../components/BuyBillsTable.vue'
import BuyDetailsTable from '../components/BuyDetailsTable.vue'

// API and base setup
const { callApi, error } = useApi()

// State management
const buyBills = ref([])
const selectedBill = ref(null)
const buyDetails = ref([])
const suppliers = ref([])
const cars = ref([])
const colors = ref([])

// Dialog controls
const showAddDialog = ref(false)
const showAddDetailDialog = ref(false)

// Form models
const newPurchase = ref({
  id_supplier: null,
  date_buy: new Date().toISOString().slice(0, 16),
  amount: 0,
  payed: 0,
  pi_path: ''
})

const newDetail = ref({
  id_car_name: null,
  id_color: null,
  amount: 0,
  notes: '',
  QTY: 1,
  year: new Date().getFullYear(),
  month: new Date().getMonth() + 1,
  is_used_car: false,
  id_buy_bill: null
})

// Data fetching functions
const fetchSuppliers = async () => {
  const result = await callApi({
    query: 'SELECT * FROM suppliers ORDER BY name ASC',
    params: []
  })
  if (result.success) {
    suppliers.value = result.data
  }
}

const addPurchase = async () => {
  const result = await callApi({
    query: `
      INSERT INTO buy_bill (id_supplier, date_buy, amount, payed, pi_path)
      VALUES (?, ?, ?, ?, ?)
    `,
    params: [
      newPurchase.value.id_supplier,
      newPurchase.value.date_buy,
      newPurchase.value.amount,
      newPurchase.value.payed,
      newPurchase.value.pi_path
    ]
  })
  
  if (result.success) {
    showAddDialog.value = false
    await fetchBuyBills()
    // Reset form
    newPurchase.value = {
      id_supplier: null,
      date_buy: new Date().toISOString().slice(0, 16),
      amount: 0,
      payed: 0,
      pi_path: ''
    }
  }
  else {
    console.error('Error adding purchase:', result.error);
  }
}

const fetchBuyBills = async () => {
  const result = await callApi({
    query: `
      SELECT bb.*, s.name as supplier_name 
      FROM buy_bill bb 
      LEFT JOIN suppliers s ON bb.id_supplier = s.id 
      ORDER BY bb.date_buy DESC
    `,
    params: []
  })
  if (result.success) {
    buyBills.value = result.data
  }
}

const fetchBuyDetails = async (billId) => {
  const result = await callApi({
    query: `
      SELECT bd.*, cn.car_name, c.color, bb.is_stock_updated
      FROM buy_details bd
      LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
      LEFT JOIN colors c ON bd.id_color = c.id
      LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
      WHERE bd.id_buy_bill = ?
    `,
    params: [billId]
  })
  if (result.success) {
    buyDetails.value = result.data
  }
  else{
    console.error('Error fetching details:', result.error);
  }
}

const selectBill = (bill) => {
  selectedBill.value = bill
  fetchBuyDetails(bill.id)
}





const fetchCars = async () => {
  const result = await callApi({
    query: 'SELECT * FROM cars_names ORDER BY car_name ASC',
    params: []
  })
  if (result.success) {
    cars.value = result.data
  }
}

const fetchColors = async () => {
  const result = await callApi({
    query: 'SELECT * FROM colors ORDER BY color ASC',
    params: []
  })
  if (result.success) {
    colors.value = result.data
  }
}

const addDetail = async () => {
  const result = await callApi({
    query: `
      INSERT INTO buy_details 
      (id_car_name, id_color, amount, notes, QTY, year, month, is_used_car, id_buy_bill, price_sell)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
    params: [
      newDetail.value.id_car_name,
      newDetail.value.id_color,
      newDetail.value.amount,
      newDetail.value.notes,
      newDetail.value.QTY,
      newDetail.value.year,
      newDetail.value.month,
      newDetail.value.is_used_car ? 1 : 0,
      selectedBill.value.id,
      newDetail.value.price_sell  // Add this
    ]
  })
  
  if (result.success) {
    showAddDetailDialog.value = false
    await fetchBuyDetails(selectedBill.value.id)
    // Reset form
    newDetail.value = {
      id_car_name: null,
      id_color: null,
      amount: 0,
      notes: '',
      QTY: 1,
      year: new Date().getFullYear(),
      month: new Date().getMonth() + 1,
      is_used_car: false,
      id_buy_bill: null
    }
  }
  else {
    console.error('Error adding detail:', result.error);
  }
}

// Lifecycle hooks
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchBuyBills()
  fetchSuppliers()
  fetchCars()
  fetchColors()
})

const handleDeleteDetail = async (detailId) => {
  const result = await callApi({
    query: 'DELETE FROM buy_details WHERE id = ?',
    params: [detailId]
  })
  
  if (result.success) {
    // Refresh the details list after successful deletion
    await fetchBuyDetails(selectedBill.value.id)
  } else {
    alert('Failed to delete detail. Please try again.')
    console.error('Error deleting detail:', result.error)
  }
}

const handleUpdateDetail = async (updatedDetail) => {
  const result = await callApi({
    query: `
      UPDATE buy_details 
      SET QTY = ?, amount = ?, year = ?, month = ?, notes = ?, price_sell = ?
      WHERE id = ?
    `,
    params: [
      updatedDetail.QTY,
      updatedDetail.amount,
      updatedDetail.year,
      updatedDetail.month,
      updatedDetail.notes,
      updatedDetail.price_sell,  // Add this
      updatedDetail.id
    ]
  })
  
  if (result.success) {
    // Refresh the details list after successful update
    await fetchBuyDetails(selectedBill.value.id)
  } else {
    alert('Failed to update detail. Please try again.')
    console.error('Error updating detail:', result.error)
  }
}
</script>

<template>
  <div class="buy-view">
    <div class="header">
      <h2>Buy Management</h2>
      <button @click="showAddDialog = true" class="add-btn">Add New Purchase</button>
    </div>

    <div class="content">
      <div class="master-detail-vertical">
        <BuyBillsTable 
          :buyBills="buyBills"
          :selectedBill="selectedBill"
          @select-bill="selectBill"
        />

        <BuyDetailsTable 
          v-if="selectedBill"
          :buyDetails="buyDetails"
          :isAdmin="isAdmin"
          @add-detail="showAddDetailDialog = true"
          @delete-detail="handleDeleteDetail"
          @update-detail="handleUpdateDetail"
        />
      </div>
    </div>

    <!-- Dialogs -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Purchase</h3>
        <form @submit.prevent="addPurchase">
          <div class="form-group">
            <label>Supplier</label>
            <select v-model="newPurchase.id_supplier" required>
              <option value="">Select Supplier</option>
              <option v-for="supplier in suppliers" 
                      :key="supplier.id" 
                      :value="supplier.id">
                {{ supplier.name }}
              </option>
            </select>
          </div>
          
          <div class="form-group">
            <label>Date</label>
            <input type="datetime-local" 
                   v-model="newPurchase.date_buy" 
                   required>
          </div>
          
          <div class="form-group">
            <label>Amount</label>
            <input type="number" 
                   v-model="newPurchase.amount" 
                   step="0.01" 
                   required>
          </div>
          
          <div class="form-group">
            <label>Paid</label>
            <input type="number" 
                   v-model="newPurchase.payed" 
                   step="0.01" 
                   required>
          </div>
          
          <div class="form-group">
            <label>PI Path</label>
            <input type="text" 
                   v-model="newPurchase.pi_path">
          </div>
          
          <div class="dialog-buttons">
            <button type="button" 
                    @click="showAddDialog = false" 
                    class="cancel-btn">
              Cancel
            </button>
            <button type="submit" class="submit-btn">
              Add Purchase
            </button>
          </div>
        </form>
      </div>
    </div>
    <div v-if="showAddDetailDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add Purchase Detail</h3>
        <form @submit.prevent="addDetail">
          <div class="form-group">
            <label>Car</label>
            <select v-model="newDetail.id_car_name" required>
              <option value="">Select Car</option>
              <option v-for="car in cars" 
                      :key="car.id" 
                      :value="car.id">
                {{ car.car_name }}
              </option>
            </select>
          </div>
          
          <div class="form-group">
            <label>Color</label>
            <select v-model="newDetail.id_color" required>
              <option value="">Select Color</option>
              <option v-for="color in colors" 
                      :key="color.id" 
                      :value="color.id">
                {{ color.color }}
              </option>
            </select>
          </div>
          
          <div class="form-group">
            <label>Quantity</label>
            <input type="number" 
                   v-model="newDetail.QTY" 
                   min="1" 
                   required>
          </div>
          
          <div class="form-group">
            <label>Amount</label>
            <input type="number" 
                   v-model="newDetail.amount" 
                   step="0.01" 
                   required>
          </div>
          
          <div class="form-row">
            <div class="form-group half">
              <label>Year</label>
              <input type="number" 
                     v-model="newDetail.year" 
                     required>
            </div>
            
            <div class="form-group half">
              <label>Month</label>
              <input type="number" 
                     v-model="newDetail.month" 
                     min="1" 
                     max="12" 
                     required>
            </div>
          </div>
          
          <div class="form-group">
            <label>Price Sell</label>
            <input type="number" 
                   v-model="newDetail.price_sell"
                   step="0.01"
                   required>
          </div>
          
          <div class="form-group">
            <label>Notes</label>
            <textarea v-model="newDetail.notes" rows="3"></textarea>
          </div>
          
          <div class="form-group checkbox">
            <label>
              <input type="checkbox" v-model="newDetail.is_used_car">
              Used Car
            </label>
          </div>
          
          <div class="dialog-buttons">
            <button type="button" 
                    @click="showAddDetailDialog = false" 
                    class="cancel-btn">
              Cancel
            </button>
            <button type="submit" class="submit-btn">
              Add Detail
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.buy-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
}

.master-detail-vertical {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.master-section {
  width: 100%;
}

.detail-section {
  width: 100%;
  border-top: 1px solid #e5e7eb;
  padding-top: 20px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.data-table tbody tr {
  cursor: pointer;
  transition: background-color 0.2s;
}

.data-table tbody tr:hover {
  background-color: #f3f4f6;
}

.data-table tbody tr.selected {
  background-color: #e5e7eb;
}

.empty-detail {
  text-align: center;
  color: #6b7280;
  padding: 40px;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.add-btn:hover {
  background-color: #059669;
}

h3 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #374151;
}

.selected-id {
  background-color: #f0fdf4;
  color: #047857;
  padding: 8px 16px;
  border-radius: 4px;
  margin-bottom: 16px;
  border: 1px solid #10b981;
  font-weight: 500;
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
  padding: 24px;
  border-radius: 8px;
  width: 100%;
  max-width: 500px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #10b981;
  ring: 2px #10b981;
}

.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn {
  padding: 8px 16px;
  background-color: #e5e7eb;
  color: #374151;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.cancel-btn:hover {
  background-color: #d1d5db;
}

.submit-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.submit-btn:hover {
  background-color: #059669;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-group.half {
  flex: 1;
}

.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  resize: vertical;
}

.form-group.checkbox {
  display: flex;
  align-items: center;
}

.form-group.checkbox input {
  width: auto;
  margin-right: 8px;
}

.form-group.checkbox label {
  display: flex;
  align-items: center;
  margin: 0;
}

</style>