package rules

import "math"

// Gate thresholds for the plain agreement report (go/no-go 90%, PR class 95%).
const (
	gateGo = 90
	gatePR = 95
)

// DisplayPct rounds agree/total to one decimal. It is not the pass bit.
func DisplayPct(agree, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(agree)*1000/float64(total)) / 10
}

// GatePct is the pass threshold for path, if it has one.
func GatePct(path string) (int, bool) {
	switch path {
	case PathGo:
		return gateGo, true
	case PathPR:
		return gatePR, true
	default:
		return 0, false
	}
}
