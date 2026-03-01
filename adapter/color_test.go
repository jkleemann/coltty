package adapter

import "testing"

func TestHexToTerminalAppRGB(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		want    string
		wantErr bool
	}{
		{
			name: "white",
			hex:  "#ffffff",
			want: "{65535, 65535, 65535}",
		},
		{
			name: "black",
			hex:  "#000000",
			want: "{0, 0, 0}",
		},
		{
			name: "dracula foreground",
			hex:  "#f8f8f2",
			want: "{63736, 63736, 62194}",
		},
		{
			name: "dracula background",
			hex:  "#282a36",
			want: "{10280, 10794, 13878}",
		},
		{
			name: "pure red",
			hex:  "#ff0000",
			want: "{65535, 0, 0}",
		},
		{
			name: "no hash prefix",
			hex:  "f8f8f2",
			want: "{63736, 63736, 62194}",
		},
		{
			name:    "short hex",
			hex:     "#fff",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			hex:     "#gggggg",
			wantErr: true,
		},
		{
			name:    "empty string",
			hex:     "",
			wantErr: true,
		},
		{
			name:    "too long",
			hex:     "#ff00ff00",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HexToTerminalAppRGB(tt.hex)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
