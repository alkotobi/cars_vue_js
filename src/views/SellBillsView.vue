<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useApi } from '../composables/useApi'
import SellBillsTable from '../components/sells/SellBillsTable.vue'
import SellBillForm from '../components/sells/SellBillForm.vue'
import SellBillCarsTable from '../components/sells/SellBillCarsTable.vue'
import UnassignedCarsTable from '../components/sells/UnassignedCarsTable.vue'
import TaskForm from '../components/car-stock/TaskForm.vue'

const { callApi } = useApi()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const selectedBillId = ref(null)
const editingBill = ref(null)
const sellBillsTableRef = ref(null)
const unassignedCarsTableRef = ref(null)
const sellBillCarsTableRef = ref(null)
const isProcessing = ref(false)

// Add task form state
const showTaskForm = ref(false)
const selectedBillForTask = ref(null)

// Add user and isAdmin
const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)
const can_create_sell_bill = computed(
  () =>
    user.value?.permissions?.some((p) => p.permission_name === 'can_sell_cars') || isAdmin.value,
)

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
})

const handleSelectBill = (billId) => {
  console.log('handleSelectBill called with billId:', billId)
  selectedBillId.value = billId

  // Update the unassigned cars table with the selected bill ID
  if (unassignedCarsTableRef.value) {
    unassignedCarsTableRef.value.setSellBillId(billId)
  }

  // Scroll to unassigned cars table after a short delay to ensure it's rendered
  setTimeout(() => {
    console.log('Attempting to scroll to unassigned cars table...')
    if (unassignedCarsTableRef.value) {
      // Get the actual DOM element, handling Vue 3 ref structure
      let element = unassignedCarsTableRef.value.$el
      console.log('Initial element:', element)

      // If it's a text node, try to get the parent element
      if (element && element.nodeType === Node.TEXT_NODE) {
        element = element.parentElement
        console.log('Using parent element:', element)
      }

      // Alternative: try to get the element by querying the DOM
      if (!element || !element.getBoundingClientRect) {
        element =
          document.querySelector('#unassigned-cars-table') ||
          document.querySelector('.unassigned-cars-table-component')
        console.log('Found element by querySelector:', element)
      }

      if (element && element.getBoundingClientRect) {
        const yOffset = -50 // Increased offset to account for any fixed headers
        const y = element.getBoundingClientRect().top + window.pageYOffset + yOffset
        console.log('Scrolling to position:', y)
        window.scrollTo({ top: y, behavior: 'smooth' })
      } else {
        console.log('Unassigned cars table element not found or invalid')
      }
    } else {
      console.log('unassignedCarsTableRef not available')
    }
  }, 100) // Reduced delay for faster response

  // Alternative scroll method if the first one doesn't work
  setTimeout(() => {
    console.log('Attempting alternative scroll method...')
    let element = unassignedCarsTableRef.value?.$el

    // If it's a text node, try to get the parent element
    if (element && element.nodeType === Node.TEXT_NODE) {
      element = element.parentElement
    }

    // Alternative: try to get the element by querying the DOM
    if (!element || !element.scrollIntoView) {
      element =
        document.querySelector('#unassigned-cars-table') ||
        document.querySelector('.unassigned-cars-table-component')
    }

    if (element && element.scrollIntoView) {
      console.log('Using scrollIntoView method')
      element.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
        inline: 'nearest',
      })
    } else {
      console.log('Element not found for scrollIntoView')
    }
  }, 200)
}

const openAddDialog = () => {
  isProcessing.value = true
  editingBill.value = {
    id_broker: null,
    date_sell: new Date().toISOString().split('T')[0],
    notes: '',
    id_user: user.value?.id,
  }
  showAddDialog.value = true
  isProcessing.value = false
}

const handleEditBill = (bill) => {
  isProcessing.value = true
  editingBill.value = { ...bill }
  showEditDialog.value = true
  isProcessing.value = false
}

const handleDeleteBill = async (id) => {
  if (!isAdmin.value) {
    alert('Only admins can delete sell bills.')
    return
  }
  if (
    !confirm(
      'Are you sure you want to delete this sell bill? This will also unassign all cars from this bill.',
    )
  ) {
    return
  }

  isProcessing.value = true
  try {
    // First, unassign all cars from this bill
    const unassignResult = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = NULL,
            id_client = NULL,
            id_port_discharge = NULL,
            freight = NULL,
            id_sell_pi = NULL,
            is_tmp_client = 0
        WHERE id_sell = ?
      `,
      params: [id],
    })

    if (!unassignResult.success) {
      console.error('Failed to unassign cars:', unassignResult.error)
      alert('Failed to unassign cars from bill: ' + unassignResult.error)
      return
    }

    // Then delete the bill
    const result = await callApi({
      query: 'DELETE FROM sell_bill WHERE id = ?',
      params: [id],
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
  } finally {
    isProcessing.value = false
  }
}

const handleSave = async (savedBill) => {
  isProcessing.value = true
  showAddDialog.value = false
  showEditDialog.value = false

  try {
    // Refresh the table
    if (sellBillsTableRef.value) {
      await sellBillsTableRef.value.fetchSellBills()

      // After refresh, select the newly created bill
      if (savedBill && savedBill.id) {
        handleSelectBill(savedBill.id)
      }
    }
  } catch (error) {
    console.error('Error refreshing table:', error)
  } finally {
    isProcessing.value = false
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

// Add task handling methods
const openTaskForBill = (bill) => {
  console.log('openTaskForBill called with bill:', bill)
  selectedBillForTask.value = bill
  showTaskForm.value = true
}

const handleTaskCreated = () => {
  showTaskForm.value = false
  // Don't set selectedBillForTask to null to avoid prop validation errors
  // Optionally refresh data if needed
}
</script>

<template>
  <div class="sell-bills-view">
    <div class="header">
      <h2>Sell Bills Management</h2>
      <div class="header-actions">
        <button
          v-if="can_create_sell_bill"
          @click="showAddDialog = true"
          class="add-btn"
          :disabled="isProcessing"
        >
          <i class="fas fa-plus"></i> Add Sell Bill
        </button>
      </div>
    </div>

    <div class="content">
      <SellBillsTable
        ref="sellBillsTableRef"
        :onEdit="handleEditBill"
        :onDelete="handleDeleteBill"
        :onSelect="handleSelectBill"
        :onTask="openTaskForBill"
        @select-bill="handleSelectBill"
        :isAdmin="isAdmin"
        :selectedBillId="selectedBillId"
      />

      <SellBillCarsTable
        ref="sellBillCarsTableRef"
        id="sell-bill-cars-table"
        :sellBillId="selectedBillId"
        @refresh="handleCarsTableRefresh"
      />

      <UnassignedCarsTable
        ref="unassignedCarsTableRef"
        id="unassigned-cars-table"
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

    <!-- Task Form -->
    <TaskForm
      v-if="selectedBillForTask"
      :entityType="'sell'"
      :entityData="selectedBillForTask"
      :isVisible="showTaskForm"
      @task-created="handleTaskCreated"
      @cancel="showTaskForm = false"
    />
  </div>
</template>

<style scoped>
.sell-bills-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  margin: 0;
  color: #2c3e50;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.add-btn:hover {
  background-color: #45a049;
}

.add-btn:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.content {
  display: flex;
  flex-direction: column;
  gap: 20px;
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
  padding: 20px;
}

.dialog {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  width: 90%;
  max-width: 500px;
}
</style>
