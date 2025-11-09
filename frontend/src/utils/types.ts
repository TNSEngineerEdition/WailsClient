import { city } from "@wails/go/models"

export type ModifiedNodes = Record<
  number,
  { neighborMaxSpeed: Record<number, number> }
>

// TODO: use types from `api` instead of manually typed ones
type GraphNeighbor = {
  id: number
  distance: number
  azimuth: number
  max_speed: number
}

export type GraphNode = {
  details: {
    id: number
    lat: number
    lon: number
    neighbors: Record<number, GraphNeighbor>

    // tram stop specific
    gtfs_stop_ids?: string[]
    name?: string
    node_type?: string
  }
}

export type CityRectangles = {
  bounds: city.CityRectangle["bounds"]
  nodes_by_id: Record<number, GraphNode>
}
