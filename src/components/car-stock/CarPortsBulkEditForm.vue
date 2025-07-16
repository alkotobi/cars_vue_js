<script setup>
import { ref, defineProps, defineEmits, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  selectedCars: {
    type: Array,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])
const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)

// Data for dropdowns
const loadingPorts = ref([])
const dischargePorts = ref([])

// Selected values
const selectedLoadingPort = ref('')
const selectedDischargePort = ref('')

const isProcessing = ref(false)
const isFetchingPorts = ref(false)

// Computed properties
const selectedCarsCount = computed(() => props.selectedCars.length)

// Fetch ports data
const fetchPorts = async () => {
  isFetchingPorts.value = true
  try {
    // Fetch loading ports
    const loadingPortsResult = await callApi({
      query: 'SELECT id, loading_port FROM loading_ports ORDER BY loading_port',
      params: [],
    })
    if (loadingPortsResult.success) {
      loadingPorts.value = loadingPortsResult.data
    }

    // Fetch discharge ports
    const dischargePortsResult = await callApi({
      query: 'SELECT id, discharge_port FROM discharge_ports ORDER BY discharge_port',
      params: [],
    })
    if (dischargePortsResult.success) {
      dischargePorts.value = dischargePortsResult.data
    }
  } catch (err) {
    error.value = t('carPortsBulkEditForm.failedToLoadPortsData')
    console.error('Error fetching ports:', err)
  } finally {
    isFetchingPorts.value = false
  }
}

