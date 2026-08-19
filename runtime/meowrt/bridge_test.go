package meowrt_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/135yshr/meow/runtime/meowrt"
)

func TestCallGoReadsWhatComesBack(t *testing.T) {
	type point struct {
		X, Y     int
		Label    string
		UserId   string
		Hidden   string
		internal int
	}
	tests := []struct {
		name string
		fn   any
		args []meowrt.Value
		want string
	}{
		{"a number", func() int { return 7 }, nil, "7"},
		{"text", func() string { return "nyan" }, nil, "nyan"},
		{"nothing at all", func() {}, nil, "catnap"},
		{"a list", func() []int { return []int{1, 2} }, nil, "[1, 2]"},
		{
			"an argument on the way in",
			func(n int) int { return n * 2 },
			[]meowrt.Value{meowrt.NewInt(21)},
			"42",
		},
		{
			"text and a number together",
			strings.Repeat,
			[]meowrt.Value{meowrt.NewString("nya"), meowrt.NewInt(2)},
			"nyanya",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := meowrt.CallGo(tt.name, tt.fn, tt.args...); got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}

	// A record comes back as a basket, under the names a Meow program writes.
	t.Run("a record", func(t *testing.T) {
		got := meowrt.CallGo("point", func() point {
			return point{X: 1, Y: 2, Label: "a", UserId: "u", Hidden: "h", internal: 9}
		})
		m, ok := got.(*meowrt.Map)
		if !ok {
			t.Fatalf("got %s, want a basket", got.Type())
		}
		if m.Items["user_id"].String() != "u" {
			t.Errorf("UserId came through as %v, want it under user_id", m.Items)
		}
		if _, leaked := m.Items["internal"]; leaked {
			t.Error("an unexported field should not come out")
		}
	})
}

// A trailing error is the failure itself, not part of the answer.
func TestCallGoTurnsAnErrorIntoAFurball(t *testing.T) {
	got := meowrt.CallGo("read", func() (string, error) {
		return "", errors.New("no such file")
	})

	f, failed := meowrt.AsFurball(got)
	if !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
	if !strings.Contains(f.Message, "no such file") {
		t.Errorf("says %q, want it to carry the reason", f.Message)
	}
	if !strings.Contains(f.Message, "read") {
		t.Errorf("says %q, want it to name the call", f.Message)
	}
}

func TestCallGoPassesOnWhatSucceeded(t *testing.T) {
	got := meowrt.CallGo("read", func() (string, error) { return "hi", nil })
	if got.String() != "hi" {
		t.Errorf("got %q, want hi", got.String())
	}
	if _, failed := meowrt.AsFurball(got); failed {
		t.Error("a call that worked should not be a furball")
	}
}

// A function asking for a context is given one rather than being asked for it
// in Meow, where there is nothing to pass.
func TestCallGoSuppliesAContext(t *testing.T) {
	got := meowrt.CallGo("wait", func(ctx context.Context, n int) (int, error) {
		if ctx == nil {
			return 0, errors.New("no context")
		}
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			return 0, errors.New("no deadline")
		}
		return n, nil
	}, meowrt.NewInt(5))

	if got.String() != "5" {
		t.Errorf("got %q, want 5", got.String())
	}
}

func TestCallGoCountsTheArguments(t *testing.T) {
	got := meowrt.CallGo("twice", func(n int) int { return n }, meowrt.NewInt(1), meowrt.NewInt(2))

	f, failed := meowrt.AsFurball(got)
	if !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
	if !strings.Contains(f.Message, "wants 1 arguments, got 2") {
		t.Errorf("says %q, want it to say how many were wanted", f.Message)
	}
}

func TestCallGoSaysWhenAnArgumentIsTheWrongShape(t *testing.T) {
	got := meowrt.CallGo("twice", func(n int) int { return n }, meowrt.NewString("nyan"))

	f, failed := meowrt.AsFurball(got)
	if !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
	if !strings.Contains(f.Message, "argument 1") {
		t.Errorf("says %q, want it to name which argument", f.Message)
	}
}

// The shape a real library has: a client made once, then methods on it.
type probeClient struct{ region string }

type probeInput struct {
	Name   string
	Detail bool
}

type probeOutput struct {
	Arn    string
	UserId string
	When   time.Time
}

func (c *probeClient) Describe(_ context.Context, in *probeInput) (*probeOutput, error) {
	if in.Name == "" {
		return nil, errors.New("no name given")
	}
	arn := "arn:" + in.Name
	if in.Detail {
		arn += "@" + c.region
	}
	return &probeOutput{Arn: arn, UserId: "AIDA", When: time.Unix(0, 0).UTC()}, nil
}

