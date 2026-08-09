package connector

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/rancher/remotedialer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type singleConnListener struct {
	conn net.Conn
	addr net.Addr
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	if l.conn == nil {
		return nil, net.ErrClosed
	}
	conn := l.conn
	l.conn = nil
	return conn, nil
}

func (l *singleConnListener) Close() error {
	return nil
}

func (l *singleConnListener) Addr() net.Addr {
	return l.addr
}

func Run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("kite connector", flag.ContinueOnError)
	klog.InitFlags(flags)
	server := flags.String("server", "", "Kite server URL")
	token := flags.String("token", "", "Kite connector token")
	kubeconfig := flags.String("kubeconfig", "", "Path to kubeconfig file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *server == "" {
		return errors.New("--server is required")
	}
	if *token == "" {
		return errors.New("--token is required")
	}

	serverURL, err := url.Parse(*server)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	if serverURL.Fragment != "" {
		return errors.New("server URL must not contain a fragment")
	}
	switch serverURL.Scheme {
	case "http":
		serverURL.Scheme = "ws"
	case "https":
		serverURL.Scheme = "wss"
	default:
		return errors.New("server URL must use http or https")
	}
	if serverURL.Host == "" {
		return errors.New("invalid connector server URL")
	}
	serverURL.Path = strings.TrimRight(serverURL.Path, "/") + "/api/v1/connector/connect"
	serverURL.RawPath = ""

	var config *rest.Config
	if *kubeconfig == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("load in-cluster Kubernetes configuration: %w", err)
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return fmt.Errorf("load kubeconfig: %w", err)
		}
	}
	target, err := url.Parse(config.Host)
	if err != nil {
		return fmt.Errorf("parse Kubernetes API URL: %w", err)
	}
	transport, err := rest.TransportFor(config)
	if err != nil {
		return fmt.Errorf("create Kubernetes API transport: %w", err)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(target)
			req.Out.Host = target.Host
			req.Out.Header.Del("Authorization")
			for name := range req.Out.Header {
				if strings.HasPrefix(strings.ToLower(name), "impersonate-") {
					req.Out.Header.Del(name)
				}
			}
		},
		Transport:     transport,
		FlushInterval: -1,
	}
	headers := http.Header{"Authorization": []string{"Bearer " + *token}}
	localDialer := func(_ context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != kubernetesAPITarget {
			return nil, fmt.Errorf("unsupported tunnel target %s/%s", network, address)
		}
		proxyConn, tunnelConn := net.Pipe()
		proxyServer := &http.Server{
			Handler:           proxy,
			ReadHeaderTimeout: 10 * time.Second,
		}
		go func() {
			_ = proxyServer.Serve(&singleConnListener{conn: proxyConn, addr: proxyConn.LocalAddr()})
		}()
		return tunnelConn, nil
	}
	authorizer := func(network, address string) bool {
		return network == "tcp" && address == kubernetesAPITarget
	}

	klog.Info("Kite connector started")
	retryCount := 0
	for {
		err := remotedialer.ConnectToProxyWithDialer(ctx, serverURL.String(), headers, authorizer, nil, localDialer, nil)
		if ctx.Err() != nil {
			return nil //nolint:nilerr // Context cancellation is a clean shutdown.
		}
		retryCount++
		klog.Warningf("Kite connector connection lost (retry %d): %v", retryCount, err)
		if retryCount >= 10 {
			return fmt.Errorf("kite connector failed to connect after %d retries: %w", retryCount, err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Second):
		}
	}
}
