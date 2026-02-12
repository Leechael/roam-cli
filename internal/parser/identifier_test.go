package parser

import "testing"

func TestParseUID(t *testing.T) {
	cases := []struct {
		in   string
		uid  string
		ok   bool
		name string
	}{
		{name: "wrapped uid", in: "((abc123xyz))", uid: "abc123xyz", ok: true},
		{name: "plain uid", in: "abc-123_xyz", uid: "abc-123_xyz", ok: true},
		{name: "page title with space", in: "My Page", uid: "", ok: false},
		{name: "cjk title", in: "测试页面", uid: "", ok: false},
		{name: "invalid chars", in: "bad/uid", uid: "", ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uid, ok := ParseUID(tc.in)
			if ok != tc.ok || uid != tc.uid {
				t.Fatalf("ParseUID(%q) = (%q, %v), want (%q, %v)", tc.in, uid, ok, tc.uid, tc.ok)
			}
		})
	}
}
