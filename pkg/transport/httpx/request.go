package httpx

import "time"

type YambolMessageRequest struct {
	Message string `json:"message"`
	TTL     int64  `json:"ttl,omitempty"`
}

type QueuesPostRequest struct {
	Name         string `json:"name"`
	MinLength    int64  `json:"min_length,omitempty"`
	MaxLength    int64  `json:"max_length,omitempty"`
	MaxSizeBytes int64  `json:"max_size_bytes,omitempty"`
	TTL          int64  `json:"ttl,omitempty"`
}

func (r *QueuesPostRequest) TTLSeconds() time.Duration {
	return time.Duration(r.TTL) * time.Second
}
