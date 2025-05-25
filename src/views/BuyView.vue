<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi } from '../composables/useApi'
import BuyBillsTable from '../components/buy/BuyBillsTable.vue'
import BuyDetailsTable from '../components/buy/BuyDetailsTable.vue'
import { useRouter } from 'vue-router'

// API and base setup
const { callApi, uploadFile, getFileUrl, error } = useApi()
const router = useRouter()

// State management
const buyBills = ref([])
const selectedBill = ref(null)
const buyDetails = ref([])
const suppliers = ref([])
const cars = ref([])
const colors = ref([])

// Dialog controls
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const showAddDetailDialog = ref(false)
const editingBill = ref(null)

// Loading states
const isSubmittingPurchase = ref(false)
const isSubmittingDetail = ref(false)

// User info
const user = ref(JSON.parse(localStorage.getItem('user')))
const isAdmin = computed(() => user.value?.role_id === 1)

// Toolbar action methods
const canUpdateStock = (bill) => {
  return bill.amount > 0 && !bill.is_stock_updated
}

const handleUpdateStock = async (bill) => {
  if (!confirm('Are you sure you want to update the stock? This action cannot be undone.')) {
    return
  }

  try {
    const result = await callApi({
      query: `
        UPDATE buy_bill 
        SET is_stock_updated = 1 
        WHERE id = ?
      `,
      params: [bill.id],
    })

    if (result.success) {
      await fetchBuyBills()
    } else {
      alert('Failed to update stock')
    }
  } catch (err) {
    console.error('Error updating stock:', err)
    alert('Failed to update stock')
  }
}

const handleDeleteBill = async (bill) => {
  if (
    !confirm(
      'Are you sure you want to delete this purchase and all its details? This action cannot be undone.',
    )
  ) {
    return
  }

  try {
    // First check if the bill can be deleted
    if (bill.is_stock_updated) {
      alert('Cannot delete a purchase that has already been updated in stock')
      return
    }

    // Start with deleting related records
    // 1. Delete payments
    const deletePayments = await callApi({
      query: `DELETE FROM buy_payments WHERE id_buy_bill = ?`,
      params: [bill.id],
    })

    if (!deletePayments.success) {
      throw new Error('Failed to delete related payments')
    }

    // 2. Delete details
    const deleteDetails = await callApi({
      query: `DELETE FROM buy_details WHERE id_buy_bill = ?`,
      params: [bill.id],
    })

    if (!deleteDetails.success) {
      throw new Error('Failed to delete purchase details')
    }

    // 3. Finally delete the bill
    const deleteBill = await callApi({
      query: `DELETE FROM buy_bill WHERE id = ? AND is_stock_updated = 0`,
      params: [bill.id],
    })

    if (deleteBill.success) {
      if (selectedBill.value?.id === bill.id) {
        selectedBill.value = null
        buyDetails.value = [] // Clear details
      }
      await fetchBuyBills()
    } else {
      throw new Error('Failed to delete purchase')
    }
  } catch (err) {
    console.error('Error deleting purchase:', err)
    alert(err.message || 'Failed to delete purchase')
  }
}

const openEditDialog = (bill) => {
  editingBill.value = bill
  newPurchase.value = {
    id_supplier: bill.id_supplier,
    date_buy: bill.date_buy.slice(0, 16), // Format for datetime-local input
    bill_ref: bill.bill_ref || '',
    pi_path: bill.pi_path || '',
    pi_file: null,
  }
  showEditDialog.value = true
}

const openAddDialog = () => {
  editingBill.value = null
  newPurchase.value = {
    id_supplier: null,
    date_buy: new Date().toISOString().slice(0, 16),
    bill_ref: '',
    pi_path: '',
    pi_file: null,
  }
  showAddDialog.value = true
}

// Form models
const newPurchase = ref({
  id_supplier: null,
  date_buy: new Date().toISOString().slice(0, 16),
  amount: 0,
  bill_ref: '',
  pi_path: '',
  pi_file: null,
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
  id_buy_bill: null,
})

