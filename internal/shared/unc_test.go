package shared

import "testing"

func TestValidateUNCPath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		allowedShares []string
		wantErr       bool
	}{
		{
			name:          "not a UNC path",
			path:          `C:\foo\bar`,
			allowedShares: []string{`\\NAS01\media`},
			wantErr:       true,
		},
		{
			name:          "valid prefix match",
			path:          `\\NAS01\media\video.mkv`,
			allowedShares: []string{`\\NAS01\media`},
			wantErr:       false,
		},
		{
			name:          "path not in allowed list",
			path:          `\\NAS02\other\file.mkv`,
			allowedShares: []string{`\\NAS01\media`},
			wantErr:       true,
		},
		{
			name:          "case insensitive prefix matching",
			path:          `\\NAS01\MEDIA\video.mkv`,
			allowedShares: []string{`\\NAS01\media`},
			wantErr:       false,
		},
		{
			name:          "case insensitive prefix matching reversed",
			path:          `\\NAS01\media\video.mkv`,
			allowedShares: []string{`\\NAS01\MEDIA`},
			wantErr:       false,
		},
		{
			name:          "empty allowed shares list",
			path:          `\\NAS01\media\video.mkv`,
			allowedShares: []string{},
			wantErr:       true,
		},
		{
			name:          "nil allowed shares list",
			path:          `\\NAS01\media\video.mkv`,
			allowedShares: nil,
			wantErr:       true,
		},
		{
			name:          "multiple allowed shares second one matches",
			path:          `\\NAS02\archive\old.mkv`,
			allowedShares: []string{`\\NAS01\media`, `\\NAS02\archive`},
			wantErr:       false,
		},
		{
			name:          "multiple allowed shares first one matches",
			path:          `\\NAS01\media\new.mkv`,
			allowedShares: []string{`\\NAS01\media`, `\\NAS02\archive`},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUNCPath(tt.path, tt.allowedShares)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUNCPath(%q, %v) error = %v, wantErr %v",
					tt.path, tt.allowedShares, err, tt.wantErr)
			}
		})
	}
}

func TestIsUNCPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "UNC path with server and share",
			path: `\\NAS01\media\file.mkv`,
			want: true,
		},
		{
			name: "UNC path bare backslashes",
			path: `\\`,
			want: true,
		},
		{
			name: "local drive path",
			path: `C:\foo\bar`,
			want: false,
		},
		{
			name: "relative path",
			path: `foo\bar`,
			want: false,
		},
		{
			name: "single backslash",
			path: `\foo`,
			want: false,
		},
		{
			name: "empty string",
			path: "",
			want: false,
		},
		{
			name: "forward slashes",
			path: "//server/share",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUNCPath(tt.path)
			if got != tt.want {
				t.Errorf("IsUNCPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
