package main

import "testing"

func TestParseFormat(t *testing.T) {
	cases := []struct {
		in      string
		want    outputFormat
		wantErr bool
	}{
		{"", formatMarkdown, false},
		{"md", formatMarkdown, false},
		{"markdown", formatMarkdown, false},
		{"txt", formatText, false},
		{"text", formatText, false},
		{"plain", formatText, false},
		{"TXT", formatText, false},
		{"json", 0, true},
		{"pdf", 0, true},
	}
	for _, c := range cases {
		got, err := parseFormat(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseFormat(%q): want error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseFormat(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseFormat(%q)=%v want %v", c.in, got, c.want)
		}
	}
}

func TestInferFormatFromPath(t *testing.T) {
	md, ok := inferFormatFromPath("out.md")
	if !ok || md != formatMarkdown {
		t.Fatalf("out.md -> %v ok=%v", md, ok)
	}
	txt, ok := inferFormatFromPath("/tmp/foo.TXT")
	if !ok || txt != formatText {
		t.Fatalf("foo.TXT -> %v ok=%v", txt, ok)
	}
	_, ok = inferFormatFromPath("out.bin")
	if ok {
		t.Fatal("out.bin should not infer")
	}
}