func TestCallGoMethodOnAHeldClient(t *testing.T) {
	client := meowrt.NewOpaque("probe.Client", &probeClient{region: "ap-northeast-1"})

	got := meowrt.CallGoMethod(context.Background(), client, "Describe",
		meowrt.NewMap(map[string]meowrt.Value{
			"name":   meowrt.NewString("nyan"),
			"detail": meowrt.NewBool(true),
		}))

	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("got %s (%s), want a basket", got.String(), got.Type())
	}
	if s := m.Items["arn"].String(); s != "arn:nyan@ap-northeast-1" {
		t.Errorf("arn is %q, want it built from both the argument and the client", s)
	}
	if s := m.Items["user_id"].String(); s != "AIDA" {
		t.Errorf("user_id is %q, want AIDA", s)
	}
	if s := m.Items["when"].String(); s != "1970-01-01T00:00:00Z" {
		t.Errorf("when is %q, want a readable time", s)
	}
}

func TestCallGoMethodPassesOnAFailure(t *testing.T) {
	client := meowrt.NewOpaque("probe.Client", &probeClient{})

	got := meowrt.CallGoMethod(context.Background(), client, "Describe",
		meowrt.NewMap(map[string]meowrt.Value{"name": meowrt.NewString("")}))

	if _, failed := meowrt.AsFurball(got); !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
}

func TestCallGoMethodSaysWhenThereIsNoSuchMethod(t *testing.T) {
	client := meowrt.NewOpaque("probe.Client", &probeClient{})

	got := meowrt.CallGoMethod(context.Background(), client, "Nope")

	f, failed := meowrt.AsFurball(got)
	if !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
	if !strings.Contains(f.Message, "Nope") {
		t.Errorf("says %q, want it to name what was asked for", f.Message)
	}
}

func TestCallGoMethodNeedsSomethingToCallOn(t *testing.T) {
	got := meowrt.CallGoMethod(context.Background(), meowrt.NewString("nyan"), "Describe")

	if _, failed := meowrt.AsFurball(got); !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
}

// A basket naming something the record has not is a mistake worth saying,
// rather than a value quietly dropped.
func TestCallGoRefusesAnUnknownField(t *testing.T) {
	client := meowrt.NewOpaque("probe.Client", &probeClient{})

	got := meowrt.CallGoMethod(context.Background(), client, "Describe",
		meowrt.NewMap(map[string]meowrt.Value{"nme": meowrt.NewString("nyan")}))

	f, failed := meowrt.AsFurball(got)
	if !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
	if !strings.Contains(f.Message, "nme") {
		t.Errorf("says %q, want it to name the field it did not know", f.Message)
	}
}

// Something being carried goes back into the next call exactly as it came.
func TestAHeldThingGoesBackInUntouched(t *testing.T) {
	c := &probeClient{region: "eu-west-1"}
	held := meowrt.NewOpaque("probe.Client", c)

	got := meowrt.CallGo("region", func(c *probeClient) string { return c.region }, held)

	if got.String() != "eu-west-1" {
		t.Errorf("got %q, want the client to arrive intact", got.String())
	}
}

func TestCallGoNeedsAFunction(t *testing.T) {
	if _, failed := meowrt.AsFurball(meowrt.CallGo("nope", 42)); !failed {
		t.Error("calling a number should be a furball")
	}
}

// The flow the bridge exists for: make the client, then call on what came
// back. A client has its state to itself, so reading it as a record would give
// an empty basket and nothing left to call.
func TestAClientComesBackAsSomethingToCallOn(t *testing.T) {
	made := meowrt.CallGo("new_client", func() *probeClient {
		return &probeClient{region: "eu-west-1"}
	})

	if _, held := meowrt.AsOpaque(made); !held {
		t.Fatalf("a client came back as %s (%s), want it held whole", made.String(), made.Type())
	}

	got := meowrt.CallGoMethod(context.Background(), made, "Describe",
		meowrt.NewMap(map[string]meowrt.Value{
			"name":   meowrt.NewString("nyan"),
			"detail": meowrt.NewBool(true),
		}))

	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("got %s (%s), want a basket", got.String(), got.Type())
	}
	if s := m.Items["arn"].String(); s != "arn:nyan@eu-west-1" {
		t.Errorf("arn is %q, want the client the constructor made to have been used", s)
	}
}

// A record still comes back as a record: it has fields to read, so reading
// them is the useful thing.
func TestARecordStillComesBackAsABasket(t *testing.T) {
	got := meowrt.CallGo("describe", func() *probeOutput {
		return &probeOutput{Arn: "arn:nyan", UserId: "AIDA"}
	})

	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("got %s (%s), want a basket", got.String(), got.Type())
	}
	if m.Items["arn"].String() != "arn:nyan" {
		t.Errorf("arn is %v, want it readable", m.Items["arn"])
	}
}

