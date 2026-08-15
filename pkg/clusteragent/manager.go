package clusteragent

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rancher/remotedialer"
	"github.com/zxh326/kite/pkg/common"
	"github.com/zxh326/kite/pkg/model"
	"golang.org/x/crypto/hkdf"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

const (
	clusterAgentTokenSize = 32
	manifestGrantTimeout  = 10 * time.Minute
	manifestGrantAAD      = "kite:cluster-agent-manifest-grant:v1"
)

var ErrInvalidManifestGrant = errors.New("invalid manifest grant")

type requestStateKey struct{}

type requestState struct {
	authenticated        bool
	clientKey            string
	connectionGeneration uint64
}

type connectionTrackingResponseWriter struct {
	http.ResponseWriter
	manager    *Manager
	state      *requestState
	connection net.Conn
}

func (w *connectionTrackingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	connection, readWriter, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}
	w.connection = connection
	if !w.manager.addConnection(w.state.clientKey, w.state.connectionGeneration, connection) {
		_ = connection.Close()
	}
	return connection, readWriter, nil
}

type manifestGrant struct {
	ClusterAgentToken string `json:"token"`
	ExpiresAt         int64  `json:"exp"`
}

type Manager struct {
	server                *remotedialer.Server
	onChange              func()
	jwtSecret             string
	mu                    sync.RWMutex
	registrations         map[string]registeredCluster
	connections           map[string]map[net.Conn]struct{}
	connectionGenerations map[string]uint64
}

func NewManager(onChange func()) *Manager {
	m := &Manager{
		onChange:              onChange,
		jwtSecret:             common.JwtSecret,
		registrations:         make(map[string]registeredCluster),
		connections:           make(map[string]map[net.Conn]struct{}),
		connectionGenerations: make(map[string]uint64),
	}
	m.server = remotedialer.New(m.authorize, remotedialer.DefaultErrorWriter)
	return m
}

func NewToken() (string, string, error) {
	value := make([]byte, clusterAgentTokenSize)
	if _, err := rand.Read(value); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(value)
	return token, tokenHash(token), nil
}

func (m *Manager) RegistrationPublicKey(token string) (string, error) {
	cluster, err := resolveToken(token)
	if err != nil {
		return "", err
	}
	if cluster == nil {
		return "", nil
	}
	if cluster.ClusterAgentPublicKey == "" {
		return "", errors.New("cluster agent registration public key is missing")
	}
	return cluster.ClusterAgentPublicKey, nil
}

func tokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func validToken(token string) bool {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
	return err == nil && len(decoded) == clusterAgentTokenSize
}

func resolveToken(token string) (*model.Cluster, error) {
	if !validToken(token) {
		return nil, nil
	}
	cluster, err := model.GetClusterByClusterAgentTokenHash(tokenHash(token))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !cluster.ClusterAgent || !cluster.Enable {
		return nil, nil
	}
	return cluster, nil
}

func manifestGrantCipher(jwtSecret string) (cipher.AEAD, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdf.New(sha256.New, []byte(jwtSecret), nil, []byte(manifestGrantAAD)), key); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (m *Manager) CreateManifestGrant(clusterAgentToken string) (string, error) {
	payload, err := json.Marshal(manifestGrant{
		ClusterAgentToken: clusterAgentToken,
		ExpiresAt:         time.Now().Add(manifestGrantTimeout).Unix(),
	})
	if err != nil {
		return "", err
	}
	aead, err := manifestGrantCipher(m.jwtSecret)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nil, nonce, payload, []byte(manifestGrantAAD))
	encrypted := make([]byte, len(nonce)+len(ciphertext))
	copy(encrypted, nonce)
	copy(encrypted[len(nonce):], ciphertext)
	return base64.RawURLEncoding.EncodeToString(encrypted), nil
}

