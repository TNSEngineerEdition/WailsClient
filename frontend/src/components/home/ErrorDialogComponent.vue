<script setup lang="ts">
const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  title: string
  error?: string
}>()
</script>

<template>
  <v-dialog v-model="model" max-width="500" max-height="400">
    <v-alert type="error" :title="props.title" class="mb-3">
      <pre class="error-content">{{ props.error }}</pre>
    </v-alert>
  </v-dialog>
</template>

<style lang="css" scoped>
.error-content {
  overflow-x: auto;
  overflow-y: auto;
  max-height: 100px;
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-word;
  margin: 0;
}
</style>
