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

type Contribution struct {
	ID             int
	Category       string
	Amount         float32
	OrganizationId int
	Date           time.Time
	Contributor    string
	Audit
}

type Organziation struct {
	ID   int
	Name string
	Audit
}
