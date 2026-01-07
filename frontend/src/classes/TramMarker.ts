import { Marker, DivIcon } from "leaflet"
import { LeafletMap } from "@classes/LeafletMap"
import { MarkerColoringMode } from "@utils/types"

export class TramMarker extends Marker {
  static coloringMode: MarkerColoringMode = "Default"
  private isOnMap = false

  constructor(
    private leafletMap: LeafletMap,
    private route: string,
  ) {
    super([0, 0], { icon: TramMarker.createIcon(route) })
  }

  private static createIcon(route: string): DivIcon {
    return new DivIcon({
      className: "",
      html: `
        <div class="tram-marker">
          <div class="tm-circle-arrow" style="transform: rotate(0);"></div>
          <div class="tm-circle"></div>
          <div class="tm-route-label">${route}</div>
        </div>
      `,
      iconSize: [24, 24],
      iconAnchor: [12, 12],
    })
  }

  private setAzimuth(azimuth: number) {
    const circleArrow =
      this.getElement()?.querySelector<HTMLElement>(".tm-circle-arrow")
    if (!circleArrow) {
      throw new Error("Tram marker arrow not found")
    }

    circleArrow.style.transform = `rotate(${azimuth + 135}deg)`
  }

  private setDelayColor(delay: number) {
    const circleElement =
      this.getElement()?.querySelector<HTMLElement>(".tm-circle")
    if (!circleElement) {
      return
    }

    const circleArrowElement =
      this.getElement()?.querySelector<HTMLElement>(".tm-circle-arrow")
    if (!circleArrowElement) {
      return
    }

    let bgColor = "rgb(11, 116, 202)"

    if (delay > 60) {
      // scale delay so 5 minute delay translates to 225
      // red value: 255-225=30, very much dark red tram marker
      const scaledDelay = delay * 0.75
      const rValue = 255 - Math.min(scaledDelay, 225)
      bgColor = `rgb(${rValue}, 7, 7)`
    }

    circleElement.style.backgroundColor = bgColor
    circleArrowElement.style.backgroundColor = bgColor
  }

  public getRoute(): string {
    return this.route
  }

  public getIsOnMap(): boolean {
    return this.isOnMap
  }

  public setHighlighted(isHighlighted: boolean) {
    const element =
      this.getElement()?.querySelector<HTMLElement>(".tram-marker")
    if (!element) {
      return
    }
    if (isHighlighted) {
      element?.classList.add("highlighted")
    } else {
      element?.classList.remove("highlighted")
    }
  }

  public setSelected(isSelected: boolean) {
    if (!this.isOnMap) return

    const element =
      this.getElement()?.querySelector<HTMLElement>(".tram-marker")

    if (!element) {
      throw new Error("Tram marker not found")
    }

    if (isSelected) {
      element?.classList.add("selected")
    } else {
      element?.classList.remove("selected")
    }
  }

  public removeCustomColoring() {
    const circleElement =
      this.getElement()?.querySelector<HTMLElement>(".tm-circle")
    if (!circleElement) {
      return
    }

    const circleArrowElement =
      this.getElement()?.querySelector<HTMLElement>(".tm-circle-arrow")
    if (!circleArrowElement) {
      return
    }

    circleElement.style.backgroundColor = ""
    circleArrowElement.style.backgroundColor = ""
  }

  public setStopped(isStopped: boolean) {
    const element =
      this.getElement()?.querySelector<HTMLElement>(".tram-marker")
    if (!element) return

    if (isStopped) {
      element.classList.add("stopped")
    } else {
      element.classList.remove("stopped")
    }
  }

  public updateCoordinates(
    lat: number,
    lon: number,
    azimuth: number,
    isStopped?: boolean,
    delay?: number,
  ) {
    if (!this.isOnMap) {
      this.leafletMap.addTram(this)
      this.isOnMap = true
    }
    this.setHighlighted(this.route === this.leafletMap.selectedRouteName)
    this.setSelected(this.leafletMap.selectedTram === this)
    this.setLatLng([lat, lon])
    this.setAzimuth(azimuth)

    if (isStopped !== undefined) {
      this.setStopped(isStopped)
    }

    if (TramMarker.coloringMode == "Delays" && delay !== undefined) {
      this.setDelayColor(delay)
    }
  }

  public removeFromMap() {
    if (!this.isOnMap) return

    this.leafletMap.removeTram(this)
    this.isOnMap = false
  }
}
