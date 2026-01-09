<script setup>
import { ref } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'
import AddItemDialog from './AddItemDialog.vue'

const { t } = useEnhancedI18n()
const { callApi } = useApi()

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['close', 'saved'])

const loading = ref(false)
const error = ref(null)
const newColor = ref({
  color: '',
  hexa: '#ffffff',
})

// Normalize hex color to always have # prefix
const normalizeHex = (hex) => {
  if (!hex) return '#ffffff'
  hex = hex.trim()
  if (!hex.startsWith('#')) {
    hex = '#' + hex
  }
  // Ensure it's a valid 6-digit hex
  if (/^#[0-9A-Fa-f]{0,6}$/.test(hex)) {
    // Pad with zeros if needed, or truncate if too long
    const hexDigits = hex.slice(1)
    if (hexDigits.length < 6) {
      return hex + '0'.repeat(6 - hexDigits.length)
    }
    return hex.slice(0, 7).toUpperCase()
  }
  return '#ffffff'
}

const handleColorPickerChange = (event) => {
  newColor.value.hexa = event.target.value.toUpperCase()
}

const handleHexInputChange = (event) => {
  const value = event.target.value
  newColor.value.hexa = normalizeHex(value)
}

const handleClose = () => {
  newColor.value = {
    color: '',
    hexa: '#ffffff',
  }
  error.value = null
  emit('close')
}

const handleSave = async () => {
  if (!newColor.value.color || newColor.value.color.trim() === '') {
    error.value = t('carStockForm.colorNameRequired') || 'Color name is required'
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        INSERT INTO colors (color, hexa)
        VALUES (?, ?)
      `,
      params: [
        newColor.value.color.trim(),
        newColor.value.hexa && newColor.value.hexa.trim() !== '' ? newColor.value.hexa.trim().toUpperCase() : null,
      ],
    })

    if (result.success) {
      const savedHexa = newColor.value.hexa && newColor.value.hexa.trim() !== '' 
        ? newColor.value.hexa.trim().toUpperCase() 
        : null
      emit('saved', {
        id: result.lastInsertId,
        color: newColor.value.color.trim(),
        hexa: savedHexa,
      })
      handleClose()
    } else {
      error.value = result.error || t('carStockForm.failedToAddColor') || 'Failed to add color'
    }
  } catch (err) {
    error.value = err.message || t('carStockForm.failedToAddColor') || 'Failed to add color'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AddItemDialog
    :show="show"
    :title="t('carStockForm.addColor') || 'Add New Color'"
    :loading="loading"
    @close="handleClose"
    @save="handleSave"
  >
    <div v-if="error" class="error-message">{{ error }}</div>
    <div class="form-group">
      <label for="color_name"
        >{{ t('carStockForm.color') || 'Color' }} <span class="required">*</span>:</label
      >
      <input
        id="color_name"
        v-model="newColor.color"
        type="text"
        :placeholder="t('carStockForm.colorNamePlaceholder') || 'Color Name'"
        required
      />
    </div>
    <div class="form-group">
      <label for="color_hexa">{{ t('carStockForm.hexColor') || 'Hex Color Code' }}:</label>
      <div class="color-picker-container">
        <input
          type="color"
          :value="newColor.hexa"
          @input="handleColorPickerChange"
          class="color-picker"
          :title="t('carStockForm.pickColor') || 'Pick a color'"
        />
        <input
          id="color_hexa"
          :value="newColor.hexa"
          @input="handleHexInputChange"
          type="text"
          :placeholder="t('carStockForm.hexColorPlaceholder') || '#ffffff'"
          pattern="^#[0-9A-Fa-f]{6}$"
          maxlength="7"
          class="hex-input"
        />
      </div>
    </div>
  </AddItemDialog>
</template>

<style scoped>
.error-message {
  color: #ef4444;
  margin-bottom: 16px;
  padding: 8px;
  background-color: #fee2e2;
  border-radius: 4px;
  font-size: 14px;
}

.color-picker-container {
  display: flex;
  gap: 12px;
  align-items: center;
}

.color-picker {
  width: 60px;
  height: 60px;
  border: 2px solid #d1d5db;
  border-radius: 8px;
  cursor: pointer;
  padding: 0;
  background: none;
  transition: border-color 0.2s;
  flex-shrink: 0;
}

.color-picker:hover {
  border-color: #10b981;
}

.color-picker::-webkit-color-swatch-wrapper {
  padding: 0;
  border-radius: 6px;
  overflow: hidden;
}

.color-picker::-webkit-color-swatch {
  border: none;
  border-radius: 6px;
}

.hex-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  font-family: 'Courier New', monospace;
  text-transform: uppercase;
}

.hex-input:focus {
  outline: none;
  border-color: #10b981;
  box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
}
</style>
