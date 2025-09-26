package iutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// JSRunner manages a long-lived Chrome tab to evaluate JavaScript in a real browser environment.
type JSRunner struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

// NewJSRunner creates and warms up a new browser context.
func NewJSRunner() (*JSRunner, error) {
	// Use an ExecAllocator with an explicit browser path when possible
	execPath := findBrowserExec()
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless, // comment to show UI when debugging
	)
	if execPath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(execPath))
	}
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	// Warm up a blank page so that subsequent Evaluate calls have a document context
	if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
		cancel()
		return nil, err
	}
	return &JSRunner{ctx: ctx, cancel: cancel}, nil
}

// findBrowserExec tries to locate Chrome/Edge on common paths, especially for Windows.
func findBrowserExec() string {
	// Env overrides first
	if v := os.Getenv("CHROME_PATH"); v != "" {
		if fileExists(v) {
			return v
		}
	}
	if v := os.Getenv("EDGE_PATH"); v != "" {
		if fileExists(v) {
			return v
		}
	}

	switch runtime.GOOS {
	case "windows":
		candidates := []string{
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("LocalAppData"), "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
		}
		for _, p := range candidates {
			if fileExists(p) {
				return p
			}
		}
	case "darwin":
		candidates := []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
		for _, p := range candidates {
			if fileExists(p) {
				return p
			}
		}
	default: // linux
		candidates := []string{
			"/usr/bin/google-chrome", "/usr/bin/google-chrome-stable", "/usr/bin/chromium", "/usr/bin/chromium-browser",
		}
		for _, p := range candidates {
			if fileExists(p) {
				return p
			}
		}
	}
	return ""
}

func fileExists(p string) bool {
	if p == "" {
		return false
	}
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return true
	}
	return false
}

// Close releases the browser resources.
func (r *JSRunner) Close() {
	if r == nil || r.cancel == nil {
		return
	}
	r.cancel()
}

// EvalJSON evaluates JS and unmarshals its JSON string result into v.
// JS must return a JSON-encoded string (e.g., JSON.stringify(obj)).
func (r *JSRunner) EvalJSON(code string, v any, timeout time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.ctx, timeout)
	defer cancel()

	var raw string
	if err := chromedp.Run(ctx, chromedp.Evaluate(code, &raw)); err != nil {
		return err
	}
	return json.Unmarshal([]byte(raw), v)
}

// CallFunctionNamed lets you pass any function name and its definition, plus any number of arguments.
// funcName: the global function name to invoke (e.g., "I"),
// funcDef: the function definition string (e.g., "function I(e){...}") which will be defined before call,
// args: arguments passed to the function.
func (r *JSRunner) CallFunctionNamed(funcName string, funcDef string, args []any, timeout time.Duration) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.ctx, timeout)
	defer cancel()

	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	js := fmt.Sprintf(`(function(){
		try {
			%[1]s;
			var __args = %[2]s;
			var __fn = (typeof %[3]s === 'function') ? %[3]s : null;
			if (!__fn) { return '%[3]s is not a function'; }
			return String(__fn.apply(null, __args));
		} catch (e) { return String(e); }
	})()`, funcDef, string(argsJSON), funcName)

	var out string
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &out)); err != nil {
		return "", err
	}
	return out, nil
}

var (
	defaultRunner     *JSRunner
	defaultRunnerOnce sync.Once
	defaultRunnerErr  error
)

// DefaultJS returns a shared singleton JSRunner, initializing it lazily.
func DefaultJS() *JSRunner {
	defaultRunnerOnce.Do(func() {
		defaultRunner, defaultRunnerErr = NewJSRunner()
		if defaultRunnerErr != nil {
			panic(defaultRunnerErr)
		}
	})
	return defaultRunner
}
