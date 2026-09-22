package restic

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// fakeCall is one recorded Run or Start call.
type fakeCall struct {
	name        string
	args        []string
	env         []string
	hasDeadline bool
	start       bool
}

// fakeResult is a scripted reply to a Run call.
type fakeResult struct {
	stdout []byte
	stderr []byte
	err    error
	delay  time.Duration
}

// fakeRunner is the shared Runner test double. It records every call,
// replies from results keyed by restic subcommand (or by the base name of
// the binary for non-restic commands such as fusermount3), and tracks the
// peak number of concurrent Run calls.
type fakeRunner struct {
	// results maps a subcommand (for example "snapshots") or a binary
	// base name to its scripted reply.
	results map[string]fakeResult
	// reply, when set, overrides results for every Run call.
	reply func(c fakeCall) fakeResult
	// newProcess, when set, builds the process each Start returns.
	newProcess func() *fakeProcess
	// startErr is returned by every Start call.
	startErr error

	mu     sync.Mutex
	calls  []fakeCall
	procs  []*fakeProcess
	active int
	peak   int
}

var _ Runner = (*fakeRunner)(nil)

func (f *fakeRunner) Run(ctx context.Context, name string, args, env []string) ([]byte, []byte, error) {
	_, hasDeadline := ctx.Deadline()
	c := fakeCall{name: name, args: slices.Clone(args), env: slices.Clone(env), hasDeadline: hasDeadline}

	f.mu.Lock()
	f.calls = append(f.calls, c)
	f.active++
	f.peak = max(f.peak, f.active)
	f.mu.Unlock()
	defer func() {
		f.mu.Lock()
		f.active--
		f.mu.Unlock()
	}()

	var r fakeResult
	if f.reply != nil {
		r = f.reply(c)
	} else if got, ok := f.results[subcommand(args)]; ok {
		r = got
	} else {
		r = f.results[filepath.Base(name)]
	}
	if r.delay > 0 {
		select {
		case <-time.After(r.delay):
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		}
	}
	return r.stdout, r.stderr, r.err
}

func (f *fakeRunner) Start(name string, args, env []string) (Process, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, fakeCall{name: name, args: slices.Clone(args), env: slices.Clone(env), start: true})
	if f.startErr != nil {
		return nil, f.startErr
	}
	p := newFakeProcess()
	if f.newProcess != nil {
		p = f.newProcess()
	}
	f.procs = append(f.procs, p)
	return p, nil
}

// recorded returns a copy of every recorded call.
func (f *fakeRunner) recorded() []fakeCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.calls)
}

// runCalls returns the recorded Run calls only.
func (f *fakeRunner) runCalls() []fakeCall {
	var out []fakeCall
	for _, c := range f.recorded() {
		if !c.start {
			out = append(out, c)
		}
	}
	return out
}

// peakConcurrent returns the highest number of Run calls in flight at once.
func (f *fakeRunner) peakConcurrent() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.peak
}

// processes returns every process handed out by Start.
func (f *fakeRunner) processes() []*fakeProcess {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.procs)
}

// subcommand returns the restic subcommand in args, skipping the global
// flags; it returns "" when args hold no subcommand.
func subcommand(args []string) string {
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--password-file" || a == "--cache-dir":
			i++
		case strings.HasPrefix(a, "-"):
		default:
			return a
		}
	}
	return ""
}

// errFakeKilled is the Wait error of a fakeProcess ended by Kill.
var errFakeKilled = errors.New("fake process killed")

// fakeProcess is a Process whose exit the test controls. When
// exitOnInterrupt is set, Signal(os.Interrupt) makes it exit cleanly.
type fakeProcess struct {
	exitOnInterrupt bool

	once    sync.Once
	done    chan struct{}
	mu      sync.Mutex
	err     error
	signals []os.Signal
	kills   int
}

var _ Process = (*fakeProcess)(nil)

func newFakeProcess() *fakeProcess { return &fakeProcess{done: make(chan struct{})} }

// exit ends the process with err; later calls have no effect.
func (p *fakeProcess) exit(err error) {
	p.once.Do(func() {
		p.mu.Lock()
		p.err = err
		p.mu.Unlock()
		close(p.done)
	})
}

func (p *fakeProcess) Wait() error {
	<-p.done
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.err
}

func (p *fakeProcess) Signal(s os.Signal) error {
	p.mu.Lock()
	p.signals = append(p.signals, s)
	p.mu.Unlock()
	if p.exitOnInterrupt && s == os.Interrupt {
		p.exit(nil)
	}
	return nil
}

func (p *fakeProcess) Kill() error {
	p.mu.Lock()
	p.kills++
	p.mu.Unlock()
	p.exit(errFakeKilled)
	return nil
}

// signalled returns a copy of the signals received.
func (p *fakeProcess) signalled() []os.Signal {
	p.mu.Lock()
	defer p.mu.Unlock()
	return slices.Clone(p.signals)
}

// killCount returns how many times Kill was called.
func (p *fakeProcess) killCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.kills
}
