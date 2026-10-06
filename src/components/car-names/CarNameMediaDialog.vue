<template>
  <BaseDialog :title="`Photos/Videos - ${carName.car_name}`" size="lg" @close="$emit('close')">
    <div class="media-upload-section">
      <input ref="fileInput" type="file" multiple accept="image/*,video/*" class="file-input" @change="onSelect" />
      <button class="btn upload-btn" :disabled="uploading || selected.length===0" @click="doUpload">
        {{ uploading ? 'Uploading...' : `Upload Selected (${selected.length})` }}
      </button>
      <!-- No SVG, deliberately: upload.php's allowlist excludes it because an SVG
           served inline is same-origin script. The hint used to advertise SVG, so
           every SVG picked here was accepted by the client and then refused by the
           server - one more silent upload. Listed types are the allowlist. -->
      <small class="form-hint">Images: JPG, PNG, GIF, WEBP. Videos: MP4, WEBM, AVI, MOV. Max 100 MB each.</small>
    </div>
    <div v-if="rejected.length>0" class="rejected-files">
      <p><strong>Rejected files:</strong></p>
      <ul>
        <li v-for="(r,i) in rejected" :key="i">{{ r.file.name }} - {{ r.reason }}</li>
      </ul>
    </div>
    <div v-if="media.length>0" class="media-grid">
      <div v-for="item in media" :key="item.id" class="media-item">
        <div class="media-preview">
          <img v-if="item.media_type==='photo'" :src="getFileUrl(item.file_path)" :alt="item.file_name" class="media-image" />
          <video v-else controls class="media-video"><source :src="getFileUrl(item.file_path)" :type="item.file_type"></video>
        </div>
        <div class="media-info">
          <p class="media-filename" :title="item.file_name">{{ item.file_name }}</p>
          <p class="media-meta">{{ formatDate(item.uploaded_at) }} by {{ item.uploaded_by_name || '-' }}</p>
          <span v-if="item.is_active===0" class="media-hidden-badge">Hidden</span>
        </div>
        <button class="btn delete-media-btn" @click="doDelete(item)">Delete</button>
      </div>
    </div>
    <div v-else class="empty-media">No media uploaded yet.</div>
    <p v-if="inactiveCount>0" class="form-hint media-hidden-note">{{ inactiveCount }} hidden item(s) not visible in gallery filters.</p>
  </BaseDialog>
</template>
<script setup>
import { onMounted, onUnmounted } from 'vue'
import BaseDialog from '../shared/BaseDialog.vue'
import { useCarNameMedia } from '../../composables/useCarNameMedia'
import { useSafeApi } from '../../composables/useSafeApi'
// notify is a prop, not a second useNotify() call. This dialog used to build its own
// instance, but useNotify's state is module-local to the instance and the <MessageBox>
// that renders it lives in the parent (CarNamesView). A dialog-local instance therefore
// had nowhere to be displayed: notify() set show = true on a ref no component was bound
// to, and every failure below was swallowed with no message and no log. Passing the
// parent's notify puts these errors back on screen.
const props = defineProps({
  carName: { type: Object, required: true },
  notify: { type: Function, required: true },
})
const emit = defineEmits(['close'])
const notify = props.notify
const { safeApi, uploadFile, getFileUrl } = useSafeApi()
const mediaApi = useCarNameMedia({ carName: props.carName, notify })
const m = mediaApi.withDeps({ safeApi, uploadFile, getFileUrl })
const media = m.media; const inactiveCount = m.inactiveCount; const selected = m.selected; const rejected = m.rejected; const uploading = m.uploading; const fileInput = m.fileInput; const formatDate = m.formatDate
function onSelect(e){ m.handleFileSelect(e.target.files) }
async function doUpload(){ await m.upload() }
async function doDelete(item){ await m.remove(item) }
onMounted(async () => { await m.fetchMedia() })
onUnmounted(() => m.reset())
</script>
<style scoped>
.media-upload-section{margin-bottom:16px;display:flex;flex-direction:column;gap:8px}
.file-input{padding:8px;border:1px solid #ddd;border-radius:4px}
.upload-btn{align-self:flex-start;padding:8px 16px;background:#10b981;color:#fff;border:none;border-radius:4px;cursor:pointer}
.upload-btn:disabled{opacity:.6;cursor:not-allowed}
.rejected-files{background:#fef2f2;border:1px solid #fecaca;color:#b91c1c;padding:12px;border-radius:4px;margin-bottom:16px}
.media-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(250px,1fr));gap:16px;max-height:60vh;overflow-y:auto;padding-right:4px}
.media-item{border:1px solid #e5e7eb;border-radius:8px;padding:8px;background:#fafafa;display:flex;flex-direction:column;gap:8px}
.media-preview{display:flex;align-items:center;justify-content:center;background:#111827;border-radius:4px;overflow:hidden;height:180px}
.media-image{max-width:100%;max-height:100%;object-fit:contain}
.media-video{width:100%;height:100%}
.media-info{display:flex;flex-direction:column;gap:4px}
.media-filename{font-size:.875rem;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.media-meta{font-size:.75rem;color:#6b7280}
.media-hidden-badge{align-self:flex-start;background:#f59e0b;color:#fff;padding:2px 6px;border-radius:4px;font-size:.75rem}
.delete-media-btn{background:#ef4444;color:#fff;border:none;border-radius:4px;padding:6px 8px;cursor:pointer}
.empty-media{text-align:center;color:#6b7280;padding:40px}
.form-hint{color:#6b7280;font-size:.875rem}
.media-hidden-note{margin-top:12px}
</style>
