package flags

import "time"

type Flag struct {
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
	Rules     []Rule    `json:"rules"`
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Rule struct {
	Attribute  string `json:"attribute"`
	Operator   string `json:"operator"`
	Value      string `json:"value"`
	Percentage int    `json:"percentage,omitempty"`
}

type EvaluationContext struct {
	SubjectID  string            `json:"subject_id"`
	Attributes map[string]string `json:"attributes"`
}

type EvaluationResult struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
	Version int64  `json:"version"`
}

type Event struct {
	Type    string `json:"type"`
	Flag    Flag   `json:"flag"`
	Version int64  `json:"version"`
}
