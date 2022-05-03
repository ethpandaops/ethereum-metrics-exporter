package exporter

import "math"

func round(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}
