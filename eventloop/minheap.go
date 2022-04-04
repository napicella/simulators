package main

type minheap []*event

func (pq minheap) Len() int { return len(pq) }

func (pq minheap) Less(i, j int) bool {
	// We want Pop to give us the lowest so we use less than equal here
	return pq[i].time <= pq[j].time
}

func (pq minheap) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *minheap) Push(x interface{}) {
	item := x.(*event)
	*pq = append(*pq, item)
}

func (pq *minheap) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
