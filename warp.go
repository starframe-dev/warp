package warp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// Warp is the root Bubbletea model. It holds a root Panel and forwards
// all messages to it without interception.
type Warp struct {
	root   Panel
	width  int
	height int

	httpServer *http.Server
	httpAddr   string
	httpMu     sync.Mutex
}

// New creates a new Warp with a TabGroup root (one default tab).
func New() *Warp {
	tg := NewTabGroup(TabTop)
	return &Warp{root: tg}
}

// SetRoot replaces the root panel. Use this to install custom layouts
// (splits, flex, nested tab groups, etc.).
func (w *Warp) SetRoot(panel Panel) {
	w.root = panel
}

// Root returns the current root panel.
func (w *Warp) Root() Panel {
	return w.root
}

// convenience delegates to the root panel if it is a *TabGroup

func (w *Warp) tabGroup() *TabGroup {
	tg, _ := w.root.(*TabGroup)
	return tg
}

// Width returns the last known width.
func (w *Warp) Width() int { return w.width }

// Height returns the last known height.
func (w *Warp) Height() int { return w.height }

// NewTab delegates to the root TabGroup (no-op if root is not a TabGroup).
func (w *Warp) NewTab(name string) *Tab {
	if tg := w.tabGroup(); tg != nil {
		return tg.NewTab(name)
	}
	return nil
}

// ActiveTab delegates to the root TabGroup.
func (w *Warp) ActiveTab() *Tab {
	if tg := w.tabGroup(); tg != nil {
		return tg.ActiveTab()
	}
	return nil
}

// SetTabPosition delegates to the root TabGroup.
func (w *Warp) SetTabPosition(pos TabPosition) {
	if tg := w.tabGroup(); tg != nil {
		tg.tabPosition = pos
	}
}

// NextTab delegates to the root TabGroup.
func (w *Warp) NextTab() {
	if tg := w.tabGroup(); tg != nil {
		tg.NextTab()
	}
}

// PrevTab delegates to the root TabGroup.
func (w *Warp) PrevTab() {
	if tg := w.tabGroup(); tg != nil {
		tg.PrevTab()
	}
}

// Init is the Bubbletea initialization.
func (w *Warp) Init() tea.Cmd {
	return nil
}

// Update forwards all messages to the root panel without interception.
func (w *Warp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
	}

	if w.root != nil {
		return w, w.root.Update(msg)
	}
	return w, nil
}

// View renders the root panel.
func (w *Warp) View() string {
	if w.root == nil {
		return ""
	}
	return w.root.View(w.width, w.height)
}

// AsPanel returns a Panel adapter for this Warp, enabling nested warps.
func (w *Warp) AsPanel() Panel {
	return &warpPanel{warp: w}
}

// warpPanel adapts Warp to the Panel interface.
type warpPanel struct {
	warp *Warp
}

func (wp *warpPanel) View(width, height int) string {
	wp.warp.width = width
	wp.warp.height = height
	return wp.warp.View()
}

func (wp *warpPanel) Update(msg tea.Msg) tea.Cmd {
	_, cmd := wp.warp.Update(msg)
	return cmd
}

// Run starts the Bubbletea program.
func (w *Warp) Run() error {
	p := tea.NewProgram(
		w,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}

// ServeHTTP starts an HTTP server exposing the element tree at /elements.
func (w *Warp) ServeHTTP(addr string) error {
	w.httpMu.Lock()
	defer w.httpMu.Unlock()

	if w.httpServer != nil {
		return nil
	}

	if addr == "" {
		if p := os.Getenv("WARP_HTTP_PORT"); p != "" {
			addr = ":" + p
		} else {
			addr = ":0"
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/elements", w.handleElements)
	mux.HandleFunc("/healthz", func(wr http.ResponseWriter, _ *http.Request) {
		wr.Header().Set("Content-Type", "text/plain")
		wr.WriteHeader(http.StatusOK)
		_, _ = wr.Write([]byte("ok"))
	})

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("warp http listen: %w", err)
	}
	w.httpAddr = ln.Addr().String()

	w.httpServer = &http.Server{Handler: mux}
	go func() {
		_ = w.httpServer.Serve(ln)
	}()
	return nil
}

// CloseHTTP stops the HTTP server.
func (w *Warp) CloseHTTP() error {
	w.httpMu.Lock()
	defer w.httpMu.Unlock()

	if w.httpServer == nil {
		return nil
	}
	err := w.httpServer.Shutdown(context.Background())
	w.httpServer = nil
	w.httpAddr = ""
	return err
}

// HTTPAddr returns the current HTTP listening address, or empty if not serving.
func (w *Warp) HTTPAddr() string {
	w.httpMu.Lock()
	defer w.httpMu.Unlock()
	return w.httpAddr
}

func (w *Warp) handleElements(wr http.ResponseWriter, _ *http.Request) {
	w.httpMu.Lock()
	width, height := w.width, w.height
	root := w.root
	w.httpMu.Unlock()

	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	var elems []Element
	if root != nil {
		elems = collectElements(root, width, height)
	}
	if elems == nil {
		elems = []Element{}
	}

	wr.Header().Set("Content-Type", "application/json")
	wr.Header().Set("Access-Control-Allow-Origin", "*")
	wr.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(wr).Encode(elems)
}

func parsePort(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return port
}