<script lang="ts" setup>
import { nextTick, onMounted, ref, useTemplateRef, watch } from "vue"
import { GetTimeBounds } from "@wails/go/city/City"
import { city, api, vehicle } from "@wails/go/models"
import {
  GetVehicleIDs,
  AdvanceVehicles,
  ResetSimulation,
} from "@wails/go/simulation/Simulation"
import { LeafletMap } from "@classes/LeafletMap"
import { VehicleMarker } from "@classes/VehicleMarker"
import { Time } from "@classes/Time"
import VehicleSidebarComponent from "@components/simulation/sidebar/VehicleSidebarComponent.vue"
import StopSidebarComponent from "@components/simulation/sidebar/StopSidebarComponent.vue"
import RouteSidebarComponent from "@components/simulation/sidebar/RouteSidebarComponent.vue"
import { MarkerColoringMode } from "@utils/types"

const mapHTMLElement = useTemplateRef("map")

const time = defineModel<number>("time", { required: true })
const loading = defineModel<boolean>("loading", { required: true })
const isRunning = defineModel<boolean>("is-running", { required: true })

const props = defineProps<{
  speed: number
  resetCounter: number
  markerColoringMode: MarkerColoringMode
}>()

const endTime = ref(0)
const leafletMap = ref<LeafletMap>()
const vehicleMarkerByID = ref<Record<number, VehicleMarker>>({})

const vehicleSidebar = ref(false)
const stopSidebar = ref(false)
const routeSidebar = ref(false)

const selectedVehicleID = ref<number>()
const followVehicle = ref(false)
const selectedStop = ref<api.ResponseGraphStop>()
const selectedRoute = ref<city.RouteInfo>()

async function setTime() {
  await GetTimeBounds().then(timeBounds => {
    time.value = timeBounds.startTime
    endTime.value = timeBounds.endTime
  })
}

async function reset() {
  vehicleSidebar.value = false
  stopSidebar.value = false
  routeSidebar.value = false
  loading.value = true

  for (const vehicleMarker of Object.values(vehicleMarkerByID.value)) {
    vehicleMarker.removeFromMap()
  }

  vehicleMarkerByID.value = await GetVehicleIDs().then(vehicles =>
    leafletMap.value!.getVehicleMarkers(vehicles, (id: number) => {
      selectedVehicleID.value = id
      vehicleSidebar.value = true
    }),
  )

  setTime()

  loading.value = false
}

function handleRouteSelected(route: city.RouteInfo) {
  selectedRoute.value = route
  routeSidebar.value = true
  const vehicleMarkersForRoute = Object.values(vehicleMarkerByID.value).filter(
    m => m.getRoute() === route.name,
  )
  leafletMap.value?.highlightVehiclesForRoute(vehicleMarkersForRoute)
  leafletMap.value?.highlightRoute(route)
}

function handleArrivalSelected(vehicleId: number) {
  if (leafletMap.value?.selectedVehicle)
    leafletMap.value.selectedVehicle.setSelected(false)
  leafletMap.value!.selectedVehicle = vehicleMarkerByID.value[vehicleId]
  leafletMap.value!.selectedVehicle.setSelected(true)
  selectedVehicleID.value = vehicleId
  vehicleSidebar.value = true
}

function handleStopSelected(stopId: number) {
  if (leafletMap.value?.selectedStop) {
    leafletMap.value.selectedStop.setSelected(false)
  }
  leafletMap.value!.selectedStop = leafletMap.value!.getStopMarker(stopId)
  leafletMap.value!.selectedStop.setSelected(true)
  selectedStop.value = leafletMap.value!.getStopMarker(stopId).getStop()
  stopSidebar.value = true
}

function handleCenterStop() {
  if (selectedStop.value && leafletMap.value) {
    handleFollowVehicle(false)
    leafletMap.value.centerOn(selectedStop.value.lat, selectedStop.value.lon)
  }
}

function handleCenterVehicle() {
  if (selectedVehicleID.value && leafletMap.value) {
    const vehicleMarker = vehicleMarkerByID.value[selectedVehicleID.value]
    const wasFollowing = followVehicle.value
    handleFollowVehicle(false)
    leafletMap.value.centerOn(
      vehicleMarker.getLatLng().lat,
      vehicleMarker.getLatLng().lng,
    )
    leafletMap.value?.getMap().once?.("moveend", () => {
      handleFollowVehicle(wasFollowing)
    })
  }
}

function handleFollowVehicle(value: boolean) {
  followVehicle.value = value
  leafletMap.value?.setFollowVehicle(value)
}

watch(() => props.resetCounter, reset)

watch(stopSidebar, isOpen => {
  if (!isOpen) {
    leafletMap.value?.deselectStop()
    selectedStop.value = undefined
  }
})

watch(vehicleSidebar, isOpen => {
  if (!isOpen) {
    leafletMap.value?.deselectVehicle()
    selectedVehicleID.value = undefined
    followVehicle.value = false
    leafletMap.value?.setFollowVehicle(false)
  }
})

watch(routeSidebar, isOpen => {
  if (!isOpen) {
    leafletMap.value?.deselectRoute()
    selectedRoute.value = undefined
  }
})

watch(selectedRoute, route => {
  if (route) {
    const vehicleMarkersForRoute = Object.values(
      vehicleMarkerByID.value,
    ).filter(m => m.getRoute() === route.name)
    leafletMap.value?.highlightVehiclesForRoute(vehicleMarkersForRoute)
  }
})

watch(
  () => props.markerColoringMode,
  mode => {
    Object.values(vehicleMarkerByID.value).forEach(vehicleMarker =>
      vehicleMarker.removeCustomColoring(),
    )
    VehicleMarker.coloringMode = mode
  },
)

