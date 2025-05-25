<script setup>
import { ref, onMounted, defineProps, defineEmits, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useRouter } from 'vue-router'
import SellBillPrintOption from './SellBillPrintOption.vue'

const props = defineProps({
  onEdit: Function,
  onDelete: Function,
  onSelect: Function,
})
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)
const emit = defineEmits(['refresh', 'select-bill'])

const router = useRouter()
const { callApi } = useApi()
const sellBills = ref([])
const loading = ref(true)
const error = ref(null)
const selectedBillId = ref(null)
const showPrintOptions = ref(false)
const selectedPrintBillId = ref(null)

// Add computed property for payments permission
const can_c_sell_payments = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_c_sell_payments')
})

// Add computed property for delete permission
const can_delete_sell_bill = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_delete_sell_bill')
})

// Add computed property for edit permission
const can_edit_sell_bill = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_sell_bill')
})

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchSellBills()
})

const fetchSellBills = async () => {
  loading.value = true
  error.value = null

  try {
    // Different query based on admin status
    const query = isAdmin.value
      ? `
      SELECT 
        sb.*,
        c.name as broker_name,
        u.username as created_by,
        (
          SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cfr
      FROM sell_bill sb
      LEFT JOIN clients c ON sb.id_broker = c.id AND c.is_broker = 1
      LEFT JOIN users u ON sb.id_user = u.id
      ORDER BY sb.date_sell DESC
    `
      : `
      SELECT 
        sb.*,
        c.name as broker_name,
        u.username as created_by,
        (
          SELECT SUM(cs.price_cell + COALESCE(cs.freight, 0))
          FROM cars_stock cs
          WHERE cs.id_sell = sb.id
        ) as total_cfr
      FROM sell_bill sb
      LEFT JOIN clients c ON sb.id_broker = c.id AND c.is_broker = 1
      LEFT JOIN users u ON sb.id_user = u.id
      WHERE sb.id_user = ?
      ORDER BY sb.date_sell DESC
    `

    const params = isAdmin.value ? [] : [user.value?.id]
    const result = await callApi({ query, params })

    if (result.success) {
      sellBills.value = result.data.map((bill) => ({
        ...bill,
        total_cfr: Number(bill.total_cfr) || 0,
      }))
    } else {
      error.value = result.error || 'Failed to fetch sell bills'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const handleEdit = (bill) => {
  if (props.onEdit) {
    props.onEdit(bill)
  }
}

const handleDelete = (id) => {
  if (props.onDelete) {
    props.onDelete(id)
  }
}

const handlePrint = (id) => {
  selectedPrintBillId.value = id
  showPrintOptions.value = true
}

const handlePrintClose = () => {
  showPrintOptions.value = false
  selectedPrintBillId.value = null
}

const handlePrintProceed = (options) => {
  console.log('Print options:', options)
  showPrintOptions.value = false
  selectedPrintBillId.value = null
}

const handlePayments = (billId) => {
  // Open payments in a new tab using router.resolve
  const route = router.resolve({
    name: 'sell-bill-payments',
    params: { id: billId },
  })
  window.open(route.href, '_blank')
}

const selectBill = (bill) => {
  selectedBillId.value = bill.id

  // Emit the selected bill ID to the parent component
  emit('select-bill', bill.id)

  if (props.onSelect) {
    props.onSelect(bill.id)
  }
}

// Expose the fetchSellBills method to parent component
defineExpose({ fetchSellBills })
</script>

<template>
  <div class="sell-bills-table-component">
    <h3>Sell Bills</h3>

    <!-- Add toolbar -->
    <div v-if="selectedBillId" class="toolbar">
      <div class="selected-bill-info">
        Selected Bill: <span class="bill-id">#{{ selectedBillId }}</span>
      </div>
      <div class="toolbar-actions">
        <button
          @click="handleEdit(sellBills.find((b) => b.id === selectedBillId))"
          class="edit-btn"
          :disabled="!can_edit_sell_bill"
          :class="{ disabled: !can_edit_sell_bill }"
        >
          Edit Bill
        </button>
        <button
          @click="handleDelete(selectedBillId)"
          class="delete-btn"
          :disabled="!can_delete_sell_bill"
          :class="{ disabled: !can_delete_sell_bill }"
        >
          Delete Bill
        </button>
        <button @click="handlePrint(selectedBillId)" class="print-btn">Print Bill</button>
        <button
          @click="handlePayments(selectedBillId)"
          class="payment-btn"
          :disabled="!can_c_sell_payments"
          :class="{ disabled: !can_c_sell_payments }"
        >
          Payments
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="sellBills.length === 0" class="no-data">No sell bills found</div>
    <table v-else class="sell-bills-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Reference</th>
          <th>Date</th>
          <th>Broker</th>
          <th>Created By</th>
          <th>Notes</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="bill in sellBills"
          :key="bill.id"
          @click="selectBill(bill)"
          :class="{ selected: selectedBillId === bill.id }"
        >
          <td>{{ bill.id }}</td>
          <td>{{ bill.bill_ref || 'N/A' }}</td>
          <td>{{ new Date(bill.date_sell).toLocaleDateString() }}</td>
          <td>{{ bill.broker_name || 'N/A' }}</td>
          <td>{{ bill.created_by || 'N/A' }}</td>
          <td>{{ bill.notes || 'N/A' }}</td>
        </tr>
      </tbody>
    </table>
    <SellBillPrintOption
      :visible="showPrintOptions"
      :billId="selectedPrintBillId"
      @close="handlePrintClose"
      @proceed="handlePrintProceed"
    />
  </div>
</template>

<style scoped>
.sell-bills-table-component {
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #f8f9fa;
  padding: 15px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.selected-bill-info {
  font-size: 1rem;
  color: #374151;
}

.bill-id {
  font-weight: 600;
  color: #3b82f6;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
}

.sell-bills-table {
  width: 100%;
  border-collapse: collapse;
}

.sell-bills-table th,
.sell-bills-table td {
  padding: 10px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.sell-bills-table th {
  background-color: #f2f2f2;
  font-weight: bold;
}

.sell-bills-table tr:hover {
  background-color: #f5f5f5;
  cursor: pointer;
}

.sell-bills-table tr.selected {
  background-color: #e3f2fd;
}

.loading,
.error,
.no-data {
  padding: 20px;
  text-align: center;
}

.error {
  color: red;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  font-weight: 500;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  font-weight: 500;
}

.print-btn {
  background-color: #6366f1;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  font-weight: 500;
}

.payment-btn {
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  font-weight: 500;
}

.edit-btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background-color: #9ca3af;
}

.edit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background-color: #9ca3af;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.delete-btn:hover {
  background-color: #dc2626;
}

.print-btn:hover {
  background-color: #4f46e5;
}

.payment-btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background-color: #9ca3af;
}

.payment-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background-color: #9ca3af;
}

.delete-btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background-color: #9ca3af;
}

.delete-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background-color: #9ca3af;
}
</style>
