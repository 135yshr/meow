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

// A record whose fields cannot all be read back — an interface, a function —
// is still worth reading and still worth passing on whole. aws.Config is the
// case in point: a basket by its shape, a handle by its use.
type settings struct {
	Region string
	Log    interface{ Write([]byte) (int, error) }
}

func TestARecordGoesOnAsItselfAsWellAsBeingRead(t *testing.T) {
	made := meowrt.CallGo("load", func() settings {
		return settings{Region: "eu-west-1", Log: nil}
	})

	m, ok := made.(*meowrt.Map)
	if !ok {
		t.Fatalf("got %s (%s), want a basket", made.String(), made.Type())
	}
	// Its fields are readable, which is the point of reading it at all.
	if m.Items["region"].String() != "eu-west-1" {
		t.Errorf("region is %v, want it readable", m.Items["region"])
	}

	// And it goes on to the next call as the record itself, which building one
	// again out of the basket could not manage.
	got := meowrt.CallGo("use", func(s settings) string { return s.Region }, made)
	if got.String() != "eu-west-1" {
		t.Errorf("got %q, want the record to have arrived whole", got.String())
	}
}

// A pointer goes back as the very same pointer, not as a new one to a copy.
// Checking a field would not tell the two apart, and a call that changes
// something has to change the thing the program holds.
func TestAPointerToARecordGoesOnAsTheSamePointer(t *testing.T) {
	held := &settings{Region: "ap-northeast-1"}
	made := meowrt.CallGo("load", func() *settings { return held })

	same := meowrt.CallGo("same", func(s *settings) bool { return s == held }, made)
	if !same.IsTruthy() {
		t.Error("got a different pointer, want the one that was read")
	}

	meowrt.CallGo("move", func(s *settings) { s.Region = "eu-west-1" }, made)
	if held.Region != "eu-west-1" {
		t.Errorf("the original still says %q, want the call to have reached it", held.Region)
	}
}

// A record kept as a pointer, asked for by value, is copied — which is what
// passing by value means, so nothing is lost that was there.
func TestAPointerAskedForByValueIsCopied(t *testing.T) {
	held := &settings{Region: "ap-northeast-1"}
	made := meowrt.CallGo("load", func() *settings { return held })

	got := meowrt.CallGo("use", func(s settings) string { return s.Region }, made)
	if got.String() != "ap-northeast-1" {
		t.Errorf("got %q, want the fields to have arrived", got.String())
	}
}

// And a record kept by value, asked for as a pointer, gets one made for it:
// there was no pointer to keep, so a new one loses nothing.
func TestARecordAskedForAsAPointerGetsOne(t *testing.T) {
	made := meowrt.CallGo("load", func() settings {
		return settings{Region: "ap-northeast-1"}
	})

	got := meowrt.CallGo("use", func(s *settings) string { return s.Region }, made)
	if got.String() != "ap-northeast-1" {
		t.Errorf("got %q, want the fields to have arrived", got.String())
	}
}

// A basket a program wrote itself came from nothing, so it is built into the
// record as before.
func TestABasketWrittenInMeowIsStillBuilt(t *testing.T) {
	got := meowrt.CallGo("use", func(in probeInput) string { return in.Name },
		meowrt.NewMap(map[string]meowrt.Value{"name": meowrt.NewString("nyan")}))

	if got.String() != "nyan" {
		t.Errorf("got %q, want nyan", got.String())
	}
}

// What it came from is only used for the very type it came from.
func TestWhatItCameFromIsNotUsedForAnotherType(t *testing.T) {
	made := meowrt.CallGo("load", func() settings {
		return settings{Region: "eu-west-1"}
	})

	// probeInput has no "region", so building one is refused rather than the
	// settings being handed over in its place.
	got := meowrt.CallGo("use", func(in probeInput) string { return in.Name }, made)

	if _, failed := meowrt.AsFurball(got); !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
}

// A dial is a record and a thing with a method, which most useful Go types
// are. Reading it is what makes its field reachable; it should not be what
// makes its method unreachable.
type dial struct{ Region string }

func (d dial) Where() string { return "at " + d.Region }

// tags is a map with a method, which is the shape url.Values has.
type tags map[string]string

