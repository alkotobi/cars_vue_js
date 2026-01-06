<template>
  <div class="notes-table-container">
    <div class="notes-table-header">
      <h4>
        <i class="fas fa-sticky-note"></i>
        {{ t('sellBills.notes') }}
      </h4>
      <button 
        v-if="mode === 'edit'" 
        @click="$emit('manage-notes')" 
        class="btn-manage-notes"
        type="button"
      >
        <i class="fas fa-edit"></i>
        {{ t('sellBills.manage_notes') || 'Manage Notes' }}
      </button>
    </div>
    
    <div v-if="notes.length === 0" class="no-notes">
      <i class="fas fa-inbox"></i>
      <p>{{ t('sellBills.no_notes') || 'No notes available' }}</p>
    </div>
    
    <table v-else class="notes-table">
      <thead>
        <tr>
          <th>{{ t('sellBills.note_number') || '#' }}</th>
          <th>{{ t('sellBills.note_user') || 'User' }}</th>
          <th>{{ t('sellBills.note_text') || 'Note' }}</th>
          <th>{{ t('sellBills.note_timestamp') || 'Date/Time' }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(note, index) in notes" :key="index">
          <td class="note-number">{{ index + 1 }}</td>
          <td class="note-user">
            <i class="fas fa-user"></i>
            {{ getUserName(note.id_user) }}
          </td>
          <td class="note-text">{{ note.note || '-' }}</td>
          <td class="note-timestamp">
            {{ formatTimestamp(note.timestamp) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  notes: {
    type: Array,
    default: () => [],
  },
  users: {
    type: Array,
    default: () => [],
  },
  mode: {
    type: String,
    default: 'view', // 'view' or 'edit'
  },
})

const emit = defineEmits(['manage-notes'])

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
</script>

<style scoped>
.notes-table-container {
  margin-top: 1rem;
}

.notes-table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.notes-table-header h4 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.notes-table-header h4 i {
  color: #3b82f6;
}

.btn-manage-notes {
  padding: 0.5rem 1rem;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-manage-notes:hover {
  background-color: #2563eb;
}

.btn-manage-notes i {
  font-size: 0.875rem;
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

.notes-table {
  width: 100%;
  border-collapse: collapse;
  background-color: white;
  border-radius: 0.5rem;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.notes-table thead {
  background-color: #f9fafb;
}

.notes-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  font-size: 0.875rem;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
}

.notes-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e5e7eb;
  font-size: 0.875rem;
  color: #1f2937;
}

.notes-table tbody tr:hover {
  background-color: #f9fafb;
}

.notes-table tbody tr:last-child td {
  border-bottom: none;
}

.note-number {
  width: 60px;
  text-align: center;
  font-weight: 600;
  color: #6b7280;
}

.note-user {
  width: 150px;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.note-user i {
  color: #9ca3af;
  font-size: 0.75rem;
}

.note-text {
  max-width: 400px;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.note-timestamp {
  width: 180px;
  color: #6b7280;
  font-size: 0.8125rem;
}
</style>

