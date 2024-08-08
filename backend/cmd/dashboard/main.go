package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	ganalytics "github.com/tdadadavid/analytics"
)

var (
	siteID string
	start  int64
	end    int64

	what           ganalytics.QueryType = ganalytics.QueryPageViews
	pos            int                 = 0
	row            int
	current        []ganalytics.Metric
	currentTitle   string
	dateRangeMode  int = 0
)

func main() {
	flag.StringVar(&siteID, "site", "", "site id")
	flag.Int64Var(&start, "start", 0, "start date as uint32")
	flag.Int64Var(&end, "end", 0, "end date as uint32")
	flag.Parse()

	if err := ui.Init(); err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	renderPageViews()

	events := ui.PollEvents()
	for {
		e := <-events
		switch e.ID {
		case "q", "<C-c>":
			return
		case "p":
			renderPageViews()
		case "b":
			what = ganalytics.QueryBrowsers
			renderPie("Browsers")
		case "o":
			what = ganalytics.QueryOs
			renderPie("OSes")
		case "c":
			what = ganalytics.QueryCountry
			renderPie("Countries")
		case "r":
			what = ganalytics.QueryReferrerHost
			metrics, err := getMetric(what)
			if err != nil {
				log.Fatal(err)
			}

			current = metrics
			title("Referrer hosts")
			renderTable()
		case "t":
			if what == ganalytics.QueryPageViews {
				metrics, err := getMetric(ganalytics.QueryPageViewList)
				if err != nil {
					log.Fatal(err)
				}

				current = metrics
			}
			renderTable()
		case "<Up>":
			row -= 1
			if row < 0 {
				row = 0
			}

			renderTable()
		case "<Down>":
			row += 1
			if row >= len(current) {
				row = len(current) - 1
			}

			renderTable()
		case "<Left>":
			pos += 1
			renderPageViews()
		case "<Right>":
			pos -= 1
			renderPageViews()
		case "z":
			dateRangeMode += 1
			if dateRangeMode > 3 {
				dateRangeMode = 0
			}

			end = int64(ganalytics.TimeToInt(time.Now()))
			days := 30
			if dateRangeMode == 1 {
				days = 90
			} else if dateRangeMode == 2 {
				days = 180
			} else if dateRangeMode == 3 {
				days = 365
			}

			start = int64(ganalytics.TimeToInt(time.Now().Add(-24 * time.Duration(days) * time.Hour)))

			ui.Clear()
			p := widgets.NewParagraph()
			p.Text = fmt.Sprintf("New time range: %d to %d", start, end)
			p.SetRect(5, 5, 45, 10)

			ui.Render(p)
		default:
			fmt.Println(e.ID)
		}
	}
}

func renderPageViews() {
	what = ganalytics.QueryPageViews

	ui.Clear()

	metrics, err := getMetric(ganalytics.QueryPageViews)
	if err != nil {
		log.Fatal(err)
	}

	current = metrics

	m := make(map[uint32]uint64)
	var keys []int
	for _, metric := range metrics {
		v, ok := m[metric.OccuredAt]
		if !ok {
			keys = append(keys, int(metric.OccuredAt))
		}

		m[metric.OccuredAt] = v + metric.Count
	}

	sort.Ints(keys)

	keys = truncate(keys)

	var data []float64
	var labels []string
	for _, key := range keys {
		v := m[uint32(key)]
		data = append(data, float64(v))
		labels = append(labels, fmt.Sprintf("%d", key)[6:])
	}

	bc := widgets.NewBarChart()
	bc.Title = title("Page views")
	bc.Data = data
	bc.Labels = labels
	//bc.BarWidth = 6
	bc.BarColors = []ui.Color{ui.ColorBlue}
	bc.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	bc.NumStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	bc.SetRect(5, 5, 90, 20)

	ui.Render(bc)
}

func truncate(keys []int) []int {
	total := len(keys)
	if total > 16 {
		offset := total - pos - 16
		if offset < 0 {
			offset = 0
		}
		last := offset + 16
		if last >= total {
			last = total - 1
		}

		return keys[offset:last]
	}
	return keys
}

func renderPie(s string) {
	ui.Clear()

	metrics, err := getMetric(what)
	if err != nil {
		log.Fatal(err)
	}

	current = metrics

	var total uint64
	for _, metric := range metrics {
		total += metric.Count
	}

	var data []float64
	var labels []string
	for _, metric := range metrics {
		data = append(data, float64(metric.Count)/float64(total))
		labels = append(labels, metric.Value)
	}

	pc := widgets.NewPieChart()
	pc.Title = title(s)
	pc.Data = data
	pc.AngleOffset = -0.5 * math.Pi
	pc.LabelFormatter = func(i int, v float64) string {
		return fmt.Sprintf("%s (%.02f%%)", labels[i], v*100)
	}
	pc.SetRect(5, 5, 70, 20)

	ui.Render(pc)
}

func title(s string) string {
	currentTitle = fmt.Sprintf("%s - start:%d end:%d", s, start, end)
	return currentTitle
}

func renderTable() {
	ui.Clear()

	var data [][]string
	total := len(current)
	last := row + 6
	if last >= total {
		last = total
	}

	data = append(data, []string{"Value", "Count"})
	for _, metric := range current[row:last] {
		data = append(data, []string{metric.Value, fmt.Sprintf("%d", metric.Count)})
	}

	dt := widgets.NewTable()
	dt.Title = currentTitle
	dt.Rows = data
	dt.SetRect(5, 5, 85, 20)

	ui.Render(dt)
}