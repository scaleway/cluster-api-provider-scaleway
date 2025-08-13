package common

import "testing"

func TestIsUpToDate(t *testing.T) {
	type args struct {
		current string
		desired string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "not semver compliant (leading v)",
			args: args{
				current: "v1.1.1",
				desired: "v1.1.1",
			},
			wantErr: true,
		},
		{
			name: "same version is up to date",
			args: args{
				current: "1.30.0",
				desired: "1.30.0",
			},
			want: true,
		},
		{
			name: "current > desired is up to date",
			args: args{
				current: "1.31.0",
				desired: "1.30.0",
			},
			want: true,
		},
		{
			name: "current < desired is not up to date",
			args: args{
				current: "1.30.0",
				desired: "1.31.0",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsUpToDate(tt.args.current, tt.args.desired)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsUpToDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsUpToDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
