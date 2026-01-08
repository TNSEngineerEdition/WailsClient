import { ref } from "vue"

export default function useTimer() {
  const timer = ref(0)

  let previousTimeInSeconds = 0
  let currentTimeStart: number | undefined = undefined
  let interval: number | undefined = undefined

  function getTimeDelta() {
    if (currentTimeStart === undefined) return 0

    return (new Date().getTime() - currentTimeStart) / 1000
  }

  function start() {
    currentTimeStart = new Date().getTime()

    interval = setInterval(() => {
      timer.value = previousTimeInSeconds + getTimeDelta()
    }, 10)
  }

  function stop() {
    if (typeof currentTimeStart == "number") {
      previousTimeInSeconds += getTimeDelta()
      currentTimeStart = undefined
    }

    timer.value = previousTimeInSeconds
    clearInterval(interval)
    interval = undefined
  }

  function reset() {
    stop()
    timer.value = 0
    previousTimeInSeconds = 0
  }

  return { timer, start, stop, reset }
}
