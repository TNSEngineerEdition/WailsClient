<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"
import useTimeUtils from "@composables/useTimeUtils"
import { simulation } from "@wails/go/models"
import { GetTramDetails } from "@wails/go/simulation/Simulation"
import { ref, watch } from "vue"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  tramID?: number,
  currentTime: number
}>()

const timeUtils = useTimeUtils()

const tramDetails = ref<simulation.TramDetails>()

function formatTime(time: number) {
  const t = timeUtils.toTimeString(time).split(':')
  return `${t[0]}:${t[1]}`
}

watch(
  () => props.tramID,
  async (id) => {
    if (id) {
      tramDetails.value = await GetTramDetails(id)
    } else {
      tramDetails.value = undefined
      model.value = false
    }
  },
  { immediate: true }
)

watch(
  () => props.currentTime,
  async (time) => {
    if (props.tramID) {
      tramDetails.value = await GetTramDetails(props.tramID)
    }
  },
  { immediate: true }
)
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="left"
    :title="tramDetails ? tramDetails.route + ' ➡ ' + tramDetails.trip_head_sign : 'Loading data...'"
    title-icon="mdi-tram"
    style="overflow-y: auto"
  >
    <template v-if="props.tramID !== undefined && tramDetails !== undefined">
      <div class="sidebarDetails">
        <span><b>TramID</b>: {{ props.tramID }}</span>
        <span><b>Speed</b>: {{ tramDetails.speed }} km/h</span>

        <div class="scrollable mt-4 ml-4">
          <div class="stopDiv">
            <b>Stop:</b>
            <b>Scheduled time:</b>
          </div>
          <ul>
            <li
              v-for="(stop, index) in tramDetails.stop_names"
              :key="index"
              :class="{ currentStop: index === tramDetails.trip_index }"
            >
              <div class="stopDiv">
                <span>{{ stop }}</span>
                <span>{{ formatTime(tramDetails.stops[index].time) }}</span>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </template>
  </SidebarComponent>
</template>

<style scoped lang="scss">
.sidebarDetails {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.scrollable {
  overflow-y: auto;
  max-height: 60vh;
  padding-right: 8px;
}

.scrollable::-webkit-scrollbar {
  width: 6px;
}
.scrollable::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.15);
  border-radius: 3px;
}

.currentStop {
  font-weight: bold;
  color: #007bff;
}

.stopDiv {
  width: 100%;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}
</style>
