package analysis

// SupportResistance holds calculated levels
type SupportResistance struct {
	Pivot float64
	R1    float64
	R2    float64
	R3    float64
	S1    float64
	S2    float64
	S3    float64
}

// CalculatePivotPoints returns standard pivot points based on High, Low, Close
func CalculatePivotPoints(high, low, close float64) SupportResistance {
	p := (high + low + close) / 3

	r1 := (2 * p) - low
	s1 := (2 * p) - high

	r2 := p + (high - low)
	s2 := p - (high - low)

	r3 := high + 2*(p-low)
	s3 := low - 2*(high-p)

	return SupportResistance{
		Pivot: p,
		R1:    r1, R2: r2, R3: r3,
		S1: s1, S2: s2, S3: s3,
	}
}

// CalculateFibLevels returns Fibonacci retracement levels for a given range
func CalculateFibLevels(high, low float64, trend string) []float64 {
	diff := high - low
	if trend == "UP" {
		return []float64{
			high,
			high - (diff * 0.236),
			high - (diff * 0.382),
			high - (diff * 0.5),
			high - (diff * 0.618),
			low,
		}
	} else {
		return []float64{
			low,
			low + (diff * 0.236),
			low + (diff * 0.382),
			low + (diff * 0.5),
			low + (diff * 0.618),
			high,
		}
	}
}
