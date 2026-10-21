package proxy

import "testing"

func TestCountry(t *testing.T) {
	cases := map[string]string{
		"🇯🇵 Tokyo 01":    "JP",
		"香港 01":          "HK",
		"Hong Kong HKT":  "HK",
		"us-la-02":       "US",
		"美国|专线洛杉矶|AI 2倍": "US",
		"韩国|专线 3倍":       "KR",
		"新加坡 IPLC":       "SG",
		"Singapore 1":    "SG",
		"UK London":      "GB",
		"Direct":         "",
		"Auto Select":    "",
		"日本 東京":          "JP",
		"台灣 HiNet":       "TW",
	}
	for name, want := range cases {
		if got := Country(name); got != want {
			t.Errorf("Country(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestDisplayNameAndFlag(t *testing.T) {
	if got := DisplayName("🇯🇵 | Tokyo"); got != "Tokyo" {
		t.Errorf("DisplayName = %q", got)
	}
	if got := DisplayName("Tokyo"); got != "Tokyo" {
		t.Errorf("DisplayName = %q", got)
	}
	if Flag("jp") != "🇯🇵" || Flag("x") != "" {
		t.Error("Flag")
	}
}
