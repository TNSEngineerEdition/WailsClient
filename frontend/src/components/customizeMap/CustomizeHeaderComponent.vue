<script lang="ts" setup>
import getColorForSpeed from "@utils/getColorForSpeed"

const props = defineProps<{
  resetChanges: () => void
  saveChanges: () => void
}>()

const speedRanges = [
  { label: "0-10", speed: 10 },
  { label: "11-20", speed: 20 },
  { label: "21-30", speed: 30 },
  { label: "31-40", speed: 40 },
  { label: "41-50", speed: 50 },
  { label: "51-60", speed: 60 },
  { label: "60+", speed: 70 },
]
</script>

<template>
  <v-footer tag="header">
    <v-row no-gutters>
      <v-col>
        <div class="left">
          <v-btn color="white" prepend-icon="mdi-restore" @click="resetChanges"
            >Reset changes</v-btn
          >
        </div>
      </v-col>
      <v-col class="legend-bar-container">
        <div class="legend-bar">
          <div
            v-for="range in speedRanges"
            :key="range.label"
            class="legend-segment"
          >
            <div
              class="color-box"
              :style="{ backgroundColor: getColorForSpeed(range.speed / 3.6) }"
            ></div>
            <span class="legend-label">{{ range.label }}</span>
          </div>
        </div>
      </v-col>

      <v-col>
        <div class="button-box-right">
          <v-btn variant="elevated" prepend-icon="mdi-backburger" to="/">
            Menu
          </v-btn>
          <v-btn variant="elevated" base-color="blue" @click="saveChanges">
            Save changes
          </v-btn>
        </div>
      </v-col>
    </v-row>
  </v-footer>
</template>

<style scoped lang="scss">
@mixin button-box {
  display: flex;
  align-items: center;
  height: 100%;
  gap: 10px;
}

.button-box-left {
  @include button-box;
  justify-content: flex-start;
}

.button-box-right {
  @include button-box;
  justify-content: flex-end;
}

.legend-bar-container {
  display: flex;
  align-items: center;
  justify-content: center;
}

.legend-bar {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
}

.legend-segment {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.color-box {
  width: 40px;
  height: 15px;
  border-radius: 4px;
}

.legend-label {
  font-size: 0.75rem;
  color: rgba(0, 0, 0, 0.7);
}
</style>
