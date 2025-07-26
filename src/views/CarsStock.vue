<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useEnhancedI18n } from '../composables/useI18n'
import CarStockTable from '../components/car-stock/CarStockTable.vue'
import CarStockForm from '../components/car-stock/CarStockForm.vue'
import CarStockFilter from '../components/car-stock/CarStockFilter.vue'

const router = useRouter()
const route = useRoute()
const { t } = useEnhancedI18n()

// Add admin logic
const user = ref(null)
const isAdmin = computed(() => {
  if (!user.value) return false
  return user.value.role_id === 1
})
const showEditDialog = ref(false)
const editingCar = ref(null)
const carStockTableRef = ref(null)
const filters = ref({
  basic: '',
  advanced: null,
})
const isLoading = ref(false)
const isProcessing = ref({
  edit: false,
  warehouses: false,
})

// Alert-specific props from URL parameters
const alertType = ref(null)
const alertDays = ref(null)
const alertTitle = ref('')

// Computed property for page title
const pageTitle = computed(() => {
  if (alertTitle.value) {
    return alertTitle.value
  }
  return t('carsStockManagementTitle')
})

onMounted(() => {
  // Get user from localStorage
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }

  // Check for alert parameters in URL
  const urlParams = new URLSearchParams(window.location.search)
  alertType.value = urlParams.get('alertType')
  alertDays.value = urlParams.get('alertDays') ? parseInt(urlParams.get('alertDays')) : null
  alertTitle.value = urlParams.get('title') || ''

  // Update page title
  document.title = pageTitle.value
})

// Watch for changes in page title and update document title
watch(pageTitle, (newTitle) => {
  document.title = newTitle
})

const handleEditCar = async (car) => {
  if (isProcessing.value.edit) return
  isProcessing.value.edit = true
  try {
    editingCar.value = { ...car }
    showEditDialog.value = true
  } finally {
    isProcessing.value.edit = false
  }
}

const handleSave = async () => {
  showEditDialog.value = false

  // Refresh the table
  if (carStockTableRef.value) {
    await carStockTableRef.value.fetchCarsStock()
  }
}

const handleFilter = async (filterData) => {
  filters.value = filterData

  // The watcher in CarStockTable will automatically call fetchCarsStock when filters change
  // Removed manual call to prevent double execution
}

const navigateToWarehouses = async () => {
  if (isProcessing.value.warehouses) return
  isProcessing.value.warehouses = true
  try {
    await router.push('/warehouses')
  } finally {
    isProcessing.value.warehouses = false
  }
}

const returnToCarsView = () => {
  router.push('/cars')
}
</script>

<template>
  <div class="cars-stock" :class="{ 'is-loading': isLoading }">
    <div class="header">
      <div class="header-left">
        <button @click="returnToCarsView" class="return-btn">
          <i class="fas fa-arrow-left"></i>
          {{ t('returnToCarsView') }}
        </button>
      </div>
      <h2>
        <i class="fas fa-warehouse"></i>
        {{ alertTitle || t('carsStockManagementTitle') }}
      </h2>
      <div class="header-right">
        <!-- Space for future buttons if needed -->
      </div>
    </div>

    <div class="content">
      <!-- Add the filter component -->
      <CarStockFilter @filter="handleFilter" ref="filterRef" :is-admin="isAdmin" />

      <CarStockTable
        ref="carStockTableRef"
        :onEdit="handleEditCar"
        :filters="filters"
        :alertType="alertType"
        :alertDays="alertDays"
      />

      <!-- Add Warehouses Button at the bottom -->
      <div class="warehouses-button-container">
        <button
          @click="navigateToWarehouses"
          class="warehouses-btn"
          :disabled="isProcessing.warehouses"
          :class="{ processing: isProcessing.warehouses }"
        >
          <i class="fas fa-building"></i>
          <span>{{ t('manageWarehousesButtonText') }}</span>
          <i v-if="isProcessing.warehouses" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <!-- Edit Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay" @click.self="showEditDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-edit"></i>
            {{ t('editCarDetailsTitle') }}
          </h3>
          <button class="close-btn" @click="showEditDialog = false" :disabled="isProcessing.edit">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <CarStockForm :carData="editingCar" @save="handleSave" @cancel="showEditDialog = false" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.cars-stock {
  padding: 20px;
  position: relative;
}

.cars-stock.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.return-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #6c757d;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background-color 0.2s;
}

.return-btn:hover {
  background-color: #5a6268;
}

.return-btn i {
  font-size: 0.8rem;
}

.header h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
}

.header h2 i {
  color: #4caf50;
}

.content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
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
  backdrop-filter: blur(4px);
}

.dialog {
  background-color: white;
  border-radius: 8px;
  padding: 0;
  width: 90%;
  max-width: 1200px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
}

.dialog-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.close-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Warehouses button styles */
.warehouses-button-container {
  margin-top: 20px;
  text-align: center;
}

.warehouses-btn {
  background-color: #4caf50;
  color: white;
  padding: 12px 24px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.warehouses-btn i:not(.loading-indicator) {
  font-size: 1.1em;
}

.warehouses-btn:hover:not(:disabled) {
  background-color: #45a049;
  transform: translateY(-1px);
}

.warehouses-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.warehouses-btn.processing {
  position: relative;
}

.warehouses-btn.processing::after {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.1);
  border-radius: inherit;
}

.loading-indicator {
  margin-left: 4px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.fa-spin {
  animation: spin 1s linear infinite;
}
</style>
