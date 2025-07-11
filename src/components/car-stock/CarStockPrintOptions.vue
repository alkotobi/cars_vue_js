<script setup>
import { ref, defineProps, defineEmits, watch } from 'vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  selectedCars: {
    type: Array,
    default: () => [],
  },
  actionType: {
    type: String,
    default: 'print', // 'print' or 'loading-order'
  },
})

const emit = defineEmits(['close', 'print', 'loading-order'])

// Storage key for user preferences - now action-specific
const getStorageKey = (actionType) => `car-stock-${actionType}-columns`

// Function to load saved column preferences for specific action
const loadSavedColumns = (actionType) => {
  try {
    const storageKey = getStorageKey(actionType)
    const saved = localStorage.getItem(storageKey)
    if (saved) {
      const savedColumns = JSON.parse(saved)
      // Update availableColumns with saved preferences
      availableColumns.value.forEach((col) => {
        const savedCol = savedColumns.find((sc) => sc.key === col.key)
        if (savedCol) {
          col.checked = savedCol.checked
        }
      })
    } else {
      // Set default preferences based on action type
      setDefaultPreferences(actionType)
    }
  } catch (error) {
    console.warn(`Failed to load saved column preferences for ${actionType}:`, error)
    setDefaultPreferences(actionType)
  }
}

// Function to save column preferences for specific action
const saveColumnPreferences = (actionType) => {
  try {
    const storageKey = getStorageKey(actionType)
    const columnsToSave = availableColumns.value.map((col) => ({
      key: col.key,
      checked: col.checked,
    }))
    localStorage.setItem(storageKey, JSON.stringify(columnsToSave))
  } catch (error) {
    console.warn(`Failed to save column preferences for ${actionType}:`, error)
  }
}

// Function to set default preferences based on action type
const setDefaultPreferences = (actionType) => {
  // Reset all to false first
  availableColumns.value.forEach((col) => {
    col.checked = false
  })

  if (actionType === 'print') {
    // Default columns for printing
    const printDefaults = [
      'id',
      'car_name',
      'vin',
      'color',
      'client_name',
      'price_cell',
      'status',
      'date_buy',
      'date_sell',
    ]
    availableColumns.value.forEach((col) => {
      if (printDefaults.includes(col.key)) {
        col.checked = true
      }
    })
  } else if (actionType === 'loading-order') {
    // Default columns for loading order
    const loadingOrderDefaults = [
      'id',
      'car_name',
      'vin',
      'color',
      'loading_port',
      'discharge_port',
      'container_ref',
      'freight',
      'warehouse_name',
    ]
    availableColumns.value.forEach((col) => {
      if (loadingOrderDefaults.includes(col.key)) {
        col.checked = true
      }
    })
  }
  // Add more action types here as needed
}

// Available columns for printing
const availableColumns = ref([
  { key: 'id', label: 'ID', checked: true },
  { key: 'date_buy', label: 'Date Buy', checked: true },
  { key: 'date_sell', label: 'Date Sell', checked: true },
  { key: 'car_name', label: 'Car Name', checked: true },
  { key: 'vin', label: 'VIN', checked: true },
  { key: 'color', label: 'Color', checked: true },
  { key: 'client_name', label: 'Client', checked: true },
  { key: 'client_id_no', label: 'Client ID', checked: false },
  { key: 'client_id_picture', label: 'Client ID Picture', checked: false },
  { key: 'loading_port', label: 'Loading Port', checked: true },
  { key: 'discharge_port', label: 'Discharge Port', checked: true },
  { key: 'container_ref', label: 'Container Ref', checked: true },
  { key: 'freight', label: 'Freight', checked: true },
  { key: 'price_cell', label: 'FOB Price', checked: true },
  { key: 'cfr_usd', label: 'CFR USD', checked: true },
  { key: 'cfr_dza', label: 'CFR DZA', checked: true },
  { key: 'rate', label: 'Rate', checked: true },
  { key: 'status', label: 'Status', checked: true },
  { key: 'warehouse_name', label: 'Warehouse', checked: true },
  { key: 'buy_bill_ref', label: 'Buy Bill Ref', checked: false },
  { key: 'sell_bill_ref', label: 'Sell Bill Ref', checked: false },
  { key: 'export_lisence_ref', label: 'Export License', checked: false },
  { key: 'notes', label: 'Notes', checked: false },
])

const isProcessing = ref(false)

// Load saved preferences when component is created
loadSavedColumns(props.actionType)

// Watch for changes in actionType to reload preferences
watch(() => props.actionType, (newActionType) => {
  loadSavedColumns(newActionType)
})

const handleClose = () => {
  emit('close')
}

const handleAction = () => {
  const selectedColumns = availableColumns.value.filter((col) => col.checked)

  if (selectedColumns.length === 0) {
    alert('Please select at least one column')
    return
  }

  isProcessing.value = true

  if (props.actionType === 'loading-order') {
    // Emit the loading order event with selected columns and cars
    emit('loading-order', {
      columns: selectedColumns,
      cars: props.selectedCars,
    })
  } else {
    // Emit the print event with selected columns and cars
    emit('print', {
      columns: selectedColumns,
      cars: props.selectedCars,
    })
  }

  isProcessing.value = false
  // Don't call handleClose() here - let the parent handle it
}

const toggleColumn = (columnKey) => {
  const column = availableColumns.value.find((col) => col.key === columnKey)
  if (column) {
    column.checked = !column.checked
    saveColumnPreferences(props.actionType)
  }
}

const selectAllColumns = () => {
  availableColumns.value.forEach((col) => {
    col.checked = true
  })
  saveColumnPreferences(props.actionType)
}

