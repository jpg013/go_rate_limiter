package ratelimit

import "github.com/segmentio/ksuid"

func genKsuid() string {
	return ksuid.New().String()
}
