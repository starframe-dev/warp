package warp

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testElementPanel struct {
	elems []Element
}

func (testElementPanel) View(width, height int) string { return "" }
func (testElementPanel) Update(msg tea.Msg) tea.Cmd { return nil }
func (p testElementPanel) Elements(width, height int) []Element {
	return p.elems
}

func TestHTTPElementsEndpoint(t *testing.T) {
	w := New()
	w.width = 80
	w.height = 24
	w.SetRoot(testElementPanel{elems: []Element{
		{Role: "button", Name: "+Folder", Bounds: Bounds{X: 0, Y: 0, W: 8, H: 1}},
	}})

	if err := w.ServeHTTP(":0"); err != nil {
		t.Fatalf("ServeHTTP: %v", err)
	}
	defer w.CloseHTTP()

	addr := w.HTTPAddr()
	resp, err := http.Get("http://" + addr + "/elements")
	if err != nil {
		t.Fatalf("get elements: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}

	var elems []Element
	if err := json.NewDecoder(resp.Body).Decode(&elems); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Role != "button" || elems[0].Name != "+Folder" {
		t.Fatalf("unexpected element: %+v", elems[0])
	}
}

func TestBoundsCenter(t *testing.T) {
	b := Bounds{X: 10, Y: 20, W: 6, H: 4}
	x, y := b.Center()
	if x != 13 || y != 22 {
		t.Fatalf("center = (%d,%d), want (13,22)", x, y)
	}
}

func TestParsePort(t *testing.T) {
	if got := parsePort("127.0.0.1:8080"); got != "8080" {
		t.Fatalf("parsePort(127.0.0.1:8080) = %q, want 8080", got)
	}
	if got := parsePort("[::]:9000"); got != "9000" {
		t.Fatalf("parsePort([::]:9000) = %q, want 9000", got)
	}
	if got := parsePort("bad"); got != "" {
		t.Fatalf("parsePort(bad) = %q, want empty", got)
	}
}

func TestHealthzEndpoint(t *testing.T) {
	w := New()
	w.width = 80
	w.height = 24
	if err := w.ServeHTTP(":0"); err != nil {
		t.Fatalf("ServeHTTP: %v", err)
	}
	defer w.CloseHTTP()

	resp, err := http.Get("http://" + w.HTTPAddr() + "/healthz")
	if err != nil {
		t.Fatalf("get healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestCollectElementsNilSafe(t *testing.T) {
	if got := collectElements(nil, 80, 24); got != nil {
		t.Fatalf("collectElements(nil) = %v, want nil", got)
	}
	if got := collectElements(testElementPanel{}, 80, 24); len(got) != 0 {
		t.Fatalf("collectElements(empty) = %v, want empty", got)
	}
}

func TestElementsDefaultsTo80x24(t *testing.T) {
	w := New()
	w.SetRoot(testElementPanel{})
	if err := w.ServeHTTP(":0"); err != nil {
		t.Fatalf("ServeHTTP: %v", err)
	}
	defer w.CloseHTTP()

	resp, err := http.Get("http://" + w.HTTPAddr() + "/elements")
	if err != nil {
		t.Fatalf("get elements: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}
}

func TestHTTPAddrEmptyWhenNotServing(t *testing.T) {
	w := New()
	if got := w.HTTPAddr(); got != "" {
		t.Fatalf("HTTPAddr() = %q, want empty", got)
	}
}

func TestHTTPCORSEnabled(t *testing.T) {
	w := New()
	w.width = 80
	w.height = 24
	w.SetRoot(testElementPanel{})
	if err := w.ServeHTTP(":0"); err != nil {
		t.Fatalf("ServeHTTP: %v", err)
	}
	defer w.CloseHTTP()

	resp, err := http.Get("http://" + w.HTTPAddr() + "/elements")
	if err != nil {
		t.Fatalf("get elements: %v", err)
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Access-Control-Allow-Origin"), "*") {
		t.Fatalf("expected CORS header, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}
