package imports

import "time"

type ImportModel struct {
	Name      string
	ReceiptNo string
	Total     float64
	BreakDown map[string]float64
	Date      time.Time
}
