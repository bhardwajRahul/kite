package internal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/zxh326/kite/pkg/cluster"
	"github.com/zxh326/kite/pkg/model"
	"github.com/zxh326/kite/pkg/rbac"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func init() {
	_ = os.Setenv("MOCKEY_CHECK_GCFLAGS", "false")
}

func TestLoadUserCreatesSuperUser(t *testing.T) {
	oldUsername := kiteUsername
	oldPassword := kitePassword
	oldSyncNow := rbac.SyncNow
	defer func() {
		kiteUsername = oldUsername
		kitePassword = oldPassword
		rbac.SyncNow = oldSyncNow
	}()

	kiteUsername = "admin"
	kitePassword = "secret"
	rbac.SyncNow = make(chan struct{}, 1)

	countMock := mockey.Mock(model.CountUsers).Return(int64(0), nil).Build()
	defer countMock.UnPatch()

	addMock := mockey.Mock(model.AddSuperUser).To(func(user *model.User) error {
		if user.Username != "admin" || user.Password != "secret" {
			t.Fatalf("unexpected user: %#v", user)
		}
		return nil
	}).Build()
	defer addMock.UnPatch()

	if err := loadUser(); err != nil {
		t.Fatalf("loadUser() error = %v", err)
	}

	select {
	case <-rbac.SyncNow:
	default:
		t.Fatal("expected rbac sync signal")
	}
}

func TestLoadUserReturnsAddSuperUserError(t *testing.T) {
	oldUsername := kiteUsername
	oldPassword := kitePassword
	oldSyncNow := rbac.SyncNow
	defer func() {
		kiteUsername = oldUsername
		kitePassword = oldPassword
		rbac.SyncNow = oldSyncNow
	}()

	kiteUsername = "admin"
	kitePassword = "secret"
	rbac.SyncNow = make(chan struct{}, 1)

	wantErr := errors.New("boom")
	countMock := mockey.Mock(model.CountUsers).Return(int64(0), nil).Build()
	defer countMock.UnPatch()

	addMock := mockey.Mock(model.AddSuperUser).Return(wantErr).Build()
	defer addMock.UnPatch()

	if err := loadUser(); !errors.Is(err, wantErr) {
		t.Fatalf("loadUser() error = %v, want %v", err, wantErr)
	}

	select {
	case <-rbac.SyncNow:
		t.Fatal("unexpected rbac sync signal")
	default:
	}
}

func TestLoadClustersSkipsWhenClustersExist(t *testing.T) {
	countMock := mockey.Mock(model.CountClusters).Return(int64(1), nil).Build()
	defer countMock.UnPatch()

	importMock := mockey.Mock(cluster.ImportClustersFromKubeconfig).To(func(*clientcmdapi.Config) int64 {
		t.Fatal("ImportClustersFromKubeconfig() should not be called")
		return 0
	}).Build()
	defer importMock.UnPatch()

	if err := loadClusters(); err != nil {
		t.Fatalf("loadClusters() error = %v", err)
	}
}

func TestLoadClustersImportsFromKubeconfig(t *testing.T) {
	countMock := mockey.Mock(model.CountClusters).Return(int64(0), nil).Build()
	defer countMock.UnPatch()

	imported := false
	importMock := mockey.Mock(cluster.ImportClustersFromKubeconfig).To(func(cfg *clientcmdapi.Config) int64 {
		imported = true
		if cfg.CurrentContext != "dev" {
			t.Fatalf("CurrentContext = %q, want %q", cfg.CurrentContext, "dev")
		}
		if len(cfg.Contexts) != 1 {
			t.Fatalf("len(Contexts) = %d, want 1", len(cfg.Contexts))
		}
		return 1
	}).Build()
	defer importMock.UnPatch()

	dir := t.TempDir()
	kubeconfigPath := filepath.Join(dir, "config")
	if err := os.WriteFile(kubeconfigPath, []byte(validKubeconfig), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	t.Setenv("KUBECONFIG", kubeconfigPath)

	if err := loadClusters(); err != nil {
		t.Fatalf("loadClusters() error = %v", err)
	}
	if !imported {
		t.Fatal("expected kubeconfig import")
	}
}

func TestLoadClustersReturnsLoadError(t *testing.T) {
	countMock := mockey.Mock(model.CountClusters).Return(int64(0), nil).Build()
	defer countMock.UnPatch()

	dir := t.TempDir()
	kubeconfigPath := filepath.Join(dir, "config")
	if err := os.WriteFile(kubeconfigPath, []byte("not: [valid"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	t.Setenv("KUBECONFIG", kubeconfigPath)

	if err := loadClusters(); err == nil {
		t.Fatal("expected loadClusters() to return error")
	}
}

const validKubeconfig = `apiVersion: v1
kind: Config
current-context: dev
clusters:
- name: dev
  cluster:
    server: https://example.com
contexts:
- name: dev
  context:
    cluster: dev
    user: dev
users:
- name: dev
  user:
    token: test-token
`
