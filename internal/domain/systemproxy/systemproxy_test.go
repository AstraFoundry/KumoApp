package systemproxy

import "testing"

func TestOwnership(t *testing.T) {
	applied := Applied{Service: "Wi-Fi", Mode: ModeManual, Host: "127.0.0.1", Port: 7890}
	kumo := Proxy{Enabled: true, Host: "localhost", Port: 7890}
	cases := []struct {
		name string
		snap Snapshot
		want Ownership
	}{
		{"owned", Snapshot{Web: kumo, Secure: kumo, SOCKS: kumo}, Owned},
		{"cleared", Snapshot{}, Cleared},
		{"other app", Snapshot{Web: Proxy{Enabled: true, Host: "127.0.0.1", Port: 7897}, Secure: kumo, SOCKS: kumo}, TakenOver},
		{"pac added", Snapshot{Web: kumo, Secure: kumo, SOCKS: kumo, AutoProxy: true, PACURL: "http://x"}, TakenOver},
	}
	for _, c := range cases {
		if got := applied.OwnershipOf(c.snap); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
	pac := Applied{Mode: ModePAC, PACURL: "http://127.0.0.1:5000/proxy.pac"}
	if pac.OwnershipOf(Snapshot{AutoProxy: true, PACURL: pac.PACURL}) != Owned {
		t.Error("pac owned")
	}
	if pac.OwnershipOf(Snapshot{AutoProxy: true, PACURL: "http://other"}) != TakenOver {
		t.Error("pac taken")
	}
}
