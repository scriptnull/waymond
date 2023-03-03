package cron

import (
	"fmt"
	"testing"

	"github.com/knadh/koanf/v2"
)

// TODO: complete this test
func TestParseConfig(t *testing.T) {
	tt := []struct {
		config    *koanf.Koanf
		shouldErr bool
	}{}

	for _, tc := range tt {
		trigger, err := ParseConfig(tc.config)
		fmt.Printf("trigger: %+v, err: %s", trigger, err)
	}
}
