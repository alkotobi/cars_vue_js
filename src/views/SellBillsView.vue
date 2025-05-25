<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'
import SellBillsTable from '../components/sells/SellBillsTable.vue'
import SellBillForm from '../components/sells/SellBillForm.vue'
import SellBillCarsTable from '../components/sells/SellBillCarsTable.vue'
import UnassignedCarsTable from '../components/sells/UnassignedCarsTable.vue'

const { callApi } = useApi()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const selectedBillId = ref(null)
const editingBill = ref(null)
const sellBillsTableRef = ref(null)
const unassignedCarsTableRef = ref(null)
const sellBillCarsTableRef = ref(null)

// Add user and isAdmin
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
})

const handleSelectBill = (billId) => {
  selectedBillId.value = billId
  
  // Update the unassigned cars table with the selected bill ID
  if (unassignedCarsTableRef.value) {
    unassignedCarsTableRef.value.setSellBillId(billId)
  }
}

const openAddDialog = () => {
  editingBill.value = {
    id_broker: null,
    date_sell: new Date().toISOString().split('T')[0],
    notes: '',
    id_user: user.value?.id
  }
  showAddDialog.value = true
}

const handleEditBill = (bill) => {
  editingBill.value = { ...bill }
  showEditDialog.value = true
}

const handleDeleteBill = async (id) => {
  if (!isAdmin.value) {
    alert('Only admins can delete sell bills.')
    return
  }
  if (!confirm('Are you sure you want to delete this sell bill? This will also unassign all cars from this bill.')) {
    return
  }
  
  try {
    // First, unassign all cars from this bill
    const unassignResult = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = NULL,
            id_client = NULL,
            id_port_discharge = NULL,
            freight = NULL,
            id_sell_pi = NULL
        WHERE id_sell = ?
      `,
      params: [id]
    })
    
    if (!unassignResult.success) {
      console.error('Failed to unassign cars:', unassignResult.error)
      alert('Failed to unassign cars from bill: ' + unassignResult.error)
      return
    }
    
    // Then delete the bill
    const result = await callApi({
      query: 'DELETE FROM sell_bill WHERE id = ?',
      params: [id]
    })
    
    if (result.success) {
      // Refresh the table
      if (sellBillsTableRef.value) {
        sellBillsTableRef.value.fetchSellBills()
      }
      
      // If this was the selected bill, clear the selection
      if (selectedBillId.value === id) {
        selectedBillId.value = null
      }
      
      // Refresh the cars tables
      handleCarsTableRefresh()
    } else {
      console.error('Failed to delete sell bill:', result.error)
      alert('Failed to delete sell bill: ' + result.error)
    }
  } catch (err) {
    console.error('Error deleting sell bill:', err)
    alert('Error deleting sell bill: ' + err.message)
  }
}

const handleSave = () => {
  console.log('handleSave called')
  showAddDialog.value = false
  showEditDialog.value = false
  
  // Refresh the table
  if (sellBillsTableRef.value) {
    console.log('Refreshing table')
    sellBillsTableRef.value.fetchSellBills()
  } else {
    console.warn('sellBillsTableRef is null')
  }
}

const handleCarsTableRefresh = async () => {
  // Refresh the main sell bills table
  if (sellBillsTableRef.value) {
    await sellBillsTableRef.value.fetchSellBills()
  }
  
  // Refresh the unassigned cars table
  if (unassignedCarsTableRef.value) {
    await unassignedCarsTableRef.value.fetchUnassignedCars()
  }
  
  // If a bill is selected, refresh the cars for that bill
  if (selectedBillId.value && sellBillCarsTableRef.value) {
    await sellBillCarsTableRef.value.fetchCarsByBillId(selectedBillId.value)
  }
}
</script>

<template>
  <div class="sell-bills-view">
    <div class="header">
      <h2>Sell Bills Management</h2>
      <div class="header-actions">
        <button @click="openAddDialog" class="add-btn">Add Sell Bill</button>
      </div>
    </div>
    
    <div class="content">
      <SellBillsTable 
        ref="sellBillsTableRef"
        :onEdit="handleEditBill"
        :onDelete="handleDeleteBill"
        :onSelect="handleSelectBill"
        @select-bill="handleSelectBill"
        :isAdmin="isAdmin" 
      />
      
      <SellBillCarsTable 
        ref="sellBillCarsTableRef"
        :sellBillId="selectedBillId" 
        @refresh="handleCarsTableRefresh"
      />
      
      <UnassignedCarsTable 
        ref="unassignedCarsTableRef"
        @refresh="handleCarsTableRefresh"
      />
    </div>
    
    <!-- Add Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <SellBillForm
          mode="add"
          :billData="editingBill"
          @save="handleSave"
          @cancel="showAddDialog = false"
        />
      </div>
    </div>
    
    <!-- Edit Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <SellBillForm
          mode="edit"
          :billData="editingBill"
          @save="handleSave"
          @cancel="showEditDialog = false"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.sell-bills-view {
  width: 100%;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.add-btn {
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
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
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}
</style>