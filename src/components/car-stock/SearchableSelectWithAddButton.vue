<script setup>
import { ElSelect, ElOption } from 'element-plus'

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

const handleChange = (value) => {
  emit('update:modelValue', value)
}
</script>

<template>
  <div class="searchable-select-with-button">
    <el-select
      :model-value="modelValue"
      @update:model-value="handleChange"
      :placeholder="placeholder"
      filterable
      :disabled="disabled"
      style="width: 100%"
      :required="required"
    >
      <el-option
        v-for="option in options"
        :key="option[optionValue]"
        :label="option[optionLabel]"
        :value="option[optionValue]"
      />
    </el-select>
    <button type="button" @click="emit('add')" class="btn-add-supplier" :title="addButtonTitle">
      <i class="fas fa-plus"></i>
    </button>
  </div>
</template>

<style scoped>
.searchable-select-with-button {
  display: flex;
  gap: 8px;
  align-items: stretch;
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
  flex-shrink: 0;
}

.btn-add-supplier:hover:not(:disabled) {
  background-color: #059669;
}

.btn-add-supplier:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}
</style>