func (m *Manager) ResolveManifestGrant(grant string) (string, error) {
	encrypted, err := base64.RawURLEncoding.Strict().DecodeString(grant)
	if err != nil {
		return "", ErrInvalidManifestGrant
	}
	aead, err := manifestGrantCipher(m.jwtSecret)
	if err != nil {
		return "", err
	}
	if len(encrypted) < aead.NonceSize() {
		return "", nil
	}
	nonce, ciphertext := encrypted[:aead.NonceSize()], encrypted[aead.NonceSize():]
	payload, err := aead.Open(nil, nonce, ciphertext, []byte(manifestGrantAAD))
	if err != nil {
		return "", ErrInvalidManifestGrant
	}
	var stored manifestGrant
	if err := json.Unmarshal(payload, &stored); err != nil {
		return "", ErrInvalidManifestGrant
	}
	if time.Now().Unix() >= stored.ExpiresAt {
		return "", ErrInvalidManifestGrant
	}
	cluster, err := resolveToken(stored.ClusterAgentToken)
	if err != nil {
		return "", err
	}
	if cluster == nil {
		return "", nil
	}
	return stored.ClusterAgentToken, nil
}

func (m *Manager) authorize(req *http.Request) (string, bool, error) {
	cluster, err := authenticateClusterAgent(req)
	if err != nil {
		klog.Errorf("Failed to validate cluster agent token: %v", err)
		return "", false, errors.New("failed to validate cluster agent token")
	}
	if cluster == nil {
		return "", false, nil
	}
	clientKey := strconv.FormatUint(uint64(cluster.ID), 10)
	if state, ok := req.Context().Value(requestStateKey{}).(*requestState); ok {
		state.authenticated = true
		state.clientKey = clientKey
		m.mu.RLock()
		state.connectionGeneration = m.connectionGenerations[clientKey]
		m.mu.RUnlock()
	}
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			if m.server.HasSession(clientKey) {
				m.onChange()
				return
			}
			select {
			case <-req.Context().Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return clientKey, true, nil
}

func authenticateClusterAgent(req *http.Request) (*model.Cluster, error) {
	authorization := req.Header.Get("Authorization")
	token, found := strings.CutPrefix(authorization, "Bearer ")
	if !found || token == "" {
		return nil, nil
	}
	cluster, err := resolveToken(token)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (m *Manager) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	state := &requestState{}
	req = req.WithContext(context.WithValue(req.Context(), requestStateKey{}, state))
	trackingWriter := &connectionTrackingResponseWriter{ResponseWriter: rw, manager: m, state: state}
	m.server.ServeHTTP(trackingWriter, req)
	if trackingWriter.connection != nil && state.clientKey != "" {
		m.removeConnection(state.clientKey, trackingWriter.connection)
	}
	if state.authenticated {
		m.onChange()
	}
}

func (m *Manager) addConnection(clientKey string, generation uint64, connection net.Conn) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.connectionGenerations[clientKey] != generation {
		return false
	}
	if m.connections[clientKey] == nil {
		m.connections[clientKey] = make(map[net.Conn]struct{})
	}
	m.connections[clientKey][connection] = struct{}{}
	return true
}

func (m *Manager) removeConnection(clientKey string, connection net.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections[clientKey], connection)
	if len(m.connections[clientKey]) == 0 {
		delete(m.connections, clientKey)
	}
}

func (m *Manager) Connected(clusterID uint) bool {
	return m.server.HasSession(strconv.FormatUint(uint64(clusterID), 10))
}

func (m *Manager) Version(clusterID uint) string {
	clientKey := strconv.FormatUint(uint64(clusterID), 10)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.registrations[clientKey].registration.AgentVersion
}

func (m *Manager) Dialer(clusterID uint) func(context.Context, string, string) (net.Conn, error) {
	clientKey := strconv.FormatUint(uint64(clusterID), 10)
	return m.server.Dialer(clientKey)
}

func (m *Manager) Disconnect(clusterID uint) {
	clientKey := strconv.FormatUint(uint64(clusterID), 10)
	m.mu.Lock()
	m.connectionGenerations[clientKey]++
	connections := m.connections[clientKey]
	delete(m.connections, clientKey)
	delete(m.registrations, clientKey)
	m.mu.Unlock()
	for connection := range connections {
		_ = connection.Close()
	}
}
