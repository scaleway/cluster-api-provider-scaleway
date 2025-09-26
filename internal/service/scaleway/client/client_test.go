package client

import (
	"errors"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	projectID = "11111111-1111-1111-1111-111111111111"
	secretKey = "11111111-1111-1111-1111-111111111111"
)

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

func TestNew(t *testing.T) {
	type args struct {
		region     scw.Region
		projectID  string
		secretData map[string][]byte
	}
	tests := []struct {
		name    string
		args    args
		asserts func(g *WithT, c *Client)
		wantErr bool
	}{
		{
			name: "empty secret",
			args: args{
				region:     scw.RegionFrPar,
				projectID:  projectID,
				secretData: map[string][]byte{},
			},
			wantErr: true,
		},
		{
			name: "invalid access key format",
			args: args{
				region:    scw.RegionFrPar,
				projectID: projectID,
				secretData: map[string][]byte{
					scw.ScwAccessKeyEnv: []byte("a"),
					scw.ScwSecretKeyEnv: []byte(secretKey),
				},
			},
			wantErr: true,
		},
		{
			name: "new client",
			args: args{
				region:    scw.RegionFrPar,
				projectID: projectID,
				secretData: map[string][]byte{
					scw.ScwAccessKeyEnv: []byte("SCWXXXXXXXXXXXXXXXXX"),
					scw.ScwSecretKeyEnv: []byte(secretKey),
					scw.ScwAPIURLEnv:    []byte("https://api.scaleway.com"),
				},
			},
			asserts: func(g *WithT, c *Client) {
				g.Expect(c).ToNot(BeNil())
				g.Expect(c.region).To(Equal(scw.RegionFrPar))
				g.Expect(c.projectID).To(Equal(projectID))
				g.Expect(c.secretKey).To(Equal(secretKey))
				g.Expect(c.vpc).ToNot(BeNil())
				g.Expect(c.vpcgw).ToNot(BeNil())
				g.Expect(c.lb).ToNot(BeNil())
				g.Expect(c.domain).ToNot(BeNil())
				g.Expect(c.instance).ToNot(BeNil())
				g.Expect(c.block).ToNot(BeNil())
				g.Expect(c.marketplace).ToNot(BeNil())
				g.Expect(c.ipam).ToNot(BeNil())
				g.Expect(c.k8s).ToNot(BeNil())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			got, err := New(tt.args.region, tt.args.projectID, tt.args.secretData)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				tt.asserts(g, got)
			}
		})
	}
}

func TestTagsWithoutCreatedBy(t *testing.T) {
	t.Parallel()
	type args struct {
		tags []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "remove created-by tag",
			args: args{
				tags: []string{"a", "b", "c", createdByTag},
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := TagsWithoutCreatedBy(tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TagsWithoutCreatedBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
