<script lang="ts" setup>
const props = defineProps<{
  description: string
  icon: string
  disabled: boolean
  parentProps?: Record<string, any>
}>()

const emit = defineEmits<{
  click: []
}>()
</script>

<template>
  <v-tooltip :text="props.description" location="bottom">
    <template v-slot:activator="{ props: tooltipProps }">
      <v-btn
        v-bind="{ ...tooltipProps, ...props.parentProps }"
        :icon="props.icon"
        :disabled="props.disabled"
        density="compact"
        size="x-large"
        variant="outlined"
        rounded="lg"
        @click="emit('click')"
      ></v-btn>
    </template>
  </v-tooltip>
</template>
