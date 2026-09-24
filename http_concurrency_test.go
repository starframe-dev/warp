package warp

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestServeHTTPDefaultsToLoopback(t *testing.T) {
	t.Setenv("WARP_HTTP_PORT", "")
	warp := New()
	if err := warp.ServeHTTP(""); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	defer warp.CloseHTTP()

	host, _, err := net.SplitHostPort(warp.HTTPAddr())
	if err != nil {
		t.Fatalf("invalid HTTP address %q: %v", warp.HTTPAddr(), err)
	}
	if host != "127.0.0.1" {
		t.Fatalf("default HTTP host=%q, want 127.0.0.1", host)
	}
}

func TestServeHTTPConcurrentElementsRequests(t *testing.T) {
	warp := New()
	panel := &serialElementsPanel{}
	warp.SetRoot(panel)
	if err := warp.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	defer warp.CloseHTTP()

	client := &http.Client{Timeout: 5 * time.Second}
	const workers, requestsPerWorker = 12, 12
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for request := 0; request < requestsPerWorker; request++ {
				resp, err := client.Get(fmt.Sprintf("http://%s/elements", warp.HTTPAddr()))
				if err != nil {
					errs <- err
					return
				}
				if resp.StatusCode != http.StatusOK {
					_ = resp.Body.Close()
					errs <- fmt.Errorf("/elements status=%d", resp.StatusCode)
					return
				}
				if err := resp.Body.Close(); err != nil {
					errs <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if got := panel.maxConcurrent.Load(); got != 1 {
		t.Fatalf("concurrent ElementProvider calls=%d, want serialized calls", got)
	}
}

type serialElementsPanel struct {
	active        atomic.Int32
	maxConcurrent atomic.Int32
}

func (*serialElementsPanel) View(_, _ int) string   { return "" }
func (*serialElementsPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *serialElementsPanel) Elements(_, _ int) []Element {
	active := p.active.Add(1)
	for {
		currentMax := p.maxConcurrent.Load()
		if active <= currentMax || p.maxConcurrent.CompareAndSwap(currentMax, active) {
			break
		}
	}
	time.Sleep(time.Millisecond)
	p.active.Add(-1)
	return nil
}

type httpShutdownProbePanel struct {
	warp     *Warp
	entered  chan struct{}
	release  chan struct{}
	observed chan string
}

func (*httpShutdownProbePanel) View(_, _ int) string   { return "" }
func (*httpShutdownProbePanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *httpShutdownProbePanel) Elements(_, _ int) []Element {
	p.entered <- struct{}{}
	<-p.release
	p.observed <- p.warp.HTTPAddr()
	return []Element{}
}

func TestCloseHTTPDoesNotHoldLockWhileShuttingDown(t *testing.T) {
	warp := New()
	probe := &httpShutdownProbePanel{
		warp:     warp,
		entered:  make(chan struct{}, 1),
		release:  make(chan struct{}),
		observed: make(chan string, 1),
	}
	warp.SetRoot(probe)
	if err := warp.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	var releaseOnce sync.Once
	releaseProbe := func() { releaseOnce.Do(func() { close(probe.release) }) }
	defer releaseProbe()

	shutdownStarted := make(chan struct{})
	warp.mu.RLock()
	server := warp.httpServer
	warp.mu.RUnlock()
	server.RegisterOnShutdown(func() { close(shutdownStarted) })

	requestDone := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + warp.HTTPAddr() + "/elements")
		if err != nil {
			requestDone <- err
			return
		}
		requestDone <- resp.Body.Close()
	}()

	select {
	case <-probe.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("/elements handler did not enter the probe panel")
	}

	closeDone := make(chan error, 1)
	go func() { closeDone <- warp.CloseHTTP() }()
	select {
	case <-shutdownStarted:
		releaseProbe()
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP server shutdown did not start")
	}

	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("CloseHTTP failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CloseHTTP deadlocked while a request inspected HTTPAddr")
	}
	select {
	case <-probe.observed:
	case <-time.After(5 * time.Second):
		t.Fatal("request handler did not observe the cleared HTTP address")
	}
	select {
	case err := <-requestDone:
		if err != nil {
			t.Errorf("/elements request failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("/elements request did not finish")
	}
}
