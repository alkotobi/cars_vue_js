<script setup>
import { ref, defineProps, defineEmits } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  car: {
    type: Object,
    required: true
  },
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['close', 'save'])
const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const success = ref(false)

const handleLoad = async () => {
  if (props.car.date_loding) return

  loading.value = true
  error.value = null

  try {
    const currentDate = new Date().toISOString().split('T')[0]
    const result = await callApi({
      query: 'UPDATE cars_stock SET date_loding = ? WHERE id = ?',
      params: [currentDate, props.car.id]
    })

    if (result.success) {
      const updatedCar = { 
        ...props.car,
        date_loding: currentDate
      }
      Object.assign(props.car, updatedCar)
      success.value = true
      setTimeout(() => {
        emit('save', updatedCar)
        emit('close')
      }, 2000) // Close after animation plays
    } else {
      throw new Error(result.error || 'Failed to update loading date')
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = false
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-content">
      <div class="modal-header">
        <h3>Car Loading</h3>
        <button class="close-btn" @click="closeModal" v-if="!success">&times;</button>
      </div>

      <div class="modal-body">
        <div v-if="!success" class="loading-status">
          <h4>Loading Status</h4>
          <div class="status-info">
            <span class="status-label">Current Status:</span>
            <span :class="['status-value', props.car.date_loding ? 'loaded' : 'pending']">
              {{ props.car.date_loding ? 'Loaded' : 'Pending' }}
            </span>
          </div>
          <div v-if="props.car.date_loding" class="date-info">
            <span class="date-label">Loading Date:</span>
            <span class="date-value">{{ new Date(props.car.date_loding).toLocaleDateString() }}</span>
          </div>
          <button 
            class="action-btn load-btn" 
            @click="handleLoad"
            :disabled="loading || !!props.car.date_loding"
            :class="{ 'disabled': !!props.car.date_loding }"
          >
            {{ loading ? 'Processing...' : 'Load Car' }}
          </button>
        </div>

        <div v-if="success" class="success-animation">
          <div class="checkmark-circle">
            <div class="checkmark"></div>
          </div>
          <h2 class="congratulations">Congratulations!</h2>
          <p class="success-message">Car has been successfully loaded</p>
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
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

.loading-status {
  background-color: #f9fafb;
  padding: 16px;
  border-radius: 6px;
}

.loading-status h4 {
  margin: 0 0 12px 0;
  color: #374151;
  font-size: 1.1rem;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.status-label, .date-label {
  font-weight: 500;
  color: #6b7280;
}

.status-value {
  font-weight: 600;
}

.loaded {
  color: #059669;
}

.pending {
  color: #dc2626;
}

.date-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.date-value {
  color: #374151;
}

.action-btn {
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  transition: all 0.2s;
}

.load-btn {
  background-color: #3b82f6;
  color: white;
}

.load-btn:hover:not(:disabled) {
  background-color: #2563eb;
  transform: translateY(-1px);
}

.load-btn:active:not(:disabled) {
  transform: translateY(0);
}

.action-btn.disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.error-message {
  color: #ef4444;
  margin-top: 16px;
  font-size: 14px;
}

/* Success Animation Styles */
.success-animation {
  text-align: center;
  padding: 20px;
}

.checkmark-circle {
  width: 100px;
  height: 100px;
  position: relative;
  display: inline-block;
  vertical-align: top;
  margin: 20px;
}

.checkmark {
  border-radius: 50%;
  display: block;
  stroke-width: 2;
  stroke: #4CAF50;
  stroke-miterlimit: 10;
  box-shadow: inset 0px 0px 0px #4CAF50;
  animation: fill .4s ease-in-out .4s forwards, scale .3s ease-in-out .9s both;
  position: relative;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  margin: auto;
  width: 50px;
  height: 50px;
}

.checkmark:before {
  content: '';
  border-radius: 50%;
  display: block;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100%;
  height: 100%;
  background-color: #4CAF50;
  animation: fillCheckmark 0.4s ease-in-out 0.4s forwards;
  opacity: 0;
}

.checkmark:after {
  content: '';
  display: block;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -75%) rotate(45deg);
  width: 25px;
  height: 50px;
  border: solid white;
  border-width: 0 4px 4px 0;
  animation: checkmark 0.2s ease-in-out 0.8s forwards;
  opacity: 0;
}

.congratulations {
  color: #059669;
  font-size: 24px;
  margin: 20px 0 10px;
  animation: slideUp 0.5s ease-out 0.7s both;
}

.success-message {
  color: #374151;
  font-size: 16px;
  margin: 0;
  animation: slideUp 0.5s ease-out 0.9s both;
}

@keyframes fillCheckmark {
  100% {
    opacity: 1;
  }
}

@keyframes checkmark {
  0% {
    opacity: 0;
    transform: translate(-50%, -75%) rotate(45deg);
  }
  100% {
    opacity: 1;
    transform: translate(-50%, -75%) rotate(45deg);
  }
}

@keyframes scale {
  0%, 100% {
    transform: none;
  }
  50% {
    transform: scale3d(1.1, 1.1, 1);
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style> 