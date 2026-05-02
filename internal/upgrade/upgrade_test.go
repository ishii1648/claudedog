package upgrade

import "testing"

func TestParseChecksum(t *testing.T) {
	body := `0123abc  hitl-metrics_darwin_arm64.tar.gz
deadbeef  hitl-metrics_linux_amd64.tar.gz
cafef00d  hitl-metrics_darwin_amd64.tar.gz
`
	got, err := parseChecksum(body, "hitl-metrics_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatalf("parseChecksum: %v", err)
	}
	if got != "0123abc" {
		t.Errorf("checksum = %q, want %q", got, "0123abc")
	}

	if _, err := parseChecksum(body, "no-such-asset.tar.gz"); err == nil {
		t.Errorf("expected error for missing asset, got nil")
	}
}

func TestPickURLs(t *testing.T) {
	assets := []releaseAsset{
		{Name: "checksums.txt", BrowserDownloadURL: "https://example/checksums.txt"},
		{Name: "hitl-metrics_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example/asset.tar.gz"},
		{Name: "hitl-metrics_linux_amd64.tar.gz", BrowserDownloadURL: "https://example/other.tar.gz"},
	}
	a, c := pickURLs(assets, "hitl-metrics_darwin_arm64.tar.gz")
	if a != "https://example/asset.tar.gz" {
		t.Errorf("asset url = %q", a)
	}
	if c != "https://example/checksums.txt" {
		t.Errorf("checksums url = %q", c)
	}

	a, _ = pickURLs(assets, "hitl-metrics_windows_arm64.tar.gz")
	if a != "" {
		t.Errorf("expected empty asset url for missing platform, got %q", a)
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"v0.0.3":   "0.0.3",
		"0.0.3":    "0.0.3",
		" v0.0.3 ": "0.0.3",
		"":         "",
		"dev":      "dev",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
