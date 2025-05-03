<script lang="ts" setup>
import { onMounted, ref, useTemplateRef } from "vue"
import { GetTimeBounds } from "@wails/go/city/City"
import {
  GetTramIDs,
  AdvanceTrams,
  FetchData,
} from "@wails/go/simulation/Simulation"
import { LeafletMap } from "@classes/LeafletMap"

const mapHTMLElement = useTemplateRef("map")

const time = ref(0)
const endTime = ref(0)

onMounted(async () => {
  if (mapHTMLElement.value === null) {
    throw new Error("Map element not found")
  }

  await FetchData("http://localhost:8000/cities/krakow")

  const leafletMap = await LeafletMap.init(mapHTMLElement.value)
  const tramMarkerByID = await GetTramIDs().then(tramIDs =>
    leafletMap.getTramMarkers(tramIDs),
  )

  await GetTimeBounds().then(timeBounds => {
    time.value = timeBounds.startTime
    endTime.value = timeBounds.endTime
  })

  while (time.value <= endTime.value || leafletMap.getEntityCount() > 0) {
    await AdvanceTrams(time.value).then(tramPositionChanges => {
      for (const stop of tramPositionChanges) {
        if (stop.lat == 0 && stop.lon == 0) {
          tramMarkerByID[stop.id].removeFromMap()
        } else {
          tramMarkerByID[stop.id].updateCoordinates(stop.lat, stop.lon)
        }
      }
    })

    time.value += 60

    await new Promise(resolve => setTimeout(resolve, 10))
  }
})
</script>

<template>
  <div id="map" ref="map"></div>
</template>

<style scoped lang="scss">
#map {
  width: 100%;
  height: 100vh;
}
</style>
