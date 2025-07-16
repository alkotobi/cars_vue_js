<script setup>
import { defineProps, defineEmits, ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  car: {
    type: Object,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])

const { callApi } = useApi()

// State
const carsByPurchaseBill = ref([])
const selectedCarId = ref(null)
const loading = ref(false)
const error = ref(null)
const searchTerm = ref('')
const filterOptions = ref({
  hasSellBill: false,
  hasClient: false,
  hasContainer: false,
  availableOnly: false,
})

// Computed
const filteredCars = computed(() => {
  let filtered = carsByPurchaseBill.value

  // Search filter
  if (searchTerm.value.trim()) {
    const term = searchTerm.value.toLowerCase().trim()
    filtered = filtered.filter(
      (car) =>
        car.vin?.toLowerCase().includes(term) ||
        car.car_name?.toLowerCase().includes(term) ||
        car.color?.toLowerCase().includes(term) ||
        car.sell_bill_ref?.toLowerCase().includes(term) ||
        car.client_name?.toLowerCase().includes(term) ||
        car.container_ref?.toLowerCase().includes(term) ||
        car.buy_bill_ref?.toLowerCase().includes(term),
    )
  }

  // Filter options
  if (filterOptions.value.hasSellBill) {
    filtered = filtered.filter((car) => car.sell_bill_ref)
  }

  if (filterOptions.value.hasClient) {
    filtered = filtered.filter((car) => car.client_name)
  }

  if (filterOptions.value.hasContainer) {
    filtered = filtered.filter((car) => car.container_ref)
  }

  if (filterOptions.value.availableOnly) {
    filtered = filtered.filter((car) => !car.sell_bill_ref) // Available = no sell bill
  }

  return filtered
})

const groupedCars = computed(() => {
  const groups = {}

  filteredCars.value.forEach((car) => {
    const purchaseBillKey = car.buy_bill_ref || 'No Purchase Bill'
    if (!groups[purchaseBillKey]) {
      groups[purchaseBillKey] = {
        buy_bill_ref: car.buy_bill_ref,
        buy_bill_id: car.buy_bill_id,
        supplier_name: car.supplier_name,
        date_buy: car.date_buy,
        cars: [],
      }
    }
    groups[purchaseBillKey].cars.push(car)
  })

  return Object.values(groups)
})

// Methods
const handleClose = () => {
  emit('close')
}

const handleSave = () => {
  console.log('Save button clicked, selectedCarId:', selectedCarId.value)

  if (!selectedCarId.value) {
    alert(t('switchBuyBill.selectCarToSwitch'))
    return
  }

  const saveData = {
    success: true,
    carId1: props.car.id,
    carId2: selectedCarId.value,
  }

  console.log('Emitting save event with data:', saveData)
  emit('save', saveData)
}

