<template>
  <div class="vin-assignment-modal" v-if="isVisible">
    <div class="vin-modal-overlay" @click="closeModal"></div>
    <div class="vin-modal-content" @click.stop>
      <div class="vin-modal-header">
        <h3>{{ t('vinAssignmentModal.vin_assignment') }}</h3>
        <button class="vin-close-btn" @click="closeModal">&times;</button>
      </div>

      <div class="vin-modal-body">
        <div class="vin-info-section">
          <p>
            <strong>{{ t('vinAssignmentModal.selected_cars') }}:</strong> {{ selectedCarsCount }}
          </p>
          <p>
            <strong>{{ t('vinAssignmentModal.vins_to_assign') }}:</strong> {{ vinList.length }}
          </p>

          <div v-if="vinList.length !== selectedCarsCount" class="vin-warning">
            <span v-if="vinList.length > selectedCarsCount">
              ⚠️
              {{
                t('vinAssignmentModal.more_vins_than_cars', {
                  vins: vinList.length,
                  cars: selectedCarsCount,
                })
              }}
            </span>
            <span v-else>
              ⚠️
              {{
                t('vinAssignmentModal.fewer_vins_than_cars', {
                  vins: vinList.length,
                  cars: selectedCarsCount,
                })
              }}
            </span>
          </div>
        </div>

        <div class="vin-input-section">
          <label for="vin-input">{{ t('vinAssignmentModal.paste_vins_label') }}</label>
          <textarea
            id="vin-input"
            v-model="vinInput"
            :placeholder="t('vinAssignmentModal.enter_vins_placeholder')"
            rows="8"
            @input="parseVins"
          ></textarea>
        </div>

        <div class="vin-preview" v-if="vinList.length > 0">
          <h4>{{ t('vinAssignmentModal.vins_to_assign') }}:</h4>
          <div class="vin-list">
            <div
              v-for="(vin, index) in vinList"
              :key="index"
              class="vin-item"
              :class="{ assigned: assignedVins.includes(vin) }"
            >
              {{ index + 1 }}. {{ vin }}
            </div>
          </div>
        </div>

        <div class="selected-cars-preview" v-if="selectedCars.length > 0">
          <h4>{{ t('vinAssignmentModal.selected_cars') }}:</h4>
          <div class="cars-list">
            <div v-for="(car, index) in selectedCars" :key="car.id" class="car-item">
              <span class="car-ref">{{ car.ref }}</span>
              <span class="car-vin" v-if="car.vin">({{ car.vin }})</span>
              <span class="assigned-vin" v-if="assignedVins[index]">
                → {{ assignedVins[index] }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="vin-modal-footer">
        <button
          class="vin-btn vin-btn-danger"
          @click="revertVins"
          :disabled="selectedCarsCount === 0 || isProcessing"
        >
          <span v-if="isProcessing">{{ t('vinAssignmentModal.reverting') }}</span>
          <span v-else>{{ t('vinAssignmentModal.revert_vins_to_null') }}</span>
        </button>
        <div class="vin-modal-footer-spacer"></div>
        <button class="vin-btn vin-btn-secondary" @click="closeModal">
          {{ t('vinAssignmentModal.cancel') }}
        </button>
        <button
          class="vin-btn vin-btn-primary"
          @click="assignVins"
          :disabled="!canAssign || isProcessing"
        >
          <span v-if="isProcessing">{{ t('vinAssignmentModal.assigning') }}</span>
          <span v-else>{{ t('vinAssignmentModal.assign_vins') }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, watch } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useApi } from '@/composables/useApi'

