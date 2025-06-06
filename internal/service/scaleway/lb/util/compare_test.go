package util

import (
	"testing"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

func TestIPsEqual(t *testing.T) {
	type args struct {
		a []*string
		b []*string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal IPs, different order",
			args: args{
				a: scw.StringSlicePtr([]string{"1.1.1.1", "2.2.2.2"}),
				b: scw.StringSlicePtr([]string{"2.2.2.2", "1.1.1.1"}),
			},
			want: true,
		},
		{
			name: "equal IPs, same order",
			args: args{
				a: scw.StringSlicePtr([]string{"1.1.1.1", "2.2.2.2"}),
				b: scw.StringSlicePtr([]string{"1.1.1.1", "2.2.2.2"}),
			},
			want: true,
		},
		{
			name: "different lengths",
			args: args{
				a: scw.StringSlicePtr([]string{"1.1.1.1"}),
				b: scw.StringSlicePtr([]string{"1.1.1.1", "2.2.2.2"}),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPsEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IPsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
