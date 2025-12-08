<script lang="ts" setup>
import { Time } from "@classes/Time"
import HeaderIconButtonComponent from "@components/simulation/header/HeaderIconButtonComponent.vue"
import useCycle from "@composables/useCycle"
import { ResetSimulation } from "@wails/go/simulation/Simulation"
import HeaderRestartConfirmationDialogComponent from "@components/simulation/header/HeaderRestartConfirmationDialogComponent.vue"
import useTimer from "@composables/useTimer"
import { ExportToFile } from "@wails/go/simulation/Simulation"
import { useRouter } from "vue-router"

const props = defineProps<{
  time: number
  loading: boolean
}>()

const router = useRouter()

const isRunning = defineModel<boolean>("is-running", { required: true })
const speed = defineModel<number>("speed", { required: true })
const resetCounter = defineModel<number>("reset-counter", { required: true })

const speedsCycle = useCycle([1, 10, 100, 1000], speed)

const timer = useTimer()

function stop() {
  timer.stop()
  isRunning.value = false
}

function updateIsRunning() {
  if (isRunning.value) {
    stop()
  } else {
    timer.start()
    isRunning.value = true
  }
}

async function reset() {
  speedsCycle.reset()
  timer.reset()
  await ResetSimulation()
  resetCounter.value++
}

async function exportData() {
  const error = await ExportToFile()

  if (error) {
    console.error(error)
  }
}

async function menu() {
  stop()
  await reset()
  router.push("/")
}
</script>

<template>
  <v-footer tag="header">
    <v-row no-gutters>
      <v-col cols="4">
        <div class="button-box">
          <HeaderIconButtonComponent
            :disabled="loading"
            :icon="isRunning ? 'mdi-pause' : 'mdi-play'"
            :description="isRunning ? 'Pause' : 'Start'"
            @click="updateIsRunning"
          ></HeaderIconButtonComponent>

          <HeaderIconButtonComponent
            :disabled="loading"
            :description="`Change speed (${speed}x)`"
            icon="mdi-fast-forward"
            @click="speedsCycle.setNextValue"
          ></HeaderIconButtonComponent>

          <HeaderRestartConfirmationDialogComponent
            :disabled="loading"
            @click="stop"
            @reset="reset"
          ></HeaderRestartConfirmationDialogComponent>
        </div>
      </v-col>

      <v-col cols="2" class="text-center text-capitalize">
        <span v-if="!loading">
          Current time <br />
          {{ new Time(props.time).toFullString() }}
        </span>
      </v-col>

      <v-col cols="2" class="text-center text-capitalize">
        <span v-if="!loading">
          Elapsed time <br />
          {{ new Time(Math.floor(timer.timer.value)).toFullString() }}
        </span>
      </v-col>

      <v-col cols="4">
        <div class="d-flex justify-end align-center h-100">
          <v-btn
            :disabled="isRunning"
            text="Export"
            class="mx-1"
            prepend-icon="mdi-file-export"
            @click="exportData"
          ></v-btn>

          <v-btn
            text="Menu"
            class="mx-1"
            prepend-icon="mdi-backburger"
            @click="menu"
          ></v-btn>
        </div>
      </v-col>
    </v-row>
  </v-footer>
</template>

<style lang="scss" scoped>
.button-box {
  display: flex;
  justify-content: left;
  align-items: center;
  height: 100%;
  gap: 10px;
}
</style>
