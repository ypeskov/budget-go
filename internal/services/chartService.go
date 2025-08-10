package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"
	"sort"

	"ypeskov/budget-go/internal/dto"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// pieSliceLabel represents a label and amount for a pie slice with a color
type pieSliceLabel struct {
	Label  string
	Amount float64
	Color  drawing.Color
}

// pieCustomRenderer draws labels outside slices and percentage text near each slice
type pieCustomRenderer struct {
	labels        []pieSliceLabel
	labelDistance float64
	pctDistance   float64
	fontColor     drawing.Color
}

func (pr *pieCustomRenderer) Render(r chart.Renderer, cb chart.Box, chartDefaults chart.Style) {
	drawPieLabels(r, cb, pr.labels, pr.labelDistance, pr.pctDistance, pr.fontColor)
}

type ChartService interface {
	GeneratePieChart(data []dto.ExpensesDiagramDataDTO, currency string) (*dto.ChartImageDTO, error)
}

type ChartServiceInstance struct{}

func NewChartService() ChartService {
	return &ChartServiceInstance{}
}

func (s *ChartServiceInstance) GeneratePieChart(data []dto.ExpensesDiagramDataDTO, currency string) (*dto.ChartImageDTO, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to generate chart")
	}

	// Sort data by amount in descending order (like Python)
	sort.Slice(data, func(i, j int) bool {
		return data[i].Amount > data[j].Amount
	})

	// Calculate total for percentages
	var total float64
	for _, item := range data {
		total += item.Amount
	}

	if total == 0 {
		return nil, fmt.Errorf("total amount is zero")
	}

	// Prepare chart values
	var values []chart.Value
	var labels []pieSliceLabel
	for _, item := range data {
		// Use consistent colors and suppress in-slice labels; rely on legend outside
		col := drawing.ColorFromHex(item.Color)
		values = append(values, chart.Value{
			Label: item.CategoryName,
			Value: item.Amount,
			Style: chart.Style{
				FillColor:   col,
				StrokeColor: col,
				FontSize:    10,
			},
		})
	}

	// Create pie chart
	// Use a square canvas to mimic matplotlib's axis('equal') circular pie
	pie := chart.PieChart{
		Width:  500,
		Height: 500,
		Values: values,
		Background: chart.Style{
			Padding: chart.Box{
				// Provide even padding so outside labels at ~1.1R fit similarly to matplotlib's labeldistance
				Top:    10,
				Left:   10,
				Right:  10,
				Bottom: 10,
			},
		},
	}

	// Render chart to buffer
	buffer := bytes.NewBuffer([]byte{})
	err := pie.Render(chart.PNG, buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to render chart: %w", err)
	}

	// Encode to base64
	base64Image := base64.StdEncoding.EncodeToString(buffer.Bytes())

	return &dto.ChartImageDTO{
		Image: fmt.Sprintf("data:image/png;base64,%s", base64Image),
	}, nil
}

// pieLabelElement draws labels outside slices and percentage text near the slice, similar to matplotlib's
// labeldistance and pctdistance configuration.
func drawPieLabels(r chart.Renderer, cb chart.Box, labels []pieSliceLabel, labelDistance, pctDistance float64, fontColor drawing.Color) {
	var total float64
	for _, l := range labels {
		total += l.Amount
	}
	if total <= 0 {
		return
	}
	cx := float64(cb.Left + cb.Width()/2)
	cy := float64(cb.Top + cb.Height()/2)
	radius := float64(chart.MinInt(cb.Width(), cb.Height()) / 2)
	angle := 0.0
	r.SetFontColor(fontColor)
	for _, l := range labels {
		if l.Amount <= 0 {
			continue
		}
		frac := l.Amount / total
		sweep := frac * 2 * math.Pi
		mid := angle + sweep/2
		// percentage text
		pct := frac * 100
		pctLabel := fmt.Sprintf("%.1f%%", pct)
		px := cx + pctDistance*radius*math.Cos(mid)
		py := cy + pctDistance*radius*math.Sin(mid)
		r.Text(pctLabel, int(px), int(py))
		// outside label
		lx := cx + labelDistance*radius*math.Cos(mid)
		ly := cy + labelDistance*radius*math.Sin(mid)
		r.Text(l.Label, int(lx), int(ly))
		angle += sweep
	}
}
