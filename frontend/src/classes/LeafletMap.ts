import { GetBounds, GetTramStops } from "@wails/go/city/City";
import { LatLngBounds, Map as LMap, tileLayer } from "leaflet";
import { TramMarker } from "@classes/TramMarker";
import { StopMarker } from "@classes/StopMarker";
import { city } from "@wails/go/models";

export class LeafletMap {
  private entityCount = 0
  public selectedStop : StopMarker | null = null

  constructor(private map: LMap) { }

  static async init(
    mapHTMLElement: HTMLElement,
    onStopClick: (stop: city.GraphNode) => void
  )
  {
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
      const marker = new StopMarker(stop.lat, stop.lon, stop.name);
      marker.onStopClick(() => {
        if (result.selectedStop && result.selectedStop !== marker) {
          result.selectedStop.setSelected(false);
        }
        marker.setSelected(true);
        result.selectedStop = marker;
        onStopClick(stop);
      });
      marker.addTo(result.map);
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