const handleSubmit = async () => {
  if (isProcessing.value || loading.value) return
  if (!selectedLoadingPort.value && !selectedDischargePort.value) {
    error.value = t('carPortsBulkEditForm.pleaseSelectAtLeastOnePort')
    return
  }

  loading.value = true
  isProcessing.value = true
  error.value = null

  try {
    const updates = []
    const params = []

    if (selectedLoadingPort.value) {
      updates.push('id_port_loading = ?')
      params.push(selectedLoadingPort.value)
    }

    if (selectedDischargePort.value) {
      updates.push('id_port_discharge = ?')
      params.push(selectedDischargePort.value)
    }

    // Add car IDs to params
    const carIds = props.selectedCars.map((car) => car.id)
    const placeholders = carIds.map(() => '?').join(',')
    params.push(...carIds)

    const result = await callApi({
      query: `UPDATE cars_stock SET ${updates.join(', ')} WHERE id IN (${placeholders})`,
      params,
    })

    if (result.success) {
      // Find the port names for the response
      const updatedCars = props.selectedCars.map((car) => {
        const updatedCar = { ...car }
        if (selectedLoadingPort.value) {
          const loadingPort = loadingPorts.value.find(
            (p) => p.id === parseInt(selectedLoadingPort.value),
          )
          updatedCar.loading_port = loadingPort?.loading_port
          updatedCar.id_port_loading = parseInt(selectedLoadingPort.value)
        }
        if (selectedDischargePort.value) {
          const dischargePort = dischargePorts.value.find(
            (p) => p.id === parseInt(selectedDischargePort.value),
          )
          updatedCar.discharge_port = dischargePort?.discharge_port
          updatedCar.id_port_discharge = parseInt(selectedDischargePort.value)
        }
        return updatedCar
      })

      emit('save', updatedCars)
      emit('close')
    } else {
      throw new Error(result.error || t('carPortsBulkEditForm.failedToUpdatePorts'))
    }
  } catch (err) {
    error.value = err.message || t('carPortsBulkEditForm.anErrorOccurred')
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

const handleRevertPorts = async () => {
  if (isProcessing.value || loading.value) return

  // Add confirmation dialog
  const confirmed = confirm(
    t('carPortsBulkEditForm.confirmRevertPorts', { count: selectedCarsCount.value }),
  )
  if (!confirmed) return

  loading.value = true
  isProcessing.value = true
  error.value = null

  try {
    // Add car IDs to params
    const carIds = props.selectedCars.map((car) => car.id)
    const placeholders = carIds.map(() => '?').join(',')

    const result = await callApi({
      query: `UPDATE cars_stock SET id_port_loading = NULL, id_port_discharge = NULL WHERE id IN (${placeholders})`,
      params: carIds,
    })

    if (result.success) {
      // Update cars with null ports
      const updatedCars = props.selectedCars.map((car) => ({
        ...car,
        loading_port: null,
        discharge_port: null,
        id_port_loading: null,
        id_port_discharge: null,
      }))

      emit('save', updatedCars)
      emit('close')
    } else {
      throw new Error(result.error || t('carPortsBulkEditForm.failedToRevertPorts'))
    }
  } catch (err) {
    error.value = err.message || t('carPortsBulkEditForm.anErrorOccurred')
  } finally {
    loading.value = false
    isProcessing.value = false
  }
}

const closeModal = () => {
  error.value = null
  selectedLoadingPort.value = ''
  selectedDischargePort.value = ''
  emit('close')
}

// Fetch ports data when component is mounted
onMounted(fetchPorts)
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content" :class="{ 'is-processing': isProcessing }">
      <div class="modal-header">
        <h3>
          <i class="fas fa-anchor"></i>
          {{ t('carPortsBulkEditForm.editPortsForCars', { count: selectedCarsCount }) }}
        </h3>
        <button class="close-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div class="info-section">
          <p>
            <strong>{{ t('carPortsBulkEditForm.selectedCars') }}:</strong> {{ selectedCarsCount }}
          </p>
          <p>
            <strong>{{ t('carPortsBulkEditForm.action') }}:</strong>
            {{ t('carPortsBulkEditForm.updatePortsForAllSelectedCars') }}
          </p>
        </div>

        <div v-if="isFetchingPorts" class="loading-state">
          <i class="fas fa-spinner fa-spin"></i>
          {{ t('carPortsBulkEditForm.loadingPortsData') }}
        </div>

        <template v-else>
          <div class="form-group">
            <label for="loading-port">
              <i class="fas fa-ship"></i>
              {{ t('carPortsBulkEditForm.loadingPorts') }}:
            </label>
            <div class="select-wrapper">
              <select
                id="loading-port"
                v-model="selectedLoadingPort"
                class="select-field"
                :disabled="isProcessing"
              >
                <option value="">{{ t('carPortsBulkEditForm.selectLoadingPort') }}</option>
                <option v-for="port in loadingPorts" :key="port.id" :value="port.id">
                  {{ port.loading_port }}
                </option>
              </select>
              <i class="fas fa-chevron-down select-arrow"></i>
            </div>
          </div>

          <div class="form-group">
            <label for="discharge-port">
              <i class="fas fa-anchor"></i>
              {{ t('carPortsBulkEditForm.dischargePorts') }}:
            </label>
            <div class="select-wrapper">
              <select
                id="discharge-port"
                v-model="selectedDischargePort"
                class="select-field"
                :disabled="isProcessing"
              >
                <option value="">{{ t('carPortsBulkEditForm.selectDischargePort') }}</option>
                <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
                  {{ port.discharge_port }}
                </option>
              </select>
              <i class="fas fa-chevron-down select-arrow"></i>
            </div>
          </div>

          <div v-if="error" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ error }}
          </div>
        </template>
      </div>

      <div class="modal-footer">
        <button
          v-if="isAdmin"
          class="revert-btn"
          @click="handleRevertPorts"
          :disabled="isProcessing || isFetchingPorts"
        >
          <i class="fas fa-undo"></i>
          <span>{{
            isProcessing
              ? t('carPortsBulkEditForm.reverting')
              : t('carPortsBulkEditForm.revertToNull')
          }}</span>
          <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
        <div class="modal-footer-spacer"></div>
        <button class="cancel-btn" @click="closeModal" :disabled="isProcessing">
          <i class="fas fa-times"></i>
          {{ t('carPortsBulkEditForm.cancel') }}
        </button>
        <button
          class="save-btn"
          @click="handleSubmit"
          :disabled="isProcessing || isFetchingPorts"
          :class="{ 'is-processing': isProcessing }"
        >
          <i class="fas fa-save"></i>
          <span>{{
            isProcessing ? t('carPortsBulkEditForm.saving') : t('carPortsBulkEditForm.saveChanges')
          }}</span>
          <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
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
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
}

.info-section {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 6px;
  border-left: 4px solid #10b981;
}

.info-section p {
  margin: 5px 0;
  color: #555;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
  text-align: center;
}

.select-wrapper {
  position: relative;
}

.select-arrow {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
  pointer-events: none;
}

.select-field {
  width: 100%;
  padding: 8px 12px;
  padding-right: 32px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  appearance: none;
  background-color: white;
  transition: all 0.2s ease;
}

.select-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.select-field:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ef4444;
  margin: 16px 0;
  padding: 12px;
  background-color: #fee2e2;
  border-radius: 4px;
  font-size: 14px;
}

.error-message i {
  font-size: 1.1em;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.modal-footer-spacer {
  flex-grow: 1;
}

.cancel-btn,
.save-btn,
.revert-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  border: none;
  transition: all 0.2s ease;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
}

.revert-btn {
  background-color: #dc3545;
  color: white;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.revert-btn:hover:not(:disabled) {
  background-color: #c82333;
}

.save-btn.is-processing {
  position: relative;
}

.save-btn.is-processing::after {
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
