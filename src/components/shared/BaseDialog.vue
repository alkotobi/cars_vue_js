<template>
  <teleport to="body">
    <div class="dialog-overlay" :style="{ zIndex }">
      <div class="dialog" :class="size" :style="dialogStyle">
        <div class="dialog-header">
          <h3 :id="titleId">{{ title }}</h3>
          <button v-if="closeable" class="close-btn" @click="$emit('close')" :aria-label="'Close ' + title">×</button>
        </div>
        <slot />
        <div v-if="$slots.footer" class="dialog-footer">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </teleport>
</template>
<script setup>
import { computed } from 'vue'
const props = defineProps({ title: { type: String, default: '' }, titleId: { type: String, default: 'dialog-title' }, size: { type: String, default: 'md' }, zIndex: { type: Number, default: 1000 }, closeable: { type: Boolean, default: true }, maxWidth: { type: String, default: null } })
defineEmits(['close'])
const dialogStyle = computed(() => (props.maxWidth ? { maxWidth: props.maxWidth } : {}))
</script>
<style scoped>
.dialog-overlay{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,.5);display:flex;align-items:center;justify-content:center;padding:20px}
.dialog{background:#fff;border-radius:8px;padding:24px;width:100%;max-height:90vh;overflow-y:auto;box-shadow:0 4px 6px rgba(0,0,0,.1)}
.dialog.sm{max-width:480px}
.dialog.md{max-width:600px}
.dialog.lg{max-width:900px}
.dialog-header{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px}
.dialog-header h3{margin:0}
.close-btn{background:transparent;border:none;font-size:24px;cursor:pointer;color:#6b7280;padding:0}
.close-btn:hover{color:#374151}
.close-btn:disabled{opacity:.6;cursor:not-allowed}
.dialog-footer{margin-top:24px;display:flex;justify-content:flex-end;gap:12px}
</style>
