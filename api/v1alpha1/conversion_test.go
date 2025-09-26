package v1alpha1

import (
	"reflect"
	"testing"

	"k8s.io/utils/ptr"
)

func Test_ptrIfNotZero(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		v    string
		want *string
	}{
		{
			name: "empty value",
			v:    "",
			want: nil,
		},
		{
			name: "value not empty",
			v:    "test",
			want: ptr.To("test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ptrIfNotZero(tt.v)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ptrIfNotZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
