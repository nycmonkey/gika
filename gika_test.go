package gika

import (
	"os"
	"testing"
)

func TestDefaultTikaAddr(t *testing.T) {
	os.Setenv("TIKA_PORT", "tcp://localhost:9998")
	tika, err := NewTikaFromDockerEnv()
	if err != nil {
		t.Fatal("Tika did not instantiate")
	}

	if tika.url != "http://localhost:9998/tika" {
		t.Fatal("Tika did not pull URL from environment")
	}
}