export default {
  name: 'VinAssignmentModal',
  props: {
    isVisible: {
      type: Boolean,
      default: false,
    },
    selectedCars: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['close', 'vins-assigned'],
  setup(props, { emit }) {
    const { t } = useEnhancedI18n()
    const { callApi } = useApi()
    const vinInput = ref('')
    const vinList = ref([])
    const assignedVins = ref([])
    const isProcessing = ref(false)

    const selectedCarsCount = computed(() => props.selectedCars.length)

    const canAssign = computed(() => {
      return (
        vinList.value.length === selectedCarsCount.value &&
        vinList.value.length > 0 &&
        selectedCarsCount.value > 0
      )
    })

    const parseVins = () => {
      if (!vinInput.value.trim()) {
        vinList.value = []
        return
      }

      // Split by multiple separators: space, comma, semicolon, newline
      const vins = vinInput.value
        .split(/[\s,;\n]+/)
        .map((vin) => vin.trim())
        .filter((vin) => vin.length > 0)

      vinList.value = vins
    }

    const assignVins = async () => {
      if (!canAssign.value) return

      isProcessing.value = true

      try {
        const assignments = props.selectedCars.map((car, index) => ({
          carId: car.id,
          vin: vinList.value[index],
        }))

        const response = await callApi({
          action: 'assign_multiple_vins',
          assignments: assignments,
        })

        if (response.success) {
          emit('vins-assigned', assignments)
          closeModal()
        } else {
          alert(
            t('vinAssignmentModal.error_assigning_vins') +
              ': ' +
              (response.message || t('vinAssignmentModal.unknown_error')),
          )
        }
      } catch (error) {
        alert(t('vinAssignmentModal.error_assigning_vins') + ': ' + error.message)
      } finally {
        isProcessing.value = false
      }
    }

    const revertVins = async () => {
      if (selectedCarsCount.value === 0) return

      // Add confirmation dialog
      const confirmed = confirm(
        t('vinAssignmentModal.confirm_revert_vins', { count: selectedCarsCount.value }),
      )
      if (!confirmed) return

      isProcessing.value = true

      try {
        const assignments = props.selectedCars.map((car) => ({
          carId: car.id,
          vin: null,
        }))

        const response = await callApi({
          action: 'assign_multiple_vins',
          assignments: assignments,
        })

        if (response.success) {
          emit('vins-assigned', assignments)
          closeModal()
        } else {
          alert(
            t('vinAssignmentModal.error_reverting_vins') +
              ': ' +
              (response.message || response.error || t('vinAssignmentModal.unknown_error')),
          )
        }
      } catch (error) {
        alert(t('vinAssignmentModal.error_reverting_vins') + ': ' + error.message)
      } finally {
        isProcessing.value = false
      }
    }

    const closeModal = () => {
      vinInput.value = ''
      vinList.value = []
      assignedVins.value = []
      emit('close')
    }

    // Watch for changes in selected cars to update assigned VINs preview
    watch(
      () => props.selectedCars,
      () => {
        assignedVins.value = new Array(props.selectedCars.length).fill('')
      },
      { immediate: true },
    )

    // Watch for changes in VIN list to update assigned VINs preview
    watch(vinList, () => {
      assignedVins.value = vinList.value.slice(0, props.selectedCars.length)
    })

    return {
      t,
      vinInput,
      vinList,
      assignedVins,
      isProcessing,
      selectedCarsCount,
      canAssign,
      parseVins,
      assignVins,
      revertVins,
      closeModal,
    }
  },
}
</script>

<style scoped>
.vin-assignment-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.vin-modal-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
}

.vin-modal-content {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 10000;
}

.vin-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
  background-color: #f8f9fa;
}

.vin-modal-header h3 {
  margin: 0;
  color: #333;
}

.vin-close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #666;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.vin-close-btn:hover {
  color: #333;
}

.vin-modal-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.vin-info-section {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 6px;
  border-left: 4px solid #007bff;
}

.vin-info-section p {
  margin: 5px 0;
  color: #555;
}

.vin-warning {
  margin-top: 10px;
  padding: 10px;
  background-color: #fff3cd;
  border: 1px solid #ffeaa7;
  border-radius: 4px;
  color: #856404;
}

.vin-input-section {
  margin-bottom: 20px;
}

.vin-input-section label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #333;
}

.vin-input-section textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: monospace;
  font-size: 14px;
  resize: vertical;
}

.vin-input-section textarea:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
}

.vin-preview,
.selected-cars-preview {
  margin-bottom: 20px;
}

.vin-preview h4,
.selected-cars-preview h4 {
  margin: 0 0 10px 0;
  color: #333;
  font-size: 16px;
}

.vin-list,
.cars-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  background-color: #f8f9fa;
}

.vin-item,
.car-item {
  padding: 8px 12px;
  border-bottom: 1px solid #e0e0e0;
  font-family: monospace;
  font-size: 14px;
}

.vin-item:last-child,
.car-item:last-child {
  border-bottom: none;
}

.vin-item.assigned {
  background-color: #d4edda;
  color: #155724;
}

.car-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.car-ref {
  font-weight: 500;
  color: #333;
}

.car-vin {
  color: #666;
  font-size: 12px;
}

.assigned-vin {
  color: #007bff;
  font-weight: 500;
}

.vin-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 20px;
  border-top: 1px solid #e0e0e0;
  background-color: #f8f9fa;
}

.vin-modal-footer-spacer {
  flex-grow: 1; /* Push buttons to the right */
}

.vin-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.vin-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.vin-btn-secondary {
  background-color: #6c757d;
  color: white;
}

.vin-btn-secondary:hover:not(:disabled) {
  background-color: #5a6268;
}

.vin-btn-primary {
  background-color: #007bff;
  color: white;
}

.vin-btn-primary:hover:not(:disabled) {
  background-color: #0056b3;
}

.vin-btn-danger {
  background-color: #dc3545;
  color: white;
}

.vin-btn-danger:hover:not(:disabled) {
  background-color: #c82333;
}
</style>
