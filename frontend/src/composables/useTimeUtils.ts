export default function useTimeUtils() {
  const SECONDS_BY_TIME_UNIT = [60 * 60, 60, 1]
  const MAX_TIME_UNIT = [24, 60, 60]

  function formatTime(time: number) {
    return SECONDS_BY_TIME_UNIT.map(
      (item, i) => Math.floor(time / item) % MAX_TIME_UNIT[i],
    ).map(item => item.toString().padStart(2, "0"))
  }

  function toFullTimeString(time: number) {
    const t = formatTime(time)
    return t.join(":")
  }

  function toShortTimeString(time: number) {
    const t = formatTime(time)
    return `${t[0]}:${t[1]}`
  }

  async function sleep(milliseconds: number) {
    await new Promise(resolve => setTimeout(resolve, milliseconds))
  }

  return { toFullTimeString, toShortTimeString, sleep }
}
