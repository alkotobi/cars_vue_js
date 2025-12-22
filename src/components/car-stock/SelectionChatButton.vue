<script setup>
import { ref } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useSelectionChat } from '@/composables/useSelectionChat'

const props = defineProps({
  selection: {
    type: Object,
    required: true,
  },
})

const { t } = useEnhancedI18n()
const { loading, getOrCreateSelectionGroup } = useSelectionChat()
const isProcessing = ref(false)

const handleChatClick = async () => {
  if (isProcessing.value || loading.value) {
    return
  }

  try {
    isProcessing.value = true

    // Get or create the group chat
    const result = await getOrCreateSelectionGroup(props.selection)

    if (result.success && result.group) {
      // Dispatch event to open chat modal and select this group
      window.dispatchEvent(
        new CustomEvent('open-chat-with-group', {
          detail: {
            groupId: result.group.id,
            groupName: result.group.name,
          },
        })
      )
    } else {
      alert(
        result.error ||
          t('carStock.failed_to_open_chat') ||
          'Failed to open chat. Please try again.'
      )
    }
  } catch (err) {
    console.error('Error opening selection chat:', err)
    alert(
      t('carStock.error_opening_chat') || 'Error opening chat. Please try again.'
    )
  } finally {
    isProcessing.value = false
  }
}
</script>

<template>
  <button
    @click="handleChatClick"
    class="action-btn chat-btn"
    :title="t('carStock.open_selection_chat') || 'Open Selection Chat'"
    :disabled="isProcessing || loading"
  >
    <i v-if="isProcessing || loading" class="fas fa-spinner fa-spin"></i>
    <i v-else class="fas fa-comments"></i>
    {{ t('carStock.chat') || 'Chat' }}
  </button>
</template>

<style scoped>
.chat-btn {
  background-color: #8b5cf6;
}

.chat-btn:hover:not(:disabled) {
  background-color: #7c3aed;
}

.chat-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>

