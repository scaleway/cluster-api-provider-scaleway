package client

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestIsForbiddenError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "forbidden error",
			args: args{
				err: &scw.ResponseError{StatusCode: http.StatusForbidden},
			},
			want: true,
		},
		{
			name: "not a forbidden error",
			args: args{
				err: &scw.InvalidArgumentsError{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsForbiddenError(tt.args.err); got != tt.want {
				t.Errorf("IsForbiddenError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "wrapped ErrNoItemFound is a NotFound error",
			args: args{
				err: fmt.Errorf("%w: no apple found", ErrNoItemFound),
			},
			want: true,
		},
		{
			name: "scaleway NotFound error",
			args: args{
				err: &scw.ResourceNotFoundError{},
			},
			want: true,
		},
		{
			name: "not a NotFound error",
			args: args{
				err: &scw.InvalidArgumentsError{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.args.err); got != tt.want {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPreconditionFailedError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "scaleway PreconditionFailed error",
			args: args{
				err: &scw.PreconditionFailedError{},
			},
			want: true,
		},
		{
			name: "not a PreconditionFailed error",
			args: args{
				err: &scw.InvalidArgumentsError{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPreconditionFailedError(tt.args.err); got != tt.want {
				t.Errorf("IsPreconditionFailedError() = %v, want %v", got, tt.want)
			}
		})
	}
}
