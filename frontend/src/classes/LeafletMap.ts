import { GetBounds, GetTramStops } from "@wails/go/city/City"
import { GetRoutePolylines } from "@wails/go/simulation/Simulation"
import L, {
  LatLngBounds,
  Map as LMap,
  tileLayer,
  FeatureGroup,
  Polyline,
} from "leaflet"
import { TramMarker } from "@classes/TramMarker"
import { StopMarker } from "@classes/StopMarker"
import { city, simulation } from "@wails/go/models"

export class LeafletMap {
  private entityCount = 0
  public selectedStop?: StopMarker
  public selectedTram?: TramMarker
  public selectedRouteName?: string
  public highlightedRouteTrams?: TramMarker[]
  private routeLayer: FeatureGroup | null = null
  private svgRenderer = L.svg()
  private antsAnims: Animation[] = []

  constructor(private map: LMap) {
    this.map.addLayer(this.svgRenderer)
  }

  static async init(
    mapHTMLElement: HTMLElement,
    handleStopSelection: (stop: city.TramStop) => void,
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

  private stopAnts() {
    this.antsAnims.forEach(a => a.cancel())
    this.antsAnims = []
  }

  private animateAlongPath(el: SVGPathElement, periodMs = 1100, dashSum = 22) {
    el.style.strokeDasharray = "12 10"
    const anim = el.animate(
      [{ strokeDashoffset: "0" }, { strokeDashoffset: String(-dashSum) }],
      { duration: periodMs, iterations: Infinity, easing: "linear" },
    )
    this.antsAnims.push(anim)
  }

  private makePolyline(coords: [number, number][], color: string) {
    return new Polyline(coords, {
      weight: 5,
      opacity: 1,
      smoothFactor: 3,
      color: color,
      renderer: this.svgRenderer,
      interactive: false,
    })
  }

  public highlightTramsForRoute(trams: TramMarker[]) {
    this.highlightedRouteTrams?.forEach(m => m.setHighlighted(false))
    this.highlightedRouteTrams = trams
    this.highlightedRouteTrams.forEach(m => m.setHighlighted(true))
  }

  public async highlightRoute(route: city.RouteInfo) {
    this.selectedRouteName = route.name
    const { forward, backward } = await GetRoutePolylines(route.name)

    const fwd = forward as [number, number][]
    const bwd = backward as [number, number][]

    if (this.routeLayer) {
      this.map.removeLayer(this.routeLayer)
      this.routeLayer = null
    }

    const layers: Polyline[] = [
      this.makePolyline(fwd, route.background_color),
      this.makePolyline(bwd, route.background_color),
    ]

    this.routeLayer = new FeatureGroup(layers).addTo(this.map)
    this.stopAnts()
    for (const l of layers) {
      const el = l.getElement() as SVGPathElement | null
      if (el) this.animateAlongPath(el, 1100, 22)
    }
  }

  public deselectRoute() {
    this.stopAnts()
    if (this.routeLayer) {
      this.map.removeLayer(this.routeLayer)
      this.routeLayer = null
      this.selectedRouteName = undefined
      this.highlightedRouteTrams?.forEach(m => m.setHighlighted(false))
      this.highlightedRouteTrams = undefined
    }
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
}
