<script setup lang="ts">
import SidebarComponent from "@components/simulation/sidebar/SidebarComponent.vue"
import { ref, computed, watch } from "vue"
import { city } from "@wails/go/models"
import { VehicleMarker } from "@classes/VehicleMarker"
import { GetPassengerCountOnRoute } from "@wails/go/simulation/Simulation"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  route?: city.RouteInfo
  vehicleMarkers: Record<number, VehicleMarker>
  currentTime: number
}>()

const passengersOnRoute = ref(0)
const vehiclesInService = computed(() => {
  if (!props.route?.name) return 0
  return Object.values(props.vehicleMarkers).filter(
    vehicle => vehicle.getRoute() === props.route!.name && vehicle.getIsOnMap(),
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
        Vehicles in service
      </div>

      <div class="value">
        {{ vehiclesInService?.valueOf() || 0 }}
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
  </SidebarComponent>
</template>
