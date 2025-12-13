import L, { FeatureGroup, Polyline, Map as LMap } from "leaflet"
import { GetSegmentsForRoute } from "@wails/go/simulation/Simulation"
import { city } from "@wails/go/models"

export class RouteHighlighter {
  private routeLayer: FeatureGroup | null = null
  private antsAnims: Animation[] = []
  private svgRenderer = L.svg()

  constructor(private map: LMap) {
    this.map.addLayer(this.svgRenderer)
  }

  public async highlight(route: city.RouteInfo) {
    const routeVariants = await GetSegmentsForRoute(route.name)

    if (this.routeLayer) {
      this.map.removeLayer(this.routeLayer)
    }

    const layers: Polyline[] = routeVariants.map(variant =>
      this.makePolyline(
        variant.polyline.map(item => [item.lat, item.lon]),
        route.background_color,
      ),
    )

    this.routeLayer = new FeatureGroup(layers).addTo(this.map)
    this.stopAnts()
    for (const l of layers) {
      const el = l.getElement() as SVGPathElement | null
      if (el) this.animateAlongPath(el, 1100, 22)
    }
  }

  public clear() {
    this.stopAnts()
    if (this.routeLayer) {
      this.map.removeLayer(this.routeLayer)
      this.routeLayer = null
    }
  }

  private makePolyline(coords: [number, number][], color: string) {
    return new Polyline(coords, {
      weight: 5,
      opacity: 1,
      smoothFactor: 3,
      color,
      renderer: this.svgRenderer,
      interactive: false,
    })
  }

  private animateAlongPath(el: SVGPathElement, periodMs = 1100, dashSum = 22) {
    el.style.strokeDasharray = "12 10"
    const anim = el.animate(
      [{ strokeDashoffset: "0" }, { strokeDashoffset: String(-dashSum) }],
      { duration: periodMs, iterations: Infinity, easing: "linear" },
    )
    this.antsAnims.push(anim)
  }

  private stopAnts() {
    this.antsAnims.forEach(a => a.cancel())
    this.antsAnims = []
  }
}
