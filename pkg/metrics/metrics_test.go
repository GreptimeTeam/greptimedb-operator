package metrics

import (
	"testing"
	"time"
)

func TestParseImagePullingDuration(t *testing.T) {
	tests := []struct {
		message  string
		duration time.Duration
	}{
		{
			message:  `Successfully pulled image "greptime/greptimedb:latest" in 314.950966ms (6.159733714s including waiting)`,
			duration: time.Duration(314950966 * time.Nanosecond),
		},
		{
			message:  `Successfully pulled image "greptime/greptimedb:latest" in 10s (6.159733714s including waiting)`,
			duration: time.Duration(10 * time.Second),
		},
	}

	for _, test := range tests {
		duration, err := parseImagePullingDuration(test.message)
		if err != nil {
			t.Fatalf("parse image pulling duration from message %q: %v", test.message, err)
		}

		if duration != test.duration {
			t.Fatalf("expected duration: %v, got: %v", test.duration, duration)
		}
	}
}

func TestParseImageName(t *testing.T) {
	tests := []struct {
		message string
		name    string
	}{
		{
			message: `Successfully pulled image "greptime/greptimedb:latest" in 314.950966ms (6.159733714s including waiting)`,
			name:    "greptime/greptimedb:latest",
		},
		{
			message: `Successfully pulled image "greptime/greptimedb:v0.9.5" in 314.950966ms (6.159733714s including waiting)`,
			name:    "greptime/greptimedb:v0.9.5",
		},
	}

	for _, test := range tests {
		name, err := parseImageName(test.message)
		if err != nil {
			t.Fatalf("parse image name from message %q: %v", test.message, err)
		}

		if name != test.name {
			t.Fatalf("expected name: %q, got: %q", test.name, name)
		}
	}
}
