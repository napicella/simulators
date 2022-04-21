package sim

// EventsQueue implements a Priority Queue of Event
type EventsQueue []*Event

func (pq EventsQueue) Len() int { return len(pq) }

func (pq EventsQueue) Less(i, j int) bool {
	// We want Pop to give us the lowest so we use less than equal here
	return pq[i].Time <= pq[j].Time
}

func (pq EventsQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *EventsQueue) Push(x interface{}) {
	item := x.(*Event)
	*pq = append(*pq, item)
}

func (pq *EventsQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return item
}
