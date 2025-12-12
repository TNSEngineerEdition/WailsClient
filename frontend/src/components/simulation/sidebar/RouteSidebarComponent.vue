<script setup lang="ts">
import SidebarComponent from "@components/simulation/sidebar/SidebarComponent.vue"
import { ref, computed, watch } from "vue"
import { city } from "@wails/go/models"
import { TramMarker } from "@classes/TramMarker"
import { GetPassengerCountOnRoute } from "@wails/go/simulation/Simulation"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  route?: city.RouteInfo
  tramMarkers: Record<number, TramMarker>
  currentTime: number
}>()

const tab = ref<"basic" | "occ" | "stops">("basic")
const passengersOnRoute = ref(0)
const tramsInService = computed(() => {
  if (!props.route?.name) return 0
  return Object.values(props.tramMarkers).filter(
    tram => tram.getRoute() === props.route!.name && tram.getIsOnMap(),
  ).length
})

watch(
  [() => props.route?.name, () => props.currentTime],
  async ([routeName]) => {
    if (routeName) {
      passengersOnRoute.value = await GetPassengerCountOnRoute(routeName)
    } else {
      passengersOnRoute.value = 0
    }
  },
  { immediate: true },
)
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="right"
    :title="'Route ' + (route?.name ?? 'Unknown route')"
    title-icon="mdi-transit-connection-horizontal"
  >
    <div class="section">
      <div class="label">
        <v-icon icon="mdi-numeric" class="mr-2"></v-icon>
        Trams in service
      </div>

      <div class="value">
        {{ tramsInService?.valueOf() || 0 }}
      </div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-counter" class="mr-2"></v-icon>
        Passengers on route
      </div>

      <div class="value">
        <span>{{ passengersOnRoute }}</span>
      </div>
    </div>

    <!-- <div class="section">
      <div class="label">
        <v-icon icon="mdi-wrench-check" class="mr-2"></v-icon>
        Route status
      </div>

      <div class="value">
        <span> TODO </span>
      </div>
    </div> -->
    <!-- <v-tabs v-model="tab" grow>
      <v-tab value="basic">Basic information</v-tab>
      <v-tab value="occ">Occupancy graph</v-tab>
      <v-tab value="stops">Stops list</v-tab>
    </v-tabs>

    <v-card-text>
      <v-tabs-window v-model="tab">
        <v-tabs-window-item value="basic">
          Basic information TODO
        </v-tabs-window-item>

        <v-tabs-window-item value="occ">
          Occupancy graph TODO
        </v-tabs-window-item>

        <v-tabs-window-item value="stops"> Stops list TODO </v-tabs-window-item>
      </v-tabs-window>
    </v-card-text> -->
  </SidebarComponent>
</template>
