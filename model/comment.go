package model

import "time"

type Comment struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	Author    string    `json:"author,omitempty"`
	Content   string    `json:"content,omitempty"`
}
