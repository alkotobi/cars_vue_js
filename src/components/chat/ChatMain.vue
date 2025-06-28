<script setup>
import { defineProps } from 'vue'

const props = defineProps({
  welcomeMessage: {
    type: String,
    default: 'Welcome to Chat!',
  },
  selectedGroup: {
    type: Object,
    default: null,
  },
})
</script>

<template>
  <div class="chat-main">
    <div class="chat-header">
      <h2>{{ welcomeMessage }}</h2>
      <div v-if="selectedGroup" class="selected-group-info">
        <h3>Selected: {{ selectedGroup.name }}</h3>
        <p>{{ selectedGroup.description }}</p>
      </div>
    </div>
    <div class="chat-content">
      <div v-if="!selectedGroup" class="select-group-message">
        <i class="fas fa-hand-pointer"></i>
        <p>Select a chat group from the sidebar to start chatting</p>
      </div>
      <div v-else class="chat-messages">
        <p>Chat functionality for "{{ selectedGroup.name }}" will be implemented here.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.chat-header {
  padding: 20px;
  background-color: white;
  border-bottom: 1px solid #e2e8f0;
}

.chat-header h2 {
  color: #06b6d4;
  margin: 0 0 10px 0;
}

.selected-group-info {
  margin-top: 10px;
  padding: 10px;
  background-color: #f0f9ff;
  border-radius: 6px;
  border-left: 3px solid #06b6d4;
}

.selected-group-info h3 {
  margin: 0 0 5px 0;
  color: #0369a1;
}

.selected-group-info p {
  margin: 0;
  color: #0c4a6e;
  font-size: 0.9rem;
}

.chat-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.select-group-message {
  text-align: center;
  color: #64748b;
  margin-top: 100px;
}

.select-group-message i {
  font-size: 3rem;
  color: #06b6d4;
  margin-bottom: 20px;
  display: block;
}

.chat-messages {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}
</style>
