package structs

import "cmp"

type priorityQueueItem[V any, P any] struct {
	value    V
	priority P
}

type PriorityQueue[V any, P any] struct {
	items   []*priorityQueueItem[V, P]
	compare func(left, right P) int
}

func NewPriorityQueue[V any, P any](compare func(left, right P) int) PriorityQueue[V, P] {
	return PriorityQueue[V, P]{
		items:   make([]*priorityQueueItem[V, P], 0),
		compare: compare,
	}
}

func NewPriorityQueueOrdered[V any, P cmp.Ordered]() PriorityQueue[V, P] {
	return PriorityQueue[V, P]{
		items:   make([]*priorityQueueItem[V, P], 0),
		compare: cmp.Compare[P],
	}
}

func (pq PriorityQueue[V, P]) Len() int {
	return len(pq.items)
}

func (pq *PriorityQueue[V, P]) Push(value V, priority P) {
	pq.items = append(pq.items, &priorityQueueItem[V, P]{
		value:    value,
		priority: priority,
	})

	pq.up(pq.Len() - 1)
}

func (pq *PriorityQueue[V, P]) Pop() V {
	lastIndex := pq.Len() - 1
	pq.swap(0, lastIndex)
	pq.down(0, lastIndex)

	item := pq.items[lastIndex]

	pq.items[lastIndex] = nil
	pq.items = pq.items[0:lastIndex]

	return item.value
}

func getParent(i int) int {
	return (i - 1) / 2
}

func getLeft(i int) int {
	return 2*i + 1
}

func (pq PriorityQueue[V, P]) less(i, j int) bool {
	return pq.compare(pq.items[i].priority, pq.items[j].priority) == -1
}

func (pq PriorityQueue[V, P]) swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// `up` and `down` methods are inspired by `container/heap`

func (pq *PriorityQueue[V, P]) up(index int) {
	for parent := getParent(index); parent != index && pq.less(index, parent); parent = getParent(index) {
		pq.swap(parent, index)
		index = parent
	}
}

func (pq *PriorityQueue[V, P]) down(index, n int) bool {
	i := index

	for left := getLeft(i); left >= 0 && left < n; left = getLeft(i) {
		j := left

		if right := left + 1; right < n && pq.less(right, left) {
			j = right
		}

		if !pq.less(j, i) {
			break
		}

		pq.swap(i, j)
		i = j
	}

	return i > index
}
