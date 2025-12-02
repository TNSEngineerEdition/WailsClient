package structs

type PQRecord[T any] struct {
	Value    T
	Priority float32
	index    int
}

type PriorityQueue[T any] []*PQRecord[T]

func (pq PriorityQueue[T]) Len() int { return len(pq) }

func (pq PriorityQueue[T]) Less(i, j int) bool { return pq[i].Priority < pq[j].Priority }

func (pq PriorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue[T]) Push(x any) {
	item := x.(*PQRecord[T])
	item.index = len(*pq)

	*pq = append(*pq, item)
}

func (pq *PriorityQueue[T]) Pop() any {
	lastIndex := pq.Len() - 1
	item := (*pq)[lastIndex]

	(*pq)[lastIndex] = nil
	*pq = (*pq)[0:lastIndex]

	return item
}
