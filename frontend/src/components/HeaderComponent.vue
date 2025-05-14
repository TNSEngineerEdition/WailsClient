<script lang="ts" setup>
import useTimeUtils from "@composables/useTimeUtils"
import HeaderIconButtonComponent from "@components/HeaderIconButtonComponent.vue"
import useCycle from "@composables/useCycle"
import { Reset } from "@wails/go/simulation/Simulation"

const props = defineProps<{
  time: number
  loading: boolean
}>()

const isRunning = defineModel<boolean>("is-running", { required: true })
const speed = defineModel<number>("speed", { required: true })
const resetCounter = defineModel<number>("reset-counter", { required: true })

const speedsCycle = useCycle([1, 10, 100, 1000], speed)

const timeUtils = useTimeUtils()

async function reset() {
  isRunning.value = false
  speedsCycle.reset()
  await Reset()
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
            @click="isRunning = !isRunning"
          ></HeaderIconButtonComponent>

          <HeaderIconButtonComponent
            :disabled="loading"
            :description="`Change speed (${speed}x)`"
            icon="mdi-fast-forward"
            @click="speedsCycle.setNextValue"
          ></HeaderIconButtonComponent>

          <HeaderIconButtonComponent
            :disabled="loading"
            description="Restart"
            icon="mdi-replay"
            @click="reset"
          ></HeaderIconButtonComponent>
        </div>
      </v-col>

      <v-col cols="2" class="text-center text-capitalize">
        Current time <br />
        {{ timeUtils.toTimeString(props.time) }}
      </v-col>

      <v-col cols="2" class="text-center text-capitalize">
        Elapsed time <br />
        {{ timeUtils.toTimeString(0) }}
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
