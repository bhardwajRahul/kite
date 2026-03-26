package version

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxh326/kite/pkg/common"
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "with v prefix", input: "v1.2.3", want: "1.2.3"},
		{name: "without v prefix", input: "1.2.3", want: "1.2.3"},
		{name: "invalid", input: "not-a-version", wantErr: true},
		{name: "empty", input: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSemver(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("unexpected version: want %q, got %q", tt.want, got.String())
			}
		})
	}
}

func TestGetVersionWithoutVersionCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origVersion := Version
	origBuildDate := BuildDate
	origCommitID := CommitID
	origEnableVersionCheck := common.EnableVersionCheck
	t.Cleanup(func() {
		Version = origVersion
		BuildDate = origBuildDate
		CommitID = origCommitID
		common.EnableVersionCheck = origEnableVersionCheck
	})

	Version = "1.2.3"
	BuildDate = "2026-03-27"
	CommitID = "abc123"
	common.EnableVersionCheck = false

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/version", nil)

	GetVersion(c)

	if recorder.Code != 200 {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}

	var got VersionInfo
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got.Version != "1.2.3" || got.BuildDate != "2026-03-27" || got.CommitID != "abc123" {
		t.Fatalf("unexpected version info: %#v", got)
	}
	if got.HasNew || got.Release != "" {
		t.Fatalf("unexpected update fields: %#v", got)
	}
}

func TestGetVersionWithCachedUpdateResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origVersion := Version
	origBuildDate := BuildDate
	origCommitID := CommitID
	origEnableVersionCheck := common.EnableVersionCheck
	origCachedUpdateResult := cachedUpdateResult
	origLastUpdateFetch := lastUpdateFetch
	t.Cleanup(func() {
		Version = origVersion
		BuildDate = origBuildDate
		CommitID = origCommitID
		common.EnableVersionCheck = origEnableVersionCheck
		cachedUpdateResult = origCachedUpdateResult
		lastUpdateFetch = origLastUpdateFetch
	})

	Version = "1.2.3"
	BuildDate = "2026-03-27"
	CommitID = "abc123"
	common.EnableVersionCheck = true
	cachedUpdateResult = updateCheckResult{
		hasNew:     true,
		releaseURL: "https://example.com/releases/v1.2.4",
	}
	lastUpdateFetch = time.Now()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest("GET", "/version", nil)

	GetVersion(c)

	if recorder.Code != 200 {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}

	var got VersionInfo
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !got.HasNew || got.Release != "https://example.com/releases/v1.2.4" {
		t.Fatalf("unexpected update fields: %#v", got)
	}
}

func TestCheckForUpdateShortCircuitsWithoutNetwork(t *testing.T) {
	origCachedUpdateResult := cachedUpdateResult
	origLastUpdateFetch := lastUpdateFetch
	t.Cleanup(func() {
		cachedUpdateResult = origCachedUpdateResult
		lastUpdateFetch = origLastUpdateFetch
	})

	cachedUpdateResult = updateCheckResult{
		hasNew:     true,
		releaseURL: "https://example.com/releases/v1.2.4",
	}
	lastUpdateFetch = time.Now()

	got := checkForUpdate(context.Background(), "1.2.3")
	if !got.hasNew || got.releaseURL != "https://example.com/releases/v1.2.4" {
		t.Fatalf("unexpected cached result: %#v", got)
	}
}

func TestCheckForUpdateSkipsBlankAndDevVersions(t *testing.T) {
	if got := checkForUpdate(context.Background(), "   "); got != (updateCheckResult{}) {
		t.Fatalf("blank version result = %#v, want zero value", got)
	}
	if got := checkForUpdate(context.Background(), "dev"); got != (updateCheckResult{}) {
		t.Fatalf("dev version result = %#v, want zero value", got)
	}
}
