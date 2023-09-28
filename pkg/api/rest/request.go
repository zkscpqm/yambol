package rest

import "time"

type YambolMessageRequest struct {
	Message string `json:"message"`
	TTL     int    `json:"ttl,omitempty"`
}

type QueuesPostRequest struct {
	Name         string `json:"name"`
	MinLength    int    `json:"min_length,omitempty"`
	MaxLength    int    `json:"max_length,omitempty"`
	MaxSizeBytes int    `json:"max_size_bytes,omitempty"`
	TTL          int    `json:"ttl,omitempty"`
}

func (r *QueuesPostRequest) TTLSeconds() time.Duration {
	return time.Duration(r.TTL) * time.Second
}
