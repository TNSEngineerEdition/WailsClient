import { GetBounds, GetStops } from "@wails/go/city/City"
import { LatLngBounds, Map as LMap, tileLayer } from "leaflet"
import { TramMarker } from "@classes/TramMarker"
import { StopMarker } from "@classes/StopMarker"
import { city, simulation, api } from "@wails/go/models"
import { RouteHighlighter } from "./RouteHighlighter"

export class LeafletMap {
  private entityCount = 0
  public selectedStop?: StopMarker
  public selectedTram?: TramMarker
  private followTram = false
  public selectedRouteName?: string
  public highlightedRouteTrams?: TramMarker[]
  private routeHighlighter: RouteHighlighter
  private stopMarkersById: Record<number, StopMarker> = {}

  constructor(private map: LMap) {
    this.routeHighlighter = new RouteHighlighter(map)
  }

  static async init(
    mapHTMLElement: HTMLElement,
    handleStopSelection: (stop: api.ResponseGraphTramStop) => void,
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

    await leafletMap.makeStops(handleStopSelection)

    tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: `&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>`,
    }).addTo(leafletMap.map)

    return leafletMap
  }

  private async makeStops(
    handleStopSelection: (stop: api.ResponseGraphTramStop) => void,
  ) {
    for (const stop of await GetStops()) {
      const marker = new StopMarker(stop)
      marker.addTo(this.map)
      marker.on("click", () => {
        if (this.selectedStop) {
          this.selectedStop.setSelected(false)
        }
        this.selectedStop = marker
        marker.setSelected(true)
        handleStopSelection(stop)
      })
      this.stopMarkersById[stop.id] = marker
    }
  }

  public highlightTramsForRoute(trams: TramMarker[]) {
    this.highlightedRouteTrams?.forEach(m => m.setHighlighted(false))
    this.highlightedRouteTrams = trams
    this.highlightedRouteTrams.forEach(m => m.setHighlighted(true))
  }

  public async highlightRoute(route: city.RouteInfo) {
    this.selectedRouteName = route.name
    await this.routeHighlighter.highlight(route)
  }

  public deselectRoute() {
    this.selectedRouteName = undefined
    this.highlightedRouteTrams?.forEach(m => m.setHighlighted(false))
    this.highlightedRouteTrams = undefined
    this.routeHighlighter.clear()
  }

  public deselectStop() {
    if (this.selectedStop) {
      this.selectedStop.setSelected(false)
      this.selectedStop = undefined
    }
  }

  public getTramMarkers(
    trams: simulation.TramIdentifier[],
    onClickHandler: (id: number) => void,
  ) {
    const result: Record<number, TramMarker> = {}

    for (const tram of trams) {
      const marker = new TramMarker(this, tram.route)
      marker.on("click", () => {
        if (this.selectedTram) this.selectedTram.setSelected(false)
        this.selectedTram = marker
        marker.setSelected(true)
        onClickHandler(tram.id)
      })
      result[tram.id] = marker
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

  public deselectTram() {
    if (this.selectedTram) {
      this.selectedTram.setSelected(false)
      this.selectedTram = undefined
    }
  }

  public getEntityCount() {
    return this.entityCount
  }

  public centerOn(lat: number, lon: number) {
    const z = 17
    this.map.flyTo([lat, lon], z, { animate: true, duration: 0.6 })
  }

  public setFollowTram(enabled: boolean) {
    this.followTram = enabled
  }

  public followTick() {
    if (!this.followTram || !this.selectedTram) return
    this.map.panTo(this.selectedTram.getLatLng(), {
      animate: true,
      duration: 0.25,
    })
  }

  public getStopMarker(stopId: number): StopMarker {
    return this.stopMarkersById[stopId]
  }
}