// Bytes are the shape a Meow program has for a run of bytes, in both
// directions, so a byte-taking Go function can be called with to_bytes and its
// answer read with to_string.
func TestBytesCrossBothWays(t *testing.T) {
	t.Run("going in", func(t *testing.T) {
		got := meowrt.CallGo("count", func(p []byte) int { return len(p) },
			meowrt.ToBytes(meowrt.NewString("nyan")))
		if got.String() != "4" {
			t.Errorf("got %q, want 4", got.String())
		}
	})

	t.Run("coming out", func(t *testing.T) {
		got := meowrt.CallGo("shout", func() []byte { return []byte("nyan") })
		if s := meowrt.ToString(got); s.String() != "nyan" {
			t.Errorf("to_string read it as %q, want nyan", s.String())
		}
	})
}

// A bool parameter takes a bool. Reading the truthiness of anything else would
// send the string "hairball" in as true.
func TestABoolParameterTakesOnlyABool(t *testing.T) {
	got := meowrt.CallGo("flag", func(b bool) bool { return b }, meowrt.NewString("hairball"))

	if _, failed := meowrt.AsFurball(got); !failed {
		t.Fatalf("got %s, want a furball rather than a flag turned the wrong way", got.String())
	}
}

// A number too big for the Go type it is going into is a mistake worth saying,
// rather than a different number arriving.
func TestANumberThatWouldNotFitIsRefused(t *testing.T) {
	got := meowrt.CallGo("byte", func(b uint8) uint8 { return b }, meowrt.NewInt(300))

	f, failed := meowrt.AsFurball(got)
	if !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
	if !strings.Contains(f.Message, "300") {
		t.Errorf("says %q, want it to name the number that would not fit", f.Message)
	}
}

func TestAVariadicFunctionTakesWhatItIsGiven(t *testing.T) {
	total := func(label string, n ...int) string {
		sum := 0
		for _, each := range n {
			sum += each
		}
		return fmt.Sprintf("%s%d", label, sum)
	}

	t.Run("several", func(t *testing.T) {
		got := meowrt.CallGo("total", total,
			meowrt.NewString("n="), meowrt.NewInt(1), meowrt.NewInt(2), meowrt.NewInt(3))
		if got.String() != "n=6" {
			t.Errorf("got %q, want n=6", got.String())
		}
	})

	t.Run("none", func(t *testing.T) {
		if got := meowrt.CallGo("total", total, meowrt.NewString("n=")); got.String() != "n=0" {
			t.Errorf("got %q, want n=0", got.String())
		}
	})

	t.Run("not even the fixed one", func(t *testing.T) {
		got := meowrt.CallGo("total", total)
		f, failed := meowrt.AsFurball(got)
		if !failed {
			t.Fatalf("got %s, want a furball", got.String())
		}
		if !strings.Contains(f.Message, "at least 1") {
			t.Errorf("says %q, want it to say how few that is", f.Message)
		}
	})
}

// The library being called is someone else's code. One that comes apart is a
// failure like any other, catchable with gag, rather than the end of the
// program.
func TestAPanicIsAFailureRatherThanTheEnd(t *testing.T) {
	got := meowrt.CallGo("boom", func() int { panic("nyaaa") })

	f, failed := meowrt.AsFurball(got)
	if !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
	if !strings.Contains(f.Message, "nyaaa") {
		t.Errorf("says %q, want it to carry what went wrong", f.Message)
	}
	if !strings.Contains(f.Message, "boom") {
		t.Errorf("says %q, want it to name the call", f.Message)
	}
}

type ownError struct{ why string }

func (e *ownError) Error() string { return e.why }

// A library that returns its own error type rather than the error interface
// still fails the call, and a nil one of those is still no failure.
func TestAnErrorOfItsOwnTypeIsStillTheFailure(t *testing.T) {
	t.Run("something went wrong", func(t *testing.T) {
		got := meowrt.CallGo("read", func() (string, *ownError) {
			return "", &ownError{why: "no such file"}
		})
		f, failed := meowrt.AsFurball(got)
		if !failed {
			t.Fatalf("got %s, want a furball", got.String())
		}
		if !strings.Contains(f.Message, "no such file") {
			t.Errorf("says %q, want it to carry the reason", f.Message)
		}
	})

	t.Run("nothing did", func(t *testing.T) {
		got := meowrt.CallGo("read", func() (string, *ownError) { return "hi", nil })
		if got.String() != "hi" {
			t.Errorf("got %q, want hi", got.String())
		}
	})
}

