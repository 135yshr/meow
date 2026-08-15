package aws

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/135yshr/meow/runtime/meowrt"
)

// fakeSTS and fakeLogs stand in for the SDK clients, so these tests never reach
// the network or need credentials.
type fakeSTS struct {
	out *sts.GetCallerIdentityOutput
	err error
}

func (f fakeSTS) GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return f.out, f.err
}

type fakeLogs struct {
	out *cloudwatchlogs.FilterLogEventsOutput
	err error
	got *cloudwatchlogs.FilterLogEventsInput
}

func (f *fakeLogs) FilterLogEvents(_ context.Context, in *cloudwatchlogs.FilterLogEventsInput, _ ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
	f.got = in
	return f.out, f.err
}

// useSTS installs a fake STS client for the duration of a test.
func useSTS(t *testing.T, client stsAPI, err error) {
	t.Helper()
	original := newSTS
	t.Cleanup(func() { newSTS = original })
	newSTS = func(context.Context) (stsAPI, error) { return client, err }
}

// useLogs installs a fake CloudWatch Logs client for the duration of a test.
func useLogs(t *testing.T, client logsAPI, err error) {
	t.Helper()
	original := newLogs
	t.Cleanup(func() { newLogs = original })
	newLogs = func(context.Context) (logsAPI, error) { return client, err }
}

func TestWhoami(t *testing.T) {
	useSTS(t, fakeSTS{out: &sts.GetCallerIdentityOutput{
		Account: aws.String("123456789012"),
		Arn:     aws.String("arn:aws:iam::123456789012:user/nyantyu"),
		UserId:  aws.String("AIDANYANTYU"),
	}}, nil)

	got := Whoami()
	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("expected a Map, got %T (%s)", got, got.String())
	}
	for key, want := range map[string]string{
		"account": "123456789012",
		"arn":     "arn:aws:iam::123456789012:user/nyantyu",
		"user_id": "AIDANYANTYU",
	} {
		v, found := m.Get(key)
		if !found {
			t.Errorf("expected the %s key", key)
			continue
		}
		if v.String() != want {
			t.Errorf("%s: got %q, want %q", key, v.String(), want)
		}
	}
}

// A field the API omits must read as an empty string rather than crashing on a
// nil pointer.
func TestWhoamiHandlesMissingFields(t *testing.T) {
	useSTS(t, fakeSTS{out: &sts.GetCallerIdentityOutput{}}, nil)

	m, ok := Whoami().(*meowrt.Map)
	if !ok {
		t.Fatal("expected a Map")
	}
	if v, _ := m.Get("account"); v.String() != "" {
		t.Errorf("got %q, want an empty string", v.String())
	}
}

func TestWhoamiReportsErrors(t *testing.T) {
	t.Run("call fails", func(t *testing.T) {
		useSTS(t, fakeSTS{err: errors.New("expired token")}, nil)
		got := Whoami()
		f, ok := got.(*meowrt.Furball)
		if !ok {
			t.Fatalf("expected a Furball, got %T", got)
		}
		if !strings.Contains(f.Message, "expired token") {
			t.Errorf("expected the cause in %q", f.Message)
		}
	})

	t.Run("client cannot be built", func(t *testing.T) {
		useSTS(t, nil, errors.New("no region configured"))
		if _, ok := Whoami().(*meowrt.Furball); !ok {
			t.Error("expected a Furball")
		}
	})

	t.Run("unexpected arguments", func(t *testing.T) {
		if _, ok := Whoami(meowrt.NewString("x")).(*meowrt.Furball); !ok {
			t.Error("expected a Furball")
		}
	})
}

func TestDig(t *testing.T) {
	logs := &fakeLogs{out: &cloudwatchlogs.FilterLogEventsOutput{
		Events: []cwltypes.FilteredLogEvent{
			{Timestamp: aws.Int64(1755266400000), Message: aws.String("nyan-marker-001 stored"), LogStreamName: aws.String("stream-a")},
			{Timestamp: aws.Int64(1755266401000), Message: aws.String("other line"), LogStreamName: aws.String("stream-b")},
		},
	}}
	useLogs(t, logs, nil)

	got := Dig(meowrt.NewString("/aws/lambda/canary"))
	l, ok := got.(*meowrt.List)
	if !ok {
		t.Fatalf("expected a List, got %T (%s)", got, got.String())
	}
	if l.Len() != 2 {
		t.Fatalf("got %d events, want 2", l.Len())
	}
	first, ok := l.Get(0).(*meowrt.Map)
	if !ok {
		t.Fatal("expected each event to be a Map")
	}
	if v, _ := first.Get("message"); v.String() != "nyan-marker-001 stored" {
		t.Errorf("message: got %q", v.String())
	}
	if v, _ := first.Get("stream"); v.String() != "stream-a" {
		t.Errorf("stream: got %q", v.String())
	}
	if v, _ := first.Get("timestamp"); v.String() != "1755266400000" {
		t.Errorf("timestamp: got %q", v.String())
	}
	if aws.ToString(logs.got.LogGroupName) != "/aws/lambda/canary" {
		t.Errorf("log group: got %q", aws.ToString(logs.got.LogGroupName))
	}
}

