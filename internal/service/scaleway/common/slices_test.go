package common

import "testing"

func TestSlicesEqualIgnoreOrder(t *testing.T) {
	type args struct {
		a []string
		b []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not equal (size mismatch)",
			args: args{
				a: []string{"a"},
				b: []string{"a", "a"},
			},
			want: false,
		},
		{
			name: "not equal",
			args: args{
				a: []string{"a", "b"},
				b: []string{"a", "c"},
			},
			want: false,
		},
		{
			name: "equal with repeated elements",
			args: args{
				a: []string{"a", "a", "a"},
				b: []string{"a", "a", "a"},
			},
			want: true,
		},
		{
			name: "equal with no repeated elements",
			args: args{
				a: []string{"a", "b", "c"},
				b: []string{"a", "b", "c"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SlicesEqualIgnoreOrder(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("SlicesEqualIgnoreOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
