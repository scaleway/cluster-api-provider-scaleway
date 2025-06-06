package scope

import "testing"

func Test_truncateString(t *testing.T) {
	type args struct {
		s      string
		maxLen int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string below max length",
			args: args{
				s:      "short",
				maxLen: 128,
			},
			want: "short",
		},
		{
			name: "exact length",
			args: args{
				s:      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				maxLen: 128,
			},
			want: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name: "long string",
			args: args{
				s:      "this is a very long string that exceeds the maximum length of 128 characters, so it should be truncated in the middle to fit within the limit imposed by Scaleway resource naming conventions.",
				maxLen: 128,
			},
			want: "this is a very long string that exceeds the maximum length of 1-ithin the limit imposed by Scaleway resource naming conventions.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateString(tt.args.s, tt.args.maxLen); got != tt.want {
				t.Errorf("truncateString() = %v, want %v", got, tt.want)
			}
		})
	}
}
