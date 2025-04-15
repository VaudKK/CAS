package models

import "time"

type Audit struct {
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

type Member struct {
	ID             int
	FirstName      string
	LastName       string
	OrganizationId int
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

type Organziation struct {
	ID   int
	Name string
	Audit
}
