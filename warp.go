package warp

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Warp is the root Bubbletea model. It holds a root Panel and forwards
// all messages to it without interception.
type Warp struct {
	root                Panel
	rootRevision        uint64
	stateRevision       uint64
	snapshotRevision    uint64
	viewSnapshotPending bool
	width               int
	height              int
	ownership           *panelOwnership
	ownsDomainRoot      bool
	setRootMu           sync.Mutex

	httpServer         *http.Server
	httpAddr           string
	httpClosing        bool
	inspectorEnabled   bool
	elementsSnapshot   *elementSnapshot
	elementsSnapshotMu sync.RWMutex
	mu                 sync.RWMutex
}

// New creates a new Warp with a TabGroup root (one default tab).
func New() *Warp {
	tg := NewTabGroup(TabTop)
	ownership := newPanelOwnership(tg)
	w := &Warp{
		root:           tg,
		ownership:      ownership,
		ownsDomainRoot: true,
	}
	attachPanelOwnership(tg, ownership)
	return w
}

// SetRoot replaces the root panel. Use this to install custom layouts
// (splits, flex, nested tab groups, etc.).
func (w *Warp) SetRoot(panel Panel) {
	w.setRootMu.Lock()
	w.mu.Lock()
	oldRoot := w.root
	ownership := w.ownership
	ownsDomainRoot := w.ownsDomainRoot
	if ownership == nil {
		ownership = newPanelOwnership(nil)
		w.ownership = ownership
		w.ownsDomainRoot = true
		ownsDomainRoot = true
	}
	w.mu.Unlock()

	candidates := collectPanelInstances(oldRoot)
	attachPanelOwnership(panel, ownership)

	w.mu.Lock()
	w.root = panel
	w.rootRevision++
	w.stateRevision++
	w.viewSnapshotPending = false
	w.mu.Unlock()
	if ownsDomainRoot {
		ownership.setRoot(panel)
	}
	w.setRootMu.Unlock()

	ownership.unmountRemoved(candidates)
}

// Root returns the current root panel.
func (w *Warp) Root() Panel {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.root
}

// convenience delegates to the root panel if it is a *TabGroup

func (w *Warp) tabGroup() *TabGroup {
	tg, _ := w.Root().(*TabGroup)
	return tg
}

// Width returns the last known width.
func (w *Warp) Width() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.width
}

// Height returns the last known height.
func (w *Warp) Height() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.height
}

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
	w.mu.Lock()
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		w.width = size.Width
		w.height = size.Height
	}
	w.stateRevision++
	root, width, height := w.root, w.width, w.height
	rootRevision, stateRevision := w.rootRevision, w.stateRevision
	inspectorEnabled := w.inspectorEnabled
	w.mu.Unlock()

	var cmd tea.Cmd
	if !isNilPanel(root) {
		cmd = root.Update(msg)
	}
	if inspectorEnabled && w.refreshElementsSnapshot(root, width, height, rootRevision, stateRevision) {
		w.mu.Lock()
		if w.inspectorEnabled && w.rootRevision == rootRevision && w.stateRevision == stateRevision {
			w.viewSnapshotPending = true
		}
		w.mu.Unlock()
	}
	return w, cmd
}

// View renders the root panel.
func (w *Warp) View() string {
	w.mu.RLock()
	root, width, height := w.root, w.width, w.height
	rootRevision, stateRevision := w.rootRevision, w.stateRevision
	inspectorEnabled := w.inspectorEnabled
	w.mu.RUnlock()
	if isNilPanel(root) {
		w.refreshSnapshotAfterView(nil, width, height, rootRevision, stateRevision, inspectorEnabled)
		return ""
	}
	if width == 0 || height == 0 {
		w.refreshSnapshotAfterView(root, width, height, rootRevision, stateRevision, inspectorEnabled)
		return "Loading..."
	}

	view := root.View(width, height)
	w.refreshSnapshotAfterView(root, width, height, rootRevision, stateRevision, inspectorEnabled)
	return view
}

func (w *Warp) refreshSnapshotAfterView(root Panel, width, height int, rootRevision, stateRevision uint64, inspectorEnabled bool) {
	if !inspectorEnabled {
		return
	}
	w.refreshElementsSnapshot(root, width, height, rootRevision, stateRevision)
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
	wp.warp.mu.Lock()
	wp.warp.width = width
	wp.warp.height = height
	wp.warp.mu.Unlock()
	return wp.warp.View()
}

func (wp *warpPanel) Update(msg tea.Msg) tea.Cmd {
	_, cmd := wp.warp.Update(msg)
	return cmd
}

