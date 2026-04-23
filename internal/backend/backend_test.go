package backend

import (
	"context"
	"testing"
)

type fakeBackend struct {
	name   string
	binary string
}

func (f fakeBackend) Name() string                                                    { return f.name }
func (f fakeBackend) BinaryName() string                                              { return f.binary }
func (f fakeBackend) Capabilities() []Capability                                      { return nil }
func (f fakeBackend) Convert(_ context.Context, _, _ string, _ Options) error         { return nil }

type fakeAvailBackend struct {
	fakeBackend
	avail map[[2]string]bool
}

func (f fakeAvailBackend) IsAvailable(from, to string) bool {
	return f.avail[[2]string{from, to}]
}

func TestIsAvailable_FallsBackToLookPath(t *testing.T) {
	// "sh" is virtually always present on POSIX; "convertr-absolutely-not-a-real-binary-xyz" is not.
	if !IsAvailable(fakeBackend{name: "shell", binary: "sh"}, "a", "b") {
		t.Error("expected sh to be reported available")
	}
	if IsAvailable(fakeBackend{name: "nope", binary: "convertr-absolutely-not-a-real-binary-xyz"}, "a", "b") {
		t.Error("expected missing binary to be reported unavailable")
	}
}

func TestIsAvailable_UsesAvailabler(t *testing.T) {
	b := fakeAvailBackend{
		fakeBackend: fakeBackend{name: "x", binary: "sh"}, // sh exists, but Availabler must win
		avail: map[[2]string]bool{
			{"a", "b"}: true,
			{"c", "d"}: false,
		},
	}
	if !IsAvailable(b, "a", "b") {
		t.Error("Availabler returning true was not honored")
	}
	if IsAvailable(b, "c", "d") {
		t.Error("Availabler returning false was not honored (LookPath fallback leaked through)")
	}
	if IsAvailable(b, "unknown", "pair") {
		t.Error("Availabler returning false by default was not honored")
	}
}
