package config

import "testing"

func TestParseModelPrefix_Valid(t *testing.T) {
	cases := []struct{ in, wantClient, wantModel string }{
		{"openai:gpt-4", "openai", "gpt-4"},
		{"local:kimi-k2.6", "local", "kimi-k2.6"},
		{"DeepSeek:ds-1", "deepseek", "ds-1"},
	}
	for _, c := range cases {
		cl, md, err := ParseModelPrefix(c.in)
		if err != nil {
			t.Fatalf("unexpected err for %s: %v", c.in, err)
		}
		if cl != c.wantClient || md != c.wantModel {
			t.Fatalf("mismatch for %s: got %s:%s", c.in, cl, md)
		}
	}
}

func TestParseModelPrefix_Invalid(t *testing.T) {
	cases := []string{"", "gpt-4", "openai:", ":gpt", "unknown:mdl"}
	for _, s := range cases {
		_, _, err := ParseModelPrefix(s)
		if err == nil {
			t.Fatalf("expected error for %q", s)
		}
	}
}
