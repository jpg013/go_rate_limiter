package ratelimit

import "github.com/segmentio/ksuid"

func GenKsuid() string {
	return ksuid.New().String()
}
