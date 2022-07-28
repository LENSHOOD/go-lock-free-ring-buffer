package bench

import (
	"bufio"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"os"
	"regexp"
	"sort"
	"strconv"
	"testing"
)

func TestGenReport(t *testing.T) {
	datFile := os.Getenv("LFRING_BENCH_CHARTS_FILE")
	if datFile == "" {
		throwErrWithReason(t, "cannot get dat file from env LFRING_BENCH_CHARTS_FILE")
	}

	dat, err := os.Open(datFile)
	throwErr(t, err)

	titleP := regexp.MustCompile(TitlePattern)
	pointP := regexp.MustCompile(PointPattern)
	endP := regexp.MustCompile(EndPattern)

	dataScanner := bufio.NewScanner(dat)
	ordinal := 1
	var currChart *lineChart = nil
	for dataScanner.Scan() {
		line := dataScanner.Text()
		// title
		if titleP.MatchString(line) {
			groups := titleP.FindStringSubmatch(line)
			title := groups[1]
			xAxisName := groups[2]
			yAxisName := groups[3]

			if currChart != nil {
				throwErrWithReason(t, "previous chart haven't been end, new chart title: "+title)
			}

			currChart = NewLineChart(title, xAxisName, yAxisName)
		}

		// add points
		if pointP.MatchString(line) {
			groups := pointP.FindStringSubmatch(line)
			lineName := groups[1]
			x, err := strconv.ParseFloat(groups[2], 64)
			throwErr(t, err)
			y, err := strconv.ParseFloat(groups[3], 64)
			throwErr(t, err)
			currChart.addPoint(lineName, x, y)
		}

		// gen report
		if endP.MatchString(line) {
			throwErr(t, currChart.genChart(fmt.Sprintf("%d.html", ordinal)))
			ordinal++
			currChart = nil
		}
	}
}

const (
	TitlePattern = "#title=(.+),xAxis=(.+),yAxis=(.+)"
	PointPattern = `(.+)=\(([+-]?[0-9]*[.]?[0-9]+),([+-]?[0-9]*[.]?[0-9]+)\)`
	EndPattern   = "#end"
)

func throwErr(t *testing.T, err error) {
	if err == nil {
		return
	}

	t.Logf(fmt.Sprintf("[Gen Report Failed] %v", err))
	t.Fail()
}

func throwErrWithReason(t *testing.T, reason string) {
	t.Logf(fmt.Sprintf("[Gen Report Failed] %s", reason))
	t.Fail()
}

type lineChart struct {
	title     string
	xAxisName string
	yAxisName string
	series    map[string]points
}

type points map[float64][]float64

func (p points) addPoint(x float64, y float64) {
	p[x] = append(p[x], y)
}

func (p points) getXAxisAndYAxisData() (xAxisData []string, yAxisData []opts.LineData) {
	getMean := func(ys []float64) (y float64) {
		for _, v := range ys {
			y += v
		}

		y = y / float64(len(ys))
		return
	}

	var sortedXs []float64
	for x, _ := range p {
		sortedXs = append(sortedXs, x)
	}
	sort.Float64s(sortedXs)

	for _, x := range sortedXs {
		xAxisData = append(xAxisData, fmt.Sprintf("%.0f", x))
		yAxisData = append(yAxisData, opts.LineData{Value: getMean(p[x])})
	}

	return
}

func NewLineChart(title string, xAxisName string, yAxisName string) *lineChart {
	return &lineChart{
		title,
		xAxisName,
		yAxisName,
		make(map[string]points),
	}
}

func (l *lineChart) addPoint(line string, x float64, y float64) {
	p := l.series[line]
	if p == nil {
		p = make(points)
		l.series[line] = p
	}

	p.addPoint(x, y)
}

func (l *lineChart) genChart(htmlFileName string) error {
	line := charts.NewLine()

	var xAxis []string
	series := make(map[string][]opts.LineData)
	for line, points := range l.series {
		var yAxis []opts.LineData
		xAxis, yAxis = points.getXAxisAndYAxisData()
		series[line] = yAxis
	}

	line.SetXAxis(xAxis)
	legend := make([]string, 0)

	var names []string
	for name, _ := range series {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		line.AddSeries(name, series[name], charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
		legend = append(legend, name)
	}

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: l.title, Right: "center", Bottom: "bottom"}),
		charts.WithXAxisOpts(opts.XAxis{Name: l.xAxisName}),
		charts.WithYAxisOpts(opts.YAxis{Name: l.yAxisName}),
		charts.WithLegendOpts(opts.Legend{Data: legend, Show: true}),
		charts.WithTooltipOpts(opts.Tooltip{Trigger: "axis", Show: true}),
	)

	f, err := os.Create(htmlFileName)
	if err != nil {
		return err
	}

	if err = line.Render(f); err != nil {
		return err
	}

	return nil
}

func TestGenChart(t *testing.T) {
	chart := NewLineChart("test chart", "xAxis", "yAxis")
	chart.addPoint("L1", 3, 44.11)
	chart.addPoint("L1", 3, 45.22)
	chart.addPoint("L1", 3, 46.33)
	chart.addPoint("L1", 6, 27.15)
	chart.addPoint("L1", 6, 28.25)
	chart.addPoint("L1", 6, 29.35)
	chart.addPoint("L1", 9, 76.6)
	chart.addPoint("L1", 9, 77.7)
	chart.addPoint("L1", 9, 78.8)
	chart.addPoint("L1", 12, 4.00)
	chart.addPoint("L1", 12, 5.00)
	chart.addPoint("L1", 12, 6.00)

	chart.addPoint("L2", 3, 9)
	chart.addPoint("L2", 3, 10)
	chart.addPoint("L2", 3, 11)
	chart.addPoint("L2", 6, 19)
	chart.addPoint("L2", 6, 20)
	chart.addPoint("L2", 6, 21)
	chart.addPoint("L2", 9, 29)
	chart.addPoint("L2", 9, 30)
	chart.addPoint("L2", 9, 31)
	chart.addPoint("L2", 12, 39)
	chart.addPoint("L2", 12, 40)
	chart.addPoint("L2", 12, 41)

	chart.addPoint("L3", 3, 3.4)
	chart.addPoint("L3", 3, 3.5)
	chart.addPoint("L3", 3, 3.6)
	chart.addPoint("L3", 6, 4.5)
	chart.addPoint("L3", 6, 4.6)
	chart.addPoint("L3", 6, 4.7)
	chart.addPoint("L3", 9, 5.6)
	chart.addPoint("L3", 9, 5.7)
	chart.addPoint("L3", 9, 5.8)
	chart.addPoint("L3", 12, 6.7)
	chart.addPoint("L3", 12, 6.8)
	chart.addPoint("L3", 12, 6.9)

	throwErr(t, chart.genChart("test.html"))
}
