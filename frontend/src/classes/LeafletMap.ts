import { GetBounds, GetTramStops } from "@wails/go/city/City";
import { CircleMarker, LatLngBounds, Map as LMap, tileLayer } from "leaflet";
import { TramMarker } from "@classes/TramMarker";

export class LeafletMap {
  private entityCount = 0

  constructor(private map: LMap) { }

  static async init(mapHTMLElement: HTMLElement) {
    const result = new LeafletMap(await GetBounds()
      .then(bounds =>
        new LatLngBounds(
          [bounds.minLat, bounds.minLon],
          [bounds.maxLat, bounds.maxLon],
        ),
      )
      .then(latLngBounds =>
        new LMap(mapHTMLElement, {
          maxBounds: latLngBounds.pad(1),
          center: latLngBounds.getCenter(),
          zoom: 13,
        }),
      )
    )

    for (const stop of await GetTramStops()) {
      new CircleMarker([stop.lat, stop.lon], {
        radius: 5,
        fill: true,
      }).addTo(result.map)
    }

    tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: `&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>`,
    }).addTo(result.map)

    return result
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
