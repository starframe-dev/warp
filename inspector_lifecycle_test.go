package warp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type countingElementPanel struct {
	calls int
}

func (*countingElementPanel) View(_, _ int) string   { return "" }
func (*countingElementPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *countingElementPanel) Elements(_, _ int) []Element {
	p.calls++
	return []Element{{Role: "status", Name: "ready"}}
}

func TestInspectorDemandAndUpdateViewSnapshotDeduplication(t *testing.T) {
	w := New()
	panel := &countingElementPanel{}
	w.SetRoot(panel)
	w.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	w.View()
	if panel.calls != 0 {
		t.Fatalf("ElementProvider calls without HTTP inspector = %d, want 0", panel.calls)
	}
	if w.elementsSnapshot != nil {
		t.Fatal("inspector-disabled Warp retained a semantic snapshot")
	}

	if err := w.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	defer w.CloseHTTP()

	w.Update(tea.KeyMsg{})
	if panel.calls != 1 {
		t.Fatalf("ElementProvider calls after Update = %d, want 1", panel.calls)
	}
	w.View()
	if panel.calls != 1 {
		t.Fatalf("same-cycle View repeated semantic traversal: calls=%d, want 1", panel.calls)
	}
	w.View()
	if panel.calls != 2 {
		t.Fatalf("later View calls=%d, want one fresh snapshot", panel.calls)
	}
}

func TestCloseHTTPClearsSnapshotAndDisablesInspectorDemand(t *testing.T) {
	w := New()
	panel := &countingElementPanel{}
	w.SetRoot(panel)
	if err := w.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	w.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if panel.calls != 1 {
		t.Fatalf("ElementProvider calls before CloseHTTP = %d, want 1", panel.calls)
	}

	if err := w.CloseHTTP(); err != nil {
		t.Fatalf("CloseHTTP failed: %v", err)
	}
	w.elementsSnapshotMu.RLock()
	snapshot := w.elementsSnapshot
	w.elementsSnapshotMu.RUnlock()
	if snapshot != nil {
		t.Fatal("CloseHTTP retained the semantic snapshot")
	}

	w.Update(tea.KeyMsg{})
	w.View()
	if panel.calls != 1 {
		t.Fatalf("ElementProvider calls after CloseHTTP = %d, want no more calls", panel.calls)
	}
}

func TestElementSnapshotCachesEncodedJSON(t *testing.T) {
	snapshot := newElementSnapshot([]Element{{Role: "button", Name: "save", Bounds: Bounds{W: 4, H: 1}}})
	first, err := snapshot.jsonBytes()
	if err != nil {
		t.Fatalf("first encode failed: %v", err)
	}
	second, err := snapshot.jsonBytes()
	if err != nil {
		t.Fatalf("second encode failed: %v", err)
	}
	if len(first) == 0 || &first[0] != &second[0] {
		t.Fatal("snapshot did not reuse its encoded JSON bytes")
	}
}

func TestHTTPServerTimeoutsAndBoundedShutdown(t *testing.T) {
	w := New()
	if err := w.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	w.mu.RLock()
	server := w.httpServer
	w.mu.RUnlock()
	if server.ReadHeaderTimeout <= 0 || server.IdleTimeout <= 0 || server.WriteTimeout <= 0 {
		t.Fatalf("HTTP timeouts not configured: header=%s idle=%s write=%s", server.ReadHeaderTimeout, server.IdleTimeout, server.WriteTimeout)
	}
	if err := w.CloseHTTP(); err != nil {
		t.Fatalf("CloseHTTP failed: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	blockedServer := &http.Server{Handler: http.HandlerFunc(func(wr http.ResponseWriter, _ *http.Request) {
		once.Do(func() { close(started) })
		<-release
		_, _ = wr.Write([]byte("done"))
	})}
	serveDone := make(chan struct{})
	go func() {
		defer close(serveDone)
		_ = blockedServer.Serve(listener)
	}()

	requestDone := make(chan error, 1)
	go func() {
		response, requestErr := (&http.Client{Timeout: time.Second}).Get("http://" + listener.Addr().String())
		if response != nil {
			_ = response.Body.Close()
		}
		requestDone <- requestErr
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("blocking HTTP handler did not start")
	}

	shutdownErr := shutdownHTTPServer(blockedServer, 20*time.Millisecond)
	if !errors.Is(shutdownErr, context.DeadlineExceeded) {
		close(release)
		t.Fatalf("bounded shutdown error=%v, want deadline exceeded", shutdownErr)
	}
	close(release)
	select {
	case <-requestDone:
	case <-time.After(time.Second):
		t.Fatal("forced close left HTTP request blocked")
	}
	select {
	case <-serveDone:
	case <-time.After(time.Second):
		t.Fatal("HTTP server did not stop after forced close")
	}
}
