<script setup lang="ts">
import SidebarComponent from "@components/simulation/sidebar/SidebarComponent.vue"
import { ref, watch } from "vue"
import { city, simulation, api } from "@wails/go/models"
import { GetRoutesForStop } from "@wails/go/city/City"
import {
  GetArrivalsForStop,
  GetPassengerCountAtStop,
} from "@wails/go/simulation/Simulation"

const ARRIVALS_IN_TABLE = 5

const headers = [
  { title: "Route", key: "route", align: "center", sortable: false },
  {
    title: "Trip head-sign",
    key: "tripHeadSign",
    align: "center",
    sortable: false,
  },
  { title: "ETA", key: "time", align: "center", sortable: false },
] as const

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  stop?: api.ResponseGraphTramStop
  currentTime: number
}>()

const routes = ref<city.RouteInfo[]>([])
const arrivalsInfo = ref<simulation.Arrival[]>([])
const routeChipColumns = ref(5)
const tab = ref<"arr" | "occ">("arr")
const passengerCount = ref(0)

const emit = defineEmits(["routeSelected"])

function onChipClick(route: city.RouteInfo) {
  emit("routeSelected", route)
}

watch(
  () => props.stop?.id,
  async id => {
    if (id) {
      routes.value = await GetRoutesForStop(id, routeChipColumns.value)
      arrivalsInfo.value = await GetArrivalsForStop(id, ARRIVALS_IN_TABLE)
    } else {
      routes.value = []
      arrivalsInfo.value = []
    }
  },
  { immediate: true },
)

watch(
  [() => props.stop?.id, () => props.currentTime],
  async ([id]) => {
    if (id) {
      passengerCount.value = await GetPassengerCountAtStop(id)
    } else {
      passengerCount.value = 0
    }
  },
  { immediate: true },
)

watch(
  () => props.currentTime,
  async () => {
    if (props.stop?.id) {
      arrivalsInfo.value = await GetArrivalsForStop(
        props.stop.id,
        ARRIVALS_IN_TABLE,
      )
    }
  },
  { immediate: true },
)
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="right"
    :title="props.stop?.name ?? 'Unknown stop'"
    title-icon="mdi-tram-side"
  >
    <div class="section">
      <div class="label">
        <v-icon icon="mdi-map-marker" class="mr-2"></v-icon>
        Coordinates
      </div>

      <div class="value">
        {{ stop?.lat.toFixed(6) }}, {{ stop?.lon.toFixed(6) }}
      </div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-zip-box" class="mr-2"></v-icon>
        GTFS ID
      </div>

      <div class="value">
        <span v-if="stop?.gtfs_stop_ids?.length">
          {{ stop.gtfs_stop_ids.join(", ") }}
        </span>
        <span v-else>Unknown</span>
      </div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-transit-connection-variant" class="mr-2"></v-icon>
        Routes
      </div>

      <div class="value">
        <div v-if="routes.length" class="route-chips">
          <span
            v-for="route in routes"
            class="chip"
            :style="{
              color: route.text_color,
              backgroundColor: route.background_color,
            }"
            @click="onChipClick(route)"
          >
            {{ route.name }}
          </span>
        </div>

        <span v-else>No routes</span>
      </div>
    </div>
    <div class="section">
      <div class="label">
        <v-icon icon="mdi-bus-stop-covered" class="mr-2"></v-icon>
        Passengers at stop
      </div>

      <div class="value">
        <span>{{ passengerCount }}</span>
      </div>
    </div>

    <v-tabs v-model="tab" grow>
      <v-tab value="arr">Arrivals</v-tab>
      <!-- <v-tab value="occ">Occupancy graph</v-tab> -->
    </v-tabs>

    <v-card-text>
      <v-tabs-window v-model="tab">
        <v-tabs-window-item value="arr">
          <v-data-table
            v-if="arrivalsInfo.length && tab === 'arr'"
            :headers="headers"
            :header-props="{
              style: 'font-weight: bold;',
            }"
            :items="arrivalsInfo"
            class="stops-table"
            density="compact"
            hide-default-footer
            hover
          >
            <template v-slot:item.time="{ item }">
              <span v-if="item.time === 0" class="blinking">
                &gt;&gt;&gt;
              </span>

              <span v-else>{{ item.time }} min</span>
            </template>
          </v-data-table>
        </v-tabs-window-item>

        <!-- <v-tabs-window-item value="occ">
          Occupancy graph TODO
        </v-tabs-window-item> -->
      </v-tabs-window>
    </v-card-text>
  </SidebarComponent>
</template>

<style scoped lang="scss">
.route-chips {
  display: grid;
  gap: 4px;
  grid-template-columns: repeat(v-bind(routeChipColumns), 1fr);
  direction: rtl;
}

.chip {
  padding: 3px 5px;
  border-radius: 4px;
  font-size: 0.85rem;
  line-height: 1;
  transition:
    background 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
  text-align: center;

  &:hover {
    background: #2896f1;
    cursor: pointer;
    transform: scale(1.1);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    z-index: 10;
  }
}

.blinking {
  animation: smooth-blink 1s ease-in-out infinite;
}

@keyframes smooth-blink {
  25%,
  75% {
    opacity: 1;
  }
  50% {
    opacity: 0;
  }
}

.stops-table {
  width: 100%;
  background-color: transparent;
}
</style>
