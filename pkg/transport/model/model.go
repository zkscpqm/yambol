package model

import "time"

type BasicInfo struct {
	Uptime  time.Duration `json:"uptime"`
	Version string        `json:"version"`
}
