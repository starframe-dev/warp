package warp

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
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

func TestServeHTTPConcurrentElementsRequestsUseSnapshot(t *testing.T) {
	warp := New()
	panel := &serialElementsPanel{}
	warp.SetRoot(panel)
	warp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
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
		t.Fatalf("concurrent ElementProvider calls=%d, want serialized UI-thread calls", got)
	}
	if got := panel.calls.Load(); got != 1 {
		t.Fatalf("ElementProvider calls after HTTP requests=%d, want 1 UI snapshot", got)
	}
}

type serialElementsPanel struct {
	active        atomic.Int32
	maxConcurrent atomic.Int32
	calls         atomic.Int32
}

func (*serialElementsPanel) View(_, _ int) string   { return "" }
func (*serialElementsPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *serialElementsPanel) Elements(_, _ int) []Element {
	p.calls.Add(1)
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

type mutableSnapshotPanel struct {
	first  int
	second int
}

func (p *mutableSnapshotPanel) View(_, _ int) string {
	return fmt.Sprintf("%d:%d", p.first, p.second)
}

func (p *mutableSnapshotPanel) Update(tea.Msg) tea.Cmd {
	p.first++
	runtime.Gosched()
	p.second = p.first
	return nil
}

func (p *mutableSnapshotPanel) Elements(_, _ int) []Element {
	first := p.first
	runtime.Gosched()
	second := p.second
	return []Element{{Role: "state", Name: fmt.Sprintf("%d:%d", first, second)}}
}

type snapshotTick struct{}

func TestElementsSnapshotAvoidsConcurrentLiveUITreeTraversal(t *testing.T) {
	warp := New()
	group := NewTabGroup(TabTop)
	firstPanel := &mutableSnapshotPanel{}
	group.ActiveTab().SetRootPanel(firstPanel)
	secondPanel := &mutableSnapshotPanel{}
	group.NewTab("second").SetRootPanel(secondPanel)
	warp.SetRoot(group)
	warp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	warp.View()

	if err := warp.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}
	defer warp.CloseHTTP()

	const uiIterations, workers, requestsPerWorker = 1200, 6, 100
	start := make(chan struct{})
	errs := make(chan error, workers+1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < uiIterations; i++ {
			warp.Update(snapshotTick{})
			if i%2 == 0 {
				group.NextTab()
			} else {
				group.PrevTab()
			}
			warp.View()
		}
	}()

	client := &http.Client{Timeout: 5 * time.Second}
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for request := 0; request < requestsPerWorker; request++ {
				resp, err := client.Get("http://" + warp.HTTPAddr() + "/elements")
				if err != nil {
					errs <- err
					return
				}
				var elems []Element
				decodeErr := json.NewDecoder(resp.Body).Decode(&elems)
				closeErr := resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					errs <- fmt.Errorf("/elements status=%d", resp.StatusCode)
					return
				}
				if decodeErr != nil {
					errs <- decodeErr
					return
				}
				if closeErr != nil {
					errs <- closeErr
					return
				}
				if len(elems) != 1 {
					errs <- fmt.Errorf("snapshot has %d elements, want 1", len(elems))
					return
				}
				parts := strings.Split(elems[0].Name, ":")
				if len(parts) != 2 || parts[0] != parts[1] {
					errs <- fmt.Errorf("incoherent element snapshot: %q", elems[0].Name)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestCloseHTTPDoesNotHoldLockWhileShuttingDown(t *testing.T) {
	warp := New()
	if err := warp.ServeHTTP("127.0.0.1:0"); err != nil {
		t.Fatalf("ServeHTTP failed: %v", err)
	}

	shutdownAddress := make(chan string, 1)
	warp.mu.RLock()
	server := warp.httpServer
	warp.mu.RUnlock()
	server.RegisterOnShutdown(func() { shutdownAddress <- warp.HTTPAddr() })

	resp, err := http.Get("http://" + warp.HTTPAddr() + "/elements")
	if err != nil {
		t.Fatalf("/elements request failed: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close /elements response: %v", err)
	}

	closeDone := make(chan error, 1)
	go func() { closeDone <- warp.CloseHTTP() }()
	select {
	case address := <-shutdownAddress:
		if address != "" {
			t.Errorf("shutdown callback observed HTTP address %q, want empty", address)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown callback could not read HTTPAddr")
	}
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("CloseHTTP failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CloseHTTP did not finish")
	}
}

func TestServeHTTPRequestsAndCloseCanRepeatDuringInspection(t *testing.T) {
	warp := New()
	warp.SetRoot(testElementPanel{elems: []Element{{Role: "status", Name: "ready"}}})
	warp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	client := &http.Client{Timeout: 2 * time.Second}
	var successes atomic.Int32

	for cycle := 0; cycle < 4; cycle++ {
		if err := warp.ServeHTTP("127.0.0.1:0"); err != nil {
			t.Fatalf("ServeHTTP cycle %d failed: %v", cycle, err)
		}
		addr := warp.HTTPAddr()
		stop := atomic.Bool{}
		firstResponse := make(chan struct{})
		requestDone := make(chan struct{})
		go func() {
			defer close(requestDone)
			var first sync.Once
			for !stop.Load() {
				resp, err := client.Get("http://" + addr + "/elements")
				if err != nil {
					return
				}
				if resp.StatusCode != http.StatusOK {
					_ = resp.Body.Close()
					return
				}
				_ = resp.Body.Close()
				successes.Add(1)
				first.Do(func() { close(firstResponse) })
				time.Sleep(time.Millisecond)
			}
		}()

		select {
		case <-firstResponse:
		case <-time.After(5 * time.Second):
			stop.Store(true)
			<-requestDone
			t.Fatalf("cycle %d did not inspect /elements", cycle)
		}

		serveDone := make(chan struct{})
		go func() {
			defer close(serveDone)
			for i := 0; i < 8; i++ {
				_ = warp.ServeHTTP("127.0.0.1:0")
				runtime.Gosched()
			}
		}()
		if err := warp.CloseHTTP(); err != nil {
			t.Errorf("CloseHTTP cycle %d failed: %v", cycle, err)
		}
		<-serveDone
		stop.Store(true)
		select {
		case <-requestDone:
		case <-time.After(5 * time.Second):
			t.Fatalf("cycle %d request worker did not stop", cycle)
		}
		if err := warp.CloseHTTP(); err != nil {
			t.Errorf("final CloseHTTP cycle %d failed: %v", cycle, err)
		}
	}

	if got := successes.Load(); got == 0 {
		t.Fatal("no /elements requests succeeded during repeated server lifecycle")
	}
}
