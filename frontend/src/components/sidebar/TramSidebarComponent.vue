<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"
import useTimeUtils from "@composables/useTimeUtils";
import { simulation } from "@wails/go/models";

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  tramID: number,
  tramDetails: simulation.TramDetails
}>()

const timeUtils = useTimeUtils()

function formatTime(time: number) {
  const t = timeUtils.toTimeString(time).split(':')
  return `${t[0]}:${t[1]}`
}
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="left"
    :title="props.tramDetails.route + ' ➡ ' + props.tramDetails.trip_head_sign"
    title-icon="mdi-tram"
    style="overflow-y:auto"
  >
  <div class="sidebarDetails">
    <span><b>TramID</b>: {{ props.tramID }}</span>
    <span><b>Speed</b>: {{ props.tramDetails.speed }} km/h</span>

    <div class="scrollable mt-4 ml-4">
      <div class="stopDiv">
        <b>Stop:</b>
        <b>Scheduled time:</b>
      </div>
      <ul>
        <li
          v-for="(stop, index) in props.tramDetails.stop_names"
          :key="index"
          :class="{ currentStop: index === props.tramDetails.trip_index }"
        >
          <div class="stopDiv">
            <span>{{ stop }}</span>
            <span>{{ formatTime(props.tramDetails.stops[index].time) }}</span>
          </div>
        </li>
      </ul>
    </div>
  </div>

  </SidebarComponent>
</template>

<style lang="scss" scoped>
.sidebarDetails {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.scrollable {
  overflow-y: auto;
  max-height: 60vh;
  padding-right: 8px;
}

/* Optional: improve look of scrollbar (WebKit browsers only) */
.scrollable::-webkit-scrollbar {
  width: 6px;
}
.scrollable::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.15);
  border-radius: 3px;
}

.currentStop {
  font-weight: bold;
  color: #007bff;
}

.stopDiv {
  width: 100%;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}
</style>
