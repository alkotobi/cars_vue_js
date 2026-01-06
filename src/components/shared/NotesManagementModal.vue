<template>
  <teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="handleClose">
      <div class="modal-container" @click.stop>
        <div class="modal-header">
          <h3>
            <i class="fas fa-sticky-note"></i>
            {{ t('sellBills.manage_notes') || 'Manage Notes' }}
          </h3>
          <button @click.stop="handleClose" class="close-btn" type="button">
            <i class="fas fa-times"></i>
          </button>
        </div>
        
        <div class="modal-content">
          <!-- Add New Note Form -->
          <div class="add-note-section">
            <h4>
              <i class="fas fa-plus-circle"></i>
              {{ t('sellBills.add_new_note') || 'Add New Note' }}
            </h4>
            <div class="form-group">
              <textarea
                v-model="newNoteText"
                :placeholder="t('sellBills.enter_note_placeholder') || 'Enter your note...'"
                rows="4"
                class="note-textarea"
              ></textarea>
            </div>
            <button 
              @click="handleAddNote" 
              :disabled="!newNoteText.trim() || saving"
              class="btn-add-note"
            >
              <i v-if="saving" class="fas fa-spinner fa-spin"></i>
              <i v-else class="fas fa-plus"></i>
              {{ saving ? t('sellBills.saving') || 'Saving...' : t('sellBills.add_note') || 'Add Note' }}
            </button>
          </div>

          <!-- Notes List -->
          <div class="notes-list-section">
            <h4>
              <i class="fas fa-list"></i>
              {{ t('sellBills.notes_list') || 'Notes List' }}
            </h4>
            
            <div v-if="localNotes.length === 0" class="no-notes">
              <i class="fas fa-inbox"></i>
              <p>{{ t('sellBills.no_notes') || 'No notes available' }}</p>
            </div>
            
            <div v-else class="notes-list">
              <div 
                v-for="(note, index) in localNotes" 
                :key="index"
                class="note-item"
                :class="{ 'editing': editingIndex === index }"
              >
                <template v-if="editingIndex === index">
                  <div class="note-edit-form">
                    <textarea
                      v-model="editNoteText"
                      rows="3"
                      class="note-textarea"
                    ></textarea>
                    <div class="note-actions">
                      <button 
                        @click="handleSaveEdit(index)" 
                        :disabled="!editNoteText.trim() || saving"
                        class="btn-save"
                      >
                        <i class="fas fa-check"></i>
                        {{ t('sellBills.save') || 'Save' }}
                      </button>
                      <button 
                        @click="cancelEdit" 
                        :disabled="saving"
                        class="btn-cancel"
                      >
                        <i class="fas fa-times"></i>
                        {{ t('sellBills.cancel') || 'Cancel' }}
                      </button>
                    </div>
                  </div>
                </template>
                <template v-else>
                  <div class="note-content">
                    <div class="note-header">
                      <div class="note-meta">
                        <span class="note-number">#{{ index + 1 }}</span>
                        <span class="note-user">
                          <i class="fas fa-user"></i>
                          {{ getUserName(note.id_user) }}
                        </span>
                        <span class="note-timestamp">
                          <i class="fas fa-clock"></i>
                          {{ formatTimestamp(note.timestamp) }}
                        </span>
                      </div>
                      <div class="note-actions">
                        <button 
                          v-if="canEditNote(note)"
                          @click="startEdit(index)"
                          class="btn-edit"
                          :title="t('sellBills.edit_note') || 'Edit Note'"
                        >
                          <i class="fas fa-edit"></i>
                        </button>
                        <button 
                          v-if="canDeleteNote(note)"
                          @click="handleDelete(index)"
                          class="btn-delete"
                          :title="t('sellBills.delete_note') || 'Delete Note'"
                        >
                          <i class="fas fa-trash"></i>
                        </button>
                      </div>
                    </div>
                    <div class="note-text">{{ note.note || '-' }}</div>
                  </div>
                </template>
              </div>
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <button @click="handleClose" class="btn-close">
            {{ t('sellBills.close') || 'Close' }}
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  notes: {
    type: Array,
    default: () => [],
  },
  users: {
    type: Array,
    default: () => [],
  },
  currentUserId: {
    type: Number,
    default: null,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
  // Generic save function - accepts (notesJson, entityId) and returns Promise
  // If not provided, modal will work in draft mode (only emit updates, don't save)
  saveFunction: {
    type: Function,
    default: null,
  },
  entityId: {
    type: [Number, String],
    default: null,
  },
  // If true, saves immediately to database. If false, only emits updates (draft mode)
  saveImmediately: {
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['close', 'notes-updated'])

const localNotes = ref([])
const newNoteText = ref('')
const editingIndex = ref(null)
const editNoteText = ref('')
const saving = ref(false)

// Watch for changes in notes prop
watch(() => props.notes, (newNotes) => {
  localNotes.value = JSON.parse(JSON.stringify(newNotes || []))
}, { immediate: true, deep: true })

const getUserName = (userId) => {
  if (!userId) return t('sellBills.unknown') || 'Unknown'
  // Handle both string and number ID comparisons
  const user = props.users.find(u => u.id == userId || u.id === userId)
  if (!user) return `User ${userId}`
  
  // Return username if available, otherwise show User ID
  return user.username || `User ${userId}`
}

const formatTimestamp = (timestamp) => {
  if (!timestamp) return '-'
  try {
    const date = new Date(timestamp)
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch (e) {
    return timestamp
  }
}

const canEditNote = (note) => {
  if (props.isAdmin) return true
  return note.id_user === props.currentUserId
}

const canDeleteNote = (note) => {
  if (props.isAdmin) return true
  return note.id_user === props.currentUserId
}

const handleAddNote = async () => {
  if (!newNoteText.value.trim() || saving.value) return
  
  saving.value = true
  try {
    const newNote = {
      id_user: props.currentUserId,
      note: newNoteText.value.trim(),
      timestamp: new Date().toISOString().slice(0, 19).replace('T', ' ')
    }
    
    localNotes.value.push(newNote)
    await saveNotesToDatabase()
    newNoteText.value = ''
  } catch (error) {
    console.error('Error adding note:', error)
    alert(error.message || 'Failed to add note')
  } finally {
    saving.value = false
  }
}

const startEdit = (index) => {
  editingIndex.value = index
  editNoteText.value = localNotes.value[index].note || ''
}

const cancelEdit = () => {
  editingIndex.value = null
  editNoteText.value = ''
}

const handleSaveEdit = async (index) => {
  if (!editNoteText.value.trim() || saving.value) return
  
  saving.value = true
  try {
    localNotes.value[index].note = editNoteText.value.trim()
    localNotes.value[index].timestamp = new Date().toISOString().slice(0, 19).replace('T', ' ')
    await saveNotesToDatabase()
    cancelEdit()
  } catch (error) {
    console.error('Error updating note:', error)
    alert(error.message || 'Failed to update note')
  } finally {
    saving.value = false
  }
}

const handleDelete = async (index) => {
  const confirmMessage = t('sellBills.confirm_delete_note') || 'Are you sure you want to delete this note?'
  if (!confirm(confirmMessage)) {
    return
  }
  
  saving.value = true
  try {
    localNotes.value.splice(index, 1)
    await saveNotesToDatabase()
  } catch (error) {
    console.error('Error deleting note:', error)
    alert(error.message || 'Failed to delete note')
  } finally {
    saving.value = false
  }
}

const saveNotesToDatabase = async () => {
  // Always emit the updated notes
  emit('notes-updated', localNotes.value)
  
  // Only save to database if saveImmediately is true and saveFunction is provided
  if (props.saveImmediately && props.saveFunction && props.entityId) {
    const notesJson = localNotes.value.length > 0 ? JSON.stringify(localNotes.value) : null
    await props.saveFunction(notesJson, props.entityId)
  }
}

const handleClose = () => {
  // Reset form state
  newNoteText.value = ''
  cancelEdit()
  emit('close')
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
  padding: 20px;
  backdrop-filter: blur(4px);
}

.modal-container {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 800px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.15);
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.modal-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
  font-size: 1.25rem;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #ef4444;
}

.modal-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.add-note-section {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 2px solid #e5e7eb;
}

.add-note-section h4 {
  margin: 0 0 1rem 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.add-note-section h4 i {
  color: #3b82f6;
}

.form-group {
  margin-bottom: 1rem;
}

.note-textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-family: inherit;
  resize: vertical;
  transition: border-color 0.2s;
}

.note-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.btn-add-note {
  padding: 0.625rem 1.25rem;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-add-note:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-add-note:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.notes-list-section h4 {
  margin: 0 0 1rem 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.notes-list-section h4 i {
  color: #3b82f6;
}

.no-notes {
  text-align: center;
  padding: 2rem;
  color: #6b7280;
}

.no-notes i {
  font-size: 2rem;
  margin-bottom: 0.5rem;
  color: #d1d5db;
}

.no-notes p {
  margin: 0;
  font-size: 0.9rem;
}

.notes-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.note-item {
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  padding: 1rem;
  transition: all 0.2s;
}

.note-item:hover {
  border-color: #d1d5db;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.note-item.editing {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.note-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.note-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.note-meta {
  display: flex;
  align-items: center;
  gap: 1rem;
  font-size: 0.8125rem;
  color: #6b7280;
}

.note-number {
  font-weight: 600;
  color: #374151;
}

.note-user,
.note-timestamp {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.note-user i,
.note-timestamp i {
  font-size: 0.75rem;
}

.note-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-edit,
.btn-delete {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.375rem 0.5rem;
  border-radius: 0.25rem;
  transition: all 0.2s;
  font-size: 0.875rem;
}

.btn-edit {
  color: #3b82f6;
}

.btn-edit:hover {
  background-color: #eff6ff;
  color: #2563eb;
}

.btn-delete {
  color: #ef4444;
}

.btn-delete:hover {
  background-color: #fef2f2;
  color: #dc2626;
}

.note-text {
  color: #1f2937;
  font-size: 0.875rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.note-edit-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.note-edit-form .note-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.btn-save {
  padding: 0.5rem 1rem;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-save:hover:not(:disabled) {
  background-color: #059669;
}

.btn-save:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-cancel {
  padding: 0.5rem 1rem;
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #4b5563;
}

.btn-cancel:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: flex-end;
}

.btn-close {
  padding: 0.625rem 1.25rem;
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s;
}

.btn-close:hover {
  background-color: #4b5563;
}
</style>

