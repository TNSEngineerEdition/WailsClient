<script setup lang="ts">
import SidebarComponent from "@components/sidebar/SidebarComponent.vue"
import { ref, watch, computed } from "vue"
import { city } from "@wails/go/models"
import { GetLinesForStop, GetArrivalsForStop } from "@wails/go/city/City"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  stop: city.GraphNode | null
  currentTime: number
}>()

const lines = ref<Array<string>>([])
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
    arrivalsInfo.value = await (GetArrivalsForStop(id, props.currentTime)) ?? []
  },
  { immediate: true }
)

const arrivals = computed(() => {
  return arrivalsInfo.value
    .filter(a => a.Departure + 30 >= props.currentTime)
    .slice(0, 5)
    .map(a => {
      const diff = a.Departure - props.currentTime
      return diff <= 0 ? { ...a, eta: null } : { ...a, eta: Math.ceil(diff / 60) }
    })
})
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
          {{ stop.gtfs_stop_ids.join(', ') }}
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
      <table class="arrivals-table">
        <thead>
          <tr>
            <th>Route</th>
            <th>Trip head-sign</th>
            <th class="eta">ETA</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="a in arrivals" :key="a.Route + a.Headsign + a.Departure">
            <td>{{ a.Route }}</td>
            <td>{{ a.Headsign }}</td>
            <td class="eta">
              <span v-if="a.eta === null">&gt;&gt;&gt;</span>
              <span v-else>{{ a.eta }}min</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </SidebarComponent>
</template>

<style scoped lang="scss">
.section {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: .4rem;

  &.arrivals {
    flex-direction: column;
    align-items: flex-start;
  }
}

.label,
.value {
  font-size: clamp(.8rem, .75rem + .2vw, 1rem);
  color: #111;
}

.label {
  display: inline-flex;
  align-items: center;
  gap: .4rem;
  font-weight: 500;
  margin-bottom: .25rem;
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
  font-size: .85rem;
  line-height: 1;
  transition: background .2s ease;
  text-align: center;
  &:hover {
    background: #2896f1;
    cursor: pointer;
  }
}

.arrivals-table {
  margin-top: 0;
  border-collapse: collapse;
  table-layout: auto;
  font-size: .9rem;
  width: 100%;

  th,
  td {
    padding: .45rem .6rem;
    white-space: nowrap;
  }

  th {
    background: #4a4d51;
    color: #fff;
    font-weight: 500;
    text-align: left;
  }

  td {
    border-top: 1px solid #dcdcdc;
    transition: background .2s ease;
  }

  tr:hover td {
    background: rgba(0, 0, 0, .05);
    cursor: pointer;
  }

  th:nth-child(1),
  td:nth-child(1) { width: 15%; }

  th:nth-child(2),
  td:nth-child(2) {
    width: 65%;
    overflow: hidden;
    text-overflow: ellipsis;
    text-align: center;
  }

  th:nth-child(3),
  td:nth-child(3) {
    width: 20%;
    text-align: right;
  }
}
</style>
