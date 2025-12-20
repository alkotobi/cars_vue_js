<script setup>
import { ref,   watch } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

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

// Storage key for subject text - action-specific
const getSubjectStorageKey = (actionType) => `car-stock-${actionType}-subject`

// Storage key for core content - action-specific
const getCoreContentStorageKey = (actionType) => `car-stock-${actionType}-core-content`

// Storage key for group by - action-specific
const getGroupByStorageKey = (actionType) => `car-stock-${actionType}-group-by`

// Subject text for the report
const subjectText = ref('')

// Core content for the report
const coreContentText = ref('')

// Group by option
const groupBy = ref('')

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

  // Also load saved subject text and core content
  loadSavedSubject(actionType)
  loadSavedCoreContent(actionType)
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

// Function to load saved subject text for specific action
const loadSavedSubject = (actionType) => {
  try {
    const storageKey = getSubjectStorageKey(actionType)
    const saved = localStorage.getItem(storageKey)
    if (saved) {
      subjectText.value = saved
    } else {
      // Set default subject based on action type
      setDefaultSubject(actionType)
    }
  } catch (error) {
    console.warn(`Failed to load saved subject for ${actionType}:`, error)
    setDefaultSubject(actionType)
  }
}

// Function to save subject text for specific action
const saveSubjectText = (actionType) => {
  try {
    const storageKey = getSubjectStorageKey(actionType)
    localStorage.setItem(storageKey, subjectText.value)
  } catch (error) {
    console.warn(`Failed to save subject for ${actionType}:`, error)
  }
}

// Function to load saved core content for specific action
const loadSavedCoreContent = (actionType) => {
  try {
    const storageKey = getCoreContentStorageKey(actionType)
    const saved = localStorage.getItem(storageKey)
    if (saved) {
      coreContentText.value = saved
    } else {
      // Set default core content based on action type
      setDefaultCoreContent(actionType)
    }
  } catch (error) {
    console.warn(`Failed to load saved core content for ${actionType}:`, error)
    setDefaultCoreContent(actionType)
  }
}

// Function to save core content for specific action
const saveCoreContent = (actionType) => {
  try {
    const storageKey = getCoreContentStorageKey(actionType)
    localStorage.setItem(storageKey, coreContentText.value)
  } catch (error) {
    console.warn(`Failed to save core content for ${actionType}:`, error)
  }
}

// Function to set default subject based on action type
const setDefaultSubject = (actionType) => {
  if (actionType === 'print') {
    subjectText.value = t('carStockPrintOptions.reportSubject')
  } else if (actionType === 'loading-order') {
    subjectText.value = t('carStockPrintOptions.loadingOrderOptions')
  } else {
    subjectText.value = ''
  }
}

// Function to set default core content based on action type
const setDefaultCoreContent = (actionType) => {
  if (actionType === 'print') {
    coreContentText.value = t('carStockPrintOptions.reportCoreContentPlaceholder')
  } else if (actionType === 'loading-order') {
    coreContentText.value = t('carStockPrintOptions.reportCoreContentPlaceholder')
  } else {
    coreContentText.value = ''
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
  { key: 'id', label: t('carStockPrintOptions.id'), checked: true },
  { key: 'date_buy', label: t('carStockPrintOptions.dateBuy'), checked: true },
  { key: 'date_sell', label: t('carStockPrintOptions.dateSell'), checked: true },
  { key: 'car_name', label: t('carStockPrintOptions.carName'), checked: true },
  { key: 'vin', label: t('carStockPrintOptions.vin'), checked: true },
  { key: 'color', label: t('carStockPrintOptions.color'), checked: true },
  { key: 'client_name', label: t('carStockPrintOptions.client'), checked: true },
  { key: 'client_id_no', label: t('carStockPrintOptions.clientId'), checked: false },
  { key: 'client_id_picture', label: t('carStockPrintOptions.clientIdPicture'), checked: false },
  { key: 'loading_port', label: t('carStockPrintOptions.loadingPort'), checked: true },
  { key: 'discharge_port', label: t('carStockPrintOptions.dischargePort'), checked: true },
  { key: 'container_ref', label: t('carStockPrintOptions.containerRef'), checked: true },
  { key: 'freight', label: t('carStockPrintOptions.freight'), checked: true },
  { key: 'price_cell', label: t('carStockPrintOptions.fobPrice'), checked: true },
  { key: 'cfr_usd', label: t('carStockPrintOptions.cfrUsd'), checked: true },
  { key: 'cfr_dza', label: t('carStockPrintOptions.cfrDza'), checked: true },
  { key: 'rate', label: t('carStockPrintOptions.rate'), checked: true },
  { key: 'status', label: t('carStockPrintOptions.status'), checked: true },
  { key: 'warehouse_name', label: t('carStockPrintOptions.warehouse'), checked: true },
  { key: 'buy_bill_ref', label: t('carStockPrintOptions.buyBillRef'), checked: false },
  { key: 'sell_bill_ref', label: t('carStockPrintOptions.sellBillRef'), checked: false },
  { key: 'export_lisence_ref', label: t('carStockPrintOptions.exportLicense'), checked: false },
  { key: 'notes', label: t('carStockPrintOptions.notes'), checked: false },
])

