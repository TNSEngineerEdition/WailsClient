<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"
import useTimeUtils from "@composables/useTimeUtils"
import { simulation } from "@wails/go/models"
import { GetTramDetails } from "@wails/go/simulation/Simulation"
import { computed, ref, watch } from "vue"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  tramId?: number
  currentTime: number
}>()

const timeUtils = useTimeUtils()

const tramDetails = ref<simulation.TramDetails>()

const headers = [
  { title: "Stop name", key: "stop", align: "center", sortable: false },
  { title: "Departure", key: "time", align: "center", sortable: false },
  { title: "Delay", key: "delay", align: "center", sortable: false },
] as const

const formattedStops = computed(
  () =>
    tramDetails.value?.stop_names.map((stop, index) => {
      const stopTime = tramDetails.value?.stops[index]?.time
      return {
        stop,
        time:
          stopTime !== undefined
            ? timeUtils.toShortTimeString(stopTime)
            : "Unknown",
        delay: "00:00",
      }
    }) ?? [],
)

function getRowProps(data: any) {
  if (data.index === tramDetails.value?.trip_index)
    return {
      style:
        "background-color: rgba(40, 150, 241, 0.2); transition: background-color 0.3s ease, font-weight 0.3s ease;",
    }
  else
    return {
      style: "transition: background-color 0.3s ease;",
    }
}

watch(
  () => props.tramId,
  async id => {
    if (id) {
      tramDetails.value = await GetTramDetails(id)
    } else {
      tramDetails.value = undefined
      model.value = false
    }
  },
  { immediate: true },
)

watch(
  () => props.currentTime,
  async () => {
    if (props.tramId) {
      tramDetails.value = await GetTramDetails(props.tramId)
    }
  },
  { immediate: true },
)
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="left"
    :title="
      tramDetails
        ? tramDetails.route + ' ➡ ' + tramDetails.trip_head_sign
        : 'Loading data...'
    "
    title-icon="mdi-tram"
  >
    <div class="section">
      <div class="label">
        <v-icon icon="mdi-identifier" class="mr-2"></v-icon>
        Tram ID
      </div>
      <div class="value">{{ props.tramId }}</div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-speedometer" class="mr-2"></v-icon>
        Speed
      </div>
      <div class="value">{{ tramDetails?.speed }} km/h</div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-map-marker-path" class="mr-2"></v-icon>
        Stops
      </div>
    </div>

    <div class="scrollable">
      <v-data-table-virtual
        v-if="tramDetails?.stop_names.length"
        :headers="headers"
        :header-props="{
          style: 'font-weight: bold;',
        }"
        :items="formattedStops"
        :row-props="getRowProps"
        class="stops-table"
        density="compact"
        hide-default-footer
        hover
      >
        <template v-slot:item.stop="{ item }">
          {{ item.stop }}
        </template>

        <template v-slot:item.time="{ item }">
          {{ item.time }}
        </template>
      </v-data-table-virtual>
    </div>
  </SidebarComponent>
</template>

<style scoped lang="scss">
.scrollable {
  overflow-y: auto;
  max-height: 60vh;
}

.scrollable::-webkit-scrollbar {
  width: 6px;
}

.scrollable::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

.section {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.4rem;
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

.stops-table {
  width: 100%;
  background-color: transparent;
}
</style>
