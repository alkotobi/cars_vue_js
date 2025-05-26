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
const isProcessing = ref(false)

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
    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Loading bills...</span>
    </div>

    <div v-if="error" class="error">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-else-if="sellBills.length === 0" class="no-data">
      <i class="fas fa-inbox fa-2x"></i>
      <p>No sell bills found</p>
    </div>

    <table v-else class="sell-bills-table">
      <thead>
        <tr>
          <th><i class="fas fa-hashtag"></i> ID</th>
          <th><i class="fas fa-barcode"></i> Reference</th>
          <th><i class="fas fa-calendar"></i> Date</th>
          <th><i class="fas fa-user-tie"></i> Broker</th>
          <th><i class="fas fa-user"></i> Created By</th>
          <th><i class="fas fa-sticky-note"></i> Notes</th>
          <th><i class="fas fa-cog"></i> Actions</th>
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
          <td class="actions">
            <button
              v-if="can_edit_sell_bill"
              @click.stop="handleEdit(bill)"
              :disabled="isProcessing"
              class="btn edit-btn"
              title="Edit Bill"
            >
              <i class="fas fa-edit"></i>
            </button>
            <button
              v-if="can_delete_sell_bill"
              @click.stop="handleDelete(bill.id)"
              :disabled="isProcessing"
              class="btn delete-btn"
              title="Delete Bill"
            >
              <i class="fas fa-trash-alt"></i>
            </button>
            <button
              @click.stop="handlePrint(bill.id)"
              :disabled="isProcessing"
              class="btn print-btn"
              title="Print Bill"
            >
              <i class="fas fa-print"></i>
            </button>
          </td>
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
  position: relative;
  margin-bottom: 20px;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  z-index: 10;
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
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
}

.sell-bills-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background-color: #f3f4f6;
}

.sell-bills-table th,
.sell-bills-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.sell-bills-table th {
  background-color: #f3f4f6;
  font-weight: 600;
  color: #374151;
}

.sell-bills-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.sell-bills-table tr:hover {
  background-color: #f9fafb;
  cursor: pointer;
}

.sell-bills-table tr.selected {
  background-color: #e5edff;
}

.loading,
.error,
.no-data {
  padding: 2rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  color: #6b7280;
}

.error {
  color: #dc2626;
  background-color: #fee2e2;
  border-radius: 8px;
  padding: 1rem;
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.print-btn {
  background-color: #10b981;
  color: white;
}

.print-btn:hover:not(:disabled) {
  background-color: #059669;
}

/* Add smooth transitions */
.sell-bills-table tr {
  transition: background-color 0.2s;
}

.btn i {
  transition: transform 0.2s;
}

.btn:hover:not(:disabled) i {
  transform: scale(1.1);
}
</style>
