package graph

import (
	"sync"
)

type NodeBlocker interface {
	TryBlocking(vehicleID uint) bool
	Unblock(vehicleID uint)
	ForceUnblock()
}

type NodeBlock struct {
	isBlocked         bool
	blockingVehicleID uint
	mu                sync.Mutex
}

func (g *NodeBlock) TryBlocking(vehicleID uint) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isBlocked && g.blockingVehicleID != vehicleID {
		return false
	}

	g.isBlocked = true
	g.blockingVehicleID = vehicleID

	return true
}

func (g *NodeBlock) unblock(condition bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isBlocked && condition {
		g.isBlocked = false
		g.blockingVehicleID = 0
	}
}

func (g *NodeBlock) Unblock(vehicleID uint) {
	g.unblock(g.blockingVehicleID == vehicleID)
}

func (g *NodeBlock) ForceUnblock() {
	g.unblock(true)
}
