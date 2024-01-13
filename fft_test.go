package pipe

import (
	"os"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// generate random data for line chart
// func generateLineItems() []opts.LineData {
// 	items := make([]opts.LineData, 0)
// 	for i := 0; i < 1024*32; i++ {
// 		items = append(items, opts.LineData{Value: rand.Intn(300)})
// 	}
// 	return items
// }

func generateLineItemsFromSlice(in []float64) []opts.LineData {
	items := make([]opts.LineData, len(in))
	for i := range in {
		items[i] = opts.LineData{Value: in[i]}
	}
	return items
}

func TestFFT(t *testing.T) {
	c := NewController(1024, 1, 0.001)
	dataWire := SrcSine(c, 10*1000.0/1024.0, 0)
	wireEnd := dataWire.NewWireEnd()

	c.Start()
	data, _ := wireEnd.GetData()
	c.Stop()

	fftResult := FFT(data, nil)
	Norm(fftResult)

	magResult := Mag(fftResult, nil)

	xAxisValues := Freq(len(data), 1000)

	/////////////////////////////////
	/////////////////////////////////

	// create a new line instance
	line := charts.NewLine()
	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    "Line example in Westeros theme",
			Subtitle: "Line chart rendered by the http server this time",
		}),

		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
	)

	// Put data into instance

	line.SetXAxis(xAxisValues)

	line.AddSeries("sine Wave", generateLineItemsFromSlice(magResult))

	f, _ := os.Create("chart.html")
	_ = line.Render(f)
}
