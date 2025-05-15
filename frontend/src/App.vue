<script lang="ts" setup>
import { ref } from "vue"
import HeaderComponent from "./components/header/HeaderComponent.vue"
import MapComponent from "./components/MapComponent.vue"

const time = ref(0)
const loading = ref(true)
const speed = ref(1)
const isRunning = ref(false)
const resetCounter = ref(0)
</script>

<template>
  <HeaderComponent
    v-model:is-running="isRunning"
    v-model:speed="speed"
    v-model:reset-counter="resetCounter"
    :time="time"
    :loading="loading"
  ></HeaderComponent>

  <MapComponent
    v-model:time="time"
    v-model:loading="loading"
    :is-running="isRunning"
    :speed="speed"
    :reset-counter="resetCounter"
  ></MapComponent>
</template>
