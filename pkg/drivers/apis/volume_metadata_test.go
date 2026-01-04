package apis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalVolumeSpec(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		excepted *VolumeSpec
		hasErr   bool
	}{
		{
			name:     "empty options",
			data:     map[string]string{},
			excepted: &VolumeSpec{},
			hasErr:   false,
		},
		{
			name: "valid options",
			data: map[string]string{
				"purgeAfterDelete": "true",
			},
			excepted: &VolumeSpec{
				PurgeAfterDelete: true},
			hasErr: false,
		},
		{
			name: "invalid value for purgeAfterDelete",
			data: map[string]string{
				"purgeAfterDelete": "yes",
			},
			excepted: &VolumeSpec{},
			hasErr:   true,
		},
		{
			name: "unknown option",
			data: map[string]string{
				"unknownOption": "value",
			},
			excepted: &VolumeSpec{},
			hasErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &VolumeSpec{}
			err := spec.Unmarshal(tt.data)
			assert.True(t, (err != nil) == tt.hasErr, "Unmarshal got not excepted error: %v", err)
			assert.Equal(t, spec, tt.excepted)
		})
	}
}
