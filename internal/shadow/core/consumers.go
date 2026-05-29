package core

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// Consumer is a downstream repo we mirror to verify the workflows draft doesn't break it.
type Consumer struct {
	Repo string `json:"repo"`
	Ref  string `json:"ref"`
}

var repoRe = regexp.MustCompile(`^[^/\s]+/[^/\s]+$`)

// ParseConsumers parses + validates the workflows' shadow-consumers.json. ref defaults to "main".
func ParseConsumers(jsonStr string) ([]Consumer, error) {
	var data any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid consumers JSON: %w", err)
	}
	arr, ok := data.([]any)
	if !ok {
		return nil, fmt.Errorf("expected an array of consumers, got %T", data)
	}
	out := make([]Consumer, 0, len(arr))
	for i, entry := range arr {
		c, err := parseConsumer(entry, i)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func parseConsumer(entry any, index int) (Consumer, error) {
	obj, ok := entry.(map[string]any)
	if !ok {
		return Consumer{}, fmt.Errorf("consumer[%d]: expected an object", index)
	}
	repo, ok := obj["repo"].(string)
	if !ok || !repoRe.MatchString(repo) {
		return Consumer{}, fmt.Errorf("consumer[%d].repo: expected \"owner/name\", got %v", index, obj["repo"])
	}
	ref := "main"
	if raw, present := obj["ref"]; present {
		s, ok := raw.(string)
		if !ok || s == "" {
			return Consumer{}, fmt.Errorf("consumer[%d].ref: expected a non-empty string, got %v", index, raw)
		}
		ref = s
	}
	return Consumer{Repo: repo, Ref: ref}, nil
}
