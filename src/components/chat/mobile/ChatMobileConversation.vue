<template>
  <div class="chat-mobile-conversation">
    <ChatMessages
      ref="chatMessagesRef"
      :group-id="groupId"
      :group-name="groupName"
      @message-sent="$emit('message-sent')"
      @new-messages-received="$emit('new-messages-received', $event)"
      @reset-triggered="$emit('reset-triggered', $event)"
    >
      <template #header-left>
        <button class="mobile-back-btn" @click="$emit('back')">
          <i class="fas fa-arrow-left"></i>
        </button>
      </template>
    </ChatMessages>
  </div>
</template>

<script setup>
import { ref, defineExpose, defineEmits } from 'vue'
import ChatMessages from '../ChatMessages.vue'

const props = defineProps({
  groupId: {
    type: Number,
    required: true,
  },
  groupName: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['back', 'message-sent', 'new-messages-received', 'reset-triggered'])

const chatMessagesRef = ref(null)

defineExpose({
  cleanup: () => chatMessagesRef.value?.cleanup?.(),
  focusMessageInput: () => chatMessagesRef.value?.focusMessageInput?.(),
  scrollToBottom: async () => await chatMessagesRef.value?.scrollToBottom?.(),
})
</script>

<style scoped>
.chat-mobile-conversation {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.mobile-back-btn {
  width: 38px;
  height: 38px;
  border: none;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.18);
  color: white;
  font-size: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  margin-right: 8px;
  flex-shrink: 0;
}

.mobile-back-btn:active {
  background-color: rgba(255, 255, 255, 0.32);
}

.chat-mobile-conversation :deep(.chat-messages) {
  border-radius: 0;
  box-shadow: none;
  height: 100%;
  min-height: 0;
}

.chat-mobile-conversation :deep(.messages-header) {
  min-height: 56px;
  padding: 8px 10px;
  padding-top: calc(8px + env(safe-area-inset-top, 0px));
  gap: 6px;
  align-items: center;
}

.chat-mobile-conversation :deep(.messages-header h3) {
  flex: 1;
  font-size: 0.98rem;
  gap: 6px;
  margin: 0;
}

.chat-mobile-conversation :deep(.group-info) {
  gap: 6px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.chat-mobile-conversation :deep(.timezone-info) {
  display: none;
}

.chat-mobile-conversation :deep(.member-count) {
  font-size: 0.7rem;
  opacity: 0.9;
}

.chat-mobile-conversation :deep(.selection-mode-btn),
.chat-mobile-conversation :deep(.bulk-edit-btn),
.chat-mobile-conversation :deep(.deselect-all-btn) {
  padding: 5px 8px;
  font-size: 0.78rem;
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  background-color: rgba(255, 255, 255, 0.16);
  color: white;
}

.chat-mobile-conversation :deep(.selection-mode-btn.active) {
  background-color: rgba(255, 255, 255, 0.28);
}

.chat-mobile-conversation :deep(.message-search-container) {
  padding: 8px 10px;
  background-color: #f8fafc;
}

.chat-mobile-conversation :deep(.messages-container) {
  padding: 12px 10px 16px;
}

.chat-mobile-conversation :deep(.message-input-container) {
  padding: 8px 10px calc(8px + env(safe-area-inset-bottom, 0px));
  background-color: white;
  border-top: 1px solid #e2e8f0;
}

.chat-mobile-conversation :deep(.input-wrapper) {
  gap: 6px;
}

.chat-mobile-conversation :deep(.message-input) {
  font-size: 15px;
  padding: 9px 12px;
  border-radius: 22px;
}

.chat-mobile-conversation :deep(.emoji-button),
.chat-mobile-conversation :deep(.file-button),
.chat-mobile-conversation :deep(.voice-button),
.chat-mobile-conversation :deep(.send-button) {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: none;
  background-color: #f1f5f9;
  color: #0f172a;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.chat-mobile-conversation :deep(.send-button) {
  background-color: #06b6d4;
  color: white;
}

.chat-mobile-conversation :deep(.send-button:disabled) {
  opacity: 0.5;
  background-color: #cbd5f5;
}
</style>
