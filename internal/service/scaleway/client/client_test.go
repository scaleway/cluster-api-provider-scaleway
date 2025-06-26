package client

import (
	"errors"
	"testing"
)

const projectID = "11111111-1111-1111-1111-111111111111"

var errAPI = errors.New("API error")

func Test_validateTags(t *testing.T) {
	type args struct {
		tags []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "empty tags is invalid",
			args:    args{},
			wantErr: true,
		},
		{
			name: "valid tags",
			args: args{tags: []string{"hello", "world"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTags(tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("validateTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_matchTags(t *testing.T) {
	type args struct {
		tags       []string
		wantedTags []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "defaults to true",
			args: args{},
			want: true,
		},
		{
			name: "partial match",
			args: args{
				tags:       []string{"hello", "world"},
				wantedTags: []string{"hello"},
			},
			want: true,
		},
		{
			name: "no match",
			args: args{
				tags:       []string{"hello", "world"},
				wantedTags: []string{"test"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchTags(tt.args.tags, tt.args.wantedTags); got != tt.want {
				t.Errorf("matchTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