// Handle PI file change
const handlePiFileChange = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // Check if file is PDF or image
  const allowedTypes = ['application/pdf', 'image/jpeg', 'image/png', 'image/gif', 'image/webp']

  if (!allowedTypes.includes(file.type)) {
    alert('Only PDF and image files (JPEG, PNG, GIF, WEBP) are allowed')
    event.target.value = ''
    return
  }

  // Store the file for upload
  newPurchase.value.pi_file = file
}

// Data fetching functions
const fetchSuppliers = async () => {
  const result = await callApi({
    query: 'SELECT * FROM suppliers ORDER BY name ASC',
    params: [],
  })
  if (result.success) {
    suppliers.value = result.data
  }
}

const addPurchase = async () => {
  // Prevent multiple submissions
  if (isSubmittingPurchase.value) {
    return
  }

  try {
    isSubmittingPurchase.value = true

    // Validate that PI document is provided
    if (!newPurchase.value.pi_file) {
      alert('PI document is required. Please select a file.')
      return
    }

    // Handle PI document upload
    try {
      // Create filename using purchase details and original extension
      const date = new Date()
      const timestamp = date.getTime()
      const dateStr = date.toISOString().split('T')[0]
      const timeStr = date.toISOString().split('T')[1].split('.')[0].replace(/:/g, '-')
      const supplierName =
        suppliers.value.find((s) => s.id === newPurchase.value.id_supplier)?.name || 'unknown'
      const fileExt = newPurchase.value.pi_file.name.split('.').pop().toLowerCase()

      // Format: pi_purchase_supplier_date_time_timestamp.ext
      const filename = `pi_purchase_${supplierName.replace(/\s+/g, '_')}_${dateStr}_${timeStr}_${timestamp}.${fileExt}`

      const uploadResult = await uploadFile(newPurchase.value.pi_file, 'purchase_pi', filename)

      if (!uploadResult.success) {
        throw new Error(uploadResult.message || 'Failed to upload PI document')
      }

      // Store just the relative path
      newPurchase.value.pi_path = `purchase_pi/${filename}`
    } catch (uploadError) {
      console.error('Upload error:', uploadError)
      throw new Error(`Failed to upload PI document: ${uploadError.message}`)
    }

    const result = await callApi({
      query: `
        INSERT INTO buy_bill (id_supplier, date_buy, bill_ref, pi_path)
        VALUES (?, ?, ?, ?)
      `,
      params: [
        newPurchase.value.id_supplier,
        newPurchase.value.date_buy,
        newPurchase.value.bill_ref,
        newPurchase.value.pi_path,
      ],
    })

    if (result.success) {
      showAddDialog.value = false
      await fetchBuyBills()
      // Reset form
      newPurchase.value = {
        id_supplier: null,
        date_buy: new Date().toISOString().slice(0, 16),
        bill_ref: '',
        pi_path: '',
        pi_file: null,
      }
    } else {
      throw new Error(result.error || 'Failed to add purchase')
    }
  } catch (err) {
    console.error('Error adding purchase:', err)
    alert(err.message)
  } finally {
    isSubmittingPurchase.value = false
  }
}

