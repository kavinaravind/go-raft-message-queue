package model

import "time"

// Comment is the model for a comment
type Comment struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	Author    string    `json:"author,omitempty"`
	Content   string    `json:"content,omitempty"`
}
