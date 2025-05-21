import { GetBounds, GetTramStops } from "@wails/go/city/City"
import { LatLngBounds, Map as LMap, tileLayer } from "leaflet"
import { TramMarker } from "@classes/TramMarker"
import { StopMarker } from "@classes/StopMarker"
import { city } from "@wails/go/models"

export class LeafletMap {
  private entityCount = 0
  public selectedStop?: StopMarker

  constructor(private map: LMap) {}

  static async init(
    mapHTMLElement: HTMLElement,
    handleStopSelection: (stop: city.GraphNode) => void,
  ) {
    const leafletMap = new LeafletMap(
      await GetBounds()
        .then(
          bounds =>
            new LatLngBounds(
              [bounds.minLat, bounds.minLon],
              [bounds.maxLat, bounds.maxLon],
            ),
        )
        .then(
          latLngBounds =>
            new LMap(mapHTMLElement, {
              maxBounds: latLngBounds.pad(1),
              center: latLngBounds.getCenter(),
              zoom: 13,
            }),
        ),
    )

    for (const stop of await GetTramStops()) {
      const marker = new StopMarker(stop.lat, stop.lon, stop.name)
      marker.addTo(leafletMap.map)
      marker.on("click", () => {
        if (leafletMap.selectedStop) {
          leafletMap.selectedStop.setSelected(false)
        }
        leafletMap.selectedStop = marker
        marker.setSelected(true)
        handleStopSelection(stop)
      })
    }

    tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: `&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>`,
    }).addTo(leafletMap.map)

    return leafletMap
  }

  public unselectStop() {
    if (this.selectedStop) {
      this.selectedStop.setSelected(false)
      this.selectedStop = undefined
    }
  }

  public getTramMarkers(tramIDs: number[]) {
    const result: Record<number, TramMarker> = {}

    for (const tramID of tramIDs) {
      result[tramID] = new TramMarker(this, {
        radius: 5,
        fill: true,
        color: "red",
      })
    }

    return result
  }

  public addTram(tramMarker: TramMarker) {
    this.entityCount++
    tramMarker.addTo(this.map)
  }

  public removeTram(tramMarker: TramMarker) {
    this.entityCount--
    tramMarker.removeFrom(this.map)
  }

  public getEntityCount() {
    return this.entityCount
  }
}