func (t tags) Only(k string) string { return t[k] }

func TestWhatWasReadOutOfGoIsStillCalledOn(t *testing.T) {
	tests := []struct {
		name   string
		fn     any
		reads  string
		member string
		args   []meowrt.Value
		want   string
	}{
		{"a record", func() dial { return dial{Region: "eu-west-1"} },
			"{region: eu-west-1}", "where", nil, "at eu-west-1"},
		{"a record behind a pointer", func() *dial { return &dial{Region: "ap-northeast-1"} },
			"{region: ap-northeast-1}", "where", nil, "at ap-northeast-1"},
		{"a time", func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
			"2026-01-02T03:04:05Z", "year", nil, "2026"},
		{"a number", func() time.Duration { return 90 * time.Minute },
			"5400000000000", "minutes", nil, "90"},
		{"a map", func() tags { return tags{"x": "1"} },
			"{x: 1}", "only", []meowrt.Value{meowrt.NewString("x")}, "1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			made := meowrt.CallGo("load", tt.fn)

			// What it reads as is what it always read as.
			if made.String() != tt.reads {
				t.Errorf("reads as %q, want %q", made.String(), tt.reads)
			}
			got := meowrt.CallMember(made, tt.member, tt.args...)
			if got.String() != tt.want {
				t.Errorf("%s gave %q, want %q", tt.member, got.String(), tt.want)
			}
		})
	}
}

// A time read as its text is still the time it was, not what the text could be
// read back as. The two differ by whatever the text does not say — here, half a
// second — and a call given the text back would be given the wrong moment
// without anything going wrong to show for it.
func TestATimeGoesOnAsTheTimeItWasRatherThanAsItsText(t *testing.T) {
	when := meowrt.CallGo("now", func() time.Time {
		return time.Date(2026, 1, 2, 3, 4, 5, 500_000_000, time.UTC)
	})

	if got := when.String(); got != "2026-01-02T03:04:05Z" {
		t.Fatalf("reads as %q, want the text to be unchanged", got)
	}

	got := meowrt.CallGo("nanos", func(at time.Time) int { return at.Nanosecond() }, when)
	if got.String() != "500000000" {
		t.Errorf("got %s nanoseconds, want the half second the text does not say", got.String())
	}
}

// What is all there remembers nothing. A plain string is the whole of what a
// call gave back, and saying it came from somewhere would say there is more to
// it than there is.
func TestWhatIsAllThereRemembersNothing(t *testing.T) {
	tests := []struct {
		name string
		of   meowrt.Value
	}{
		{"a plain string", meowrt.CallGo("name", func() string { return "nyan" })},
		{"a plain number", meowrt.CallGo("paws", func() int { return 4 })},
		{"a basket written in Meow", meowrt.NewMap(map[string]meowrt.Value{
			"host": meowrt.NewString("example.com"),
		})},
		{"a Meow value that made the round trip", meowrt.CallGo("same",
			func(v meowrt.Value) meowrt.Value { return v },
			meowrt.NewMap(map[string]meowrt.Value{"host": meowrt.NewString("example.com")}))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, ok := tt.of.(meowrt.Origin)
			if !ok {
				t.Fatalf("%s cannot remember at all", tt.of.Type())
			}
			if from := o.Origin(); from != nil {
				t.Errorf("remembers %#v, want nothing", from)
			}

			// And so there is nothing to call on it.
			got := meowrt.CallMember(tt.of, "string")
			f, failed := meowrt.AsFurball(got)
			if !failed {
				t.Fatalf("got %s, want a furball", got.String())
			}
			if !strings.Contains(f.Message, "Cannot call") {
				t.Errorf("says %q, want it to say there is nothing to call", f.Message)
			}
		})
	}
}

// What was read out of Go is named as Go names it when a member is asked for
// that it does not have, since that is the thing the member was looked for on.
func TestWhatWasReadSaysWhatItIsWhenThereIsNoSuchMember(t *testing.T) {
	made := meowrt.CallGo("load", func() dial { return dial{Region: "eu-west-1"} })

	got := meowrt.CallMember(made, "purr")

	f, failed := meowrt.AsFurball(got)
	if !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
	if !strings.Contains(f.Message, "dial") || !strings.Contains(f.Message, "Purr") {
		t.Errorf("says %q, want it to name the record and what was asked of it", f.Message)
	}
}

