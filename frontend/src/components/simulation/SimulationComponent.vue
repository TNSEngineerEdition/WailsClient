<script lang="ts" setup>
import { ref } from "vue"
import HeaderComponent from "@components/simulation/header/HeaderComponent.vue"
import MapComponent from "@components/simulation/MapComponent.vue"
import { MarkerColoringMode } from "@utils/types"
import MarkerColoringSelectorComponent from "./MarkerColoringSelectorComponent.vue"

const time = ref(0)
const loading = ref(true)
const speed = ref(1)
const isRunning = ref(false)
const resetCounter = ref(0)
const markerColoringMode = ref<MarkerColoringMode>("Default")
</script>

<template>
  <HeaderComponent
    v-model:is-running="isRunning"
    v-model:speed="speed"
    v-model:reset-counter="resetCounter"
    v-model:marker-coloring-mode="markerColoringMode"
    :time="time"
    :loading="loading"
  ></HeaderComponent>

  <div class="position-relative">
    <MapComponent
      v-model:time="time"
      v-model:loading="loading"
      v-model:is-running="isRunning"
      :speed="speed"
      :reset-counter="resetCounter"
      :marker-coloring-mode="markerColoringMode"
    ></MapComponent>
  </div>
</template>
