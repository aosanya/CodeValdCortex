package models

import "time"

// Overview represents the overview document in an agency's database
type Overview struct {
	Key          string    `json:"_key,omitempty"`
	AgencyID     string    `json:"agency_id"`
	Introduction string    `json:"introduction"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpdateOverviewRequest is the request body for updating overview
type UpdateOverviewRequest struct {
	Introduction string `json:"introduction"`
}
