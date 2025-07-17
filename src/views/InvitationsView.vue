<template>
  <div class="invitations-view">
    <InvitationsTable />
  </div>
</template>

<script setup>
import InvitationsTable from '../components/InvitationsTable.vue'
</script>

<style scoped>
.invitations-view {
  min-height: 100vh;
  background-color: #f5f5f5;
}
</style>
