package story

// Median gives the low, median and high of values, so the sector overview shares the strips' median.
func Median(values []float64) Range { return rangeOf(values) }

// Profitable keeps positive values: PE medians exclude loss-makers.
func Profitable(values []float64) []float64 { return positive(values) }
