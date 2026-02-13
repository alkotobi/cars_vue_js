<script setup>
import { ref, computed, onMounted, watch } from 'vue'

const editMode = ref('form') // 'form' | 'raw'
const options = ref(null)
const rawJson = ref('')
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const saveSuccess = ref(false)

const hasCrossDev = computed(() => typeof window !== 'undefined' && window.CrossDev?.invoke)

const htmlLoading = computed({
  get: () => options.value?.htmlLoading ?? { method: 'file', filePath: '', url: '', htmlContent: '', preloadPath: '' },
  set: (v) => {
    if (!options.value) options.value = {}
    options.value.htmlLoading = v
  }
})

const method = computed({
  get: () => htmlLoading.value.method ?? 'file',
  set: (v) => { htmlLoading.value = { ...htmlLoading.value, method: v } }
})
const filePath = computed({
  get: () => htmlLoading.value.filePath ?? '',
  set: (v) => { htmlLoading.value = { ...htmlLoading.value, filePath: v } }
})
const url = computed({
  get: () => htmlLoading.value.url ?? '',
  set: (v) => { htmlLoading.value = { ...htmlLoading.value, url: v } }
})
const htmlContent = computed({
  get: () => htmlLoading.value.htmlContent ?? '',
  set: (v) => { htmlLoading.value = { ...htmlLoading.value, htmlContent: v } }
})
const preloadPath = computed({
  get: () => htmlLoading.value.preloadPath ?? '',
  set: (v) => { htmlLoading.value = { ...htmlLoading.value, preloadPath: v } }
})

function syncRawFromOptions() {
  try {
    rawJson.value = JSON.stringify(options.value ?? {}, null, 2)
  } catch (_) {
    rawJson.value = '{}'
  }
}

function syncOptionsFromRaw() {
  try {
    options.value = JSON.parse(rawJson.value)
    error.value = ''
    return true
  } catch (e) {
    error.value = 'Invalid JSON: ' + (e.message || String(e))
    return false
  }
}

watch(options, () => {
  if (editMode.value === 'raw') syncRawFromOptions()
}, { deep: true })

watch(editMode, (m) => {
  if (m === 'raw') syncRawFromOptions()
  else if (m === 'form') syncOptionsFromRaw()
})

async function loadOptions() {
  if (!hasCrossDev.value) {
    error.value = 'CrossDev API not available'
    loading.value = false
    return
  }
  loading.value = true
  error.value = ''
  try {
    const r = await window.CrossDev.invoke('readOptions', {})
    if (r?.success && r.options) {
      options.value = r.options
      syncRawFromOptions()
    } else {
      error.value = r?.error || 'Failed to load options'
    }
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    loading.value = false
  }
}

async function saveOptions() {
  if (!hasCrossDev.value) return
  if (editMode.value === 'raw' && !syncOptionsFromRaw()) return
  saving.value = true
  error.value = ''
  saveSuccess.value = false
  try {
    const r = await window.CrossDev.invoke('writeOptions', { options: options.value })
    if (r?.success) {
      saveSuccess.value = true
      setTimeout(() => { saveSuccess.value = false }, 3000)
    } else {
      error.value = r?.error || 'Failed to save'
    }
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    saving.value = false
  }
}

async function browseFile(filterKey) {
  if (!hasCrossDev.value) return
  const filters = {
    html: 'HTML Files (*.html)|*.html|All Files (*.*)|*.*',
    js: 'JavaScript (*.js)|*.js|All Files (*.*)|*.*',
    any: 'All Files (*.*)|*.*'
  }
  const titles = {
    html: 'Select HTML file',
    js: 'Select preload script',
    any: 'Select file'
  }
  try {
    const r = await window.CrossDev.invoke('openFileDialog', {
      title: titles[filterKey] || titles.any,
      filter: filters[filterKey] || filters.any
    })
    if (r?.success && r.filePath) {
      if (filterKey === 'html') filePath.value = r.filePath
      else if (filterKey === 'js') preloadPath.value = r.filePath
    }
  } catch (e) {
    error.value = e?.message || String(e)
  }
}

onMounted(() => {
  document.title = 'Settings - CrossDev Options'
  loadOptions()
})
</script>

