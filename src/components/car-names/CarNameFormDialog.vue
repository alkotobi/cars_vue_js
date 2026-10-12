<template>
  <BaseDialog :title="isEdit ? 'Edit Car Name' : 'Add Car Name'" @close="$emit('close')">
    <div class="form-group">
      <label for="car-name">Car Name</label>
      <input
        id="car-name"
        v-model="form.car_name"
        type="text"
        class="input-field"
        placeholder="Enter car name (will be uppercase)"
      />
    </div>
    <div class="form-group">
      <label for="car-brand">Brand</label>
      <select id="car-brand" v-model="form.id_brand" class="input-field">
        <option disabled value="">Select a brand</option>
        <option v-for="b in brands" :key="b.id" :value="b.id">{{ b.brand }}</option>
      </select>
    </div>
    <div class="form-group">
      <label for="car-notes">Notes</label>
      <textarea
        id="car-notes"
        v-model="form.notes"
        class="input-field textarea"
        rows="2"
        placeholder="Notes (optional)"
      ></textarea>
    </div>
    <div class="form-group">
      <label for="car-size">Size <span class="required">*</span></label>
      <select id="car-size" v-model="form.is_big_car" class="input-field">
        <option disabled value="">Select big or small</option>
        <option :value="0">Small car</option>
        <option :value="1">Big car</option>
      </select>
    </div>
    <template #footer>
      <button class="btn save-btn" :disabled="saving || !isValid" @click="save">
        {{ saving ? 'Saving...' : 'Save' }}
      </button>
      <button class="btn cancel-btn" @click="$emit('close')">Cancel</button>
    </template>
  </BaseDialog>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import BaseDialog from '../shared/BaseDialog.vue'
const props = defineProps({
  carName: { type: Object, default: null },
  brands: { type: Array, default: () => [] },
  saving: { type: Boolean, default: false },
})
const emit = defineEmits(['close', 'save'])
const form = ref({ car_name: '', notes: '', is_big_car: '', id_brand: '' })
const isEdit = computed(() => !!props.carName)
const isValid = computed(
  () =>
    form.value.car_name.trim().length > 0 &&
    !!form.value.id_brand &&
    (form.value.is_big_car === 0 || form.value.is_big_car === 1),
)
watch(
  () => props.carName,
  (cn) => {
    form.value = {
      car_name: cn?.car_name || '',
      notes: cn?.notes || '',
      is_big_car: cn?.is_big_car === 0 || cn?.is_big_car === 1 ? (cn.is_big_car ? 1 : 0) : '',
      id_brand: cn?.id_brand || '',
    }
  },
  { immediate: true },
)
function save() {
  if (!isValid.value) return
  emit('save', { id: props.carName?.id, ...form.value })
}
</script>
<style scoped>
.form-group {
  margin-bottom: 16px;
}
.form-group label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
}
.input-field {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}
.textarea {
  resize: vertical;
}
.required {
  color: #ef4444;
}
.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  margin-right: 8px;
  cursor: pointer;
}
.save-btn {
  background: #10b981;
  color: #fff;
}
.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.cancel-btn {
  background: #6b7280;
  color: #fff;
}
</style>
