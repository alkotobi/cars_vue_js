<script setup>
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
  title: {
    type: String,
    required: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  maxWidth: {
    type: String,
    default: '500px',
  },
  minWidth: {
    type: String,
    default: null,
  },
})

const emit = defineEmits(['close', 'save'])

const handleClose = () => {
  emit('close')
}

const handleSave = () => {
  emit('save')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="dialog-overlay" @click.self="handleClose">
      <div class="dialog" :style="{ maxWidth: props.maxWidth, minWidth: props.minWidth }">
        <div class="dialog-header">
          <h3>{{ title }}</h3>
          <button @click="handleClose" class="dialog-close-btn" type="button">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="dialog-body">
          <slot></slot>
        </div>
        <div class="dialog-actions">
          <button @click="handleSave" class="btn save-btn" :disabled="loading">
            {{ loading ? t('carStockForm.saving') || 'Saving...' : t('carStockForm.save') || 'Save' }}
          </button>
          <button @click="handleClose" class="btn cancel-btn" :disabled="loading">
            {{ t('carStockForm.cancel') || 'Cancel' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 3000;
}

.dialog {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.dialog-close-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.dialog-close-btn:hover {
  background-color: #f3f4f6;
  color: #1f2937;
}

.dialog-body {
  padding: 20px;
  flex: 1;
  overflow-y: auto;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #059669;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #4b5563;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 500;
  color: #374151;
  font-size: 14px;
}

.form-group label .required {
  color: #ef4444;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  font-family: inherit;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}

.form-group textarea {
  resize: vertical;
  min-height: 80px;
}
</style>
