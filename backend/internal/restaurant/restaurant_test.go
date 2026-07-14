package restaurant

import (
	"errors"
	"testing"
)

func TestLastCheckFromError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus LastCheckStatus
		wantDetail string
	}{
		{name: "nil", err: nil, wantStatus: LastCheckStatusSuccess, wantDetail: ""},
		{name: "manual upload only", err: ErrManualUploadOnly, wantStatus: LastCheckStatusNotChanged, wantDetail: ""},
		{name: "menu unchanged", err: ErrMenuUnchanged, wantStatus: LastCheckStatusNotChanged, wantDetail: ""},
		{name: "generic error", err: errors.New("boom"), wantStatus: LastCheckStatusError, wantDetail: "boom"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, detail := LastCheckFromError(tc.err)
			if status != tc.wantStatus {
				t.Fatalf("unexpected status, got %q want %q", status, tc.wantStatus)
			}
			if detail != tc.wantDetail {
				t.Fatalf("unexpected detail, got %q want %q", detail, tc.wantDetail)
			}
		})
	}
}
