<script lang="ts" setup>
import getColorForSpeed from "@utils/getColorForSpeed"
import { ref } from "vue"

const showLegend = ref(false)

const speedRanges = [
  { label: "0-10 km/h", speed: 10 },
  { label: "11-20 km/h", speed: 20 },
  { label: "21-30 km/h", speed: 30 },
  { label: "31-40 km/h", speed: 40 },
  { label: "41-50 km/h", speed: 50 },
  { label: "51-60 km/h", speed: 60 },
  { label: "60+ km/h", speed: 70 },
]
</script>

<template>
  <div
    class="legend-container"
    @mouseenter="showLegend = true"
    @mouseleave="showLegend = false"
  >
    <transition name="fade">
      <v-card v-if="showLegend" class="legend-card" elevation="8">
        <v-card-title class="text-subtitle-1 pb-2">Colors meaning</v-card-title>
        <v-divider class="mb-2" />
        <v-card-text>
          <div
            v-for="range in speedRanges"
            :key="range.label"
            class="legend-item"
          >
            <div
              class="color-box"
              :style="{ backgroundColor: getColorForSpeed(range.speed / 3.6) }"
            />
            <span>{{ range.label }}</span>
          </div>
        </v-card-text>
      </v-card>
    </transition>

    <v-btn icon="mdi-information-outline" size="large" elevation="3"></v-btn>
  </div>
</template>

<style scoped lang="scss">
.legend-container {
  position: absolute;
  bottom: 24px;
  right: 24px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  z-index: 1000;
}

.legend-card {
  margin-bottom: 8px;
  width: 200px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.color-box {
  width: 20px;
  height: 12px;
  border-radius: 4px;
}

.fade-enter-active,
.fade-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
</style>
