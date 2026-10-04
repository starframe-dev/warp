package integration_test

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type inspectorPanel struct {
	name string
}

func (p *inspectorPanel) View(_, _ int) string {
	p.name = "view snapshot"
	return "inspector test"
}

func (p *inspectorPanel) Update(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyMsg); ok && len(key.Runes) > 0 && key.Runes[0] == 'u' {
		p.name = "update snapshot"
	}
	return nil
}

func (p *inspectorPanel) Elements(_, _ int) []warp.Element {
	if p.name == "" {
		return []warp.Element{}
	}
	return []warp.Element{{Role: "status", Name: p.name}}
}

func inspectorRequest(t *testing.T, client *http.Client, method, url, token, origin string) (*http.Response, []byte) {
	t.Helper()
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("create %s request: %v", method, err)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("perform %s %s: %v", method, url, err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return response, body
}

func decodeInspectorElements(t *testing.T, body []byte) []warp.Element {
	t.Helper()
	var elements []warp.Element
	if err := json.Unmarshal(body, &elements); err != nil {
		t.Fatalf("decode elements %q: %v", body, err)
	}
	return elements
}

func TestHTTPInspectorIntegrationLifecycle(t *testing.T) {
	w := warp.New()
	panel := &inspectorPanel{}
	w.SetRoot(panel)
	if err := w.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("start inspector: %v", err)
	}
	defer func() { _ = w.Close() }()
	baseURL := "http://" + w.HTTPAddr()
	client := &http.Client{}

	response, body := inspectorRequest(t, client, http.MethodGet, baseURL+"/healthz", "", "")
	if response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != "text/plain" || strings.TrimSpace(string(body)) != "ok" {
		t.Fatalf("health response: status=%d content-type=%q body=%q", response.StatusCode, response.Header.Get("Content-Type"), body)
	}

	response, body = inspectorRequest(t, client, http.MethodGet, baseURL+"/elements", "", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("empty elements status=%d", response.StatusCode)
	}
	if got := decodeInspectorElements(t, body); len(got) != 0 {
		t.Fatalf("initial elements=%+v, want empty snapshot", got)
	}

	w.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	w.View()
	response, body = inspectorRequest(t, client, http.MethodGet, baseURL+"/elements", "", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("populated elements status=%d", response.StatusCode)
	}
	elements := decodeInspectorElements(t, body)
	if len(elements) != 1 || elements[0].Role != "status" || elements[0].Name != "view snapshot" {
		t.Fatalf("elements after View=%+v, want view snapshot", elements)
	}

	w.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	response, body = inspectorRequest(t, client, http.MethodGet, baseURL+"/elements", "", "")
	elements = decodeInspectorElements(t, body)
	if response.StatusCode != http.StatusOK || len(elements) != 1 || elements[0].Name != "update snapshot" {
		t.Fatalf("elements after Update: status=%d elements=%+v", response.StatusCode, elements)
	}

	w.SetRoot(&inspectorPanel{})
	response, body = inspectorRequest(t, client, http.MethodGet, baseURL+"/elements", "", "")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("cleared elements status=%d", response.StatusCode)
	}
	if got := decodeInspectorElements(t, body); len(got) != 0 {
		t.Fatalf("elements immediately after SetRoot=%+v, want empty snapshot", got)
	}

	if err := w.CloseHTTP(); err != nil {
		t.Fatalf("stop inspector: %v", err)
	}
	if w.HTTPAddr() != "" {
		t.Fatalf("HTTPAddr after shutdown=%q, want empty", w.HTTPAddr())
	}
	request, _ := http.NewRequest(http.MethodGet, baseURL+"/healthz", nil)
	if response, err := client.Do(request); err == nil {
		_ = response.Body.Close()
		t.Fatal("health endpoint remained reachable after shutdown")
	}
}

func TestHTTPInspectorIntegrationAccessControlsMethodsAndCORS(t *testing.T) {
	w := warp.New()
	w.SetRoot(&inspectorPanel{name: "authorized"})
	const allowedOrigin = "https://inspector.example"
	if err := w.ServeHTTPWithOptions("127.0.0.1:0", warp.InspectorOptions{
		AllowedOrigin: allowedOrigin,
		BearerToken:   "integration-secret",
	}); err != nil {
		t.Fatalf("start protected inspector: %v", err)
	}
	defer func() { _ = w.Close() }()
	baseURL := "http://" + w.HTTPAddr()
	client := &http.Client{}
	w.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	w.View()

	response, _ := inspectorRequest(t, client, http.MethodGet, baseURL+"/elements", "", allowedOrigin)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d, want 401", response.StatusCode)
	}
	if got := response.Header.Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Fatalf("allowed CORS origin=%q, want %q", got, allowedOrigin)
	}

	response, body := inspectorRequest(t, client, http.MethodGet, baseURL+"/elements", "integration-secret", allowedOrigin)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("authorized status=%d, want 200 (body %q)", response.StatusCode, body)
	}
	if got := response.Header.Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Fatalf("authorized CORS origin=%q, want %q", got, allowedOrigin)
	}
	if elements := decodeInspectorElements(t, body); len(elements) != 1 || elements[0].Name != "view snapshot" {
		t.Fatalf("authorized elements=%+v", elements)
	}

	request, err := http.NewRequest(http.MethodOptions, baseURL+"/elements", nil)
	if err != nil {
		t.Fatalf("create preflight request: %v", err)
	}
	request.Header.Set("Origin", allowedOrigin)
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	response, err = client.Do(request)
	if err != nil {
		t.Fatalf("send preflight request: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("OPTIONS status=%d, want 204", response.StatusCode)
	}
	if got := response.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodGet) {
		t.Fatalf("preflight allow-methods=%q, want GET", got)
	}

	response, _ = inspectorRequest(t, client, http.MethodPost, baseURL+"/elements", "integration-secret", "")
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST status=%d, want 405", response.StatusCode)
	}
	response, _ = inspectorRequest(t, client, http.MethodPut, baseURL+"/elements", "integration-secret", "")
	if response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("PUT status=%d, want 405", response.StatusCode)
	}

	response, _ = inspectorRequest(t, client, http.MethodGet, baseURL+"/elements", "integration-secret", "https://other.example")
	if got := response.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unconfigured origin got CORS header %q", got)
	}
}
