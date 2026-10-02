<template>
  <BaseDialog title="Brands Management" @close="$emit('close')">
    <div class="modal-actions">
      <button class="add-btn" @click="$emit('add')">Add Brand</button>
    </div>
    <BrandsTable :brands="brands" :loading="loading" :error="error" :is-admin="isAdmin" :deleting="deleting" @select="onSelect" @edit="onEdit" @delete="onDelete" @retry="$emit('retry')" />
    <template #footer>
      <button class="btn cancel-btn" @click="$emit('close')">Close</button>
    </template>
  </BaseDialog>
</template>
<script setup>
import BaseDialog from '../shared/BaseDialog.vue'
import BrandsTable from './BrandsTable.vue'
const props = defineProps({ brands: { type: Array, default: () => [] }, loading: { type: Boolean, default: false }, error: { type: String, default: '' }, isAdmin: { type: Boolean, default: false }, deleting: { type: Boolean, default: false } })
const emit = defineEmits(['close', 'add', 'edit', 'delete', 'select', 'retry'])
function onSelect(b){ emit('select', b); emit('close') }
function onEdit(b){ emit('edit', b) }
function onDelete(b){ emit('delete', b) }
</script>
<style scoped>
.modal-actions{margin-bottom:12px}
.add-btn{padding:8px 16px;border:none;border-radius:4px;cursor:pointer;background-color:#10b981;color:#fff}
.add-btn:hover{background-color:#059669}
.btn{padding:6px 12px;border:none;border-radius:4px;margin-right:8px;cursor:pointer}
.cancel-btn{background:#6b7280;color:#fff}
</style>
