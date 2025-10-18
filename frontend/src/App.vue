<script lang="ts" setup>
import { ref } from "vue"
import HeaderComponent from "./components/header/HeaderComponent.vue"
import MapComponent from "./components/MapComponent.vue"
import EditMapComponent from "./components/editMap/EditMapComponent.vue"

const time = ref(0)
const loading = ref(true)
const speed = ref(1)
const isRunning = ref(false)
const resetCounter = ref(0)
const isEditMap = ref(true)
</script>

<template>
  <HeaderComponent
    v-if="!isEditMap"
    v-model:is-running="isRunning"
    v-model:speed="speed"
    v-model:reset-counter="resetCounter"
    :time="time"
    :loading="loading"
  ></HeaderComponent>

  <MapComponent
    v-if="!isEditMap"
    v-model:time="time"
    v-model:loading="loading"
    :is-running="isRunning"
    :speed="speed"
    :reset-counter="resetCounter"
  ></MapComponent>

  <EditMapComponent
    v-if="isEditMap"
    v-model:isEditMap="isEditMap"
    :loading="loading"
  ></EditMapComponent>
</template>
