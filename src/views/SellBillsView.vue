<script setup>
import { ref, onMounted } from 'vue'
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
const sellBillCarsTableRef = ref(null)  // Add this line

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
    notes: ''
  }
  showAddDialog.value = true
}

const handleEditBill = (bill) => {
  editingBill.value = { ...bill }
  showEditDialog.value = true
}

const handleDeleteBill = async (id) => {
  if (!confirm('Are you sure you want to delete this sell bill?')) {
    return
  }
  
  try {
    const result = await callApi({
      query: 'DELETE FROM sell_bill WHERE id = ?',
      params: [id]
    })
    
    if (result.success) {
      // Refresh the table
      if (sellBillsTableRef.value) {
        sellBillsTableRef.value.fetchSellBills()
      }
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
  showAddDialog.value = false
  showEditDialog.value = false
  
  // Refresh the table
  if (sellBillsTableRef.value) {
    sellBillsTableRef.value.fetchSellBills()
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
      <button @click="openAddDialog" class="add-btn">Add Sell Bill</button>
    </div>
    
    <div class="content">
      <SellBillsTable 
        ref="sellBillsTableRef"
        :onEdit="handleEditBill"
        :onDelete="handleDeleteBill"
        :onSelect="handleSelectBill"
        @select-bill="handleSelectBill"
      />
      
      <!-- Update to add ref for SellBillCarsTable -->
      <SellBillCarsTable 
        ref="sellBillCarsTableRef"
        :sellBillId="selectedBillId" 
        @refresh="handleCarsTableRefresh"
      />
      
      <!-- Unassigned cars that can be added to the bill -->
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