const isProcessing = ref(false)

// Load saved preferences when component is created
loadSavedColumns(props.actionType)

// Watch for changes in actionType to reload preferences
watch(
  () => props.actionType,
  (newActionType) => {
    loadSavedColumns(newActionType)
  },
)

const handleClose = () => {
  emit('close')
}

const handleAction = () => {
  const selectedColumns = availableColumns.value.filter((col) => col.checked)

  if (selectedColumns.length === 0) {
    alert(t('carStockPrintOptions.pleaseSelectAtLeastOneColumn'))
    return
  }

  // Save subject text, core content, and group by before emitting
  saveSubjectText(props.actionType)
  saveCoreContent(props.actionType)
  saveGroupBy(props.actionType)

  isProcessing.value = true

  if (props.actionType === 'loading-order') {
    // Emit the loading order event with selected columns, cars, subject, core content, and group by
    emit('loading-order', {
      columns: selectedColumns,
      cars: props.selectedCars,
      subject: subjectText.value,
      coreContent: coreContentText.value,
      groupBy: groupBy.value,
    })
  } else {
    // Emit the print event with selected columns, cars, subject, core content, and group by
    emit('print', {
      columns: selectedColumns,
      cars: props.selectedCars,
      subject: subjectText.value,
      coreContent: coreContentText.value,
      groupBy: groupBy.value,
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

// Function to handle subject text changes
const handleSubjectChange = () => {
  saveSubjectText(props.actionType)
}

// Function to handle core content changes
const handleCoreContentChange = () => {
  saveCoreContent(props.actionType)
}

// Function to load saved group by for specific action
const loadSavedGroupBy = (actionType) => {
  try {
    const storageKey = getGroupByStorageKey(actionType)
    const saved = localStorage.getItem(storageKey)
    if (saved) {
      groupBy.value = saved
    } else {
      groupBy.value = ''
    }
  } catch (error) {
    console.warn(`Failed to load saved group by for ${actionType}:`, error)
    groupBy.value = ''
  }
}

// Function to save group by for specific action
const saveGroupBy = (actionType) => {
  try {
    const storageKey = getGroupByStorageKey(actionType)
    localStorage.setItem(storageKey, groupBy.value)
  } catch (error) {
    console.warn(`Failed to save group by for ${actionType}:`, error)
  }
}

// Function to handle group by changes
const handleGroupByChange = () => {
  saveGroupBy(props.actionType)
}

// Available group by options
const groupByOptions = ref([
  { value: '', label: t('carStockPrintOptions.no_grouping') || 'No Grouping' },
  { value: 'buy_bill_ref', label: t('carStockPrintOptions.groupByBuyBillRef') || 'Buy Bill Ref' },
  { value: 'container_ref', label: t('carStockPrintOptions.groupByContainerRef') || 'Container Ref' },
  { value: 'client_name', label: t('carStockPrintOptions.groupByClient') || 'Client' },
  { value: 'loading_port', label: t('carStockPrintOptions.groupByLoadingPort') || 'Loading Port' },
  { value: 'discharge_port', label: t('carStockPrintOptions.groupByDischargePort') || 'Discharge Port' },
  { value: 'warehouse_name', label: t('carStockPrintOptions.groupByWarehouse') || 'Warehouse' },
  { value: 'status', label: t('carStockPrintOptions.groupByStatus') || 'Status' },
  { value: 'color', label: t('carStockPrintOptions.groupByColor') || 'Color' },
])
</script>

<template>
  <div v-if="show" class="print-options-overlay" @click="handleClose">
    <div class="print-options-modal" @click.stop>
      <div class="modal-header">
        <h3>
          <i :class="actionType === 'loading-order' ? 'fas fa-list-alt' : 'fas fa-print'"></i>
          {{
            actionType === 'loading-order'
              ? t('carStockPrintOptions.loadingOrderOptions')
              : t('carStockPrintOptions.printOptions')
          }}
        </h3>
        <button @click="handleClose" class="close-btn" :title="t('carStockPrintOptions.close')">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-content">
        <div class="selection-info">
          <i class="fas fa-info-circle"></i>
          <span>
            {{
              actionType === 'loading-order'
                ? selectedCars.length === 1
                  ? t('carStockPrintOptions.loadingOrderInfo', { count: selectedCars.length })
                  : t('carStockPrintOptions.loadingOrderInfoPlural', { count: selectedCars.length })
                : selectedCars.length === 1
                  ? t('carStockPrintOptions.selectedCarsInfo', { count: selectedCars.length })
                  : t('carStockPrintOptions.selectedCarsInfoPlural', { count: selectedCars.length })
            }}
          </span>
        </div>

        <div class="subject-section">
          <div class="section-header">
            <h4>
              <i class="fas fa-file-alt"></i>
              {{ t('carStockPrintOptions.reportSubject') }}
            </h4>
          </div>
          <div class="subject-input-container">
            <input
              v-model="subjectText"
              @input="handleSubjectChange"
              type="text"
              class="subject-input"
              :placeholder="t('carStockPrintOptions.reportSubjectPlaceholder')"
              maxlength="100"
            />
            <div class="subject-help">
              <i class="fas fa-info-circle"></i>
              <span>{{ t('carStockPrintOptions.reportSubjectHelp') }}</span>
            </div>
          </div>
        </div>

        <div class="core-content-section">
          <div class="section-header">
            <h4>
              <i class="fas fa-align-left"></i>
              {{ t('carStockPrintOptions.reportCoreContent') }}
            </h4>
          </div>
          <div class="core-content-container">
            <textarea
              v-model="coreContentText"
              @input="handleCoreContentChange"
              class="core-content-textarea"
              :placeholder="t('carStockPrintOptions.reportCoreContentPlaceholder')"
              rows="4"
              maxlength="500"
            ></textarea>
            <div class="core-content-help">
              <i class="fas fa-info-circle"></i>
              <span>{{ t('carStockPrintOptions.reportCoreContentHelp') }}</span>
            </div>
          </div>
        </div>

        <div class="group-by-section">
          <div class="section-header">
            <h4>
              <i class="fas fa-layer-group"></i>
              {{ t('carStockPrintOptions.groupBy') || 'Group By' }}
            </h4>
          </div>
          <div class="group-by-container">
            <select
              v-model="groupBy"
              @change="handleGroupByChange"
              class="group-by-select"
            >
              <option
                v-for="option in groupByOptions"
                :key="option.value"
                :value="option.value"
              >
                {{ option.label }}
              </option>
            </select>
            <div class="group-by-help">
              <i class="fas fa-info-circle"></i>
              <span>{{ t('carStockPrintOptions.groupByHelp') || 'Group printed rows by a specific column' }}</span>
            </div>
          </div>
        </div>

        <div class="columns-section">
          <div class="section-header">
            <h4>
              <i class="fas fa-columns"></i>
              {{ t('carStockPrintOptions.selectColumnsToPrint') }}
            </h4>
            <div class="column-actions">
              <button
                @click="selectAllColumns"
                class="action-btn"
                :title="t('carStockPrintOptions.selectAllColumnsTitle')"
              >
                <i class="fas fa-check-square"></i>
                {{ t('carStockPrintOptions.selectAllColumns') }}
              </button>
              <button
                @click="deselectAllColumns"
                class="action-btn"
                :title="t('carStockPrintOptions.deselectAllColumnsTitle')"
              >
                <i class="fas fa-square"></i>
                {{ t('carStockPrintOptions.deselectAllColumns') }}
              </button>
              <button
                @click="selectCommonColumns"
                class="action-btn"
                :title="t('carStockPrintOptions.selectCommonColumnsTitle')"
              >
                <i class="fas fa-star"></i>
                {{ t('carStockPrintOptions.selectCommonColumns') }}
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
          {{ t('carStockPrintOptions.cancel') }}
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
          <span>{{
            actionType === 'loading-order'
              ? t('carStockPrintOptions.generateLoadingOrder')
              : t('carStockPrintOptions.print')
          }}</span>
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

.subject-section {
  margin-bottom: 20px;
}

.subject-input-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.subject-input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  transition: all 0.2s ease;
  background: white;
}

.subject-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.subject-input::placeholder {
  color: #9ca3af;
}

.subject-help {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #6b7280;
  font-style: italic;
}

.subject-help i {
  color: #9ca3af;
  font-size: 10px;
}

.core-content-section {
  margin-bottom: 20px;
}

.core-content-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.core-content-textarea {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  transition: all 0.2s ease;
  background: white;
  resize: vertical;
  min-height: 100px;
  line-height: 1.5;
}

.core-content-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.core-content-textarea::placeholder {
  color: #9ca3af;
}

.core-content-help {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #6b7280;
  font-style: italic;
}

.core-content-help i {
  color: #9ca3af;
  font-size: 10px;
}

.group-by-section {
  margin-bottom: 20px;
}

.group-by-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group-by-select {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  transition: all 0.2s ease;
  background: white;
  cursor: pointer;
}

.group-by-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.group-by-help {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #6b7280;
  font-style: italic;
}

.group-by-help i {
  color: #9ca3af;
  font-size: 10px;
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
