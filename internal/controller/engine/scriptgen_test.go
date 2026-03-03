package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEscapeBat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "ampersand", input: "a&b", want: "a^&b"},
		{name: "pipe", input: "a|b", want: "a^|b"},
		{name: "less than", input: "a<b", want: "a^<b"},
		{name: "greater than", input: "a>b", want: "a^>b"},
		{name: "caret", input: "a^b", want: "a^^b"},
		{name: "no special chars", input: "hello world", want: "hello world"},
		{name: "multiple special chars", input: "echo foo & bar | baz > out < in ^ end", want: "echo foo ^& bar ^| baz ^> out ^< in ^^ end"},
		{name: "empty string", input: "", want: ""},
		{name: "all specials adjacent", input: "&|<>^", want: "^&^|^<^>^^"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeBat(tt.input)
			if got != tt.want {
				t.Errorf("escapeBat(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTrimAvs(t *testing.T) {
	fn := templateFuncs["trimAvs"].(func(int, int) string)

	tests := []struct {
		name       string
		start, end int
		want       string
	}{
		{name: "normal range", start: 0, end: 1000, want: "Trim(0, 1000)"},
		{name: "same frame", start: 500, end: 500, want: "Trim(500, 500)"},
		{name: "large values", start: 100000, end: 200000, want: "Trim(100000, 200000)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("trimAvs(%d, %d) = %q, want %q", tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestTrimVpy(t *testing.T) {
	fn := templateFuncs["trimVpy"].(func(int, int) string)

	tests := []struct {
		name       string
		start, end int
		want       string
	}{
		{name: "normal range", start: 0, end: 1000, want: "[0:1000]"},
		{name: "same frame", start: 500, end: 500, want: "[500:500]"},
		{name: "large values", start: 100000, end: 200000, want: "[100000:200000]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(tt.start, tt.end)
			if got != tt.want {
				t.Errorf("trimVpy(%d, %d) = %q, want %q", tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestGpuFlag(t *testing.T) {
	fn := templateFuncs["gpuFlag"].(func(string) string)

	tests := []struct {
		name   string
		vendor string
		want   string
	}{
		{name: "nvidia", vendor: "nvidia", want: "--hwaccel nvenc --hwaccel_output_format cuda"},
		{name: "nvidia uppercase", vendor: "NVIDIA", want: "--hwaccel nvenc --hwaccel_output_format cuda"},
		{name: "amd", vendor: "amd", want: "--hwaccel amf"},
		{name: "amd mixed case", vendor: "Amd", want: "--hwaccel amf"},
		{name: "intel", vendor: "intel", want: "--hwaccel qsv"},
		{name: "intel uppercase", vendor: "INTEL", want: "--hwaccel qsv"},
		{name: "unknown vendor", vendor: "qualcomm", want: ""},
		{name: "empty string", vendor: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(tt.vendor)
			if got != tt.want {
				t.Errorf("gpuFlag(%q) = %q, want %q", tt.vendor, got, tt.want)
			}
		})
	}
}

func TestDefaultFunc(t *testing.T) {
	fn := templateFuncs["default"].(func(string, string) string)

	tests := []struct {
		name string
		dflt string
		val  string
		want string
	}{
		{name: "empty val returns default", dflt: "fallback", val: "", want: "fallback"},
		{name: "non-empty val returns val", dflt: "fallback", val: "actual", want: "actual"},
		{name: "both empty", dflt: "", val: "", want: ""},
		{name: "default empty but val set", dflt: "", val: "actual", want: "actual"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fn(tt.dflt, tt.val)
			if got != tt.want {
				t.Errorf("default(%q, %q) = %q, want %q", tt.dflt, tt.val, got, tt.want)
			}
		})
	}
}

func TestRenderToFile(t *testing.T) {
	t.Run("simple variable substitution", func(t *testing.T) {
		dir := t.TempDir()
		outPath := filepath.Join(dir, "output.bat")

		data := map[string]string{
			"SOURCE_PATH": `\\NAS01\test.mkv`,
		}

		err := renderToFile("test", "{{.SOURCE_PATH}}", data, outPath)
		if err != nil {
			t.Fatalf("renderToFile() error = %v", err)
		}

		got, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("reading output file: %v", err)
		}

		want := `\\NAS01\test.mkv`
		if string(got) != want {
			t.Errorf("file contents = %q, want %q", string(got), want)
		}
	})

	t.Run("escapeBat template function", func(t *testing.T) {
		dir := t.TempDir()
		outPath := filepath.Join(dir, "output.bat")

		data := map[string]string{
			"V": "a&b",
		}

		err := renderToFile("test", "{{ escapeBat .V }}", data, outPath)
		if err != nil {
			t.Fatalf("renderToFile() error = %v", err)
		}

		got, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("reading output file: %v", err)
		}

		want := "a^&b"
		if string(got) != want {
			t.Errorf("file contents = %q, want %q", string(got), want)
		}
	})

	t.Run("bad template syntax returns error", func(t *testing.T) {
		dir := t.TempDir()
		outPath := filepath.Join(dir, "output.bat")

		data := map[string]string{}

		err := renderToFile("bad", "{{ .Foo", data, outPath)
		if err == nil {
			t.Fatal("renderToFile() expected error for bad template syntax, got nil")
		}
		if !strings.Contains(err.Error(), "parse template") {
			t.Errorf("error = %q, want it to contain %q", err.Error(), "parse template")
		}
	})

	t.Run("multiple template functions combined", func(t *testing.T) {
		dir := t.TempDir()
		outPath := filepath.Join(dir, "output.bat")

		data := map[string]string{
			"CMD": "echo hello & goodbye",
		}

		err := renderToFile("test", `{{ escapeBat .CMD }}`, data, outPath)
		if err != nil {
			t.Fatalf("renderToFile() error = %v", err)
		}

		got, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("reading output file: %v", err)
		}

		want := "echo hello ^& goodbye"
		if string(got) != want {
			t.Errorf("file contents = %q, want %q", string(got), want)
		}
	})
}
