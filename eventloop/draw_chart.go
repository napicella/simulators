package main

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"math"
	"os"
)

func draw(
	latencies requestLatenciesVsFailureRateByStrategy,
	loadVsFailureRate loadVsFailureRateByStrategy) {

	latenciesChart := drawLatencies(latencies)
	loadChart := drawLoad(loadVsFailureRate)

	page := components.NewPage()
	page.PageTitle = "Retriers"
	page.AddCharts(loadChart)
	page.AddCharts(latenciesChart)

	f, err := os.Create("./build/graphs/stats.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

func drawLatencies(latencies requestLatenciesVsFailureRateByStrategy) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Latency over failure rate"}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Failure Rate",
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Request latency (log)",
			Type: "value",
			Max:  10,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Right: "80%"}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Type:  "png",
					Name:  "latency",
					Title: "png",
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true, Right: "150", Orient: "vertical"}),
	)

	line.SetXAxis(latencies.failureRate)
	for strategyName, latencyArray := range latencies.requestLatencyByStrategy {
		logLatencies := logE(latencyArray)
		line = line.AddSeries(
			fmt.Sprintf("%s - Latency", strategyName),
			generateLineItems(logLatencies),
			charts.WithLineStyleOpts(opts.LineStyle{Color: getLineColorForStrategy(strategyName)}),
			charts.WithItemStyleOpts(opts.ItemStyle{Color: getLineColorForStrategy(strategyName)}),
		)
	}

	return line
}

func drawLoad(loadVsFailureRate loadVsFailureRateByStrategy) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Load over failure rate"}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Failure Rate",
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Load",
			Type: "value",
			Min:  100,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Right: "80%"}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true,
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show:  true,
					Type:  "png",
					Name:  "failure-rate",
					Title: "png",
				},
			},
		}),
		charts.WithLegendOpts(opts.Legend{Show: true, Right: "150", Orient: "vertical"}),
	)

	line.SetXAxis(loadVsFailureRate.failureRate)
	for strategyName, loadArray := range loadVsFailureRate.loadByRetryStrategy {
		line = line.AddSeries(
			fmt.Sprintf("%s - Load ", strategyName),
			generateLineItems(loadArray),
			charts.WithLineStyleOpts(opts.LineStyle{Color: getLineColorForStrategy(strategyName)}),
			charts.WithItemStyleOpts(opts.ItemStyle{Color: getLineColorForStrategy(strategyName)}),
		)
	}

	return line
}

func generateLineItems(data []float64) []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < len(data); i++ {
		items = append(items, opts.LineData{Value: data[i]})
	}
	return items
}

func logE(input []float64) []float64 {
	var out []float64
	for _, f := range input {
		// adding a bias of 1.0 to avoid getting negative numbers when the input is less
		// than 1
		out = append(out, math.Log(f+1.0))
	}
	return out
}

func getLineColorForStrategy(retryStrategy retrierFactoryName) string {
	switch retryStrategy {
	case fixedRetry:
		return "grey"
	case circuitBreaker:
		return "blue"
	case tokenBucket:
		return "red"
	case tokenBucketFixedRetry:
		return "green"
	default:
		panic(fmt.Sprintf("invalid retryStrategy name %s", retryStrategy))
	}
}
