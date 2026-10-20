package downloader

import (
	"testing"
)

func TestLiveURLAndCookieAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in short mode")
	}
}
