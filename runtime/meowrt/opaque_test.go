package meowrt_test

import (
	"testing"

	"github.com/135yshr/meow/runtime/meowrt"
)

// A Go library hands back things Meow has no shape for — a client, a
// connection, a handle. With nowhere to put one, every library had to be
// wrapped by hand before Meow could reach it.
func TestOpaqueCarriesWhatMeowHasNoShapeFor(t *testing.T) {
	type client struct{ region string }
	c := &client{region: "ap-northeast-1"}

	o := meowrt.NewOpaque("sts.Client", c)

	if o.Type() != "Opaque" {
		t.Errorf("type is %q, want Opaque", o.Type())
	}
	if o.String() != "<sts.Client>" {
		t.Errorf("reads as %q, want <sts.Client>", o.String())
	}
	if !o.IsTruthy() {
		t.Error("a held thing should read as true")
	}
	got, ok := meowrt.AsOpaque(o)
	if !ok || got.V != any(c) {
		t.Errorf("got %v, want the very thing that went in", got.V)
	}
}

// A handle that was never made is nothing, and reads as false.
func TestOpaqueHoldingNothingIsFalse(t *testing.T) {
	o := meowrt.NewOpaque("sts.Client", nil)
	if o.IsTruthy() {
		t.Error("holding nothing should read as false")
	}
}

// Named by nothing, it says what it is holding rather than staying silent.
func TestOpaqueNamedByNothingFallsBackToItsGoType(t *testing.T) {
	if got := meowrt.NewOpaque("", 42).String(); got != "<int>" {
		t.Errorf("reads as %q, want <int>", got)
	}
}