const updatePurchase = async () => {
  // Prevent multiple submissions
  if (isSubmittingPurchase.value) {
    return
  }

  try {
    isSubmittingPurchase.value = true

    // Validate that PI document exists (either existing or new file)
    if (!newPurchase.value.pi_file && !newPurchase.value.pi_path) {
      alert('PI document is required. Please select a file.')
      return
    }

    // Handle PI document upload if there's a new file
    if (newPurchase.value.pi_file) {
      try {
        // Create filename using purchase details and original extension
        const date = new Date()
        const timestamp = date.getTime()
        const dateStr = date.toISOString().split('T')[0]
        const timeStr = date.toISOString().split('T')[1].split('.')[0].replace(/:/g, '-')
        const supplierName =
          suppliers.value.find((s) => s.id === newPurchase.value.id_supplier)?.name || 'unknown'
        const fileExt = newPurchase.value.pi_file.name.split('.').pop().toLowerCase()

        // Format: pi_purchase_supplier_date_time_timestamp.ext
        const filename = `pi_purchase_${supplierName.replace(/\s+/g, '_')}_${dateStr}_${timeStr}_${timestamp}.${fileExt}`

        const uploadResult = await uploadFile(newPurchase.value.pi_file, 'purchase_pi', filename)

        if (!uploadResult.success) {
          throw new Error(uploadResult.message || 'Failed to upload PI document')
        }

        // Store just the relative path
        newPurchase.value.pi_path = `purchase_pi/${filename}`
      } catch (uploadError) {
        console.error('Upload error:', uploadError)
        throw new Error(`Failed to upload PI document: ${uploadError.message}`)
      }
    }

    const result = await callApi({
      query: `
        UPDATE buy_bill 
        SET id_supplier = ?, date_buy = ?, bill_ref = ?, pi_path = ?
        WHERE id = ?
      `,
      params: [
        newPurchase.value.id_supplier,
        newPurchase.value.date_buy,
        newPurchase.value.bill_ref,
        newPurchase.value.pi_path,
        editingBill.value.id,
      ],
    })

    if (result.success) {
      showEditDialog.value = false
      editingBill.value = null
      await fetchBuyBills()
      // Reset form
      newPurchase.value = {
        id_supplier: null,
        date_buy: new Date().toISOString().slice(0, 16),
        bill_ref: '',
        pi_path: '',
        pi_file: null,
      }
    } else {
      throw new Error(result.error || 'Failed to update purchase')
    }
  } catch (err) {
    console.error('Error updating purchase:', err)
    alert(err.message)
  } finally {
    isSubmittingPurchase.value = false
  }
}

const fetchBuyBills = async () => {
  const result = await callApi({
    query: `
      SELECT 
        bb.*,
        s.name as supplier_name,
        COALESCE((
          SELECT SUM(amount * QTY)
          FROM buy_details
          WHERE id_buy_bill = bb.id
        ), 0) as calculated_amount
      FROM buy_bill bb 
      LEFT JOIN suppliers s ON bb.id_supplier = s.id 
      ORDER BY bb.date_buy DESC
    `,
    params: [],
  })
  if (result.success) {
    buyBills.value = result.data.map((bill) => ({
      ...bill,
      amount: Number(bill.calculated_amount),
      payed: Number(bill.payed),
    }))
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
    params: [billId],
  })
  if (result.success) {
    buyDetails.value = result.data
  } else {
    console.error('Error fetching details:', result.error)
  }
}

const selectBill = (bill) => {
  selectedBill.value = bill
  fetchBuyDetails(bill.id)
}

const fetchCars = async () => {
  const result = await callApi({
    query: 'SELECT * FROM cars_names ORDER BY car_name ASC',
    params: [],
  })
  if (result.success) {
    cars.value = result.data
  }
}

const fetchColors = async () => {
  const result = await callApi({
    query: 'SELECT * FROM colors ORDER BY color ASC',
    params: [],
  })
  if (result.success) {
    colors.value = result.data
  }
}

const updateBillAmount = async (billId) => {
  const result = await callApi({
    query: `
      UPDATE buy_bill 
      SET amount = (
        SELECT COALESCE(SUM(amount * QTY), 0)
        FROM buy_details
        WHERE id_buy_bill = ?
      )
      WHERE id = ?
    `,
    params: [billId, billId],
  })

  if (!result.success) {
    console.error('Error updating bill amount:', result.error)
  }
}

const addDetail = async () => {
  // Prevent multiple submissions
  if (isSubmittingDetail.value) {
    return
  }

  try {
    isSubmittingDetail.value = true

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
        newDetail.value.price_sell,
      ],
    })

    if (result.success) {
      await updateBillAmount(selectedBill.value.id) // Update bill amount
      showAddDetailDialog.value = false
      await fetchBuyDetails(selectedBill.value.id)
      await fetchBuyBills() // Refresh bills to show updated amount
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
        id_buy_bill: null,
      }
    } else {
      console.error('Error adding detail:', result.error)
    }
  } catch (err) {
    console.error('Error adding detail:', err)
    alert('Failed to add detail. Please try again.')
  } finally {
    isSubmittingDetail.value = false
  }
}