const fetchCarsByPurchaseBill = async () => {
  loading.value = true
  error.value = null

  try {
    console.log('Fetching cars for car ID:', props.car.id)

    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cn.car_name,
          c.color,
          cs.vin,
          cs.id_buy_details,
          bd.id_buy_bill,
          bb.bill_ref as buy_bill_ref,
          bb.date_buy,
          s.name as supplier_name,
          cs.container_ref,
          sb.bill_ref as sell_bill_ref,
          cl.name as client_name
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors c ON cs.id_color = c.id
        LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
        LEFT JOIN suppliers s ON bb.id_supplier = s.id
        LEFT JOIN sell_bill sb ON cs.id_sell = sb.id
        LEFT JOIN clients cl ON cs.id_client = cl.id
        WHERE cs.id != ?
        AND cn.car_name = (
          SELECT cn2.car_name 
          FROM cars_stock cs2
          LEFT JOIN buy_details bd2 ON cs2.id_buy_details = bd2.id
          LEFT JOIN cars_names cn2 ON bd2.id_car_name = cn2.id
          WHERE cs2.id = ?
        )
        ORDER BY bb.bill_ref, cs.id
      `,
      params: [props.car.id, props.car.id],
    })

    console.log('API Response:', result)

    if (result.success) {
      carsByPurchaseBill.value = result.data
      console.log('Cars loaded:', result.data?.length || 0)
    } else {
      console.error('API Error:', result)
      error.value = result.error || result.message || t('switchBuyBill.fetchError')
    }
  } catch (err) {
    console.error('Error fetching cars:', err)
    error.value = err.message || t('switchBuyBill.serverError')
  } finally {
    loading.value = false
  }
}

const selectCar = (carId) => {
  selectedCarId.value = carId
}

const clearFilters = () => {
  searchTerm.value = ''
  filterOptions.value = {
    hasSellBill: false,
    hasClient: false,
    hasContainer: false,
    availableOnly: false,
  }
}

// Watch for show prop changes
onMounted(() => {
  if (props.show) {
    fetchCarsByPurchaseBill()
  }
})

// Watch for show prop changes
import { watch } from 'vue'
watch(
  () => props.show,
  (newShow) => {
    if (newShow) {
      fetchCarsByPurchaseBill()
      selectedCarId.value = null
    }
  },
)
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="handleClose">
    <div class="modal">
      <div class="modal-header">
        <h3>
          <i class="fas fa-exchange-alt"></i>
          {{ t('switchBuyBill.title') }}
        </h3>
        <button class="close-btn" @click="handleClose">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="car-info">
          <p>
            <strong>{{ t('switchBuyBill.carId') }}:</strong> {{ car?.id }}
          </p>
          <p>
            <strong>{{ t('switchBuyBill.carName') }}:</strong> {{ car?.car_name }}
          </p>
          <p>
            <strong>{{ t('switchBuyBill.currentPurchaseBill') }}:</strong>
            {{ car?.buy_bill_ref || t('switchBuyBill.none') }}
          </p>
          <p v-if="car?.sell_bill_ref">
            <strong>{{ t('switchBuyBill.sellBill') }}:</strong> {{ car.sell_bill_ref }}
          </p>
          <p v-if="car?.client_name">
            <strong>{{ t('switchBuyBill.client') }}:</strong> {{ car.client_name }}
          </p>
          <p v-if="car?.container_ref">
            <strong>{{ t('switchBuyBill.container') }}:</strong> {{ car.container_ref }}
          </p>
          <p class="filter-note">
            <strong>{{ t('switchBuyBill.note') }}:</strong> {{ t('switchBuyBill.sameModelOnly') }}:
            <em>{{ car?.car_name }}</em>
          </p>
        </div>

        <div class="selection-section">
          <h4>{{ t('switchBuyBill.selectCarToSwitch') }}</h4>

          <!-- Search and Filter Controls -->
          <div class="filter-controls">
            <div class="search-box">
              <i class="fas fa-search"></i>
              <input
                v-model="searchTerm"
                type="text"
                :placeholder="t('switchBuyBill.searchPlaceholder')"
                class="search-input"
              />
            </div>

            <div class="filter-options">
              <label class="filter-checkbox">
                <input v-model="filterOptions.hasSellBill" type="checkbox" />
                <span>{{ t('switchBuyBill.hasSellBill') }}</span>
              </label>

              <label class="filter-checkbox">
                <input v-model="filterOptions.hasClient" type="checkbox" />
                <span>{{ t('switchBuyBill.hasClient') }}</span>
              </label>

              <label class="filter-checkbox">
                <input v-model="filterOptions.hasContainer" type="checkbox" />
                <span>{{ t('switchBuyBill.hasContainer') }}</span>
              </label>

              <label class="filter-checkbox">
                <input v-model="filterOptions.availableOnly" type="checkbox" />
                <span>{{ t('switchBuyBill.availableOnly') }}</span>
              </label>

              <button
                v-if="searchTerm || Object.values(filterOptions).some((v) => v)"
                @click="clearFilters"
                class="clear-filters-btn"
              >
                <i class="fas fa-times"></i>
                {{ t('switchBuyBill.clearFilters') }}
              </button>
            </div>
          </div>

          <div v-if="loading" class="loading">
            <i class="fas fa-spinner fa-spin"></i>
            {{ t('switchBuyBill.loadingCars') }}
          </div>

          <div v-else-if="error" class="error">
            <i class="fas fa-exclamation-circle"></i>
            <div class="error-message">
              <strong>{{ t('switchBuyBill.errorLoadingCars') }}:</strong>
              <p>{{ error }}</p>
              <small>{{ t('switchBuyBill.checkConsoleForDetails') }}</small>
            </div>
          </div>

          <div v-else-if="groupedCars.length === 0" class="empty-state">
            <i class="fas fa-car"></i>
            <p v-if="searchTerm || Object.values(filterOptions).some((v) => v)">
              {{ t('switchBuyBill.noCarsMatchCriteria') }}
            </p>
            <p v-else>{{ t('switchBuyBill.noOtherCarsSameModel') }}</p>
          </div>

          <div v-else class="results-info">
            <p>
              {{
                t('switchBuyBill.foundCars', {
                  count: filteredCars.length,
                  bills: groupedCars.length,
                })
              }}
            </p>
          </div>

          <div v-else class="cars-list">
            <div v-for="group in groupedCars" :key="group.buy_bill_ref" class="purchase-bill-group">
              <div class="group-header">
                <h5>
                  <i class="fas fa-file-invoice-dollar"></i>
                  {{ group.buy_bill_ref || t('switchBuyBill.noPurchaseBill') }}
                </h5>
                <div class="group-details">
                  <span v-if="group.supplier_name">
                    <i class="fas fa-building"></i>
                    {{ group.supplier_name }}
                  </span>
                  <span v-if="group.date_buy">
                    <i class="fas fa-calendar"></i>
                    {{ new Date(group.date_buy).toLocaleDateString() }}
                  </span>
                </div>
              </div>

              <div class="cars-grid">
                <div
                  v-for="car in group.cars"
                  :key="car.id"
                  class="car-item"
                  :class="{ selected: selectedCarId === car.id }"
                  @click="selectCar(car.id)"
                >
                  <div class="car-header">
                    <span class="car-id">#{{ car.id }}</span>
                    <span v-if="selectedCarId === car.id" class="selected-indicator">
                      <i class="fas fa-check-circle"></i>
                    </span>
                  </div>
                  <div class="car-details">
                    <div class="car-name">{{ car.car_name }}</div>
                    <div v-if="car.color" class="car-color">
                      <i class="fas fa-palette"></i>
                      {{ car.color }}
                    </div>
                    <div v-if="car.vin" class="car-vin">
                      <i class="fas fa-barcode"></i>
                      {{ car.vin }}
                    </div>
                    <div v-if="car.sell_bill_ref" class="car-sell-bill">
                      <i class="fas fa-file-invoice"></i>
                      <span class="label">{{ t('switchBuyBill.sellBill') }}:</span>
                      {{ car.sell_bill_ref }}
                    </div>
                    <div v-if="car.client_name" class="car-client">
                      <i class="fas fa-user"></i>
                      <span class="label">{{ t('switchBuyBill.client') }}:</span>
                      {{ car.client_name }}
                    </div>
                    <div v-if="car.container_ref" class="car-container">
                      <i class="fas fa-shipping-fast"></i>
                      <span class="label">{{ t('switchBuyBill.container') }}:</span>
                      {{ car.container_ref }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="cancel-btn" @click="handleClose">{{ t('switchBuyBill.cancel') }}</button>
          <button class="save-btn" @click="handleSave" :disabled="!selectedCarId">
            {{ t('switchBuyBill.switchPurchaseBill') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
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

.modal {
  background-color: white;
  border-radius: 8px;
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
}

.modal-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.125rem;
}

.modal-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background-color: #f3f4f6;
  color: #1f2937;
}

.modal-body {
  padding: 24px;
}

.car-info {
  margin-bottom: 24px;
}

.car-info p {
  margin: 8px 0;
  color: #374151;
  font-size: 0.875rem;
}

.car-info strong {
  color: #1f2937;
  font-weight: 600;
}

.filter-note {
  margin-top: 12px;
  padding: 8px 12px;
  background-color: #f0f7ff;
  border: 1px solid #bfdbfe;
  border-radius: 4px;
  font-size: 0.875rem;
  color: #1e40af;
}

.filter-note em {
  font-style: italic;
  font-weight: 600;
}

/* Selection section styles */
.selection-section {
  margin-top: 24px;
}

.selection-section h4 {
  margin: 0 0 16px 0;
  color: #1f2937;
  font-size: 1rem;
  font-weight: 600;
}

/* Filter controls styles */
.filter-controls {
  margin-bottom: 20px;
  padding: 16px;
  background-color: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
}

.search-box {
  position: relative;
  margin-bottom: 12px;
}

.search-box i {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
  z-index: 1;
}

.search-input {
  width: 100%;
  padding: 10px 12px 10px 36px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  background-color: white;
  transition: border-color 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.filter-options {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.filter-checkbox {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.875rem;
  color: #374151;
  cursor: pointer;
  user-select: none;
}

.filter-checkbox input[type='checkbox'] {
  width: 16px;
  height: 16px;
  accent-color: #3b82f6;
}

.clear-filters-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #f3f4f6;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  color: #6b7280;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.clear-filters-btn:hover {
  background-color: #e5e7eb;
  color: #374151;
}

.results-info {
  margin-bottom: 16px;
  padding: 8px 12px;
  background-color: #f0f7ff;
  border: 1px solid #bfdbfe;
  border-radius: 4px;
  font-size: 0.875rem;
  color: #1e40af;
}

.loading,
.error,
.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.loading i {
  font-size: 1.5em;
  color: #3b82f6;
}

.error {
  color: #ef4444;
}

.error i {
  font-size: 1.5em;
}

.error-message {
  text-align: left;
  max-width: 100%;
}

.error-message strong {
  display: block;
  margin-bottom: 8px;
  color: #dc2626;
}

.error-message p {
  margin: 8px 0;
  word-break: break-word;
}

.error-message small {
  color: #9ca3af;
  font-style: italic;
}

.empty-state i {
  font-size: 2em;
  color: #9ca3af;
}

.empty-state p {
  margin: 0;
  font-size: 0.875rem;
}

.cars-list {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
}

.purchase-bill-group {
  border-bottom: 1px solid #e5e7eb;
}

.purchase-bill-group:last-child {
  border-bottom: none;
}

.group-header {
  background-color: #f8fafc;
  padding: 12px 16px;
  border-bottom: 1px solid #e5e7eb;
}

.group-header h5 {
  margin: 0 0 8px 0;
  color: #1f2937;
  font-size: 0.875rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.group-header h5 i {
  color: #3b82f6;
}

.group-details {
  display: flex;
  gap: 16px;
  font-size: 0.75rem;
  color: #6b7280;
}

.group-details span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.group-details i {
  color: #9ca3af;
}

.cars-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 12px;
  padding: 12px;
}

.car-item {
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  background-color: white;
  min-height: 120px;
}

.car-item:hover {
  border-color: #3b82f6;
  background-color: #f0f7ff;
}

.car-item.selected {
  border-color: #3b82f6;
  background-color: #eff6ff;
  box-shadow: 0 0 0 1px #3b82f6;
}

.car-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.car-id {
  font-weight: 600;
  color: #1f2937;
  font-size: 0.875rem;
}

.selected-indicator {
  color: #3b82f6;
  font-size: 1rem;
}

.car-details {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.car-name {
  font-weight: 500;
  color: #1f2937;
  font-size: 0.875rem;
}

.car-color,
.car-vin,
.car-sell-bill,
.car-client,
.car-container {
  font-size: 0.75rem;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 4px;
}

.car-color i,
.car-vin i,
.car-sell-bill i,
.car-client i,
.car-container i {
  color: #9ca3af;
  width: 12px;
}

.car-sell-bill .label,
.car-client .label,
.car-container .label {
  font-weight: 500;
  color: #4b5563;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn {
  padding: 8px 16px;
  background-color: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.cancel-btn:hover {
  background-color: #e5e7eb;
  color: #1f2937;
}

.save-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.save-btn:hover {
  background-color: #2563eb;
}

.save-btn:active {
  transform: translateY(1px);
}
</style>