// Close releases resources owned by a root Warp and stops its HTTP inspector.
// Embedded Warps participate in their outer ownership domain and therefore do
// not unmount their root when Close is called.
func (w *Warp) Close() error {
	httpErr := w.CloseHTTP()
	w.mu.RLock()
	ownsDomainRoot := w.ownsDomainRoot
	w.mu.RUnlock()
	if ownsDomainRoot {
		w.SetRoot(nil)
	}
	return httpErr
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

const (
	httpReadHeaderTimeout = 5 * time.Second
	httpIdleTimeout       = 60 * time.Second
	httpWriteTimeout      = 60 * time.Second
	httpShutdownTimeout   = 5 * time.Second
)

// InspectorOptions configures optional HTTP inspector access controls.
// Cross-origin access is disabled by default. Set AllowedOrigin to an exact
// origin (or "*" explicitly) to opt in. BearerToken protects /elements when set.
type InspectorOptions struct {
	AllowedOrigin string
	BearerToken   string
}

// ServeHTTP starts an HTTP server exposing the element tree at /elements.
// It uses secure defaults: no cross-origin access and no authentication token.
func (w *Warp) ServeHTTP(addr string) error {
	return w.ServeHTTPWithOptions(addr, InspectorOptions{})
}

// ServeHTTPWithOptions starts the inspector with explicit access controls.
func (w *Warp) ServeHTTPWithOptions(addr string, options InspectorOptions) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.httpServer != nil || w.httpClosing {
		return nil
	}

	if addr == "" {
		port := os.Getenv("WARP_HTTP_PORT")
		if port == "" {
			port = "0"
		}
		addr = net.JoinHostPort("127.0.0.1", port)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/elements", func(wr http.ResponseWriter, request *http.Request) {
		w.handleElementsRequest(wr, request, options)
	})
	mux.HandleFunc("/healthz", func(wr http.ResponseWriter, _ *http.Request) {
		wr.Header().Set("Content-Type", "text/plain")
		wr.WriteHeader(http.StatusOK)
		_, _ = wr.Write([]byte("ok"))
	})

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("warp http listen: %w", err)
	}
	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		IdleTimeout:       httpIdleTimeout,
		WriteTimeout:      httpWriteTimeout,
	}
	w.httpAddr = ln.Addr().String()
	w.httpServer = server
	w.inspectorEnabled = true
	go w.serveHTTP(server, ln)
	return nil
}

func (w *Warp) serveHTTP(server *http.Server, listener net.Listener) {
	_ = server.Serve(listener)

	w.mu.Lock()
	if w.httpServer != server || w.httpClosing {
		w.mu.Unlock()
		return
	}
	w.httpServer = nil
	w.httpAddr = ""
	w.inspectorEnabled = false
	w.viewSnapshotPending = false
	w.elementsSnapshotMu.Lock()
	w.elementsSnapshot = nil
	w.elementsSnapshotMu.Unlock()
	w.mu.Unlock()
}

// CloseHTTP stops the HTTP server.
func (w *Warp) CloseHTTP() error {
	w.mu.Lock()
	server := w.httpServer
	if server == nil {
		w.mu.Unlock()
		return nil
	}
	w.httpServer = nil
	w.httpAddr = ""
	w.httpClosing = true
	w.inspectorEnabled = false
	w.viewSnapshotPending = false
	w.elementsSnapshotMu.Lock()
	w.elementsSnapshot = nil
	w.elementsSnapshotMu.Unlock()
	w.mu.Unlock()

	shutdownErr := shutdownHTTPServer(server, httpShutdownTimeout)

	w.mu.Lock()
	w.httpClosing = false
	w.mu.Unlock()
	return shutdownErr
}

func shutdownHTTPServer(server *http.Server, timeout time.Duration) error {
	if server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return errors.Join(err, server.Close())
	}
	return nil
}

// HTTPAddr returns the current HTTP listening address, or empty if not serving.
func (w *Warp) HTTPAddr() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.httpAddr
}

func (w *Warp) handleElementsRequest(wr http.ResponseWriter, request *http.Request, options InspectorOptions) {
	origin := request.Header.Get("Origin")
	if options.AllowedOrigin != "" && origin != "" &&
		(options.AllowedOrigin == "*" || options.AllowedOrigin == origin) {
		allowedOrigin := origin
		if options.AllowedOrigin == "*" {
			allowedOrigin = "*"
		}
		wr.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		wr.Header().Add("Vary", "Origin")
	}
	if request.Method == http.MethodOptions {
		wr.Header().Set("Access-Control-Allow-Methods", http.MethodGet)
		wr.Header().Set("Access-Control-Allow-Headers", "Authorization")
		wr.WriteHeader(http.StatusNoContent)
		return
	}
	if request.Method != http.MethodGet {
		http.Error(wr, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !inspectorBearerAuthorized(request.Header.Get("Authorization"), options.BearerToken) {
		wr.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(wr, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.handleElements(wr, request)
}

func inspectorBearerAuthorized(header, token string) bool {
	if token == "" {
		return true
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	candidate := strings.TrimPrefix(header, prefix)
	if len(candidate) != len(token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(candidate), []byte(token)) == 1
}

func (w *Warp) handleElements(wr http.ResponseWriter, _ *http.Request) {
	w.elementsSnapshotMu.RLock()
	snapshot := w.elementsSnapshot
	w.elementsSnapshotMu.RUnlock()

	encoded, err := snapshot.jsonBytes()
	if err != nil {
		http.Error(wr, "failed to encode element snapshot", http.StatusInternalServerError)
		return
	}
	wr.Header().Set("Content-Type", "application/json")
	wr.WriteHeader(http.StatusOK)
	_, _ = wr.Write(encoded)
}

func (w *Warp) refreshElementsSnapshot(root Panel, width, height int, rootRevision, stateRevision uint64) bool {
	w.mu.RLock()
	enabled := w.inspectorEnabled && w.rootRevision == rootRevision && w.stateRevision == stateRevision
	w.mu.RUnlock()
	if !enabled {
		return false
	}

	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	var elems []Element
	if !isNilPanel(root) {
		elems = cloneElements(collectElements(root, width, height))
	}
	snapshot := newElementSnapshot(elems)

	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.inspectorEnabled || rootRevision != w.rootRevision || stateRevision != w.stateRevision {
		return false
	}
	w.elementsSnapshotMu.Lock()
	w.elementsSnapshot = snapshot
	w.elementsSnapshotMu.Unlock()
	w.snapshotRevision = stateRevision
	w.viewSnapshotPending = false
	return true
}

func cloneElements(elems []Element) []Element {
	if elems == nil {
		return nil
	}
	cloned := make([]Element, len(elems))
	for i, elem := range elems {
		cloned[i] = elem
		cloned[i].Children = cloneElements(elem.Children)
	}
	return cloned
}

func parsePort(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return port
}
