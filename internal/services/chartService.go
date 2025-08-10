package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"

	"github.com/wcharczuk/go-chart/v2"
	"ypeskov/budget-go/internal/dto"
)

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
	for _, item := range data {
        // Use Python-compatible label field name
        values = append(values, chart.Value{
            Label: item.CategoryName,
			Value: item.Amount,
		})
	}

	// Create pie chart
	pie := chart.PieChart{
		Width:  800,
		Height: 600,
		Values: values,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    20,
				Left:   20,
				Right:  20,
				Bottom: 20,
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