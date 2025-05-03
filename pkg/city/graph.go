package city

type graphNodeNeighbor struct {
	ID      uint64  `json:"id"`
	Length  float32 `json:"length"`
	Azimuth float32 `json:"azimuth"`
}

type GraphNode struct {
	ID          uint64              `json:"id"`
	Latitude    float32             `json:"lat"`
	Longitude   float32             `json:"lon"`
	Neighbors   []graphNodeNeighbor `json:"neighbors"`
	Name        *string             `json:"name"`
	GTFSStopIDs *[]string           `json:"gtfs_stop_ids"`
}

func (g *GraphNode) isTramStop() bool {
	return g.Name != nil && g.GTFSStopIDs != nil
}
