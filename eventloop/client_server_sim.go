package main

import (
	mathstats "github.com/montanaflynn/stats"
	"math"
	mathrand "math/rand"
	"napicella.com/simulators/simulation"
)

type client struct {
	stats              *stats
	server             *server
	requestsPerSeconds int
	currentAttempt     int

	retrierFactory retrierFactory

	drain bool
}

func (t *client) genLoad(time float64, payload interface{}) []sim.Event {
	if t.drain {
		return nil
	}

	desiredStdDev := 0.1
	desiredMean := float64(t.requestsPerSeconds)
	nextCall := math.Abs(mathrand.NormFloat64()*desiredStdDev + desiredMean)

	return []sim.Event{
		{
			Time:        time + nextCall,
			CallbackFun: t.genLoad,
			Payload:     nil,
		},
		{
			Time:        time,
			CallbackFun: t.call,
			Payload:     nil,
		},
	}
}

func (t *client) stopLoadGen() {
	t.drain = true
}

func (t *client) call(time float64, payload interface{}) []sim.Event {
	t.stats.uniqueCalls++
	t.stats.attempts++

	retrier := t.retrierFactory.get()
	retrier.initCall()

	c := &call{
		r:              retrier,
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

func (t *call) callSuccess(time float64, payload interface{}) []sim.Event {
	t.r.recordSuccess()
	t.stats.reqSuccessCount++
	req := payload.(*request)
	t.stats.requestLatency(time - req.time)

	return nil
}

func (t *call) callFailed(time float64, payload interface{}) []sim.Event {
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

func (t *server) sendRequest(t_ float64, payload interface{}) []sim.Event {
	c := payload.(*call)

	t.requests = append(t.requests, request{time: t_, client: c})
	if !t.isBusy {
		return []sim.Event{
			{
				Time:        t_,
				CallbackFun: t.processRequest,
				Payload:     nil,
			},
		}
	}
	return nil
}

// server can process one request at the time
func (t *server) processRequest(t_ float64, payload interface{}) []sim.Event {
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
		return []sim.Event{
			{
				Time:        requestEndTime,
				CallbackFun: req.client.callFailed,
				Payload:     &req,
			},
			{
				Time:        requestEndTime,
				CallbackFun: t.processRequest,
				Payload:     nil,
			},
		}
	}

	return []sim.Event{
		{
			Time:        requestEndTime,
			CallbackFun: req.client.callSuccess,
			Payload:     &req,
		},
		{
			Time:        requestEndTime,
			CallbackFun: t.processRequest,
			Payload:     nil,
		},
	}
}

func (t *server) startServer(t_ float64, payload interface{}) []sim.Event {
	return []sim.Event{
		{
			Time:        t_,
			CallbackFun: t.processRequest,
			Payload:     nil,
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

func (t *stats) requestLatency(latency float64) {
	t.reqLatencies = append(t.reqLatencies, latency)
}

func (t *stats) getLoad() float64 {
	return (float64(t.attempts) / float64(t.uniqueCalls)) * 100
}

func (t *stats) getp90Latency() float64 {
	if len(t.reqLatencies) == 0 {
		return 0
	}
	p90, e := mathstats.Percentile(t.reqLatencies, 90.0)
	if e != nil {
		panic(e)
	}
	return p90
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

	q := &sim.EventsQueue{
		{
			Time:        t,
			CallbackFun: server.startServer,
			Payload:     nil,
		},
		{
			Time:        t,
			CallbackFun: c.genLoad,
			Payload:     nil,
		},
	}

	sim.Run(maxTime, q, func() {
		c.stopLoadGen()
	})
}

func main() {
	// using  a fixed seed to make the simulation deterministic across runs
	var seed int64 = 1650543745
	mathrand.Seed(seed)
	failureRates := rangeInterval(0, 1, 0.01)

	loadVsRate := loadVsFailureRateByStrategy{
		failureRate:         failureRates,
		loadByRetryStrategy: make(map[retrierFactoryName][]float64),
	}
	latencyVsRate := requestLatenciesVsFailureRateByStrategy{
		failureRate:              failureRates,
		requestLatencyByStrategy: make(map[retrierFactoryName][]float64),
	}

	for _, retryStrategyName := range []retrierFactoryName{
		fixedRetry, circuitBreaker, tokenBucket, tokenBucketFixedRetry} {

		var loads []float64
		var p90Latencies []float64
		for _, failureRate := range failureRates {

			s := &stats{}
			runSimulation(s, failureRate, retryStrategyName)

			loads = append(loads, s.getLoad())
			p90Latencies = append(p90Latencies, s.getp90Latency())
		}
		loadVsRate.loadByRetryStrategy[retryStrategyName] = loads
		latencyVsRate.requestLatencyByStrategy[retryStrategyName] = p90Latencies
	}

	draw(latencyVsRate, loadVsRate)
}

type loadVsFailureRateByStrategy struct {
	// array of the failure rates used in the simulation
	failureRate []float64
	// the load that server experienced with each retry strategy.
	// It's a map between retry strategy name and an array of load (one for each failure
	// rate). This also means that len(loadByRetryStrategy[x]) == len(failureRate)
	loadByRetryStrategy map[retrierFactoryName][]float64
}

type requestLatenciesVsFailureRateByStrategy struct {
	// array of the failure rates used in the simulation
	failureRate []float64
	// p90 latency requests experienced with each strategy
	requestLatencyByStrategy map[retrierFactoryName][]float64
}
