package kube

import (
	"net/http"
	"reflect"
	"testing"
)

func TestBuildProxyURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		host      string
		kind      string
		namespace string
		resource  string
		path      string
		rawQuery  string
		want      string
		wantErr   bool
	}{
		{
			name:      "encodes and joins proxy url",
			host:      "https://example.com/",
			kind:      "pods",
			namespace: "default",
			resource:  "nginx",
			path:      "/logs/",
			rawQuery:  "b=2&a=1",
			want:      "https://example.com/api/v1/namespaces/default/pods/nginx/proxy/logs/?a=1&b=2",
		},
		{
			name:      "rejects invalid query",
			host:      "https://example.com",
			kind:      "pods",
			namespace: "default",
			resource:  "nginx",
			path:      "",
			rawQuery:  "%",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := buildProxyURL(tt.host, tt.kind, tt.namespace, tt.resource, tt.path, tt.rawQuery)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got url %q", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("buildProxyURL() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("buildProxyURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCloneHeader(t *testing.T) {
	t.Parallel()

	src := http.Header{}
	src.Add("X-Test", "one")
	src.Add("X-Test", "two")
	src.Add("Authorization", "token")

	got := cloneHeader(src)
	src.Set("X-Test", "changed")

	if !reflect.DeepEqual(got.Values("X-Test"), []string{"one", "two"}) {
		t.Fatalf("cloneHeader() copied values = %v, want %v", got.Values("X-Test"), []string{"one", "two"})
	}
	if got.Get("Authorization") != "token" {
		t.Fatalf("cloneHeader() authorization = %q, want %q", got.Get("Authorization"), "token")
	}
}

func TestCopyHeader(t *testing.T) {
	t.Parallel()

	dst := http.Header{}
	src := http.Header{}
	src.Add("X-Test", "one")
	src.Add("X-Test", "two")

	copyHeader(dst, src)
	src.Set("X-Test", "changed")

	if !reflect.DeepEqual(dst.Values("X-Test"), []string{"one", "two"}) {
		t.Fatalf("copyHeader() copied values = %v, want %v", dst.Values("X-Test"), []string{"one", "two"})
	}
}
