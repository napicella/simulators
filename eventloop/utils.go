package main

import "math"

func max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

func rangeInterval(start, end, increment float64) []float64 {
	var res []float64
	for i := start; math.Round(i*100)/100 <= end; i = i + increment {
		res = append(res, math.Round(i*100)/100)
	}
	return res
}
