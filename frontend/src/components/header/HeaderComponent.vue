<script lang="ts" setup>
import { Time } from "@classes/Time"
import HeaderIconButtonComponent from "@components/header/HeaderIconButtonComponent.vue"
import useCycle from "@composables/useCycle"
import { ResetSimulation } from "@wails/go/simulation/Simulation"
import HeaderRestartConfirmationDialogComponent from "@components/header/HeaderRestartConfirmationDialogComponent.vue"
import useTimer from "@composables/useTimer"

const props = defineProps<{
  time: number
  loading: boolean
}>()

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
        Current time <br />
        {{ new Time(props.time).toFullString() }}
      </v-col>

      <v-col cols="2" class="text-center text-capitalize">
        Elapsed time <br />
        {{ new Time(timer.timer.value).toFullString() }}
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
