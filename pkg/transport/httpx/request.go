package httpx

import (
	"time"
	"yambol/config"
	"yambol/pkg/util"
)

type MessageRequest struct {
	Message string `json:"message"`
	TTL     int64  `json:"ttl,omitempty"`
}

type QueuesPostRequest struct {
	Name string `json:"name"`
	config.QueueConfig
}

func (r *QueuesPostRequest) TTLSeconds() time.Duration {
	return util.Seconds(r.TTL)
}