const deselectAllColumns = () => {
  availableColumns.value.forEach((col) => {
    col.checked = false
  })
  saveColumnPreferences(props.actionType)
}

const selectCommonColumns = () => {
  // Reset all to false first
  availableColumns.value.forEach((col) => {
    col.checked = false
  })

  // Select common/essential columns based on action type
  let commonColumns = []
  if (props.actionType === 'print') {
    commonColumns = ['id', 'car_name', 'vin', 'color', 'client_name', 'price_cell', 'status']
  } else if (props.actionType === 'loading-order') {
    commonColumns = [
      'id',
      'car_name',
      'vin',
      'color',
      'loading_port',
      'discharge_port',
      'container_ref',
    ]
  }

  availableColumns.value.forEach((col) => {
    if (commonColumns.includes(col.key)) {
      col.checked = true
    }
  })
  saveColumnPreferences(props.actionType)
}
</script>

<template>
  <div v-if="show" class="print-options-overlay" @click="handleClose">
    <div class="print-options-modal" @click.stop>
      <div class="modal-header">
        <h3>
          <i :class="actionType === 'loading-order' ? 'fas fa-list-alt' : 'fas fa-print'"></i>
          {{ actionType === 'loading-order' ? 'Loading Order' : 'Print' }} Options
        </h3>
        <button @click="handleClose" class="close-btn" title="Close">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-content">
        <div class="selection-info">
          <i class="fas fa-info-circle"></i>
          <span
            >{{ actionType === 'loading-order' ? 'Loading Order' : 'Printing' }}
            {{ selectedCars.length }} selected car{{ selectedCars.length === 1 ? '' : 's' }}</span
          >
        </div>

        <div class="columns-section">
          <div class="section-header">
            <h4>
              <i class="fas fa-columns"></i>
              Select Columns to Print
            </h4>
            <div class="column-actions">
              <button @click="selectAllColumns" class="action-btn" title="Select All Columns">
                <i class="fas fa-check-square"></i>
                All
              </button>
              <button @click="deselectAllColumns" class="action-btn" title="Deselect All Columns">
                <i class="fas fa-square"></i>
                None
              </button>
              <button @click="selectCommonColumns" class="action-btn" title="Select Common Columns">
                <i class="fas fa-star"></i>
                Common
              </button>
            </div>
          </div>

          <div class="columns-grid">
            <label
              v-for="column in availableColumns"
              :key="column.key"
              class="column-checkbox"
              :class="{ checked: column.checked }"
            >
              <input type="checkbox" :checked="column.checked" @change="toggleColumn(column.key)" />
              <span class="checkbox-label">{{ column.label }}</span>
            </label>
          </div>
        </div>
      </div>

      <div class="modal-actions">
        <button @click="handleClose" class="cancel-btn">
          <i class="fas fa-times"></i>
          Cancel
        </button>
        <button
          @click="handleAction"
          :class="[
            actionType === 'loading-order' ? 'loading-order-btn' : 'print-btn',
            { processing: isProcessing },
          ]"
          :disabled="isProcessing || availableColumns.filter((col) => col.checked).length === 0"
        >
          <i :class="actionType === 'loading-order' ? 'fas fa-list-alt' : 'fas fa-print'"></i>
          <span>{{ actionType === 'loading-order' ? 'Generate Loading Order' : 'Print' }}</span>
          <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.print-options-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 20px;
}

.print-options-modal {
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  background: #f8f9fa;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  font-size: 18px;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #374151;
}

.modal-content {
  padding: 24px;
  flex: 1;
  overflow-y: auto;
}

.selection-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #eff6ff;
  border: 1px solid #dbeafe;
  border-radius: 8px;
  margin-bottom: 20px;
  color: #1e40af;
  font-weight: 500;
}

.selection-info i {
  color: #3b82f6;
}

.columns-section {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h4 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header h4 i {
  color: #6b7280;
}

.column-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  padding: 6px 12px;
  background: #f3f4f6;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  color: #374151;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 4px;
}

.action-btn:hover {
  background: #e5e7eb;
  border-color: #9ca3af;
}

.columns-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.column-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: white;
}

.column-checkbox:hover {
  background: #f9fafb;
  border-color: #d1d5db;
}

.column-checkbox.checked {
  background: #eff6ff;
  border-color: #3b82f6;
}

.column-checkbox input[type='checkbox'] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
}

.checkbox-label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  cursor: pointer;
  user-select: none;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
  background: #f8f9fa;
}

.cancel-btn,
.print-btn,
.loading-order-btn {
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
  border: none;
}

.cancel-btn {
  background: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover {
  background: #e5e7eb;
}

.print-btn {
  background: #3b82f6;
  color: white;
}

.print-btn:hover:not(:disabled) {
  background: #2563eb;
}

.print-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.print-btn.processing {
  position: relative;
  pointer-events: none;
}

.loading-order-btn {
  background: #059669;
  color: white;
}

.loading-order-btn:hover:not(:disabled) {
  background: #047857;
}

.loading-order-btn:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.loading-order-btn.processing {
  position: relative;
  pointer-events: none;
}

.loading-indicator {
  margin-left: 4px;
}

@media (max-width: 768px) {
  .print-options-modal {
    margin: 10px;
    max-height: 95vh;
  }

  .modal-header,
  .modal-content,
  .modal-actions {
    padding: 16px;
  }

  .columns-grid {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .column-actions {
    width: 100%;
    justify-content: space-between;
  }

  .modal-actions {
    flex-direction: column;
  }

  .cancel-btn,
  .print-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
