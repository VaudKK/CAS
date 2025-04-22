package models

import "time"

type Audit struct {
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
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

type Organziation struct {
	ID   int
	Name string
	Audit
}
