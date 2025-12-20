<script setup>
import { ref, computed } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useApi } from '../../composables/useApi'

const { t } = useEnhancedI18n()
const { callApi } = useApi()

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  selectedCarIds: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['close', 'saved'])

const name = ref('')
const description = ref('')
const deadline = ref('')
const priority = ref('medium')
const isSaving = ref(false)
const error = ref(null)

const isValid = computed(() => {
  return name.value.trim().length > 0 && props.selectedCarIds.length > 0
})

const handleSave = async () => {
  if (!isValid.value) return

  isSaving.value = true
  error.value = null

  try {
    // Get current user from localStorage
    const userStr = localStorage.getItem('user')
    const user = userStr ? JSON.parse(userStr) : null

    if (!user || !user.id) {
      throw new Error('User not found')
    }

    const result = await callApi({
      query: `
        INSERT INTO car_selections (
          name,
          description,
          user_create_selection,
          selection_data,
          owned_by,
          status,
          priority,
          deadline,
          status_changed_at
        ) VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, NOW())
      `,
      params: [
        name.value.trim(),
        description.value.trim() || null,
        user.id,
        JSON.stringify(props.selectedCarIds),
        JSON.stringify([user.id]), // Creator is added to owned_by
        priority.value,
        deadline.value.trim() || null,
      ],
    })

    if (result.success) {
      emit('saved', {
        id: result.insertId,
        name: name.value.trim(),
        description: description.value.trim(),
      })
      // Reset form
      name.value = ''
      description.value = ''
      deadline.value = ''
      priority.value = 'medium'
      emit('close')
    } else {
      error.value = result.error || t('carStock.failed_to_save_selection')
    }
  } catch (err) {
    error.value = err.message || t('carStock.error_saving_selection')
  } finally {
    isSaving.value = false
  }
}

const handleClose = () => {
  name.value = ''
  description.value = ''
  deadline.value = ''
  priority.value = 'medium'
  error.value = null
  emit('close')
}
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="handleClose">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>{{ t('carStock.save_selection') }}</h3>
        <button @click="handleClose" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div class="form-group">
          <label for="selection-name">
            {{ t('carStock.selection_name') }}
            <span class="required">*</span>
          </label>
          <input
            id="selection-name"
            v-model="name"
            type="text"
            :placeholder="t('carStock.enter_selection_name')"
            class="form-input"
            :disabled="isSaving"
          />
        </div>

        <div class="form-group">
          <label for="selection-description">
            {{ t('carStock.description') }}
          </label>
          <textarea
            id="selection-description"
            v-model="description"
            :placeholder="t('carStock.enter_description')"
            class="form-textarea"
            rows="4"
            :disabled="isSaving"
          ></textarea>
        </div>

        <div class="form-group">
          <label for="selection-deadline">
            {{ t('carStock.deadline') || 'Deadline' }}
          </label>
          <input
            id="selection-deadline"
            v-model="deadline"
            type="datetime-local"
            class="form-input"
            :disabled="isSaving"
          />
        </div>

        <div class="form-group">
          <label for="selection-priority">
            {{ t('carStock.priority') || 'Priority' }}
          </label>
          <select
            id="selection-priority"
            v-model="priority"
            class="form-input"
            :disabled="isSaving"
          >
            <option value="low">{{ t('carStock.priority_low') || 'Low' }}</option>
            <option value="medium">{{ t('carStock.priority_medium') || 'Medium' }}</option>
            <option value="high">{{ t('carStock.priority_high') || 'High' }}</option>
            <option value="urgent">{{ t('carStock.priority_urgent') || 'Urgent' }}</option>
          </select>
        </div>

        <div class="selection-info">
          <i class="fas fa-car"></i>
          <span>
            {{ selectedCarIds.length }}
            {{ selectedCarIds.length === 1 ? t('carStockToolbar.car') : t('carStockToolbar.cars') }}
            {{ t('carStock.selected') }}
          </span>
        </div>
      </div>

      <div class="modal-footer">
        <button @click="handleClose" class="btn btn-cancel" :disabled="isSaving">
          {{ t('carStock.cancel') }}
        </button>
        <button
          @click="handleSave"
          class="btn btn-save"
          :disabled="!isValid || isSaving"
        >
          <i v-if="isSaving" class="fas fa-spinner fa-spin"></i>
          <i v-else class="fas fa-save"></i>
          {{ t('carStock.save') }}
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
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.modal-body {
  padding: 24px;
  flex: 1;
}

.error-message {
  background-color: #fef2f2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.required {
  color: #ef4444;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s ease;
  font-family: inherit;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:disabled,
.form-textarea:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.selection-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  background-color: #f0f9ff;
  border-radius: 6px;
  color: #0c4a6e;
  font-size: 14px;
  font-weight: 500;
}

.selection-info i {
  color: #3b82f6;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
}

.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.btn-save {
  background-color: #3b82f6;
  color: white;
}

.btn-save:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>