<template>
  <div class="settings-view">
    <header class="settings-header">
      <h1>CrossDev Settings</h1>
      <p class="settings-subtitle">Edit options.json – HTML loading configuration</p>
    </header>

    <div v-if="!hasCrossDev" class="settings-error">
      CrossDev API is not available. This view is intended to run inside the CrossDev desktop app.
    </div>
    <template v-else>
      <div class="settings-tabs">
        <button
          type="button"
          class="tab"
          :class="{ active: editMode === 'form' }"
          @click="editMode = 'form'"
        >
          Form
        </button>
        <button
          type="button"
          class="tab"
          :class="{ active: editMode === 'raw' }"
          @click="editMode = 'raw'"
        >
          Raw JSON
        </button>
      </div>

      <div v-if="loading" class="settings-loading">Loading options...</div>
      <template v-else>
        <div v-if="error" class="settings-error">{{ error }}</div>
        <div v-if="saveSuccess" class="settings-success">Saved successfully. Restart the app to apply changes.</div>

        <div v-if="editMode === 'form'" class="settings-form">
          <section class="form-section">
            <h2>HTML Loading</h2>
            <div class="form-group">
              <label>Method</label>
              <select v-model="method">
                <option value="file">File</option>
                <option value="url">URL</option>
                <option value="html">HTML</option>
              </select>
              <small>How the main window content is loaded</small>
            </div>
            <div v-if="method === 'file'" class="form-group">
              <label>File path</label>
              <div class="input-with-browse">
                <input v-model="filePath" type="text" placeholder="e.g. demo.html or dist/index.html" />
                <button type="button" class="browse-btn" @click="browseFile('html')">Browse...</button>
              </div>
              <small>Relative to app working directory or absolute path</small>
            </div>
            <div v-if="method === 'url'" class="form-group">
              <label>URL</label>
              <input v-model="url" type="text" placeholder="e.g. http://localhost:5173/" />
            </div>
            <div v-if="method === 'html'" class="form-group">
              <label>HTML content</label>
              <textarea v-model="htmlContent" rows="4" placeholder="Inline HTML string"></textarea>
            </div>
            <div class="form-group">
              <label>Preload script path (optional)</label>
              <div class="input-with-browse">
                <input v-model="preloadPath" type="text" placeholder="Path to custom preload script" />
                <button type="button" class="browse-btn" @click="browseFile('js')">Browse...</button>
              </div>
              <small>Leave empty to use built-in bridge</small>
            </div>
          </section>
        </div>

        <div v-else class="settings-raw">
          <textarea
            v-model="rawJson"
            class="raw-json"
            spellcheck="false"
            placeholder="{}"
            @blur="syncOptionsFromRaw()"
          ></textarea>
          <p class="raw-hint">Edit JSON directly. Switch to Form to use structured controls.</p>
        </div>

        <div class="settings-actions">
          <button type="button" class="btn-primary" :disabled="saving" @click="saveOptions">
            {{ saving ? 'Saving...' : 'Save' }}
          </button>
          <button type="button" class="btn-secondary" :disabled="loading" @click="loadOptions">
            Reload
          </button>
        </div>
      </template>
    </template>
  </div>
</template>

<style scoped>
.settings-view {
  min-height: 100vh;
  padding: 24px;
  font-family: system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif;
  background: #1e1e1e;
  color: #e0e0e0;
}
.settings-header {
  margin-bottom: 24px;
}
.settings-header h1 {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: #fff;
}
.settings-subtitle {
  font-size: 0.9rem;
  color: #888;
  margin: 0;
}
.settings-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 20px;
}
.tab {
  padding: 8px 16px;
  background: #2d2d2d;
  border: 1px solid #444;
  border-radius: 6px;
  color: #b0b0b0;
  cursor: pointer;
  font-size: 0.9rem;
}
.tab:hover {
  background: #383838;
  color: #e0e0e0;
}
.tab.active {
  background: #0d6efd;
  border-color: #0d6efd;
  color: #fff;
}
.settings-loading,
.settings-error,
.settings-success {
  padding: 12px 16px;
  border-radius: 6px;
  margin-bottom: 16px;
}
.settings-loading {
  background: #2d2d2d;
  color: #888;
}
.settings-error {
  background: rgba(220, 53, 69, 0.2);
  border: 1px solid #dc3545;
  color: #f8a5a5;
}
.settings-success {
  background: rgba(25, 135, 84, 0.2);
  border: 1px solid #198754;
  color: #86c9a0;
}
.settings-form {
  margin-bottom: 24px;
}
.form-section {
  background: #2d2d2d;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 16px;
}
.form-section h2 {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #fff;
}
.form-group {
  margin-bottom: 16px;
}
.form-group:last-child {
  margin-bottom: 0;
}
.form-group label {
  display: block;
  font-size: 0.85rem;
  font-weight: 500;
  margin-bottom: 6px;
  color: #b0b0b0;
}
.input-with-browse {
  display: flex;
  gap: 8px;
  align-items: stretch;
}
.input-with-browse input {
  flex: 1;
  min-width: 0;
}
.browse-btn {
  flex-shrink: 0;
  padding: 8px 14px;
  background: #444;
  border: 1px solid #555;
  border-radius: 6px;
  color: #e0e0e0;
  font-size: 0.85rem;
  cursor: pointer;
}
.browse-btn:hover {
  background: #555;
  border-color: #666;
}
.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 6px;
  color: #e0e0e0;
  font-size: 0.9rem;
  caret-color: #fff;
}
.form-group .input-with-browse input {
  width: auto;
}
.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #0d6efd;
}
.form-group textarea {
  resize: vertical;
  font-family: ui-monospace, monospace;
}
.form-group small {
  display: block;
  margin-top: 4px;
  font-size: 0.75rem;
  color: #666;
}
.settings-raw {
  margin-bottom: 24px;
}
.raw-json {
  width: 100%;
  min-height: 320px;
  padding: 16px;
  background: #1e1e1e;
  border: 1px solid #444;
  border-radius: 8px;
  color: #e0e0e0;
  font-family: ui-monospace, monospace;
  font-size: 13px;
  line-height: 1.5;
  resize: vertical;
  caret-color: #fff;
}
.raw-json:focus {
  outline: none;
  border-color: #0d6efd;
}
.raw-hint {
  font-size: 0.8rem;
  color: #666;
  margin: 8px 0 0 0;
}
.settings-actions {
  display: flex;
  gap: 12px;
}
.btn-primary,
.btn-secondary {
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  border: none;
}
.btn-primary {
  background: #0d6efd;
  color: #fff;
}
.btn-primary:hover:not(:disabled) {
  background: #0b5ed7;
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn-secondary {
  background: #444;
  color: #e0e0e0;
}
.btn-secondary:hover:not(:disabled) {
  background: #555;
}
.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
