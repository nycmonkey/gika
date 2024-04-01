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
	want := "http://localhost:9998"
	if tika.url != want {
		t.Fatalf("Tika did not pull URL from environment: expected %s, got %s", want, tika.url)
	}
}
