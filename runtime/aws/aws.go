// Package aws exposes Amazon Web Services to Meow programs, through the
// official aws-sdk-go-v2.
//
// Credentials and region are resolved by the SDK's own default chain —
// environment variables, shared config and credentials files, SSO, container
// and instance metadata — so a Meow program authenticates exactly the way the
// AWS CLI does on the same machine, and never has to be handed a secret
// directly.
//
// Results are returned as ordinary Meow Maps and Lists rather than as opaque
// handles, so a program can index, print and compare them with the operators it
// already has.
package aws

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/135yshr/meow/runtime/meowrt"
)

// callTimeout bounds every AWS call, so a program cannot hang indefinitely on
// an unreachable endpoint.
const callTimeout = 30 * time.Second

// maxEvents bounds how many log events a single filter call returns, matching
// the ceiling the CloudWatch Logs API itself applies per page.
const maxEvents = 10_000

// furball wraps an error as a Meow Furball value with the "Hiss! ... nya~" form.
func furball(format string, args ...any) meowrt.Value {
	return &meowrt.Furball{Message: fmt.Sprintf("Hiss! "+format+", nya~", args...)}
}

// stsAPI and logsAPI narrow the SDK clients to the calls this package makes, so
// tests can substitute them without reaching the network.
type stsAPI interface {
	GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type logsAPI interface {
	FilterLogEvents(context.Context, *cloudwatchlogs.FilterLogEventsInput, ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
}

// newSTS and newLogs are swapped out in tests.
var (
	newSTS = func(ctx context.Context) (stsAPI, error) {
		cfg, err := loadConfig(ctx)
		if err != nil {
			return nil, err
		}
		return sts.NewFromConfig(cfg), nil
	}
	newLogs = func(ctx context.Context) (logsAPI, error) {
		cfg, err := loadConfig(ctx)
		if err != nil {
			return nil, err
		}
		return cloudwatchlogs.NewFromConfig(cfg), nil
	}
	resolveRegion = func(ctx context.Context) (string, error) {
		cfg, err := loadConfig(ctx)
		if err != nil {
			return "", err
		}
		return cfg.Region, nil
	}
)

// loadConfig resolves credentials and region through the SDK's default chain.
func loadConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, err
	}
	if cfg.Region == "" {
		return aws.Config{}, errors.New("no AWS region configured; set AWS_REGION or a profile region")
	}
	return cfg, nil
}

// deref reads through a *string, treating nil as empty.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// expectString extracts a Go string from a Meow value.
func expectString(fn, what string, v meowrt.Value) (string, meowrt.Value) {
	if f, ok := v.(*meowrt.Furball); ok {
		return "", f
	}
	s, ok := v.(*meowrt.String)
	if !ok {
		typeName := "catnap"
		if v != nil {
			typeName = v.Type()
		}
		return "", furball("%s expects a String %s, got %s", fn, what, typeName)
	}
	return s.Val, nil
}

// Whoami reports the identity the program is authenticated as, as a Map of
// account, arn and user_id.
//
// It is the cheapest way to answer "are my credentials working, and am I in the
// account I think I am" before doing anything that matters.
func Whoami(args ...meowrt.Value) meowrt.Value {
	if len(args) != 0 {
		return furball("whoami expects no arguments, got %d", len(args))
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	client, err := newSTS(ctx)
	if err != nil {
		return furball("%s", err)
	}
	out, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return furball("%s", err)
	}
	return meowrt.NewMap(map[string]meowrt.Value{
		"account": meowrt.NewString(deref(out.Account)),
		"arn":     meowrt.NewString(deref(out.Arn)),
		"user_id": meowrt.NewString(deref(out.UserId)),
	})
}

// filterOptions carries the optional arguments of Dig.
type filterOptions struct {
	pattern   string
	startTime *int64
	endTime   *int64
	limit     int32
}

// extractFilterOptions reads the options Map accepted by Dig.
func extractFilterOptions(m *meowrt.Map) (filterOptions, meowrt.Value) {
	opts := filterOptions{limit: maxEvents}
	if v, found := m.Get("pattern"); found {
		s, fb := expectString("dig", "pattern", v)
		if fb != nil {
			return opts, fb
		}
		opts.pattern = s
	}
	// Times are milliseconds since the epoch, which is what the API uses;
	// clock.now() * 1000 lines up with it.
	for _, field := range []struct {
		key  string
		dest **int64
	}{
		{"start", &opts.startTime},
		{"end", &opts.endTime},
	} {
		v, found := m.Get(field.key)
		if !found {
			continue
		}
		n, fb := meowrt.TryAsInt(v)
		if fb != nil {
			return opts, fb
		}
		ms := n
		*field.dest = &ms
	}
	if v, found := m.Get("limit"); found {
		n, fb := meowrt.TryAsInt(v)
		if fb != nil {
			return opts, fb
		}
		if n <= 0 || n > maxEvents {
			return opts, furball("dig expects a limit between 1 and %d, got %d", maxEvents, n)
		}
		opts.limit = int32(n)
	}
	return opts, nil
}

// Dig searches a CloudWatch Logs group and returns the matching events as a
// litter of Maps, each holding timestamp, message and stream.
//
// Arguments are (group [, options]), where options may carry "pattern",
// "start", "end" and "limit". Times are milliseconds since the Unix epoch, the
// unit the API uses.
func Dig(args ...meowrt.Value) meowrt.Value {
	if len(args) < 1 || len(args) > 2 {
		return furball("dig expects 1-2 arguments (group [, options]), got %d", len(args))
	}
	group, fb := expectString("dig", "log group", args[0])
	if fb != nil {
		return fb
	}

	opts := filterOptions{limit: maxEvents}
	if len(args) == 2 {
		m, ok := args[1].(*meowrt.Map)
		if !ok {
			typeName := "catnap"
			if args[1] != nil {
				typeName = args[1].Type()
			}
			return furball("dig expects a Map of options, got %s", typeName)
		}
		var optFb meowrt.Value
		if opts, optFb = extractFilterOptions(m); optFb != nil {
			return optFb
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	client, err := newLogs(ctx)
	if err != nil {
		return furball("%s", err)
	}
	in := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(group),
		Limit:        aws.Int32(opts.limit),
		StartTime:    opts.startTime,
		EndTime:      opts.endTime,
	}
	if opts.pattern != "" {
		in.FilterPattern = aws.String(opts.pattern)
	}
	out, err := client.FilterLogEvents(ctx, in)
	if err != nil {
		return furball("%s", err)
	}
	return eventsToList(out.Events)
}

// eventsToList renders log events as a litter of Maps.
func eventsToList(events []cwltypes.FilteredLogEvent) meowrt.Value {
	values := make([]meowrt.Value, 0, len(events))
	for _, e := range events {
		var ts int64
		if e.Timestamp != nil {
			ts = *e.Timestamp
		}
		values = append(values, meowrt.NewMap(map[string]meowrt.Value{
			"timestamp": meowrt.NewInt(ts),
			"message":   meowrt.NewString(deref(e.Message)),
			"stream":    meowrt.NewString(deref(e.LogStreamName)),
		}))
	}
	return meowrt.NewList(values...)
}

// Region reports the region the SDK resolved, so a program can confirm where it
// is about to act before acting.
func Region(args ...meowrt.Value) meowrt.Value {
	if len(args) != 0 {
		return furball("region expects no arguments, got %d", len(args))
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()

	region, err := resolveRegion(ctx)
	if err != nil {
		return furball("%s", err)
	}
	return meowrt.NewString(region)
}
