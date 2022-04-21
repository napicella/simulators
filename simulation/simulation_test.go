package sim

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math"
	mathrand "math/rand"
	"testing"
)

func TestSimulationCore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Simulation Suite")
}

var _ = Describe("Client side load balancing", func() {
	It("distributes traffic more or less evenly among the servers", func() {
		stats := newStats()
		runSimulation(stats)

		Expect(stats.callsByEndpoint["server-1"]).To(Equal(1682))
		Expect(stats.callsByEndpoint["server-2"]).To(Equal(1591))
		Expect(stats.callsByEndpoint["server-3"]).To(Equal(1725))
	})
})

// A very simple simulation to show how to use the simulator.
//
// It runs an sample simulation with one client and one server. The client calls the server
// periodically according to a normal variable with mean 1 and std 0.1; the client
// simulates load balancing by distributing traffic uniformly among the three backend
// servers (server-1, server-2 and server-3). The server records stats on the call made to
// each of the backend server.
// We expect the calls of each server to be more or less even :)
func runSimulation(stats *stats) {
	// For the test to be deterministic, the simulation uses math/rand (instead of
	// crypto/rand) and a constant value for the seed
	var seed int64 = 1650536787
	mathrand.Seed(seed)

	server := &server{
		stats: stats,
	}
	client := client{
		requestsPerSecond: 1.0,
		server:            server,
	}

	maxTime := 5000.0
	q := &EventsQueue{
		{
			// starting event in the simulation
			Time:        0,
			CallbackFun: client.genLoad,
			Payload:     nil,
		},
	}

	Run(maxTime, q, func() {
		client.stopLoadGen()
	})
}

type client struct {
	requestsPerSecond int
	server            *server
	drain             bool
}

// genLoad calls the server periodically according to a normal variable with mean 1 and
// std 0.1
func (t *client) genLoad(time float64, payload interface{}) []Event {
	if t.drain {
		// stop generating load, that is do not return any more events
		return nil
	}

	// pick time for next call
	desiredStdDev := 0.1
	desiredMean := float64(t.requestsPerSecond)
	nextCall := math.Abs(mathrand.NormFloat64()*desiredStdDev + desiredMean)

	// pick server to call
	var req request
	l := mathrand.Float64()
	switch {
	case l <= 0.33:
		req.endpoint = "server-1"
	case l > 0.33 && l <= 0.66:
		req.endpoint = "server-2"
	default:
		req.endpoint = "server-3"
	}

	// return two events: one event to call the server at the `time` and another one to
	// generate more load to the server in `time + nextCall`
	return []Event{
		{
			// generate an event to call the server
			Time:        time,
			CallbackFun: t.server.call,
			Payload:     req,
		},
		{
			// generate load again in time + nextCall
			Time:        time + nextCall,
			CallbackFun: t.genLoad,
			Payload:     nil,
		},
	}
}

func (t *client) stopLoadGen() {
	t.drain = true
}

type server struct {
	stats *stats
}

func (t *server) call(time float64, payload interface{}) []Event {
	req, ok := payload.(request)
	if !ok {
		panic("server payload is not a request")
	}
	t.stats.recordCall(req.endpoint)

	return nil
}

type request struct {
	endpoint string
}

type stats struct {
	callsByEndpoint map[string]int
}

func newStats() *stats {
	return &stats{callsByEndpoint: make(map[string]int)}
}

func (t *stats) recordCall(endpoint string) {
	t.callsByEndpoint[endpoint]++
}
