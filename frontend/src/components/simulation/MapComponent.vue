<script lang="ts" setup>
import { onMounted, ref, useTemplateRef, watch } from "vue"
import { GetTimeBounds } from "@wails/go/city/City"
import { city, api, tram } from "@wails/go/models"
import { GetTramIDs, AdvanceTrams } from "@wails/go/simulation/Simulation"
import { LeafletMap } from "@classes/LeafletMap"
import { TramMarker } from "@classes/TramMarker"
import { Time } from "@classes/Time"
import TramSidebarComponent from "@components/simulation/sidebar/TramSidebarComponent.vue"
import StopSidebarComponent from "@components/simulation/sidebar/StopSidebarComponent.vue"
import RouteSidebarComponent from "@components/simulation/sidebar/RouteSidebarComponent.vue"

const mapHTMLElement = useTemplateRef("map")

const time = defineModel<number>("time", { required: true })
const loading = defineModel<boolean>("loading", { required: true })
const isRunning = defineModel<boolean>("is-running", { required: true })

const props = defineProps<{
  speed: number
  resetCounter: number
}>()

const endTime = ref(0)
const leafletMap = ref<LeafletMap>()
const tramMarkerByID = ref<Record<number, TramMarker>>({})

const tramSidebar = ref(false)
const stopSidebar = ref(false)
const routeSidebar = ref(false)

const selectedTramID = ref<number>()
const selectedStop = ref<api.ResponseGraphTramStop>()
const selectedRoute = ref<city.RouteInfo>()

async function reset() {
  tramSidebar.value = false
  stopSidebar.value = false
  routeSidebar.value = false
  loading.value = true

  for (const tramMarker of Object.values(tramMarkerByID.value)) {
    tramMarker.removeFromMap()
  }

  tramMarkerByID.value = await GetTramIDs().then(trams =>
    leafletMap.value!.getTramMarkers(trams, (id: number) => {
      selectedTramID.value = id
      tramSidebar.value = true
    }),
  )

  await GetTimeBounds().then(timeBounds => {
    time.value = timeBounds.startTime
    endTime.value = timeBounds.endTime
  })

  loading.value = false
}

function handleRouteSelected(route: city.RouteInfo) {
  selectedRoute.value = route
  routeSidebar.value = true
  const tramMarkersForRoute = Object.values(tramMarkerByID.value).filter(
    m => m.getRoute() === route.name,
  )
  leafletMap.value?.highlightTramsForRoute(tramMarkersForRoute)
  leafletMap.value?.highlightRoute(route)
}

watch(() => props.resetCounter, reset)

watch(stopSidebar, isOpen => {
  if (!isOpen) {
    leafletMap.value?.deselectStop()
    selectedStop.value = undefined
  }
})

watch(tramSidebar, isOpen => {
  if (!isOpen) {
    leafletMap.value?.deselectTram()
    selectedTramID.value = undefined
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
    const tramMarkersForRoute = Object.values(tramMarkerByID.value).filter(
      m => m.getRoute() === route.name,
    )
    leafletMap.value?.highlightTramsForRoute(tramMarkersForRoute)
  }
})

onMounted(async () => {
  if (mapHTMLElement.value === null) {
    throw new Error("Map element not found")
  }

  leafletMap.value = await LeafletMap.init(mapHTMLElement.value, stop => {
    selectedStop.value = stop
    stopSidebar.value = true
  })

  await reset()

  while (
    time.value <= endTime.value ||
    leafletMap.value!.getEntityCount() > 0
  ) {
    while (!isRunning.value) {
      await Time.sleep(1)
    }

    for (const tramPositionChange of await AdvanceTrams(time.value)) {
      if (tramPositionChange.lat == 0 && tramPositionChange.lon == 0) {
        tramMarkerByID.value[tramPositionChange.id].removeFromMap()
        continue
      }

      const isStopped =
        tramPositionChange.state === tram.TramState.STOPPED ||
        tramPositionChange.state === tram.TramState.STOPPING
      tramMarkerByID.value[tramPositionChange.id].updateCoordinates(
        tramPositionChange.lat,
        tramPositionChange.lon,
        tramPositionChange.azimuth,
        isStopped,
      )
    }

    time.value += 1

    await Time.sleep(1000 / props.speed)
  }

  isRunning.value = false
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
    <TramSidebarComponent
      v-model="tramSidebar"
      :tram-id="selectedTramID"
      :tram-marker="selectedTramID ? tramMarkerByID[selectedTramID] : undefined"
      :current-time="time"
    />
  </div>
  <div class="sidebar-stack right">
    <StopSidebarComponent
      v-model="stopSidebar"
      :stop="selectedStop"
      :current-time="time"
      @routeSelected="handleRouteSelected"
    />
    <RouteSidebarComponent
      v-model="routeSidebar"
      :route="selectedRoute"
      :tram-markers="tramMarkerByID"
      :current-time="time"
    />
  </div>
</template>

<style lang="scss">
#map {
  width: 100%;
  height: calc(100vh - 64px);
}

.tram-marker {
  position: relative;
  width: 24px;
  height: 24px;
  pointer-events: auto;
  transition:
    transform 0.2s ease,
    background-color 0.3s ease;
}

.tram-marker.highlighted {
  transform: scale(1.1);
}

.tram-marker.selected {
  transform: scale(1.2);
}

.tm-circle-arrow {
  position: absolute;
  width: 24px;
  height: 24px;
  background-color: #2896f1;
  border-radius: 50% 50% 50% 0%;
  z-index: 1;
}

.tm-circle {
  position: absolute;
  width: 18px;
  height: 18px;
  top: 3px;
  left: 3px;
  background-color: #2896f1;
  border-radius: 50%;
  z-index: 2;
}

.tm-route-label {
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

.tram-marker.highlighted .tm-circle-arrow,
.tram-marker.highlighted .tm-circle {
  background-color: orange;
}

.tram-marker.selected .tm-circle-arrow,
.tram-marker.selected .tm-circle {
  background-color: #67ad2f;
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

.tram-marker.stopped .tm-circle-arrow,
.tram-marker.stopped .tm-circle {
  background-color: red;
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

.tram-marker.stopped.selected .tm-circle-arrow,
.tram-marker.stopped.selected .tm-circle {
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
