<script lang="ts" setup>
const props = defineProps<{
  running: boolean
  disabled?: boolean
}>()

const emit = defineEmits(["click"])
</script>

<template>
  <v-btn
    :color="!props.disabled ? (props.running ? 'red' : 'green') : undefined"
    :disabled="props.disabled"
    :class="props.disabled ? 'btn-disabled' : ''"
    variant="outlined"
    density="comfortable"
    size="default"
    rounded="lg"
    @click="emit('click')"
  >
    {{
      props.disabled
        ? "Unavailable"
        : props.running
          ? "Stop Tram"
          : "Resume Tram"
    }}
  </v-btn>
</template>

<style scoped>
.btn-disabled {
  background-color: #e0e0e0 !important;
  color: #9e9e9e !important;
  border-color: #bdbdbd !important;
}
</style>
