<script lang="ts" setup>
import { ref } from "vue"

const showGuide = ref(false)
</script>

<template>
  <div
    class="guide-container"
    @mouseenter="showGuide = true"
    @mouseleave="showGuide = false"
  >
    <transition name="fade">
      <v-card v-if="showGuide" class="guide-card" elevation="8">
        <v-card-title class="text-subtitle-1 pb-2">
          How to use the map
        </v-card-title>
        <v-divider class="mb-2" />
        <v-card-text>
          <ul class="guide-list">
            <li>
              Select the <b>part of the network</b> you want to work with
              (rectangle).
            </li>
            <li>
              Click on the tracks to choose the <b>starting point</b> of your
              selection.
            </li>
            <li>
              Click again to choose the <b>end point</b> of your selection.
            </li>
            <li>
              Remember to select according to the
              <b>direction of the tracks!</b>
              Otherwise the end of the selection is going to be selected
              automatically (see the paragraph below)
            </li>
            <li>
              The selection will always end at:
              <ul>
                <li>stop</li>
                <li>switch</li>
                <li>tracks crossing</li>
                <li>outside of the selected network part</li>
              </ul>
            </li>
            <li>
              After selecting, you can <b>change maximum speed</b> and
              <b>save changes</b> or start a new selection.
            </li>
          </ul>
        </v-card-text>
      </v-card>
    </transition>

    <v-btn icon="mdi-information-outline" size="large" elevation="3"></v-btn>
  </div>
</template>

<style scoped lang="scss">
.guide-container {
  position: absolute;
  bottom: 24px;
  right: 24px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  z-index: 1000;
}

.guide-card {
  margin-bottom: 8px;
  width: 350px;
}

.guide-list {
  margin: 0;
  padding-left: 18px;
  font-size: 0.9rem;
  line-height: 1.4;
  list-style-type: "🔷 ";
}

.guide-list > li {
  margin-bottom: 10px;
}

.guide-list li::marker {
  color: #2896f1;
}

.guide-list ul {
  padding-left: 16px;
  list-style-type: circle;
  font-size: 0.85rem;
  color: rgba(0, 0, 0, 0.75);
  list-style-type: disc;
}

.fade-enter-active,
.fade-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
</style>
