<template>
  <div v-if="visible" class="dialog-overlay" @click.self="$emit('cancel')">
    <div class="dialog">
      <div class="dialog-header">
        <h3><i class="fas fa-sticky-note"></i> {{ title || 'Edit Notes' }}</h3>
        <button class="close-btn" @click="$emit('cancel')"><i class="fas fa-times"></i></button>
      </div>
      <div class="dialog-content">
        <textarea
          v-model="localValue"
          rows="6"
          class="notes-textarea"
          placeholder="Enter notes..."
        ></textarea>
      </div>
      <div class="dialog-actions">
        <button class="cancel-btn" @click="$emit('cancel')">Cancel</button>
        <button class="save-btn" @click="saveNotes">Save</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
const props = defineProps({
  modelValue: String,
  visible: Boolean,
  title: String,
})
const emit = defineEmits(['update:modelValue', 'save', 'cancel'])
const localValue = ref(props.modelValue || '')

watch(
  () => props.modelValue,
  (val) => {
    localValue.value = val || ''
  },
)

const saveNotes = () => {
  emit('update:modelValue', localValue.value)
  emit('save', localValue.value)
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.dialog {
  background: #fff;
  border-radius: 8px;
  min-width: 350px;
  max-width: 90vw;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  padding: 0;
  overflow: hidden;
}
.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px 8px 20px;
  border-bottom: 1px solid #eee;
}
.dialog-content {
  padding: 20px;
}
.notes-textarea {
  width: 100%;
  min-height: 120px;
  border-radius: 6px;
  border: 1px solid #ccc;
  padding: 10px;
  font-size: 1em;
  resize: vertical;
}
.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 20px 16px 20px;
  border-top: 1px solid #eee;
}
.save-btn {
  background: #3498db;
  color: #fff;
  border: none;
  border-radius: 4px;
  padding: 8px 18px;
  font-size: 1em;
  cursor: pointer;
}
.cancel-btn {
  background: #eee;
  color: #333;
  border: none;
  border-radius: 4px;
  padding: 8px 18px;
  font-size: 1em;
  cursor: pointer;
}
.close-btn {
  background: none;
  border: none;
  font-size: 1.2em;
  color: #888;
  cursor: pointer;
}
</style>