// nanos is a number that is also a fmt.Stringer, which a great many named Go
// types are.
type nanos int64

func (n nanos) String() string { return fmt.Sprintf("%dns", int64(n)) }

// Only something from Go can satisfy an interface with methods, and what was
// read out of Go is something from Go. Being read is not what should stop it
// answering for the interface it implements.
func TestWhatWasReadSatisfiesAnInterfaceItImplements(t *testing.T) {
	made := meowrt.CallGo("wait", func() nanos { return nanos(5) })

	if made.String() != "5" {
		t.Fatalf("reads as %q, want the number it is", made.String())
	}

	got := meowrt.CallGo("show", func(s fmt.Stringer) string { return s.String() }, made)
	if got.String() != "5ns" {
		t.Errorf("got %q, want the interface to have been satisfied", got.String())
	}
}

// An empty interface is a different question. It asks for the plain value
// behind the Meow one, which is what a call like json.Marshal is reading for —
// a record marshaled as a basket keeps the names a Meow program writes, rather
// than turning back into the Go ones on the way past.
func TestAnEmptyInterfaceStillTakesThePlainValue(t *testing.T) {
	made := meowrt.CallGo("load", func() dial { return dial{Region: "eu-west-1"} })

	got := meowrt.CallGo("marshal", func(v any) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	}, made)

	if got.String() != `{"region":"eu-west-1"}` {
		t.Errorf("got %s, want the basket's own names", got.String())
	}
}

// ledger is something a call is made on to do rather than to say. Go writes a
// great many of these — Set, Add, Reset, Close.
type ledger struct{ N int }

func (l *ledger) Inc()          { l.N++ }
func (l *ledger) Note() error   { l.N++; return nil }
func (l *ledger) Fail() error   { return errors.New("no") }
func (l *ledger) Find() *ledger { return nil }

// A method with nothing of its own to say gives back what it was called on,
// read afresh. Handing back catnap would say nothing at all and leave the doing
// with nowhere to show; a new value next to the old one is how a language whose
// values do not change says that something happened.
func TestAMethodWithNothingToSayGivesBackWhatItWasCalledOn(t *testing.T) {
	made := meowrt.CallGo("open", func() *ledger { return &ledger{N: 1} })

	t.Run("a method that returns nothing", func(t *testing.T) {
		got := meowrt.CallMember(made, "inc")

		if got.String() != "{n: 2}" {
			t.Errorf("got %s, want the reading taken after the call", got.String())
		}
		// And the reading taken before it says what it always said.
		if made.String() != "{n: 1}" {
			t.Errorf("the older reading now says %s, want it unchanged", made.String())
		}
	})

	t.Run("a method that returns only a failure that did not happen", func(t *testing.T) {
		got := meowrt.CallMember(made, "note")

		if got.String() != "{n: 3}" {
			t.Errorf("got %s, want the reading taken after the call", got.String())
		}
	})

	t.Run("a method that failed", func(t *testing.T) {
		got := meowrt.CallMember(made, "fail")

		if _, failed := meowrt.AsFurball(got); !failed {
			t.Errorf("got %s, want the failure rather than the receiver", got.String())
		}
	})

	// The rule is about a method with no answer, not about an answer that is
	// nothing. A search that found nothing found nothing.
	t.Run("a method that answered with nothing", func(t *testing.T) {
		got := meowrt.CallMember(made, "find")

		if got.Type() != "Nil" {
			t.Errorf("got %s (%s), want catnap", got.String(), got.Type())
		}
	})
}

// A handle is not read, so there is nothing to read again. Giving back the same
// handle is what lets one call follow another.
func TestAMethodWithNothingToSayGivesBackTheHandleItself(t *testing.T) {
	held := meowrt.NewOpaque("probe.Ledger", &ledger{})

	got := meowrt.CallMember(held, "inc")

	if got != meowrt.Value(held) {
		t.Errorf("got %s, want the handle it was called on", got.String())
	}
}