// Lifecycle hooks
onMounted(() => {
  fetchBuyBills()
  fetchSuppliers()
  fetchCars()
  fetchColors()
})

const handleDeleteDetail = async (detailId) => {
  const result = await callApi({
    query: 'DELETE FROM buy_details WHERE id = ?',
    params: [detailId],
  })

  if (result.success) {
    await updateBillAmount(selectedBill.value.id) // Update bill amount
    await fetchBuyDetails(selectedBill.value.id)
    await fetchBuyBills() // Refresh bills to show updated amount
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
      updatedDetail.price_sell,
      updatedDetail.id,
    ],
  })

  if (result.success) {
    await updateBillAmount(selectedBill.value.id) // Update bill amount
    await fetchBuyDetails(selectedBill.value.id)
    await fetchBuyBills() // Refresh bills to show updated amount
  } else {
    alert('Failed to update detail. Please try again.')
    console.error('Error updating detail:', result.error)
  }
}

// Replace handlePaymentClick with new method to open in new tab
const openPayments = (bill) => {
  window.open(`/buy-payments/${bill.id}`, '_blank')
}
</script>

<template>
  <div class="buy-view">
    <div class="header">
      <h2>Buy Management</h2>
      <button @click="openAddDialog" class="add-btn">Add New Purchase</button>
    </div>

    <div class="content">
      <div class="master-detail-vertical">
        <BuyBillsTable :buyBills="buyBills" :selectedBill="selectedBill" @select-bill="selectBill">
          <!-- Toolbar actions -->
          <template #actions="{ bill }">
            <button @click.stop="openPayments(bill)" class="payment-btn">Payments</button>
            <button
              @click.stop="openEditDialog(bill)"
              class="action-btn edit-btn"
              :disabled="bill.is_stock_updated"
            >
              Edit
            </button>
            <button
              v-if="false"
              @click.stop="handleUpdateStock(bill)"
              class="action-btn update-btn"
              :disabled="!canUpdateStock(bill)"
            >
              Update Stock
            </button>
            <button
              v-if="isAdmin"
              @click.stop="handleDeleteBill(bill)"
              class="action-btn delete-btn"
              :disabled="bill.is_stock_updated"
            >
              Delete
            </button>
          </template>
        </BuyBillsTable>

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
              <option v-for="supplier in suppliers" :key="supplier.id" :value="supplier.id">
                {{ supplier.name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>Date</label>
            <input type="datetime-local" v-model="newPurchase.date_buy" required />
          </div>

          <div class="form-group">
            <label>Bill Reference <span class="required">*</span></label>
            <input
              type="text"
              v-model="newPurchase.bill_ref"
              placeholder="Enter bill reference number"
              required
            />
          </div>

          <div class="form-group">
            <label>PI Document (PDF or Image) <span class="required">*</span></label>
            <input
              type="file"
              @change="handlePiFileChange"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp"
              required
            />
            <a
              v-if="newPurchase.pi_path"
              :href="getFileUrl(newPurchase.pi_path)"
              target="_blank"
              class="current-file-link"
            >
              View Current PI Document
            </a>
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="showAddDialog = false" class="cancel-btn">Cancel</button>
            <button type="submit" class="submit-btn" :disabled="isSubmittingPurchase">
              <span v-if="isSubmittingPurchase" class="spinner"></span>
              {{ isSubmittingPurchase ? 'Adding...' : 'Add Purchase' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Edit Purchase Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Purchase</h3>
        <form @submit.prevent="updatePurchase">
          <div class="form-group">
            <label>Supplier</label>
            <select v-model="newPurchase.id_supplier" required>
              <option value="">Select Supplier</option>
              <option v-for="supplier in suppliers" :key="supplier.id" :value="supplier.id">
                {{ supplier.name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>Date</label>
            <input type="datetime-local" v-model="newPurchase.date_buy" required />
          </div>

          <div class="form-group">
            <label>Bill Reference <span class="required">*</span></label>
            <input
              type="text"
              v-model="newPurchase.bill_ref"
              placeholder="Enter bill reference number"
              required
            />
          </div>

          <div class="form-group">
            <label>PI Document (PDF or Image) <span class="required">*</span></label>
            <input
              type="file"
              @change="handlePiFileChange"
              accept=".pdf,.jpg,.jpeg,.png,.gif,.webp"
            />
            <a
              v-if="newPurchase.pi_path"
              :href="getFileUrl(newPurchase.pi_path)"
              target="_blank"
              class="current-file-link"
            >
              View Current PI Document
            </a>
            <div v-if="!newPurchase.pi_path && !newPurchase.pi_file" class="validation-message">
              PI document is required
            </div>
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="showEditDialog = false" class="cancel-btn">Cancel</button>
            <button type="submit" class="submit-btn" :disabled="isSubmittingPurchase">
              <span v-if="isSubmittingPurchase" class="spinner"></span>
              {{ isSubmittingPurchase ? 'Updating...' : 'Update Purchase' }}
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
              <option v-for="car in cars" :key="car.id" :value="car.id">
                {{ car.car_name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>Color</label>
            <select v-model="newDetail.id_color" required>
              <option value="">Select Color</option>
              <option v-for="color in colors" :key="color.id" :value="color.id">
                {{ color.color }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>Quantity</label>
            <input type="number" v-model="newDetail.QTY" min="1" required />
          </div>

          <div class="form-group">
            <label>Amount</label>
            <input type="number" v-model="newDetail.amount" step="0.01" required />
          </div>

          <div class="form-row">
            <div class="form-group half">
              <label>Year</label>
              <input type="number" v-model="newDetail.year" required />
            </div>

            <div class="form-group half">
              <label>Month</label>
              <input type="number" v-model="newDetail.month" min="1" max="12" required />
            </div>
          </div>

          <div class="form-group">
            <label>Price Sell</label>
            <input type="number" v-model="newDetail.price_sell" step="0.01" required />
          </div>

          <div class="form-group">
            <label>Notes</label>
            <textarea v-model="newDetail.notes" rows="3"></textarea>
          </div>

          <div class="form-group checkbox">
            <label>
              <input type="checkbox" v-model="newDetail.is_used_car" />
              Used Car
            </label>
          </div>

          <div class="dialog-buttons">
            <button type="button" @click="showAddDetailDialog = false" class="cancel-btn">
              Cancel
            </button>
            <button type="submit" class="submit-btn" :disabled="isSubmittingDetail">
              <span v-if="isSubmittingDetail" class="spinner"></span>
              {{ isSubmittingDetail ? 'Adding...' : 'Add Detail' }}
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

.submit-btn:hover:not(:disabled) {
  background-color: #059669;
}

.submit-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.6;
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

.payment-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s;
}

.payment-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.payment-btn:before {
  content: '💰';
  font-size: 1rem;
}

.payment-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.update-btn {
  background-color: #059669;
}

.update-btn:hover:not(:disabled) {
  background-color: #047857;
}

.update-btn:before {
  content: '📦';
  font-size: 1rem;
}

.delete-btn {
  background-color: #dc2626;
}

.delete-btn:hover:not(:disabled) {
  background-color: #b91c1c;
}

.delete-btn:before {
  content: '🗑';
  font-size: 1rem;
}

.edit-btn {
  background-color: #3b82f6;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.edit-btn:before {
  content: '✏️';
  font-size: 1rem;
}

.current-file-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.875rem;
  margin-top: 0.5rem;
  display: inline-block;
}

.current-file-link:hover {
  text-decoration: underline;
}

.required {
  color: #ef4444;
  font-weight: 600;
}

.validation-message {
  color: #ef4444;
  font-size: 0.875rem;
  margin-top: 0.25rem;
  font-style: italic;
}

.spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid #ffffff;
  border-radius: 50%;
  border-top-color: transparent;
  animation: spin 1s ease-in-out infinite;
  margin-right: 8px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
