package router

import (
	"git.mark1708.ru/me/convertr/internal/backend"
)

// Step is a single conversion hop in a route.
type Step struct {
	From    string
	To      string
	Backend backend.Backend
}

// Route is an ordered sequence of steps from source to target format.
type Route struct {
	Steps []Step
}

// TotalCost returns the sum of edge costs across all steps.
func (r *Route) TotalCost() int {
	total := 0
	for _, s := range r.Steps {
		for _, c := range s.Backend.Capabilities() {
			if c.From == s.From && c.To == s.To {
				total += c.Cost
				break
			}
		}
	}
	return total
}
