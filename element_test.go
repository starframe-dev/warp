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
func (testElementPanel) Update(msg tea.Msg) tea.Cmd    { return nil }
func (p testElementPanel) Elements(width, height int) []Element {
	return p.elems
}

type testNonProviderPanel struct{}

func (testNonProviderPanel) View(width, height int) string { return "" }
func (testNonProviderPanel) Update(msg tea.Msg) tea.Cmd    { return nil }

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

func TestBoundsCenterOdd(t *testing.T) {
	b := Bounds{X: 0, Y: 0, W: 5, H: 3}
	x, y := b.Center()
	if x != 2 || y != 1 {
		t.Fatalf("center = (%d,%d), want (2,1)", x, y)
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
	if got := collectElements(testNonProviderPanel{}, 80, 24); got != nil {
		t.Fatalf("collectElements(nonProvider) = %v, want nil", got)
	}
}

func TestElementProviderFunc(t *testing.T) {
	fn := ElementProviderFunc(func(width, height int) []Element {
		return []Element{
			{Role: "label", Name: "ok", Bounds: Bounds{X: 0, Y: 0, W: width, H: height}},
		}
	})

	elems := fn.Elements(100, 50)
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Role != "label" || elems[0].Bounds.W != 100 || elems[0].Bounds.H != 50 {
		t.Fatalf("unexpected element: %+v", elems[0])
	}

	var ep ElementProvider = fn
	elems = ep.Elements(10, 20)
	if elems[0].Bounds.W != 10 || elems[0].Bounds.H != 20 {
		t.Fatalf("unexpected element via interface: %+v", elems[0])
	}
}

func TestFindElement(t *testing.T) {
	elems := []Element{
		{
			Role:   "button",
			Name:   "Save",
			Action: "save",
			Bounds: Bounds{X: 0, Y: 0, W: 4, H: 1},
		},
		{
			Role: "group",
			Name: "Toolbar",
			Bounds: Bounds{X: 0, Y: 1, W: 10, H: 1},
			Children: []Element{
				{
					Role:   "button",
					Name:   "Open",
					Action: "open",
					Bounds: Bounds{X: 0, Y: 1, W: 4, H: 1},
				},
			},
		},
	}

	cases := []struct {
		role, name, action string
		wantName           string
		wantOK             bool
	}{
		{"button", "Save", "save", "Save", true},
		{"button", "Open", "", "Open", true},
		{"group", "", "", "Toolbar", true},
		{"button", "", "", "Save", true},
		{"", "", "missing", "", false},
		{"", "", "", "Save", true},
	}

	for _, tc := range cases {
		got, ok := FindElement(elems, tc.role, tc.name, tc.action)
		if ok != tc.wantOK {
			t.Fatalf("FindElement(%q,%q,%q) ok=%v, want %v", tc.role, tc.name, tc.action, ok, tc.wantOK)
		}
		if ok && got.Name != tc.wantName {
			t.Fatalf("FindElement(%q,%q,%q) name=%q, want %q", tc.role, tc.name, tc.action, got.Name, tc.wantName)
		}
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
