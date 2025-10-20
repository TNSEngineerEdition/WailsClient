package graph

import (
	"sync"
)

type NodeBlocker interface {
	TryBlocking(tramID uint) bool
	Unblock(tramID uint)
	ForceUnblock()
}

type NodeBlock struct {
	isBlocked      bool
	blockingTramID uint
	mu             sync.Mutex
}

func (g *NodeBlock) TryBlocking(tramID uint) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isBlocked && g.blockingTramID != tramID {
		return false
	}

	g.isBlocked = true
	g.blockingTramID = tramID

	return true
}

func (g *NodeBlock) Unblock(tramID uint) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isBlocked && g.blockingTramID == tramID {
		g.isBlocked = false
		g.blockingTramID = 0
	}
}

func (g *NodeBlock) ForceUnblock() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isBlocked {
		g.isBlocked = false
		g.blockingTramID = 0
	}
}
