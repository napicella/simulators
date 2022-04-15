package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestClientServerSim(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Server Simulation Suite")
}

var _ = When("failure rate is 1 (all calls fail)", func() {
	It("the load is (number_of_retries + 1) * 100 %", func() {
		s := &stats{}
		failureRate := 1.0
		runSimulation(s, failureRate)

		load := (float64(s.attempts) / float64(s.uniqueCalls)) * 100
		// assumes 3 retries
		Expect(load).To(Equal(400.0))
	})
})
