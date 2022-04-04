package main

import (
	"io"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func generateLineItems(data []float64) []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < len(data); i++ {
		items = append(items, opts.LineData{Value: data[i]})
	}
	return items
}

func itemRange(length int) []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < length; i++ {
		items = append(items, opts.LineData{Value: i})
	}
	return items
}

func draw(latencies []float64) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic line example"}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Time",
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Request latency",
			Type: "value",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "axis"}),
		charts.WithLegendOpts(opts.Legend{Right: "80%"}),
	)

	line.SetXAxis(itemRange(len(latencies))).
		AddSeries("Request latency", generateLineItems(latencies))

	page := components.NewPage()
	page.AddCharts(line)
	f, err := os.Create("line.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}
