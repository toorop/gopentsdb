package openstdb

// Point represent a "point" (really ?)
type Point struct {
	// A generic name for the time series such as sys.cpu.user,
	// stock.quote or env.probe.temp.
	Metric string `json:"metric"`

	// A Unix/POSIX Epoch timestamp in second or milliseconds defined as the number
	// of milliseconds that have elapsed since January 1st, 1970 at 00:00:00
	// UTC time. Only positive timestamps are supported at this time.
	Timestamp int64 `json:"timestamp"`

	// A numeric value to store at the given timestamp for the time series.
	Value float64 `json:"value"`

	// A key/value pair consisting of a tagk (the key) and a tagv (the value).
	// Each data point must have at least one tag.
	Tags map[string]string `json:"tags"`
}

// NewPoint init aand retuen a new openstdb Point
func NewPoint() Point {
	p := &Point{}
	p.Tags = make(map[string]string)
	return *p
}

// Getters and setter: there goals are to insure that data are correctly
// formated & typed.
// Of course you can, as your owns risks, directly use Point struct as there
// fields are exported.

// TODO
