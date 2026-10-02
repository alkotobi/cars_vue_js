<template>
  <BaseDialog :title="isEdit ? 'Edit Brand' : 'Add Brand'" @close="$emit('close')">
    <div class="form-group">
      <label for="brand-name">Brand Name</label>
      <input id="brand-name" v-model="form.brand" type="text" class="input-field" placeholder="Enter brand name" />
    </div>
    <div class="form-group">
      <label for="brand-logo">Logo</label>
      <input id="brand-logo" type="file" accept="image/*" @change="onFileChange" class="input-field" />
      <div v-if="previewUrl" class="logo-preview">
        <img :src="previewUrl" alt="Logo preview" class="brand-logo-preview" />
        <button type="button" class="btn delete-btn-small" @click="clearLogo">Remove</button>
      </div>
      <div v-else-if="isEdit && currentLogoUrl" class="logo-preview">
        <img :src="currentLogoUrl" alt="Current logo" class="brand-logo-preview" />
      </div>
      <small class="form-hint">Optional. PNG, JPG, GIF, SVG, WEBP</small>
    </div>
    <template #footer>
      <button class="btn save-btn" :disabled="saving || !isValid" @click="save">{{ saving ? 'Saving...' : 'Save' }}</button>
      <button class="btn cancel-btn" @click="$emit('close')">Cancel</button>
    </template>
  </BaseDialog>
</template>
<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import BaseDialog from '../shared/BaseDialog.vue'
import { useApi } from '../../composables/useApi'
const { getFileUrl } = useApi()
const props = defineProps({ brand: { type: Object, default: null }, saving: { type: Boolean, default: false } })
const emit = defineEmits(['close', 'save'])
const form = ref({ brand: '' })
const logoFile = ref(null)
const previewUrl = ref('')
const isEdit = computed(() => !!props.brand)
const isValid = computed(() => form.value.brand.trim().length > 0)
const currentLogoUrl = computed(() => (props.brand?.logo_path ? getFileUrl(props.brand.logo_path) : ''))
watch(() => props.brand, (b) => {
  form.value.brand = b?.brand || ''
  logoFile.value = null
  if (previewUrl.value) { URL.revokeObjectURL(previewUrl.value); previewUrl.value = '' }
}, { immediate: true })
function onFileChange(e) {
  const f = e.target.files[0]
  if (!f) return
  logoFile.value = f
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  previewUrl.value = URL.createObjectURL(f)
}
function clearLogo() {
  logoFile.value = null
  if (previewUrl.value) { URL.revokeObjectURL(previewUrl.value); previewUrl.value = '' }
}
function save() { if (!isValid.value) return; emit('save', { brand: form.value.brand, logoFile: logoFile.value, id: props.brand?.id, editingBrandLogoFile: logoFile.value }) }
onUnmounted(() => { if (previewUrl.value) URL.revokeObjectURL(previewUrl.value) })
</script>
<style scoped>
.form-group{margin-bottom:16px}
.form-group label{display:block;margin-bottom:4px;font-weight:500}
.input-field{width:100%;padding:8px;border:1px solid #ddd;border-radius:4px}
.logo-preview{margin-top:8px;display:flex;align-items:center;gap:12px}
.brand-logo-preview{width:100px;height:100px;object-fit:contain;border:1px solid #ddd;border-radius:4px;padding:8px}
.btn{padding:6px 12px;border:none;border-radius:4px;margin-right:8px;cursor:pointer}
.save-btn{background:#10b981;color:#fff}
.save-btn:disabled{opacity:.6;cursor:not-allowed}
.cancel-btn{background:#6b7280;color:#fff}
.delete-btn-small{background:rgba(239,68,68,.9);color:#fff;border:none;border-radius:4px;padding:4px 8px;cursor:pointer}
.form-hint{color:#6b7280;font-size:.875rem}
</style>
