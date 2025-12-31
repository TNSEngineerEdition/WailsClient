const delayClasses = ["delay-0", "delay-1", "delay-2", "delay-3", "delay-4"]

export function getClassForDelay(delay: number): string {
  if (delay <= 60) return delayClasses[0] // <=1min
  if (delay <= 120) return delayClasses[1] // <=2min
  if (delay <= 180) return delayClasses[2] // <=3min
  if (delay <= 300) return delayClasses[3] // <=5min
  return delayClasses[4] // >5min
}

export function getDelayClasses(): string[] {
  return delayClasses
}
