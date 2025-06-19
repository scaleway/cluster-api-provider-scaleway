package vpcgw

import "testing"

func Test_canUpgradeTypes(t *testing.T) {
	t.Parallel()

	type args struct {
		types   []string
		current string
		desired string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "upgrade from VPC-GW-S to VPC-GW-XL",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "VPC-GW-S",
				desired: "VPC-GW-XL",
			},
			want: true,
		},
		{
			name: "upgrade from VPC-GW-S to VPC-GW-M",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "VPC-GW-S",
				desired: "VPC-GW-XL",
			},
			want: true,
		},
		{
			name: "current equals desired, not upgradable",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "VPC-GW-S",
				desired: "VPC-GW-S",
			},
			want: false,
		},
		{
			name: "unknown current type, not upgradable",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "UNKNOWN-S",
				desired: "VPC-GW-L",
			},
			want: false,
		},
		{
			name: "unknown current and desired type, not upgradable",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "UNKNOWN-S",
				desired: "UNKNOWN-M",
			},
			want: false,
		},
		{
			name: "unknown desired type, not upgradable",
			args: args{
				types:   []string{"VPC-GW-S", "VPC-GW-M", "VPC-GW-L", "VPC-GW-XL"},
				current: "VPC-GW-S",
				desired: "UNKNOWN-M",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := canUpgradeTypes(tt.args.types, tt.args.current, tt.args.desired); got != tt.want {
				t.Errorf("canUpgradeTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}
