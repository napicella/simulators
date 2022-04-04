package main

import (
	"encoding/csv"
	"fmt"
	"math"
	mathrand "math/rand"
	"os"
)

// runs a simulation with 3 servers:
// parallel: send the request to all three servers and wait until the last one has answered
// sequential: server 1 calls server 2 that callers server 3
// Store the latency experienced in a cvs file. The cvs file is used to get statics and
// plots with the "r" language.
//
// After running this simulation, run 'Rscript ./plot.r' to generate a plot
func main() {
	_ = runSimulationParallel()
	_ = runSimulationSerial()
}

const numberOfIterations = 5000

func runSimulationSerial() error {
	csvFile, err := os.Create("/tmp/result-serial.csv")
	if err != nil {
		return err
	}
	defer csvFile.Close()
	writer := csv.NewWriter(csvFile)
	writer.Write([]string{"x", "y"})

	for i := 0; i < numberOfIterations; i++ {
		firstSvr := sampleLatency()
		secondSvr := sampleLatency()
		thirdSvr := sampleLatency()
		latency := firstSvr + secondSvr + thirdSvr
		writer.Write([]string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%f", latency)})
	}
	writer.Flush()
	return nil
}

func runSimulationParallel() error {
	csvFile, err := os.Create("/tmp/result-parallel.csv")
	if err != nil {
		return err
	}
	defer csvFile.Close()
	csvwriter := csv.NewWriter(csvFile)
	csvwriter.Write([]string{"x", "y"})

	for i := 0; i < numberOfIterations; i++ {
		firstSvr := sampleLatency()
		secondSvr := sampleLatency()
		thirdSvr := sampleLatency()
		maxLatency := max(firstSvr, secondSvr, thirdSvr)
		csvwriter.Write([]string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%f", maxLatency)})
	}
	csvwriter.Flush()
	return nil
}

func max(a, b, c float64) float64 {
	return math.Max(a, math.Max(b, c))
}

func sampleLatency() float64 {
	desiredStdDev := 0.5
	desiredMean := 0.5
	return math.Abs(mathrand.NormFloat64()*desiredStdDev + desiredMean)
}
