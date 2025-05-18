<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"
import { ref, watch, computed } from "vue"
import { city } from "@wails/go/models"
import { GetLinesForStop, GetArrivalsForStop } from "@wails/go/city/City"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  stop?: city.GraphNode
  currentTime: number
}>()

const lines = ref<string[]>([])
const arrivalsInfo = ref<city.Arrival[]>([])

watch(
  () => props.stop?.id,
  async id => {
    if (!id) {
      lines.value = []
      arrivalsInfo.value = []
      return
    }
    lines.value = await GetLinesForStop(id)
    arrivalsInfo.value = await GetArrivalsForStop(id, props.currentTime)
  },
  { immediate: true },
)

const arrivals = computed(() => {
  return arrivalsInfo.value
    .filter(a => a.Departure + 30 >= props.currentTime)
    .slice(0, 5)
    .map(a => {
      const diff = a.Departure - props.currentTime
      return diff <= 0
        ? { ...a, eta: null }
        : { ...a, eta: Math.ceil(diff / 60) }
    })
})

const headers = [
  { title: "Route", key: "Route", sortable: false },
  {
    title: "Trip head-sign",
    key: "Headsign",
    align: "center",
    sortable: false,
  },
  { title: "ETA", key: "eta", align: "end", sortable: false },
] as const
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
        <template v-if="lines.length">
          <div class="line-chips">
            <span v-for="l in lines" :key="l" class="chip">
              {{ l }}
            </span>
          </div>
        </template>
        <span v-else>No lines</span>
      </div>
    </div>
    <div v-if="arrivals.length" class="section arrivals">
      <div class="label">
        <v-icon icon="mdi-clock-time-four" class="mr-2"></v-icon>
        Arrivals
      </div>
      <v-data-table
        class="arrivals-table"
        :headers="headers"
        :items="arrivals"
        disable-pagination
        hide-default-footer
        density="compact"
        :items-per-page="5"
      >
        <template v-slot:item.eta="{ item }">
          <span v-if="item.eta === null">&gt;&gt;&gt;</span>
          <span v-else>{{ item.eta }}min</span>
        </template>
      </v-data-table>
    </div>
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
  font-weight: 500;
  margin-bottom: 0.25rem;
}

.value {
  font-weight: 450;
  text-align: right;
}

.line-chips {
  display: grid;
  flex-wrap: wrap;
  gap: 4px;
  grid-template-columns: repeat(4, 1fr);
}

.line-chips:has(.chip:nth-child(1):nth-last-child(3)),
.line-chips:has(.chip:nth-child(1):nth-last-child(2)),
.line-chips:has(.chip:nth-child(1):nth-last-child(1)) {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 4px;
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
  margin-top: 0;
  font-size: 0.9rem;
  width: 100%;
  background-color: transparent;

  :deep(th),
  :deep(td) {
    padding: 0.45rem 0.6rem;
    white-space: nowrap;
  }

  :deep(th) {
    background: #4a4d51;
    color: #fff;
    font-weight: 500;
    text-align: left;
  }

  :deep(td) {
    transition: background 0.2s ease;
  }

  :deep(tr:hover td) {
    background: rgba(0, 0, 0, 0.05);
    cursor: pointer;
  }

  :deep(th:nth-child(1)),
  :deep(td:nth-child(1)) {
    width: 15%;
  }

  :deep(th:nth-child(2)),
  :deep(td:nth-child(2)) {
    width: 65%;
    overflow: hidden;
    text-overflow: ellipsis;
    text-align: center;
  }

  :deep(th:nth-child(3)),
  :deep(td:nth-child(3)) {
    width: 20%;
    text-align: right;
  }
}
</style>
