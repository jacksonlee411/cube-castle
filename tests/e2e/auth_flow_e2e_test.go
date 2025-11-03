package e2e

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// waits for an HTTP GET to return 200 within timeout
func waitHealthy(t *testing.T, endpoint string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(endpoint)
		if err == nil && resp.StatusCode == 200 {
			_ = resp.Body.Close()
			return
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for healthy: %s", endpoint)
}

func startService(t *testing.T, dir string, env []string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	// pipe output to help debugging if test fails
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start %v: %v", args, err)
	}
	return cmd
}

func killIfRunning(_ *testing.T, cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// tests/e2e -> repo root
	return filepath.Clean(filepath.Join(wd, "../.."))
}

func TestAuthFlow_RealHTTP_RS256_JWKS_and_TenantChecks(t *testing.T) {
	if os.Getenv("E2E_RUN") != "1" {
		t.Skip("set E2E_RUN=1 to enable end-to-end HTTP test")
	}
	// Require Postgres unless DATABASE_URL points elsewhere and port is open
	if os.Getenv("DATABASE_URL") == "" && !portOpen("localhost:5432") {
		t.Skip("PostgreSQL not available on localhost:5432; run `make docker-up` and re-run with E2E_RUN=1")
	}

	root := projectRoot(t)

	// Ensure dev key exists for RS256 mint
	_ = os.MkdirAll(filepath.Join(root, "secrets"), 0o755)
	pkPath := filepath.Join(root, "secrets/dev-jwt-private.pem")
	if _, err := os.Stat(pkPath); os.IsNotExist(err) {
		// generate RSA private key via Go
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("generate rsa: %v", err)
		}
		b := x509.MarshalPKCS1PrivateKey(key)
		pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b})
		if err := os.WriteFile(pkPath, pemBytes, 0o600); err != nil {
			t.Fatalf("write pem: %v", err)
		}
		pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})
		_ = os.WriteFile(filepath.Join(root, "secrets/dev-jwt-public.pem"), pubPem, 0o644)
	}

	// Start command service (RS256 mint + OIDC simulate)
	cmdEnv := []string{
		"JWT_MINT_ALG=RS256",
		"JWT_PRIVATE_KEY_PATH=" + pkPath,
		"JWT_KEY_ID=bff-key-1",
		"OIDC_SIMULATE=true",
	}
	cmdSrv := startService(t, root, cmdEnv, "go", "run", "./cmd/hrms-server/command/main.go")
	defer killIfRunning(t, cmdSrv)

	// Start query service (JWKS verify)
	gqlEnv := []string{
		"JWT_ALG=RS256",
		"JWT_JWKS_URL=http://localhost:9090/.well-known/jwks.json",
	}
	gqlSrv := startService(t, root, gqlEnv, "go", "run", "./cmd/hrms-server/query/main.go")
	defer killIfRunning(t, gqlSrv)

	waitHealthy(t, "http://localhost:9090/health", 20*time.Second)
	waitHealthy(t, "http://localhost:8090/health", 20*time.Second)

	// Cookie jar client
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

	// Trigger login (OIDC simulate path will set sid/csrf and redirect)
	resp, err := client.Get("http://localhost:9090/auth/login?redirect=%2F")
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	_ = resp.Body.Close()

	// Get session
	resp, err = client.Get("http://localhost:9090/auth/session")
	if err != nil {
		t.Fatalf("session failed: %v", err)
	}
	var sessBody map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&sessBody); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	_ = resp.Body.Close()
	data := sessBody["data"].(map[string]any)
	accessToken := data["accessToken"].(string)
	tenantID := data["tenantId"].(string)

	// GraphQL OK with matching tenant
	reqBody := strings.NewReader(`{"query":"query { __typename }"}`)
	req, _ := http.NewRequest("POST", "http://localhost:8090/graphql", reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Tenant-ID", tenantID)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("graphql request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("graphql expected 200, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// GraphQL 403 with mismatched tenant
	reqBody2 := strings.NewReader(`{"query":"query { __typename }"}`)
	req2, _ := http.NewRequest("POST", "http://localhost:8090/graphql", reqBody2)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+accessToken)
	req2.Header.Set("X-Tenant-ID", "mismatch-tenant")
	resp, err = client.Do(req2)
	if err != nil {
		t.Fatalf("graphql mismatch failed: %v", err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("graphql expected 403 on tenant mismatch, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Refresh token
	// Extract csrf from cookie jar
	u, _ := url.Parse("http://localhost:9090/")
	var csrf string
	for _, c := range jar.Cookies(u) {
		if c.Name == "csrf" {
			csrf = c.Value
			break
		}
	}
	if csrf == "" {
		t.Fatalf("missing csrf cookie")
	}
	req3, _ := http.NewRequest("POST", "http://localhost:9090/auth/refresh", nil)
	req3.Header.Set("X-CSRF-Token", csrf)
	resp, err = client.Do(req3)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 refresh, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func portOpen(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}
