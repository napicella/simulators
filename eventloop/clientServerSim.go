package main

import (
	"container/heap"
	"log"
	"math"
	mathrand "math/rand"
	"time"
)

var drain bool

type client struct {
	stats              *stats
	server             *server
	requestsPerSeconds int
	currentAttempt     int

	retrierFactory retrierFactory
}

func (t *client) genLoad(time float64, payload interface{}) []event {
	desiredStdDev := 0.1
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
	t.stats.uniqueCalls++
	t.stats.attempts++

	c := &call{
		r:              t.retrierFactory.get(),
		stats:          t.stats,
		server:         t.server,
		currentAttempt: 0,
	}

	return t.server.sendRequest(time, c)
}

type call struct {
	r              retrier
	stats          *stats
	server         *server
	currentAttempt int
}

func (t *call) callSuccess(time float64, payload interface{}) []event {
	t.r.recordSuccess()
	t.stats.reqSuccessCount++
	req := payload.(*request)
	t.stats.requestLatency(time - req.time)

	return nil
}

func (t *call) callFailed(time float64, payload interface{}) []event {
	t.r.recordFailure()

	if t.r.shouldRetry() {
		t.stats.attempts++
		return t.server.sendRequest(time, t)
	}

	// request failed after exhausting all attempts
	t.stats.reqFailedCount++

	return nil
}

type server struct {
	requests []request
	isBusy   bool
	stats    *stats
	// how long it takes for the server to fulfill a request (average, normally distributed)
	requestLatency float64
	// the server failure rate
	failureRate float64
}

type request struct {
	time   float64
	client *call
}

func (t *server) sendRequest(t_ float64, payload interface{}) []event {
	c := payload.(*call)

	t.requests = append(t.requests, request{time: t_, client: c})
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
	// request is done at requestEndTime
	requestEndTime := t_ + requestComputeTime

	// failure rate
	if mathrand.Float64() < t.failureRate {
		return []event{
			{
				time:        requestEndTime,
				callbackFun: req.client.callFailed,
				payload:     &req,
			},
			{
				time:        requestEndTime,
				callbackFun: t.processRequest,
				payload:     nil,
			},
		}
	}

	return []event{
		{
			time:        requestEndTime,
			callbackFun: req.client.callSuccess,
			payload:     &req,
		},
		{
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

type stats struct {
	uniqueCalls     int
	attempts        int
	reqLatencies    []float64
	reqSuccessCount int
	reqFailedCount  int
}

func (t *stats) try() {
	t.attempts = t.attempts + 1
}

func (t *stats) requestLatency(latency float64) {
	t.reqLatencies = append(t.reqLatencies, latency)
}

func runSimulation(s *stats, failureRate float64, factoryName retrierFactoryName) {
	server := &server{
		requests:       nil,
		isBusy:         false,
		stats:          s,
		requestLatency: 0.5,
		failureRate:    failureRate,
	}
	c := client{
		requestsPerSeconds: 1,
		stats:              s,
		server:             server,
		retrierFactory:     getFactory(factoryName),
	}
	t := 0.0
	maxTime := 5000.0

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
}

func main() {
	seed := time.Now().Unix()
	log.Printf("seed: %d", seed)
	mathrand.Seed(seed)

	failureRates := []float64{
		0.0, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45,
		0.5, 0.55, 0.6, 0.65, 0.7, 0.75, 0.8, 0.85, 0.9, 1}

	loadVsRate := loadVsFailureRateByStrategy{
		failureRate:         failureRates,
		loadByRetryStrategy: make(map[retrierFactoryName][]float64),
	}

	for _, retryStrategyName := range []retrierFactoryName{fixedRetry, circuitBreaker} {
		var loads []float64

		for _, failureRate := range failureRates {
			s := &stats{}
			drain = false
			runSimulation(s, failureRate, retryStrategyName)

			load := (float64(s.attempts) / float64(s.uniqueCalls)) * 100
			loads = append(loads, load)
		}
		loadVsRate.loadByRetryStrategy[retryStrategyName] = loads
	}

	drawLoad(loadVsRate)
}

type loadVsFailureRateByStrategy struct {
	// array of the failure rates used in the simulation
	failureRate []float64
	// the load that server experienced with each retry strategy.
	// It's a map between retry strategy name and an array of load (one for each failure
	// rate). This also means that len(loadByRetryStrategy[x]) == len(failureRate)
	loadByRetryStrategy map[retrierFactoryName][]float64
}
