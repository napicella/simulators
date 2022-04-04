package main

import (
	"container/heap"
	"fmt"
	"math"
	mathrand "math/rand"
	"time"
)

var drain bool

type client struct {
	stats              *stats
	server             *server
	requestsPerSeconds int
}

func (t *client) genLoad(time float64, payload interface{}) []event {
	desiredStdDev := 0.2
	desiredMean := float64(t.requestsPerSeconds)
	nextCall := math.Abs(mathrand.NormFloat64()*desiredStdDev + desiredMean)

	if drain {
		return nil
	}

	return []event{
		{
			time:        time + nextCall,
			callbackFun: t.genLoad,
			payload:     nil,
		},
		{
			time:        time,
			callbackFun: t.call,
			payload:     nil,
		},
	}
}

func (t *client) call(time float64, payload interface{}) []event {
	t.stats.try()
	return t.server.sendRequest(time, payload)
}

type stats struct {
	attempts     int
	reqLatencies []float64
}

func (t *stats) try() {
	t.attempts = t.attempts + 1
}

func (t *stats) getAttempts() int {
	return t.attempts
}

func (t *stats) requestLatency(t_ float64, payload interface{}) []event {
	latency := payload.(float64)
	t.reqLatencies = append(t.reqLatencies, latency)
	return nil
}

type server struct {
	requests []request
	isBusy   bool
	stats    *stats
	// how long it takes for the server to fulfill a request
	requestLatency float64
}

type request struct {
	time float64
}

func (t *server) sendRequest(t_ float64, payload interface{}) []event {
	t.requests = append(t.requests, request{time: t_})
	if !t.isBusy {
		return []event{
			{
				time:        t_,
				callbackFun: t.processRequest,
				payload:     nil,
			},
		}
	}
	return nil
}

// server can process one request at the time
func (t *server) processRequest(t_ float64, payload interface{}) []event {
	if len(t.requests) == 0 {
		t.isBusy = false
		return nil
	}

	var req request
	req, t.requests = t.requests[0], t.requests[1:]
	t.isBusy = true

	requestComputeTime := math.Abs(mathrand.NormFloat64()*0.1 + t.requestLatency)
	requestEndTime := t_ + requestComputeTime

	return []event{
		{
			// request is done at requestEndTime
			time:        requestEndTime,
			callbackFun: t.stats.requestLatency,
			payload:     requestEndTime - req.time,
		},
		{
			// request is done at requestEndTime
			time:        requestEndTime,
			callbackFun: t.processRequest,
			payload:     nil,
		},
	}
}

func (t *server) startServer(t_ float64, payload interface{}) []event {
	return []event{
		{
			time:        t_,
			callbackFun: t.processRequest,
			payload:     nil,
		},
	}
}

// A different example of a simulation is in:
// https://github.com/mbrooker/simulator_example/blob/main/ski_sim.py
func main() {
	mathrand.Seed(time.Now().Unix())
	s := &stats{}
	server := &server{
		requests:       nil,
		isBusy:         false,
		stats:          s,
		requestLatency: 0.5,
	}
	c := client{
		requestsPerSeconds: 1,
		stats:              s,
		server:             server,
	}
	t := 0.0
	maxTime := 100.0

	h := &minheap{
		{
			time:        t,
			callbackFun: server.startServer,
			payload:     nil,
		},
		{
			time:        t,
			callbackFun: c.genLoad,
			payload:     nil,
		},
	}
	heap.Init(h)

	for h.Len() > 0 {
		item := heap.Pop(h)
		e := item.(*event)
		t = e.time

		events := e.callbackFun(t, e.payload)
		for _, ev := range events {
			evCopy := ev
			heap.Push(h, &evCopy)
		}

		if t > maxTime {
			drain = true
		}
	}
	fmt.Println(s.reqLatencies)
	draw(s.reqLatencies)
}
