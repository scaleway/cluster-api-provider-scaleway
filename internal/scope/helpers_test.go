package scope

import "testing"

func Test_base36TruncatedHash(t *testing.T) {
	t.Parallel()
	type args struct {
		str     string
		hashLen int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "hash value",
			args: args{
				str:     "test",
				hashLen: 16,
			},
			want: "wo9lxbufgjnn34bj",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := base36TruncatedHash(tt.args.str, tt.args.hashLen)
			if (err != nil) != tt.wantErr {
				t.Errorf("base36TruncatedHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("base36TruncatedHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
