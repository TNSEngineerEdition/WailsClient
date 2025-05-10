<script setup lang="ts">
import { ref, watch, computed } from "vue"
import { GetLinesForStop, GetArrivalsForStop } from "@wails/go/city/City"
import { city } from "@wails/go/models"
import SvgIcon from '@jamescoyle/vue-icon'
import {
  mdiClose,
  mdiTramSide,
  mdiMapMarker,
  mdiZipBox,
  mdiClockTimeFour,
  mdiTransitConnectionVariant
} from '@mdi/js'

const props = defineProps<{
  stop: city.GraphNode | null
  currentTime: number
}>()

const lines = ref<string[]>([])
const arrivalsInfo = ref<city.Arrival[]>([])
const visible = ref(true)

watch(
  () => props.stop?.id,
  async id => {
    if (!id) {
      lines.value = []
      arrivalsInfo.value = []
      return
    }
    lines.value = await GetLinesForStop(id)
    arrivalsInfo.value = await (GetArrivalsForStop(id, props.currentTime)) || []
    visible.value = true
  },
  { immediate: true }
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
</script>

<template>
  <transition name="sidebar-slide" appear @after-leave="$emit('close')">
    <aside v-if="visible && stop" class="sidebar">

      <button class="close-btn" @click="visible = false">
        <svg-icon type="mdi" :path="mdiClose"/>
      </button>

      <transition name="content-fade" mode="out-in">
        <div :key="stop.id">
          <h2 class="stop-title">
              <svg-icon type="mdi" :path="mdiTramSide "/>
              {{ stop.name || "Unknown stop" }}
          </h2>

          <div class="section">
            <div class="label">
              <svg-icon type="mdi" :path="mdiMapMarker"/>
              Coordinates
            </div>
            <div class="value">
              {{ stop.lat.toFixed(6) }}, {{ stop.lon.toFixed(6) }}
            </div>
          </div>

          <div class="section">
            <div class="label">
              <svg-icon type="mdi" :path="mdiZipBox"/> GTFS ID
            </div>
            <div class="value">
              <span v-if="stop.gtfs_stop_ids?.length">
                {{ stop.gtfs_stop_ids.join(', ') }}
              </span>
              <span v-else>Unknown</span>
            </div>
          </div>

          <div class="section">
            <div class="label">
              <svg-icon type="mdi" :path="mdiTransitConnectionVariant"/> Lines
            </div>
            <div class="value line-chips">
              <template v-if="lines.length">
                <span v-for="l in lines" :key="l" class="chip">
                  {{ l }}
                </span>
              </template>
              <span v-else class="value">No lines</span>
            </div>
          </div>
          <div v-if="arrivals.length" class="section arrivals">
            <div class="label arrivals">
              <svg-icon type="mdi" :path="mdiClockTimeFour"/> Arrivals
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
        </div>
      </transition>
    </aside>
  </transition>
</template>

<style scoped lang="scss">
.sidebar-slide-enter-active,
.sidebar-slide-leave-active {
  transition: opacity .3s ease, transform .3s ease;
}

.sidebar-slide-enter-from,
.sidebar-slide-leave-to {
  opacity: 0;
  transform: translateX(30px);
}

.sidebar-slide-enter-to,
.sidebar-slide-leave-from {
  opacity: 1;
  transform: translateX(0);
}

.content-fade-enter-active,
.content-fade-leave-active {
  transition: opacity .25s ease;
}

.content-fade-enter-from,
.content-fade-leave-to {
  opacity: 0;
}

.content-fade-enter-to,
.content-fade-leave-from {
  opacity: 1;
}

.sidebar {
  position: fixed;
  top: 1rem;
  right: 1rem;
  width: clamp(260px, 28vw, 420px);
  max-height: calc(100vh - 2rem);
  overflow-y: auto;
  z-index: 1000;
  padding: clamp(.8rem, 1.2vw, 1.5rem);
  border: 1px solid rgba(100, 100, 100, .3);
  border-radius: 16px;
  background: rgba(235, 235, 235, .85);
  box-shadow: 0 0 14px rgba(0, 0, 0, .5);
  backdrop-filter: blur(1.4px);
  font-family: 'Segoe UI', sans-serif;

  .stop-title {
    display: flex;
    align-items: center;
    gap: .5rem;
    margin: 0 0 1.2rem;
    font-size: clamp(1.2rem, 1rem + .4vw, 1.6rem);
    font-weight: 600;
    color: #111;

    svg {
      flex: 0 0 auto;
      width: 1.6em;
      height: 1.6em;
      margin-top: .15em;
    }
  }

  .close-btn {
    float: right;
    background: none;
    border: none;
    font-size: 1.5rem;
    color: #333;
    cursor: pointer;
  }

  .section {
    display: flex;
    align-items: center;
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
    white-space: nowrap;
  }

  .value {
    font-weight: 450;
    text-align: right;
  }

  .line-chips {
    display: flex;
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

    &:hover {
      background: #2896f1;
      cursor: pointer;
    }
  }

  .arrivals-table {
    width: 100%;
    margin-top: 0;
    border-collapse: collapse;
    table-layout: fixed;
    font-size: .9rem;

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
}

</style>
