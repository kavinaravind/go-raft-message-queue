package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUnmarshalComment(t *testing.T) {
	jsonStr := `{
        "timestamp": "2022-01-01T00:00:00Z",
        "author": "test author",
        "content": "test content"
    }`

	var comment Comment
	err := json.Unmarshal([]byte(jsonStr), &comment)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2022-01-01T00:00:00Z")
	if !comment.Timestamp.Equal(expectedTime) {
		t.Errorf("expected %v, got %v", expectedTime, comment.Timestamp)
	}
	if comment.Author != "test author" {
		t.Errorf("expected %s, got %s", "test author", comment.Author)
	}
	if comment.Content != "test content" {
		t.Errorf("expected %s, got %s", "test content", comment.Content)
	}
}