// A time goes out as text, so that text is what it goes back in as.
func TestATimeMakesTheRoundTrip(t *testing.T) {
	when := meowrt.CallGo("now", func() time.Time { return time.Unix(0, 0).UTC() })

	got := meowrt.CallGo("year", func(at time.Time) int { return at.Year() }, when)

	if got.String() != "1970" {
		t.Errorf("got %q, want 1970", got.String())
	}
}

// A member call is written the same way whatever it lands on, so which it is
// gets told apart here rather than where it is written.
func TestCallMemberTellsApartWhatItLandsOn(t *testing.T) {
	t.Run("a method on something held", func(t *testing.T) {
		held := meowrt.NewOpaque("probe.Client", &probeClient{region: "ap-northeast-1"})

		got := meowrt.CallMember(held, "describe",
			meowrt.NewMap(map[string]meowrt.Value{"name": meowrt.NewString("nyan")}))

		m, ok := got.(*meowrt.Map)
		if !ok {
			t.Fatalf("got %s (%s), want a basket", got.String(), got.Type())
		}
		if m.Items["arn"].String() != "arn:nyan" {
			t.Errorf("arn is %v, want it built by the method", m.Items["arn"])
		}
	})

	t.Run("a failure carries on being one", func(t *testing.T) {
		furball := meowrt.NewFurball("Hiss! earlier, nya~")

		if got := meowrt.CallMember(furball, "describe"); got != meowrt.Value(furball) {
			t.Errorf("got %s, want the failure itself", got.String())
		}
	})

	t.Run("something with no members at all", func(t *testing.T) {
		got := meowrt.CallMember(meowrt.NewInt(1), "describe")

		f, failed := meowrt.AsFurball(got)
		if !failed {
			t.Fatalf("got %s, want a furball", got.String())
		}
		if !strings.Contains(f.Message, "describe") {
			t.Errorf("says %q, want it to name what was asked for", f.Message)
		}
	})
}

// FromGo is how a Go package's own values — its constants, its variables —
// come across.
func TestFromGoReadsAPlainValue(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{"a number", 42, "42"},
		{"text", "nyan", "nyan"},
		{"nothing", nil, "catnap"},
		{"a list", []string{"a", "b"}, "[a, b]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := meowrt.FromGo(tt.v); got.String() != tt.want {
				t.Errorf("got %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// A great many Go functions take an empty interface — Sprintf, Marshal — and
// what goes in is the plain Go value behind the Meow one.
func TestAnEmptyInterfaceTakesThePlainValue(t *testing.T) {
	t.Run("several of them", func(t *testing.T) {
		got := meowrt.CallGo("sprintf", fmt.Sprintf,
			meowrt.NewString("%s has %d"), meowrt.NewString("nyan"), meowrt.NewInt(4))
		if got.String() != "nyan has 4" {
			t.Errorf("got %q, want \"nyan has 4\"", got.String())
		}
	})

	t.Run("a basket becomes a map", func(t *testing.T) {
		got := meowrt.CallGo("marshal", json.Marshal,
			meowrt.NewMap(map[string]meowrt.Value{"name": meowrt.NewString("nyan")}))
		if s := meowrt.ToString(got); s.String() != `{"name":"nyan"}` {
			t.Errorf("got %q, want the basket written out", s.String())
		}
	})

	t.Run("a litter becomes a slice", func(t *testing.T) {
		got := meowrt.CallGo("marshal", json.Marshal,
			meowrt.NewList(meowrt.NewInt(1), meowrt.NewString("a")))
		if s := meowrt.ToString(got); s.String() != `[1,"a"]` {
			t.Errorf("got %q, want the litter written out", s.String())
		}
	})

	t.Run("something held goes in as what it holds", func(t *testing.T) {
		held := meowrt.NewOpaque("probe.Client", &probeClient{region: "eu-west-1"})
		got := meowrt.CallGo("type", func(v any) string {
			c, ok := v.(*probeClient)
			if !ok {
				return "not a client"
			}
			return c.region
		}, held)
		if got.String() != "eu-west-1" {
			t.Errorf("got %q, want the client itself to have arrived", got.String())
		}
	})
}

// An interface with methods is another matter: only something held from Go can
// satisfy one, so a Meow value is refused rather than guessed at.
func TestAnInterfaceWithMethodsNeedsSomethingThatHasThem(t *testing.T) {
	got := meowrt.CallGo("write", func(w io.Writer) int { return 0 }, meowrt.NewString("nyan"))

	if _, failed := meowrt.AsFurball(got); !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
}
