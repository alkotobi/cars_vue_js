<script setup>
defineProps({
  modelValue: {
    type: [String, Number, null],
    default: null,
  },
  options: {
    type: Array,
    required: true,
  },
  optionValue: {
    type: String,
    default: 'id',
  },
  optionLabel: {
    type: String,
    default: 'name',
  },
  placeholder: {
    type: String,
    default: 'Select...',
  },
  required: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  addButtonTitle: {
    type: String,
    default: 'Add New',
  },
})

const emit = defineEmits(['update:modelValue', 'add'])

const handleChange = (event) => {
  const value = event.target.value
  if (value === '') {
    emit('update:modelValue', null)
  } else {
    // Try to convert to number if it's a numeric string, otherwise keep as string
    const numValue = Number(value)
    emit('update:modelValue', isNaN(numValue) ? value : numValue)
  }
}
</script>

<template>
  <div class="input-with-button">
    <select
      :value="modelValue === null || modelValue === '' ? '' : String(modelValue)"
      @change="handleChange($event)"
      :required="required"
      :disabled="disabled"
    >
      <option value="">{{ placeholder }}</option>
      <option v-for="option in options" :key="option[optionValue]" :value="option[optionValue]">
        {{ option[optionLabel] }}
      </option>
    </select>
    <button type="button" @click="emit('add')" class="btn-add-supplier" :title="addButtonTitle">
      <i class="fas fa-plus"></i>
    </button>
  </div>
</template>

<style scoped>
.input-with-button {
  display: flex;
  gap: 8px;
  align-items: stretch;
}

.input-with-button select {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
  background-color: white;
}

.input-with-button select:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.btn-add-supplier {
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  transition: background-color 0.2s;
}

.btn-add-supplier:hover:not(:disabled) {
  background-color: #059669;
}

.btn-add-supplier:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}
</style>
