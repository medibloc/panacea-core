package main

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"
)

func TestFaultConfigValidation(t *testing.T) {
	t.Parallel()
	valid := faultConfig{
		ListenAddress: ":26656",
		TargetAddress: "validator:26656",
		DialTimeout:   time.Second,
		Seed:          1,
	}
	if err := valid.validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	invalid := valid
	invalid.TargetAddress = ""
	invalid.Jitter = -time.Millisecond
	invalid.DialTimeout = 0
	invalid.Seed = 0
	if err := invalid.validate(); err == nil {
		t.Fatal("invalid config unexpectedly accepted")
	}
}

func TestJitteredDelayIsDeterministicAndBounded(t *testing.T) {
	t.Parallel()
	first := rand.New(rand.NewSource(11))
	second := rand.New(rand.NewSource(11))
	for index := 0; index < 100; index++ {
		got := jitteredDelay(50*time.Millisecond, 20*time.Millisecond, first)
		want := jitteredDelay(50*time.Millisecond, 20*time.Millisecond, second)
		if got != want {
			t.Fatalf("delay %d = %s, want deterministic %s", index, got, want)
		}
		if got < 30*time.Millisecond || got > 70*time.Millisecond {
			t.Fatalf("delay %d outside [30ms,70ms]: %s", index, got)
		}
	}
}

func TestProxyCopyInjectsBoundedEffectiveLoss(t *testing.T) {
	t.Parallel()
	input := bytes.Repeat([]byte("x"), proxyBufferSize*3)
	var output bytes.Buffer
	config := faultConfig{DropEvery: 2}
	err := proxyCopy(
		context.Background(),
		&output,
		bytes.NewReader(input),
		config,
		rand.New(rand.NewSource(1)),
		nil,
		1,
		"test",
		func(time.Duration) {},
	)
	if !errors.Is(err, errInjectedLoss) {
		t.Fatalf("proxyCopy error = %v, want injected loss", err)
	}
	if output.Len() != proxyBufferSize {
		t.Fatalf("forwarded bytes = %d, want %d before second chunk loss", output.Len(), proxyBufferSize)
	}
}
