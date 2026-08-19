package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/135yshr/meow/runtime/meowrt"
)

// What Whoami is written by hand to produce, the bridge produces from the SDK's
// own type without any wrapper at all.
//
// This is the case the wrapping was written for, so it is the one worth
// checking: reading *sts.GetCallerIdentityOutput needs no knowledge of AWS,
// only of Go.
func TestTheBridgeReadsTheSDKsOwnOutput(t *testing.T) {
	account, arn, userID := "123456789012", "arn:aws:iam::123456789012:user/nyan", "AIDAEXAMPLE"

	got := meowrt.CallGo("whoami", func() (*sts.GetCallerIdentityOutput, error) {
		return &sts.GetCallerIdentityOutput{Account: &account, Arn: &arn, UserId: &userID}, nil
	})

	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("got %s (%s), want a basket", got.String(), got.Type())
	}
	// The very three keys Whoami builds by hand.
	for key, want := range map[string]string{"account": account, "arn": arn, "user_id": userID} {
		v, present := m.Items[key]
		if !present {
			t.Errorf("no %q came through", key)
			continue
		}
		if v.String() != want {
			t.Errorf("%s is %q, want %q", key, v.String(), want)
		}
	}
}

// A pointer the SDK left nil is nothing, not an empty string pretending to be
// an answer — the same reading `deref` was written to give.
func TestTheBridgeReadsAnAbsentFieldAsNothing(t *testing.T) {
	got := meowrt.CallGo("whoami", func() (*sts.GetCallerIdentityOutput, error) {
		return &sts.GetCallerIdentityOutput{}, nil
	})

	m, ok := got.(*meowrt.Map)
	if !ok {
		t.Fatalf("got %s, want a basket", got.Type())
	}
	if v := m.Items["account"]; v.String() != "catnap" {
		t.Errorf("an absent account reads as %q, want catnap", v.String())
	}
}

// An SDK failure is the failure of the call.
func TestTheBridgePassesOnAnSDKFailure(t *testing.T) {
	got := meowrt.CallGo("whoami", func() (*sts.GetCallerIdentityOutput, error) {
		return nil, errRefused
	})

	if _, failed := meowrt.AsFurball(got); !failed {
		t.Fatalf("got %s, want a furball", got.String())
	}
}

type refused struct{}

func (refused) Error() string { return "credentials refused" }

var errRefused = refused{}
