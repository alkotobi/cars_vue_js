<template>
  <div v-if="modelValue" class="dialog-overlay" @click.self="handleClose">
    <div class="dialog print-options-dialog">
      <div class="dialog-header">
        <h3>
          <i class="fas fa-print"></i>
          Print options
        </h3>
        <button type="button" class="close-btn" @click="handleClose" :disabled="isConfirming">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="dialog-content">
        <p class="print-options-intro">
          Choose which columns and sections to include when printing this loading record.
        </p>

        <fieldset class="print-options-group">
          <legend>Car table columns</legend>
          <label class="print-option-label">
            <input type="checkbox" v-model="localOptions.carId" />
            Car ID
          </label>
          <label class="print-option-label">
            <input type="checkbox" v-model="localOptions.carName" />
            Car Name
          </label>
          <label class="print-option-label">
            <input type="checkbox" v-model="localOptions.color" />
            Color
          </label>
          <label class="print-option-label">
            <input type="checkbox" v-model="localOptions.vin" />
            VIN
          </label>
          <label class="print-option-label">
            <input type="checkbox" v-model="localOptions.paymentStatus" />
            Payment Status
          </label>
          <label class="print-option-label">
            <input type="checkbox" v-model="localOptions.client" />
            Client
          </label>
        </fieldset>

        <fieldset class="print-options-group">
          <legend>Sections</legend>
          <label class="print-option-label">
            <input type="checkbox" v-model="localOptions.loadingInfo" />
            Loading info (header details)
          </label>
          <label class="print-option-label">
            <input type="checkbox" v-model="localOptions.summary" />
            Summary (containers / cars / status)
          </label>
        </fieldset>
      </div>

      <div class="dialog-footer">
        <button type="button" class="cancel-btn" @click="handleClose" :disabled="isConfirming">
          Cancel
        </button>
        <button
          type="button"
          class="confirm-btn"
          @click="handleConfirm"
          :disabled="isConfirming || !hasAnyOption"
        >
          <i class="fas fa-print"></i>
          Print
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { getDefaultPrintOptions } from '@/lib/loadingPrintOptions.js'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  initialOptions: {
    type: Object,
    default: () => getDefaultPrintOptions(),
  },
})

const emit = defineEmits(['update:modelValue', 'close', 'confirm'])

const localOptions = ref({ ...getDefaultPrintOptions(), ...props.initialOptions })
const isConfirming = ref(false)

const hasAnyOption = computed(() => {
  const o = localOptions.value
  return (
    o.carId ||
    o.carName ||
    o.color ||
    o.vin ||
    o.paymentStatus ||
    o.client ||
    o.loadingInfo ||
    o.summary
  )
})

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      localOptions.value = { ...getDefaultPrintOptions(), ...props.initialOptions }
    }
  },
)

watch(
  () => props.initialOptions,
  (opts) => {
    if (props.modelValue) {
      localOptions.value = { ...getDefaultPrintOptions(), ...(opts || {}) }
    }
  },
  { deep: true },
)

const handleClose = () => {
  emit('update:modelValue', false)
  emit('close')
}

const handleConfirm = () => {
  if (!hasAnyOption.value) return
  isConfirming.value = true
  try {
    const options = { ...localOptions.value }
    emit('confirm', options)
  } finally {
    isConfirming.value = false
  }
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.dialog {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 420px;
  max-height: 90vh;
  overflow-y: auto;
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
  font-size: 1.1rem;
}

.close-btn {
  background: none;
  border: none;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  color: #6b7280;
  font-size: 1.1rem;
}

.close-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #374151;
}

.dialog-content {
  padding: 24px;
}

.print-options-intro {
  margin: 0 0 1rem 0;
  color: #555;
  font-size: 0.95rem;
}

.print-options-group {
  margin-bottom: 1.25rem;
  padding: 0.75rem 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: #f9fafb;
}

.print-options-group legend {
  font-weight: 600;
  color: #374151;
  padding: 0 0.25rem;
}

.print-option-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.5rem 0;
  cursor: pointer;
  font-size: 0.9rem;
}

.print-option-label input {
  margin: 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 24px;
  border-top: 1px solid #e5e7eb;
  background: #f8fafc;
  border-bottom-left-radius: 8px;
  border-bottom-right-radius: 8px;
}
</style>

