<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi, colorErrorText } from '../composables/useApi'
import { useSubmitGuard } from '../composables/useSubmitGuard'
import { useEnhancedI18n } from '../composables/useI18n'

const { t } = useEnhancedI18n()
const { getColors, createColor, updateColor, deleteColor } = useApi()
const { guard, isBusy } = useSubmitGuard()

const colors = ref([])
const isLoading = ref(false)
const loadFailed = ref(false)
const pageError = ref(null)
const statusMessage = ref(null)
const dialogError = ref(null)

const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingColor = ref(null)

const user = ref(null)

// Delete is the only admin-gated action, and the button mirrors that rather than
// replacing it: create_color / update_color accept any signed-in user because a
// colour is picked while entering a car, while delete_color is admin-only in
// api/actions/colors.php. Hiding the button is a courtesy; the server refuses it.
const isAdmin = computed(() => user.value?.role_id === 1)

const newColor = ref({
  color: '',
  hexa: '#000000',
})

// colors.hexa is a free-text varchar, so a row can hold something the browser's
// colour input cannot represent - binding that to <input type="color"> snaps the
// picker to #000000 and a save would overwrite the stored value with no sign it
// ever existed. The value is shown as text as well as a swatch so a bad value is
// visible and fixable rather than silently replaced.
const HEX_PATTERN = /^#[0-9A-Fa-f]{6}$/

const isHexValid = (value) => HEX_PATTERN.test((value || '').trim())

const fetchColors = async () => {
  isLoading.value = true
  loadFailed.value = false
  try {
    colors.value = await getColors()
  } catch (err) {
    loadFailed.value = true
    pageError.value = colorErrorText(t, err) || t('colorsView.errors.loadFailed')
    console.error('fetchColors', err)
  } finally {
    isLoading.value = false
  }
}

// One funnel for the three writes. Keeps the dialog open on failure so the typed
// values are not lost, and turns every refusal into a visible line instead of the
// silent no-op this view used to be: callApi rejects on transport errors, and a
// duplicate colour comes back as {success: false} - both were being discarded.
const runWrite = async (work) => {
  pageError.value = null
  statusMessage.value = null
  try {
    await work()
    await fetchColors()
    return true
  } catch (err) {
    pageError.value = colorErrorText(t, err) || t('colorsView.errors.saveFailed')
    console.error('colors write', err)
    return false
  }
}

const addColor = guard('add', async () => {
  dialogError.value = null
  const ok = await runWrite(() => createColor({ ...newColor.value }))
  if (ok) {
    showAddDialog.value = false
    newColor.value = { color: '', hexa: '#000000' }
    statusMessage.value = t('colorsView.added')
  }
})

const openEditDialog = (color) => {
  editingColor.value = { ...color }
  dialogError.value = null
  showEditDialog.value = true
}

const saveEdit = guard('update', async () => {
  dialogError.value = null
  const ok = await runWrite(() => updateColor(editingColor.value.id, { ...editingColor.value }))
  if (ok) {
    showEditDialog.value = false
    editingColor.value = null
    statusMessage.value = t('colorsView.updated')
  }
})

// confirm() runs before the guarded call, not inside it. Inside, the guard holds
// the key for as long as the modal is open, so every row's button reads
// "Deleting..." and is disabled while the user is still being asked.
const confirmDelete = (color) => {
  if (!window.confirm(t('colorsView.confirmDelete', { name: color.color }))) {
    return
  }
  removeColor(color)
}

const removeColor = guard('delete', async (color) => {
  await runWrite(() => deleteColor(color.id))
  if (!pageError.value) {
    statusMessage.value = t('colorsView.deleted', { name: color.color })
  }
})

const closeAddDialog = () => {
  showAddDialog.value = false
  dialogError.value = null
  newColor.value = { color: '', hexa: '#000000' }
}

const closeEditDialog = () => {
  showEditDialog.value = false
  dialogError.value = null
  editingColor.value = null
}

onMounted(async () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    try {
      user.value = JSON.parse(userStr)
    } catch {
      user.value = null
    }
  }
  await fetchColors()
})
</script>

