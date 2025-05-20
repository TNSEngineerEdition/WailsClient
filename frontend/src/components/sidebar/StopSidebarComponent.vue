<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"
import { ref, watch } from "vue"
import { city } from "@wails/go/models"
import { GetLinesForStop, GetArrivalsForStop } from "@wails/go/city/City"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  stop?: city.GraphNode
  currentTime: number
}>()

const lines = ref<string[]>([])
const arrivalsInfo = ref<city.Arrival[]>([])

const headers = [
  { title: "Route", key: "Route", align: "center", sortable: false },
  {
    title: "Trip head-sign",
    key: "Headsign",
    align: "center",
    sortable: false,
  },
  { title: "ETA", key: "eta", align: "center", sortable: false },
] as const

const ARRIVALS_IN_TABLE = 5
const LINE_CHIP_COLUMNS = ref(5)

watch(
  () => props.stop?.id,
  async id => {
    if (id) {
      lines.value = await GetLinesForStop(id, LINE_CHIP_COLUMNS.value)
      arrivalsInfo.value = await GetArrivalsForStop(
        id,
        props.currentTime,
        ARRIVALS_IN_TABLE,
      )
    } else {
      lines.value = []
      arrivalsInfo.value = []
    }
  },
  { immediate: true },
)

watch(
  () => props.currentTime,
  async currentTime => {
    if (props.stop?.id) {
      arrivalsInfo.value = await GetArrivalsForStop(
        props.stop.id,
        currentTime,
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
        Lines
      </div>

      <div class="value">
        <div v-if="lines.length" class="line-chips">
          <span v-for="line in lines" class="chip">
            {{ line }}
          </span>
        </div>

        <span v-else>No lines</span>
      </div>
    </div>

    <v-data-table
      v-if="arrivalsInfo.length"
      :headers="headers"
      :items="arrivalsInfo"
      class="arrivals-table"
      density="compact"
      hide-default-footer
      hover
    >
      <template v-slot:top>
        <div class="label">
          <v-icon icon="mdi-clock-time-four" class="mr-2"></v-icon>
          Arrivals
        </div>
      </template>

      <template v-slot:item.eta="{ item }">
        <span v-if="item.ETA === 0" class="blinking"> &gt;&gt;&gt; </span>

        <span v-else>{{ item.ETA }} min</span>
      </template>
    </v-data-table>
  </SidebarComponent>
</template>

<style scoped lang="scss">
.section {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 0.4rem;

  &.arrivals {
    flex-direction: column;
    align-items: flex-start;
  }
}

.label,
.value {
  font-size: clamp(0.8rem, 0.75rem + 0.2vw, 1rem);
  color: #111;
}

.label {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: bold;
  margin-bottom: 0.25rem;
}

.value {
  font-weight: 450;
  text-align: right;
}

.line-chips {
  display: grid;
  gap: 4px;
  grid-template-columns: repeat(v-bind(LINE_CHIP_COLUMNS), 1fr);
  direction: rtl;
}

.chip {
  padding: 3px 5px;
  background: #0078d4;
  border-radius: 4px;
  color: #fff;
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

.arrivals-table {
  width: 100%;
  background-color: transparent;
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
</style>
