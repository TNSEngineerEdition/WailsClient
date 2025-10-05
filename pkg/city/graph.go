package city

import (
	"sync"
)

type graphNodeNeighbor struct {
	ID       uint64  `json:"id"`
	Distance float32 `json:"distance"`
	Azimuth  float32 `json:"azimuth"`
	MaxSpeed float32 `json:"max_speed"`
}

type TramStop struct {
	ID          uint64    `json:"id"`
	Latitude    float32   `json:"lat"`
	Longitude   float32   `json:"lon"`
	Name        *string   `json:"name"`
	GTFSStopIDs *[]string `json:"gtfs_stop_ids"`
}

type GraphNode struct {
	ID             uint64                       `json:"id"`
	Latitude       float32                      `json:"lat"`
	Longitude      float32                      `json:"lon"`
	Neighbors      map[uint64]graphNodeNeighbor `json:"neighbors"`
	Name           *string                      `json:"name"`
	GTFSStopIDs    *[]string                    `json:"gtfs_stop_ids"`
	isBlocked      bool
	blockingTramID uint
	mu             sync.Mutex
}

func (g *GraphNode) IsTramStop() bool {
	return g.Name != nil && g.GTFSStopIDs != nil
}

func (g *GraphNode) getTramStopDetails() TramStop {
	return TramStop{
		ID:          g.ID,
		Latitude:    g.Latitude,
		Longitude:   g.Longitude,
		Name:        g.Name,
		GTFSStopIDs: g.GTFSStopIDs,
	}
}

func (g *GraphNode) TryBlocking(tramID uint) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isBlocked && g.blockingTramID != tramID {
		return false
	}

	g.isBlocked = true
	g.blockingTramID = tramID
	return true
}

func (g *GraphNode) Unblock(tramID uint) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.isBlocked || g.blockingTramID == tramID {
		g.isBlocked = false
		g.blockingTramID = 0
	}
}