onMounted(async () => {
  if (mapHTMLElement.value === null) {
    throw new Error("Map element not found")
  }

  leafletMap.value = await LeafletMap.init(mapHTMLElement.value, stop => {
    selectedStop.value = stop
    stopSidebar.value = true
  })

  await reset()

  while (true) {
    await ResetSimulation()
    await setTime()

    while (
      time.value <= endTime.value ||
      leafletMap.value!.getEntityCount() > 0
    ) {
      while (!isRunning.value) {
        await Time.sleep(1)
      }

      for (const vehiclePositionChange of await AdvanceVehicles(time.value)) {
        if (vehiclePositionChange.lat == 0 && vehiclePositionChange.lon == 0) {
          vehicleMarkerByID.value[vehiclePositionChange.id].removeFromMap()
          continue
        }

        const isStopped =
          vehiclePositionChange.state === vehicle.VehicleState.STOPPED ||
          vehiclePositionChange.state === vehicle.VehicleState.STOPPING

        vehicleMarkerByID.value[vehiclePositionChange.id].updateCoordinates(
          vehiclePositionChange.lat,
          vehiclePositionChange.lon,
          vehiclePositionChange.azimuth,
          isStopped,
          vehiclePositionChange.delay,
        )
      }
      leafletMap.value?.followTick()

      time.value += 1

      await Time.sleep(1000 / props.speed)
    }

    isRunning.value = false
    await nextTick() // Wait until assignment is finished

    while (!isRunning.value) {
      await Time.sleep(1)
    }
  }
})
</script>

<template>
  <v-overlay
    v-model="loading"
    opacity="0"
    class="d-flex justify-center align-center"
    persistent
    contained
  >
    <v-progress-circular indeterminate size="128"></v-progress-circular>
  </v-overlay>

  <div id="map" ref="map"></div>

  <div class="sidebar-stack left">
    <VehicleSidebarComponent
      v-model="vehicleSidebar"
      :vehicle-id="selectedVehicleID"
      :vehicle-marker="
        selectedVehicleID ? vehicleMarkerByID[selectedVehicleID] : undefined
      "
      :current-time="time"
      :follow-vehicle="followVehicle"
      @stopSelected="handleStopSelected"
      @centerVehicle="handleCenterVehicle"
      @followVehicle="handleFollowVehicle"
    />
  </div>
  <div class="sidebar-stack right">
    <StopSidebarComponent
      v-model="stopSidebar"
      :stop="selectedStop"
      :current-time="time"
      @routeSelected="handleRouteSelected"
      @arrivalSelected="handleArrivalSelected"
      @centerStop="handleCenterStop"
    />
    <RouteSidebarComponent
      v-model="routeSidebar"
      :route="selectedRoute"
      :vehicle-markers="vehicleMarkerByID"
      :current-time="time"
    />
  </div>
</template>

<style lang="scss">
#map {
  width: 100%;
  height: calc(100vh - 64px);
}

.vehicle-marker {
  position: relative;
  width: 24px;
  height: 24px;
  pointer-events: auto;
  transition:
    transform 0.2s ease,
    background-color 0.3s ease;
}

.vehicle-marker.highlighted {
  transform: scale(1.1);
}

.vehicle-marker.selected {
  transform: scale(1.2);
}

.vm-circle-arrow {
  position: absolute;
  width: 24px;
  height: 24px;
  background-color: #2896f1;
  border-radius: 50% 50% 50% 0%;
  z-index: 1;
}

.vm-circle {
  position: absolute;
  width: 18px;
  height: 18px;
  top: 3px;
  left: 3px;
  background-color: #2896f1;
  border-radius: 50%;
  z-index: 2;
}

.vm-route-label {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) rotate(0deg);
  font-size: 12px;
  font-weight: bold;
  color: white;
  pointer-events: none;
  user-select: none;
  z-index: 3;
}

.vehicle-marker.selected .vm-circle-arrow,
.vehicle-marker.selected .vm-circle {
  background-color: #67ad2f;
}

.vehicle-marker.highlighted .vm-circle-arrow,
.vehicle-marker.highlighted .vm-circle {
  background-color: orange !important;
}

@keyframes pulse-red {
  0% {
    box-shadow: 0 0 0 0 rgba(255, 0, 0, 0.6);
  }
  70% {
    box-shadow: 0 0 0 10px rgba(255, 0, 0, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(255, 0, 0, 0);
  }
}

.vehicle-marker.stopped .vm-circle-arrow,
.vehicle-marker.stopped .vm-circle {
  background-color: red !important;
  animation: pulse-red 1.5s infinite;
}

@keyframes pulse-red-selected {
  0% {
    box-shadow: 0 0 0 0 rgba(255, 0, 0, 0.6);
    background-color: red;
  }
  50% {
    box-shadow: 0 0 10px 4px rgba(255, 0, 0, 0.8);
    background-color: red;
  }
  100% {
    box-shadow: 0 0 0 0 rgba(255, 0, 0, 0.6);
    background-color: red;
  }
}

.vehicle-marker.stopped.selected .vm-circle-arrow,
.vehicle-marker.stopped.selected .vm-circle {
  background-color: red;
  animation: pulse-red-selected 1.5s infinite;
}

.sidebar-stack {
  position: fixed;
  top: calc(60px + 20px);
  z-index: 1001;
  display: flex;
  flex-direction: column;
  gap: 12px;
  pointer-events: none;
}
.sidebar-stack > * {
  pointer-events: auto;
}

.sidebar-stack.left {
  left: 54px;
}
.sidebar-stack.right {
  right: 20px;
}
</style>
