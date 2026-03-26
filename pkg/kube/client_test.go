package kube

import (
	"context"
	"errors"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type getErrorClient struct {
	client.Client
	err error
}

func (c getErrorClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return c.err
}

func TestWaitForResourceDeletionReturnsNilWhenResourceIsMissing(t *testing.T) {
	obj := &corev1.ConfigMap{}
	obj.SetNamespace("default")
	obj.SetName("demo")

	err := WaitForResourceDeletion(context.Background(), fake.NewClientBuilder().Build(), obj, 600*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForResourceDeletion() error = %v, want nil", err)
	}
}

func TestWaitForResourceDeletionReturnsWrappedGetError(t *testing.T) {
	obj := &corev1.ConfigMap{}
	obj.SetNamespace("default")
	obj.SetName("demo")

	wantErr := errors.New("boom")
	err := WaitForResourceDeletion(context.Background(), getErrorClient{Client: fake.NewClientBuilder().Build(), err: wantErr}, obj, 600*time.Millisecond)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("WaitForResourceDeletion() error = %v, want wrapped %v", err, wantErr)
	}
}
