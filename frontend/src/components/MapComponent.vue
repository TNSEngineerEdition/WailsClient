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
const selectedStop = ref<city.GraphNode | null>(null)
const stopSidebar = ref(false)

const timeUtils = useTimeUtils()

async function reset() {
  tramSidebar.value = false
  stopSidebar.value = false
  loading.value = true

  for (const tramMarker of Object.values(tramMarkerByID.value)) {
    tramMarker.removeFromMap()
  }

  tramMarkerByID.value = await GetTramIDs().then(tramIDs =>
    leafletMap.value!.getTramMarkers(tramIDs),
  )

  await GetTimeBounds().then(timeBounds => {
    time.value = timeBounds.startTime
    endTime.value = timeBounds.endTime
  })

  loading.value = false
}

watch(() => props.resetCounter, reset)

onMounted(async () => {
  if (mapHTMLElement.value === null) {
    throw new Error("Map element not found")
  }

  await FetchData("http://localhost:8000/cities/krakow")
  leafletMap.value = await LeafletMap.init(
    mapHTMLElement.value,
    (stop) => {
      selectedStop.value = stop
      stopSidebar.value = true
    }
  )

  await reset()

  while (
    time.value <= endTime.value ||
    leafletMap.value!.getEntityCount() > 0
  ) {
    while (!props.isRunning) {
      await timeUtils.sleep(1)
    }

    await AdvanceTrams(time.value).then(tramPositionChanges => {
      for (const stop of tramPositionChanges) {
        if (stop.lat == 0 && stop.lon == 0) {
          tramMarkerByID.value[stop.id].removeFromMap()
        } else {
          tramMarkerByID.value[stop.id].updateCoordinates(stop.lat, stop.lon)
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

  <TramSidebarComponent v-model="tramSidebar"></TramSidebarComponent>
  <StopSidebarComponent v-model="stopSidebar" :stop="selectedStop" :current-time="time"></StopSidebarComponent>
</template>

<style scoped lang="scss">
#map {
  width: 100%;
  height: calc(100vh - 64px);
}
</style>
