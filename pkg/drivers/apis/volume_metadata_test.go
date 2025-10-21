package apis

import (
	"reflect"
	"testing"
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
				"purgeAfterDelete":   "true",
				"allowMultipleMount": "false",
			},
			excepted: &VolumeSpec{
				PurgeAfterDelete:   true,
				AllowMultipleMount: false,
			},
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
			name: "invalid value for allowMultipleMount",
			data: map[string]string{
				"allowMultipleMount": "yes",
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
			if !reflect.DeepEqual(spec, tt.excepted) {
				t.Errorf("Unmarshal expected %v, got %v", tt.excepted, spec)
			}
			if (err != nil) != tt.hasErr {
				t.Errorf("Unmarshal got not excepted error: %v", err)
			}
		})
	}
}
