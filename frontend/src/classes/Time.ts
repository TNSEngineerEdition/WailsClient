export class Time {
  static SECONDS_BY_TIME_UNIT = [60 * 60, 60, 1]
  static MAX_TIME_UNIT = [24, 60, 60]

  private isNegative: boolean
  private seconds: number
  private minutes: number
  private hours: number

  constructor(
    time: number,
    private isSigned: boolean = false,
  ) {
    if (!Number.isInteger(time)) {
      throw new Error("Time must be an integer")
    }

    this.isNegative = time < 0
    time = Math.abs(time)

    this.seconds = time % 60
    time = Math.floor(time / 60)

    this.minutes = time % 60
    time = Math.floor(time / 60)

    this.hours = time
  }

  public static async sleep(milliseconds: number) {
    await new Promise(resolve => setTimeout(resolve, milliseconds))
  }

  private static toPadded(value: number) {
    return value.toString().padStart(2, "0")
  }

  private static toTimeString(values: number[]) {
    return values.map(Time.toPadded).join(":")
  }

  private format(values: number[]) {
    const timeString = Time.toTimeString(values)

    if (this.isNegative) {
      return `-${timeString}`
    } else if (this.isSigned && values.some(x => x > 0)) {
      return `+${timeString}`
    } else {
      return timeString
    }
  }

  public toFullString() {
    return this.format([this.hours, this.minutes, this.seconds])
  }

  public toShortMinuteString() {
    return this.format([this.hours, this.minutes])
  }

  public toShortSecondString() {
    return this.format([this.hours * 60 + this.minutes, this.seconds])
  }
}