func TestDigPassesOptions(t *testing.T) {
	logs := &fakeLogs{out: &cloudwatchlogs.FilterLogEventsOutput{}}
	useLogs(t, logs, nil)

	got := Dig(meowrt.NewString("group"), meowrt.NewMap(map[string]meowrt.Value{
		"pattern": meowrt.NewString("nyan-marker-001"),
		"start":   meowrt.NewInt(1755266400000),
		"end":     meowrt.NewInt(1755266500000),
		"limit":   meowrt.NewInt(50),
	}))
	if _, ok := got.(*meowrt.Furball); ok {
		t.Fatalf("unexpected Furball: %s", got.String())
	}
	if aws.ToString(logs.got.FilterPattern) != "nyan-marker-001" {
		t.Errorf("pattern: got %q", aws.ToString(logs.got.FilterPattern))
	}
	if aws.ToInt64(logs.got.StartTime) != 1755266400000 {
		t.Errorf("start: got %d", aws.ToInt64(logs.got.StartTime))
	}
	if aws.ToInt64(logs.got.EndTime) != 1755266500000 {
		t.Errorf("end: got %d", aws.ToInt64(logs.got.EndTime))
	}
	if aws.ToInt32(logs.got.Limit) != 50 {
		t.Errorf("limit: got %d", aws.ToInt32(logs.got.Limit))
	}
}

// An omitted pattern must not be sent as an empty one, which the API would
// treat as a filter rather than as "everything".
func TestDigOmitsEmptyPattern(t *testing.T) {
	logs := &fakeLogs{out: &cloudwatchlogs.FilterLogEventsOutput{}}
	useLogs(t, logs, nil)

	Dig(meowrt.NewString("group"))
	if logs.got.FilterPattern != nil {
		t.Errorf("expected no filter pattern, got %q", aws.ToString(logs.got.FilterPattern))
	}
}

func TestDigRejectsBadArguments(t *testing.T) {
	logs := &fakeLogs{out: &cloudwatchlogs.FilterLogEventsOutput{}}
	useLogs(t, logs, nil)

	tests := []struct {
		name string
		args []meowrt.Value
	}{
		{"no arguments", nil},
		{"too many arguments", []meowrt.Value{
			meowrt.NewString("g"), meowrt.NewMap(nil), meowrt.NewString("extra"),
		}},
		{"non-string group", []meowrt.Value{meowrt.NewInt(1)}},
		{"non-map options", []meowrt.Value{meowrt.NewString("g"), meowrt.NewString("nope")}},
		{"non-string pattern", []meowrt.Value{meowrt.NewString("g"), meowrt.NewMap(map[string]meowrt.Value{
			"pattern": meowrt.NewInt(1),
		})}},
		{"non-int start", []meowrt.Value{meowrt.NewString("g"), meowrt.NewMap(map[string]meowrt.Value{
			"start": meowrt.NewString("soon"),
		})}},
		{"limit below range", []meowrt.Value{meowrt.NewString("g"), meowrt.NewMap(map[string]meowrt.Value{
			"limit": meowrt.NewInt(0),
		})}},
		{"limit above range", []meowrt.Value{meowrt.NewString("g"), meowrt.NewMap(map[string]meowrt.Value{
			"limit": meowrt.NewInt(maxEvents + 1),
		})}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := Dig(tt.args...).(*meowrt.Furball); !ok {
				t.Error("expected a Furball")
			}
		})
	}
}

func TestDigReportsCallErrors(t *testing.T) {
	useLogs(t, &fakeLogs{err: errors.New("group not found")}, nil)

	got := Dig(meowrt.NewString("missing"))
	f, ok := got.(*meowrt.Furball)
	if !ok {
		t.Fatalf("expected a Furball, got %T", got)
	}
	if !strings.Contains(f.Message, "group not found") {
		t.Errorf("expected the cause in %q", f.Message)
	}
}

func TestRegion(t *testing.T) {
	original := resolveRegion
	t.Cleanup(func() { resolveRegion = original })

	resolveRegion = func(context.Context) (string, error) { return "ap-northeast-1", nil }
	if got := Region(); got.String() != "ap-northeast-1" {
		t.Errorf("got %q, want %q", got.String(), "ap-northeast-1")
	}

	resolveRegion = func(context.Context) (string, error) { return "", errors.New("no region") }
	if _, ok := Region().(*meowrt.Furball); !ok {
		t.Error("expected a Furball when the region cannot be resolved")
	}
	if _, ok := Region(meowrt.NewString("x")).(*meowrt.Furball); !ok {
		t.Error("expected a Furball for unexpected arguments")
	}
}