<template>
  <div class="colors-view">
    <div class="header">
      <h2>{{ t('colorsView.title') }}</h2>
      <button @click="showAddDialog = true" class="add-btn">{{ t('colorsView.addColor') }}</button>
    </div>

    <p v-if="pageError" class="banner banner-error" role="alert">{{ pageError }}</p>
    <p v-if="statusMessage" class="banner banner-ok" role="status">{{ statusMessage }}</p>

    <div class="content">
      <table class="colors-table">
        <thead>
          <tr>
            <th>{{ t('colorsView.colorName') }}</th>
            <th>{{ t('colorsView.colorPreview') }}</th>
            <th>{{ t('colorsView.hexCode') }}</th>
            <th>{{ t('colorsView.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="isLoading">
            <td colspan="4" class="muted">{{ t('colorsView.loading') }}</td>
          </tr>
          <tr v-else-if="loadFailed">
            <td colspan="4" class="muted">
              {{ t('colorsView.loadFailed') }}
              <button class="btn retry-btn" @click="fetchColors">
                {{ t('colorsView.retry') }}
              </button>
            </td>
          </tr>
          <tr v-else-if="!colors.length">
            <td colspan="4" class="muted">{{ t('colorsView.noColors') }}</td>
          </tr>
          <template v-else>
            <tr v-for="color in colors" :key="color.id">
              <td>{{ color.color }}</td>
              <td>
                <div
                  class="color-preview"
                  :style="{ backgroundColor: isHexValid(color.hexa) ? color.hexa : 'transparent' }"
                  :title="color.hexa || ''"
                ></div>
              </td>
              <td>
                <span class="hex-code">{{ color.hexa || t('colorsView.notSet') }}</span>
              </td>
              <td>
                <button @click="openEditDialog(color)" class="btn edit-btn">
                  {{ t('colorsView.edit') }}
                </button>
                <button
                  v-if="isAdmin"
                  @click="confirmDelete(color)"
                  class="btn delete-btn"
                  :disabled="isBusy('delete')"
                >
                  {{ isBusy('delete') ? t('colorsView.deleting') : t('colorsView.delete') }}
                </button>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <!-- Add Color Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>{{ t('colorsView.addTitle') }}</h3>
        <p v-if="dialogError" class="banner banner-error" role="alert">{{ dialogError }}</p>
        <div class="form-group">
          <label>{{ t('colorsView.colorName') }}:</label>
          <input
            v-model="newColor.color"
            :placeholder="t('colorsView.colorNamePlaceholder')"
            class="input-field"
          />
        </div>
        <div class="form-group">
          <label>{{ t('colorsView.colorPicker') }}:</label>
          <div class="color-picker-container">
            <input type="color" v-model="newColor.hexa" class="color-picker" />
            <input v-model="newColor.hexa" class="hex-input input-field" maxlength="7" />
          </div>
        </div>
        <div class="dialog-actions">
          <button @click="addColor" class="btn save-btn" :disabled="isBusy('add')">
            {{ isBusy('add') ? t('colorsView.saving') : t('colorsView.save') }}
          </button>
          <button @click="closeAddDialog" class="btn cancel-btn">
            {{ t('colorsView.cancel') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Color Dialog -->
    <div v-if="showEditDialog && editingColor" class="dialog-overlay">
      <div class="dialog">
        <h3>{{ t('colorsView.editTitle') }}</h3>
        <p v-if="dialogError" class="banner banner-error" role="alert">{{ dialogError }}</p>
        <div class="form-group">
          <label>{{ t('colorsView.colorName') }}:</label>
          <input
            v-model="editingColor.color"
            :placeholder="t('colorsView.colorNamePlaceholder')"
            class="input-field"
          />
        </div>
        <div class="form-group">
          <label>{{ t('colorsView.colorPicker') }}:</label>
          <div class="color-picker-container">
            <input type="color" v-model="editingColor.hexa" class="color-picker" />
            <input v-model="editingColor.hexa" class="hex-input input-field" maxlength="7" />
          </div>
        </div>
        <div class="dialog-actions">
          <button @click="saveEdit" class="btn save-btn" :disabled="isBusy('update')">
            {{ isBusy('update') ? t('colorsView.saving') : t('colorsView.save') }}
          </button>
          <button @click="closeEditDialog" class="btn cancel-btn">
            {{ t('colorsView.cancel') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.colors-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.colors-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.colors-table th,
.colors-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.colors-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.colors-table tbody tr:hover {
  background-color: #f5f5f5;
}

.muted {
  color: #6b7280;
  text-align: center;
}

.banner {
  margin: 0 0 12px;
  padding: 10px 12px;
  border-radius: 4px;
  font-size: 14px;
}

.banner-error {
  color: #991b1b;
  background-color: #fee2e2;
  border: 1px solid #fecaca;
}

.banner-ok {
  color: #065f46;
  background-color: #d1fae5;
  border: 1px solid #a7f3d0;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.retry-btn {
  margin-left: 8px;
  background-color: #3b82f6;
  color: white;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.dialog {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  min-width: 400px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.input-field {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.color-preview {
  width: 30px;
  height: 30px;
  border-radius: 4px;
  border: 1px solid #ddd;
  display: inline-block;
}

.hex-code {
  font-family: monospace;
  font-size: 14px;
}

.color-picker-container {
  display: flex;
  align-items: center;
  gap: 12px;
}

.color-picker {
  width: 50px;
  height: 40px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  padding: 0;
  background: none;
}

.color-picker::-webkit-color-swatch-wrapper {
  padding: 0;
}

.color-picker::-webkit-color-swatch {
  border: 1px solid #ddd;
}

.hex-input {
  font-family: monospace;
  text-transform: uppercase;
}
</style>
