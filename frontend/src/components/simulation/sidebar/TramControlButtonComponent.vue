<script lang="ts" setup>
import { computed } from "vue"
const props = defineProps<{
  running: boolean
  disabled?: boolean
}>()

const emit = defineEmits(["click"])

const tramButtonLabel = computed(() => {
  if (props.disabled) return "Unavailable"
  return props.running ? "Stop Tram" : "Resume Tram"
})

const buttonColor = computed(() => {
  return props.disabled ? undefined : props.running ? "red" : "green"
})

const iconName = computed(() => {
  if (props.disabled) return "mdi-tram"
  return props.running ? "mdi-pause" : "mdi-play"
})
</script>

<template>
  <v-btn
    :color="buttonColor"
    :disabled="props.disabled"
    :class="props.disabled && 'btn-disabled'"
    variant="outlined"
    density="comfortable"
    size="default"
    rounded="lg"
    @click="emit('click')"
  >
    <v-icon class="mr-2" size="20">{{ iconName }}</v-icon>
    <span>{{ tramButtonLabel }}</span>
  </v-btn>
</template>

<style scoped>
.btn-disabled {
  background-color: #e0e0e0 !important;
  color: #9e9e9e !important;
  border-color: #bdbdbd !important;
}
</style>
