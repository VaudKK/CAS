package models

import "time"

type Audit struct {
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
	CreatedBy  string    `json:"createdBy"`
	ModifiedBy string    `json:"modifiedBy"`
}

type User struct {
	ID             int       `json:"id"`
	UserName       string    `json:"userName"`
	Email          string    `json:"email"`
	OrganizationId int       `json:"organizationId"`
	Password       string    `json:"password"`
	Verified       bool      `json:"verified"`
	Active         bool      `json:"active"`
	LastLogin      time.Time `json:"lastLogin"`
	Audit
}

type Fund struct {
	ID             int                `json:"id"`
	ReceiptNo      string             `json:"receiptNo"`
	BreakDown      map[string]float64 `json:"breakDown"`
	Total          float64            `json:"total"`
	OrganizationId int                `json:"organizationId"`
	Date           string             `json:"date"`
	Contributor    string             `json:"contributor"`
	Audit
}

type UpdateFund struct {
	Contributor string             `json:"contributor"`
	Date        string             `json:"date"`
	Total       float64            `json:"total"`
	BreakDown   map[string]float64 `json:"breakDown"`
}

type Organziation struct {
	ID   int
	Name string
	Audit
}

type Otp struct {
	ID               int       `json:"id,omitempty"`
	VerificationMode string    `json:"verificationMode,omitempty"`
	Otp              string    `json:"otp,omitempty"`
	Expiry           time.Time `json:"expiry,omitempty"`
	Used             bool      `json:"used,omitempty"`
	Subject          string    `json:"subject,omitempty"`
	SessionId        string    `json:"sessionId,omitempty"`
}

type MonthlyStats struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type MonthlySummations struct {
	Category string
	Date     string
	Total    float64
}

type StatisticalVariance struct {
	Category      string
	Total         float64
	PreviousTotal float64
	Difference    float64
}

type Variance struct {
	Category     string  `json:"category"`
	CurrentValue float64 `json:"currentValue"`
	Percentage   float32 `json:"percentage"`
	Direction    int8    `json:"direction"`
}
