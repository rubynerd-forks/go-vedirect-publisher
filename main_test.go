package main

import (
	"reflect"
	"testing"
)

func TestParseExtras(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      map[string]string
		wantError bool
	}{
		{
			name:      "valid single pair",
			input:     "hostname=shedtofu",
			want:      map[string]string{"hostname": "shedtofu"},
			wantError: false,
		},
		{
			name:      "valid multiple pairs",
			input:     "hostname=shedtofu,installation=shed",
			want:      map[string]string{"hostname": "shedtofu", "installation": "shed"},
			wantError: false,
		},
		{
			name:      "empty string",
			input:     "",
			want:      map[string]string{},
			wantError: false,
		},
		{
			name:      "value with spaces",
			input:     "name=Shed TOFU",
			want:      map[string]string{"name": "Shed TOFU"},
			wantError: false,
		},
		{
			name:      "trimmed spaces",
			input:     " hostname = shedtofu , installation = shed ",
			want:      map[string]string{"hostname": "shedtofu", "installation": "shed"},
			wantError: false,
		},
		{
			name:      "value with equals sign",
			input:     "query=select=value",
			want:      map[string]string{"query": "select=value"},
			wantError: false,
		},
		{
			name:      "invalid format - no equals",
			input:     "invalid",
			want:      nil,
			wantError: true,
		},
		{
			name:      "invalid format - empty key",
			input:     "=value",
			want:      nil,
			wantError: true,
		},
		{
			name:      "invalid format - mixed valid and invalid",
			input:     "hostname=shedtofu,invalid",
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExtras(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("parseExtras() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseExtras() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceList(t *testing.T) {
	t.Run("String method empty list", func(t *testing.T) {
		var dl DeviceList
		if got := dl.String(); got != "" {
			t.Errorf("DeviceList.String() = %q, want empty string", got)
		}
	})

	t.Run("String method single device", func(t *testing.T) {
		dl := DeviceList{"/dev/ttyUSB0"}
		want := "/dev/ttyUSB0"
		if got := dl.String(); got != want {
			t.Errorf("DeviceList.String() = %q, want %q", got, want)
		}
	})

	t.Run("String method multiple devices", func(t *testing.T) {
		dl := DeviceList{"/dev/ttyUSB0", "/dev/ttyUSB1", "/dev/ttyUSB2"}
		want := "/dev/ttyUSB0,/dev/ttyUSB1,/dev/ttyUSB2"
		if got := dl.String(); got != want {
			t.Errorf("DeviceList.String() = %q, want %q", got, want)
		}
	})

	t.Run("Set method adds devices", func(t *testing.T) {
		var dl DeviceList
		devices := []string{"/dev/ttyUSB0", "/dev/ttyUSB1", "/dev/ttyUSB2"}

		for _, dev := range devices {
			if err := dl.Set(dev); err != nil {
				t.Errorf("DeviceList.Set(%q) error = %v", dev, err)
			}
		}

		if len(dl) != len(devices) {
			t.Errorf("DeviceList length = %d, want %d", len(dl), len(devices))
		}

		for i, want := range devices {
			if dl[i] != want {
				t.Errorf("DeviceList[%d] = %q, want %q", i, dl[i], want)
			}
		}
	})
}
