<script lang="ts" setup>
import { onMounted, ref, useTemplateRef, watch } from "vue"
import { GetTimeBounds } from "@wails/go/city/City"
import { city } from "@wails/go/models"
import {
  GetTramIDs,
  AdvanceTrams,
  FetchData,
} from "@wails/go/simulation/Simulation"
import { LeafletMap } from "@classes/LeafletMap"
import { TramMarker } from "@classes/TramMarker"
import useTimeUtils from "@composables/useTimeUtils"
import TramSidebarComponent from "@components/sidebar/TramSidebarComponent.vue"
import StopSidebarComponent from "@components/sidebar/StopSidebarComponent.vue"

const mapHTMLElement = useTemplateRef("map")

const time = defineModel<number>("time", { required: true })
const loading = defineModel<boolean>("loading", { required: true })

const props = defineProps<{
  speed: number
  isRunning: boolean
  resetCounter: number
}>()

const endTime = ref(0)
const leafletMap = ref<LeafletMap>()
const tramMarkerByID = ref<Record<number, TramMarker>>({})

const tramSidebar = ref(false)
const stopSidebar = ref(false)

const selectedTramID = ref<number>()
const selectedStop = ref<city.GraphNode>()

const timeUtils = useTimeUtils()

async function reset() {
  tramSidebar.value = false
  stopSidebar.value = false
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

onMounted(async () => {
  if (mapHTMLElement.value === null) {
    throw new Error("Map element not found")
  }

  await FetchData("http://localhost:8000/cities/krakow")
  leafletMap.value = await LeafletMap.init(mapHTMLElement.value, stop => {
    selectedStop.value = stop
    stopSidebar.value = true
  })

  await reset()

  while (
    time.value <= endTime.value ||
    leafletMap.value!.getEntityCount() > 0
  ) {
    while (!props.isRunning) {
      await timeUtils.sleep(1)
    }

    await AdvanceTrams(time.value).then(tramPositionChanges => {
      for (const tram of tramPositionChanges) {
        if (tram.lat == 0 && tram.lon == 0) {
          tramMarkerByID.value[tram.id].removeFromMap()
        } else {
          tramMarkerByID.value[tram.id].updateCoordinates(
            tram.lat,
            tram.lon,
            tram.azimuth,
          )
        }
      }
    })

    time.value += 1

    await timeUtils.sleep(1000 / props.speed)
  }
})
</script>

<template>
  <v-overlay
    v-model="loading"
    opacity="0"
    class="d-flex justify-center align-center"
    persistent
  >
    <v-progress-circular indeterminate size="128"></v-progress-circular>
  </v-overlay>

  <div id="map" ref="map"></div>

  <TramSidebarComponent
    v-model="tramSidebar"
    :tram-id="selectedTramID"
    :current-time="time"
  />
  <StopSidebarComponent
    v-model="stopSidebar"
    :stop="selectedStop"
    :current-time="time"
  ></StopSidebarComponent>
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

.tram-marker.selected .tm-circle-arrow,
.tram-marker.selected .tm-circle {
  background-color: #67ad2f;
}
</style>
