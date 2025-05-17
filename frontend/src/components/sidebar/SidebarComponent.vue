<script setup lang="ts">
import { computed } from "vue"

const model = defineModel({ required: true })

const props = defineProps<{
  position: "left" | "right"
  titleIcon: string
  title: string
}>()

const horizontalPosition = computed(() => {
  if (props.position == "left") {
    return { left: "54px" }
  } else {
    return { right: "20px" }
  }
})
</script>

<template>
  <transition name="fade-scale">
    <v-card v-if="model" class="side-bar-card" :style="horizontalPosition">
      <v-card-title class="d-flex align-center justify-space-between">
        <div class="d-flex align-center justify-space-between">
          <v-icon :icon="props.titleIcon" class="mr-2"></v-icon>

          {{ props.title }}
        </div>

        <v-btn
          icon="mdi-close"
          variant="text"
          density="compact"
          class="ml-4"
          @click="model = false"
        ></v-btn>
      </v-card-title>

      <v-card-text>
        <slot></slot>
      </v-card-text>
    </v-card>
  </transition>
</template>

<style lang="scss" scoped>
.side-bar-card {
  position: absolute;
  top: calc(60px + 20px);
  z-index: 1001;
  background-color: rgba(255, 255, 255, 0.8);
}

.fade-scale-enter-active,
.fade-scale-leave-active {
  transition: all 0.3s ease;
}

.fade-scale-enter-from,
.fade-scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
