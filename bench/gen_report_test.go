package bench

import (
	"bufio"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"os"
	"regexp"
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

		if pointP.MatchString(line) {
			groups := pointP.FindStringSubmatch(line)
			lineName := groups[1]
			x, err := strconv.ParseFloat(groups[2], 64)
			throwErr(t, err)
			y, err := strconv.ParseFloat(groups[3], 64)
			throwErr(t, err)
			currChart.addPoint(lineName, x, y)
		}

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
	series    map[string][]point
}

type point struct {
	x float64
	y float64
}

func NewLineChart(title string, xAxisName string, yAxisName string) *lineChart {
	return &lineChart{
		title,
		xAxisName,
		yAxisName,
		make(map[string][]point),
	}
}

func (l *lineChart) addPoint(line string, x float64, y float64) {
	l.series[line] = append(l.series[line], point{x, y})
}

func (l *lineChart) genChart(htmlFileName string) error {
	line := charts.NewLine()

	var xAxis []string
	series := make(map[string][]opts.LineData)
	for line, points := range l.series {
		data := make([]opts.LineData, 0)
		xAxis = make([]string, 0)
		for _, p := range points {
			xAxis = append(xAxis, fmt.Sprintf("%.0f", p.x))
			data = append(data, opts.LineData{Value: p.y})
		}
		series[line] = data
	}

	line.SetXAxis(xAxis)
	legend := make([]string, 0)
	for name, data := range series {
		line.AddSeries(name, data, charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
		legend = append(legend, name)
	}

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: l.title}),
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
	chart.addPoint("L1", 3, 45.33)
	chart.addPoint("L1", 6, 28.15)
	chart.addPoint("L1", 9, 77.7)
	chart.addPoint("L1", 12, 5.00)

	chart.addPoint("L2", 3, 10)
	chart.addPoint("L2", 6, 20)
	chart.addPoint("L2", 9, 30)
	chart.addPoint("L2", 12, 40)

	chart.addPoint("L3", 3, 3.5)
	chart.addPoint("L3", 6, 4.6)
	chart.addPoint("L3", 9, 5.7)
	chart.addPoint("L3", 12, 6.8)

	throwErr(t, chart.genChart("test.html"))
}
