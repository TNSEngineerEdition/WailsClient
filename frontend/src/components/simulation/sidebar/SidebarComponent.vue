<script setup lang="ts">
import { computed } from "vue"

const model = defineModel({ required: true })
const props = defineProps<{
  position: "left" | "right"
  titleIcon: string
  title: string
}>()

const slideDirection = computed(() => `sidebar-slide-${props.position}`)
</script>

<template>
  <transition :name="slideDirection">
    <v-card v-if="model" class="side-bar-card">
      <v-card-title class="d-flex align-center justify-space-between my-1">
        <transition name="content-fade" mode="out-in">
          <div class="title-left" :key="props.title">
            <v-icon :icon="props.titleIcon" class="mr-2" />
            <span class="font-weight-bold">{{ props.title }}</span>

            <slot name="title-actions" />
          </div>
        </transition>

        <v-btn
          icon="mdi-close"
          variant="text"
          density="compact"
          class="ml-1"
          @click="model = false"
        />
      </v-card-title>
      <v-card-text>
        <transition name="content-fade" mode="out-in">
          <div :key="props.title"><slot /></div>
        </transition>
      </v-card-text>
    </v-card>
  </transition>
</template>

<style scoped lang="scss">
.side-bar-card {
  width: 100%;
  max-height: calc(100vh - 120px);
  overflow: auto;
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

.title-left {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.title-left span {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>

<style lang="scss">
.section {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.4rem;

  &.arrivals {
    flex-direction: column;
    align-items: flex-start;
  }
}

.label,
.value {
  font-size: clamp(0.8rem, 0.75rem + 0.2vw, 1rem);
  color: #111;
}

.label {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: bold;
  margin-bottom: 0.25rem;
}

.value {
  font-weight: 450;
  text-align: right;
}
</style>
