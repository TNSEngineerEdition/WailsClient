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

const slideDirection = computed(() => `sidebar-slide-${props.position}`)
</script>

<template>
  <transition :name="slideDirection">
    <v-card v-if="model" class="side-bar-card" :style="horizontalPosition">
      <v-card-title class="d-flex align-center justify-space-between my-1">
        <transition name="content-fade" mode="out-in">
          <div
            class="d-flex align-center justify-space-between"
            :key="props.title"
          >
            <v-icon :icon="props.titleIcon" class="mr-2"></v-icon>
            <span class="font-weight-bold">{{ props.title }}</span>
          </div>
        </transition>

        <v-btn
          icon="mdi-close"
          variant="text"
          density="compact"
          class="ml-4"
          @click="model = false"
        ></v-btn>
      </v-card-title>
      <v-card-text>
        <transition name="content-fade" mode="out-in">
          <div :key="props.title">
            <slot></slot>
          </div>
        </transition>
      </v-card-text>
    </v-card>
  </transition>
</template>

<style lang="scss" scoped>
.side-bar-card {
  position: absolute;
  top: calc(60px + 20px);
  z-index: 1001;
  background-color: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(100, 100, 100, 0.3);
  border-radius: 16px;
  box-shadow: 0 0 14px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(1.4px);
}

.sidebar-slide-left-enter-active,
.sidebar-slide-left-leave-active,
.sidebar-slide-right-enter-active,
.sidebar-slide-right-leave-active {
  transition:
    opacity 0.3s ease,
    transform 0.3s ease;
}

.sidebar-slide-left-enter-from,
.sidebar-slide-left-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}
.sidebar-slide-left-enter-to,
.sidebar-slide-left-leave-from {
  opacity: 1;
  transform: translateX(0);
}

.sidebar-slide-right-enter-from,
.sidebar-slide-right-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
.sidebar-slide-right-enter-to,
.sidebar-slide-right-leave-from {
  opacity: 1;
  transform: translateX(0);
}

.content-fade-enter-active,
.content-fade-leave-active {
  transition: opacity 0.25s ease;
}

.content-fade-enter-from,
.content-fade-leave-to {
  opacity: 0;
}

.content-fade-enter-to,
.content-fade-leave-from {
  opacity: 1;
}
</style>
