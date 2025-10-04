<script setup lang="ts">
import SidebarComponent from "@components/simulation/sidebar/SidebarComponent.vue"
import { Time } from "@classes/Time"
import { simulation } from "@wails/go/models"
import { GetTramDetails } from "@wails/go/simulation/Simulation"
import { computed, ref, watch } from "vue"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  tramId?: number
  currentTime: number
}>()

const tramDetails = ref<simulation.TramDetails>()
const tab = ref<"stops" | "occ" | "delay">("stops")

const headers = [
  { title: "Stop name", key: "stop", align: "center", sortable: false },
  { title: "Departure", key: "time", align: "center", sortable: false },
  { title: "Arrival", key: "arrival", align: "center", sortable: false },
  { title: "Departure", key: "departure", align: "center", sortable: false },
] as const

const stopsTableData = computed(
  () =>
    tramDetails.value?.stop_names.map((stop, index) => {
      const time = tramDetails.value?.stops[index]?.time ?? 0
      const tripIndex = tramDetails.value?.trip_index ?? 0

      return {
        stop,
        time,
        arrival:
          index <= tripIndex
            ? (tramDetails.value?.arrivals[index] ?? 0) - time
            : null,
        departure:
          index <= tripIndex - 1
            ? (tramDetails.value?.departures[index] ?? 0) - time
            : null,
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

function getDelayTextColorClass(delay: number) {
  if (delay > 0) {
    return "text-red font-weight-bold"
  } else if (delay < 0) {
    return "text-info font-weight-bold"
  } else {
    return ""
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
        ? `${tramDetails.route} ➡ ${tramDetails.trip_head_sign}`
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
    <v-tabs v-model="tab" grow>
      <v-tab value="stops">Stops table</v-tab>
      <v-tab value="occ">Occupancy graph</v-tab>
      <v-tab value="delay">Delay graph</v-tab>
    </v-tabs>

    <v-card-text>
      <v-tabs-window v-model="tab">
        <v-tabs-window-item value="stops">
          <div class="scrollable">
            <v-data-table-virtual
              v-if="tramDetails?.stop_names.length"
              :headers="headers"
              :header-props="{
                style: 'font-weight: bold;',
              }"
              :items="stopsTableData"
              :row-props="getRowProps"
              class="stops-table"
              density="compact"
              hide-default-footer
              hover
            >
              <template v-slot:item.time="{ item }">
                {{ new Time(item.time).toShortMinuteString() }}
              </template>

              <template v-slot:item.arrival="{ item }">
                <span
                  v-if="item.arrival != null"
                  :class="getDelayTextColorClass(item.arrival)"
                >
                  {{ new Time(item.arrival, true).toShortSecondString() }}
                </span>
              </template>

              <template v-slot:item.departure="{ item }">
                <span
                  v-if="item.departure != null"
                  :class="getDelayTextColorClass(item.departure)"
                >
                  {{ new Time(item.departure, true).toShortSecondString() }}
                </span>
              </template>
            </v-data-table-virtual>
          </div>
        </v-tabs-window-item>

        <v-tabs-window-item value="occ">
          Occupancy graph TODO
        </v-tabs-window-item>

        <v-tabs-window-item value="delay">
          Delay graph TODO
        </v-tabs-window-item>
      </v-tabs-window>
    </v-card-text>
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

.stops-table {
  width: 100%;
  background-color: transparent;
}
</style>
