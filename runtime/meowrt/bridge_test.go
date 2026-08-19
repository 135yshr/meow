package meowrt_test

import (
	"context"
	"errors"
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
