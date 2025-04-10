package models

import "time"

type Audit struct {
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type Member struct {
	ID             int
	FirstName      string
	LastName       string
	OrganizationId int
	Audit
}

type Fund struct {
	ID             int
	BreakDown      map[string]float64
	Total         float64
	OrganizationId int
	Date           string
	Contributor    string
	Audit
}

type Organziation struct {
	ID   int
	Name string
	Audit
}
