//go:build tdjson

// Command tdlib-mcp is an MCP server over stdio exposing a live TDLib
// session. It authorizes like the login example (same env vars and
// /data/code.txt poller) and starts serving MCP frames immediately, before
// authorization completes, so tdlib_auth_status / tdlib_submit_code work
// during login.
//
// Environment (see examples/login):
//
//	TDLIB_API_ID, TDLIB_API_HASH  required, from https://my.telegram.org
//	TDLIB_PHONE_NUMBER            required
//	TDLIB_PASSWORD                optional 2FA password
//	TDLIB_CODE                    optional login code (skips the wait)
//	TDLIB_DATA_DIR                optional, default /data
//
// stdout carries MCP frames only; all logs go to stderr.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	tdlib "github.com/aliforever/go-tdlib/v2"
	"github.com/aliforever/go-tdlib/v2/td"
)

func main() {
	log.SetOutput(os.Stderr)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	apiID, err := strconv.ParseInt(os.Getenv("TDLIB_API_ID"), 10, 32)
	if err != nil {
		return fmt.Errorf("TDLIB_API_ID: %w", err)
	}
	apiHash := os.Getenv("TDLIB_API_HASH")
	if apiHash == "" {
		return fmt.Errorf("TDLIB_API_HASH is required")
	}
	phone := os.Getenv("TDLIB_PHONE_NUMBER")
	if phone == "" {
		return fmt.Errorf("TDLIB_PHONE_NUMBER is required")
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Keep TDLib's own logging to errors; see examples/login for why
	// td_execute is used instead of the log callback.
	tdlib.Execute([]byte(`{"@type":"setLogVerbosityLevel","new_verbosity_level":1}`))

	manager := tdlib.NewManager()
	client := manager.NewClient()

	a := &app{
		client:   client,
		apiID:    int32(apiID),
		apiHash:  apiHash,
		phone:    phone,
		password: os.Getenv("TDLIB_PASSWORD"),
		dataDir:  envOr("TDLIB_DATA_DIR", "/data"),
		codeCh:   make(chan string, 1),
		closedCh: make(chan struct{}),
	}

	srv := NewServer(Config{
		Lookup:     lookupAdapter,
		Search:     searchAdapter,
		Call:       callAdapter(client),
		SubmitCode: a.submitCode,
	})

	// One untyped update hook feeding the ring buffer, waiters and the
	// authorization-state tracker. Must not block; HandleUpdate copies raw.
	client.OnRaw(func(_ context.Context, tdType string, raw json.RawMessage) {
		srv.HandleUpdate(tdType, raw)
	})
	client.On[*td.UpdateAuthorizationState](a.onAuthorizationState)

	// TDLib's lifetime is governed by runCtx, not the signal context, so the
	// graceful Close handshake below still has a working transport after a
	// signal cancels sigCtx.
	runCtx, stopRun := context.WithCancel(context.Background())
	defer stopRun()
	runDone := make(chan error, 1)
	go func() {
		runDone <- manager.RunTDJSON(runCtx)
	}()

	// Serve MCP frames immediately — before authorization completes.
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- srv.Serve(runCtx, os.Stdin, os.Stdout)
	}()

	select {
	case <-sigCtx.Done():
		log.Printf("tdlib-mcp: signal received, shutting down")
	case err := <-serveDone:
		if err != nil {
			log.Printf("tdlib-mcp: stdio loop ended: %v", err)
		} else {
			log.Printf("tdlib-mcp: stdin closed, shutting down")
		}
	case err := <-runDone:
		stopRun()
		if err != nil {
			return fmt.Errorf("tdjson manager: %w", err)
		}
		return nil
	}

	// Graceful Close so the session is flushed to the database and the next
	// run resumes without a new login code.
	closeCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := client.Do(closeCtx, td.Close{}); err != nil {
		log.Printf("tdlib-mcp: close: %v", err)
	}
	select {
	case <-a.closedCh:
	case <-closeCtx.Done():
		log.Printf("tdlib-mcp: timed out waiting for authorizationStateClosed")
	}
	stopRun()
	return nil
}

type app struct {
	client *tdlib.Client

	apiID    int32
	apiHash  string
	phone    string
	password string
	dataDir  string

	// codeCh feeds the WaitCode arm; written by tdlib_submit_code and
	// (indirectly, via the file poller) by /data/code.txt.
	codeCh         chan string
	waitingForCode atomic.Bool
	closedCh       chan struct{}
	closeOnce      sync.Once
}

func (a *app) onAuthorizationState(ctx context.Context, u *td.UpdateAuthorizationState) {
	log.Printf("tdlib-mcp: authorization state: %s", u.AuthorizationState.Kind())

	err := u.AuthorizationState.Match(td.AuthorizationStateCases[error]{
		AuthorizationStateWaitTdlibParameters: func(*td.AuthorizationStateWaitTdlibParameters) error {
			dbDir := a.dataDir + "/tdlib-db"
			filesDir := a.dataDir + "/tdlib-files"
			// TDLib does not create these directories itself.
			for _, dir := range []string{dbDir, filesDir} {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					return err
				}
			}
			_, err := a.client.Do(ctx, td.SetTdlibParameters{
				UseTestDc:           false,
				DatabaseDirectory:   dbDir,
				FilesDirectory:      filesDir,
				UseFileDatabase:     true,
				UseChatInfoDatabase: true,
				UseMessageDatabase:  true,
				UseSecretChats:      false,
				ApiId:               a.apiID,
				ApiHash:             a.apiHash,
				SystemLanguageCode:  "en",
				DeviceModel:         "tdlib-mcp",
				ApplicationVersion:  serverVersion,
			})
			return err
		},
		AuthorizationStateWaitPhoneNumber: func(*td.AuthorizationStateWaitPhoneNumber) error {
			_, err := a.client.SetAuthenticationPhoneNumber(ctx, a.phone, nil)
			return err
		},
		AuthorizationStateWaitCode: func(*td.AuthorizationStateWaitCode) error {
			code, err := a.waitForCode(ctx)
			if err != nil {
				return err
			}
			_, err = a.client.CheckAuthenticationCode(ctx, code)
			return err
		},
		AuthorizationStateWaitPassword: func(*td.AuthorizationStateWaitPassword) error {
			if a.password == "" {
				return fmt.Errorf("2FA password required: set TDLIB_PASSWORD")
			}
			_, err := a.client.CheckAuthenticationPassword(ctx, a.password)
			return err
		},
		AuthorizationStateWaitRegistration: func(*td.AuthorizationStateWaitRegistration) error {
			return fmt.Errorf("account is not registered; registration is not supported")
		},
		AuthorizationStateReady: func(*td.AuthorizationStateReady) error {
			log.Printf("tdlib-mcp: authorized")
			return nil
		},
		AuthorizationStateClosed: func(*td.AuthorizationStateClosed) error {
			// Reached after our own td.Close during graceful shutdown.
			a.closeOnce.Do(func() { close(a.closedCh) })
			return nil
		},
		Unknown: func(kind string, _ json.RawMessage) error {
			log.Printf("tdlib-mcp: unknown authorization state %q; ignoring", kind)
			return nil
		},
	})
	if err != nil {
		log.Printf("tdlib-mcp: authorization step failed: %v", err)
	}
}

// submitCode implements the tdlib_submit_code tool: it hands the code to
// the WaitCode arm through the same channel the /data/code.txt poller uses.
func (a *app) submitCode(code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return fmt.Errorf("empty code")
	}
	if !a.waitingForCode.Load() {
		return fmt.Errorf("the authorization flow is not waiting for a code right now")
	}
	select {
	case a.codeCh <- code:
		return nil
	default:
		return fmt.Errorf("a code was already submitted and is being processed")
	}
}

// waitForCode returns TDLIB_CODE if set, otherwise blocks on the first of:
// a code from tdlib_submit_code, or a code appearing in <dataDir>/code.txt
// (polled every 2s; the file is removed after reading).
func (a *app) waitForCode(ctx context.Context) (string, error) {
	if code := os.Getenv("TDLIB_CODE"); code != "" {
		return code, nil
	}
	a.waitingForCode.Store(true)
	defer a.waitingForCode.Store(false)

	path := a.dataDir + "/code.txt"
	log.Printf("tdlib-mcp: waiting for login code: use the tdlib_submit_code tool or write it to %s", path)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case code := <-a.codeCh:
			if code = strings.TrimSpace(code); code != "" {
				return code, nil
			}
		case <-ticker.C:
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			code := strings.TrimSpace(string(data))
			if code == "" {
				continue
			}
			_ = os.Remove(path)
			return code, nil
		}
	}
}

// ---------------------------------------------------------------------------
// Backend adapters: generated td registry → server.go contract
// ---------------------------------------------------------------------------

func convertSchema(m td.MethodSchema) MethodSchema {
	out := MethodSchema{
		Name:        m.Name,
		Description: m.Description,
		ReturnType:  m.ReturnType,
	}
	for _, p := range m.Params {
		out.Params = append(out.Params, ParamSchema{
			Name:        p.Name,
			GoType:      p.GoType,
			IsUnion:     p.IsUnion,
			Variants:    p.Variants,
			Description: p.Description,
		})
	}
	return out
}

func lookupAdapter(name string) (MethodSchema, bool) {
	m, ok := td.LookupMethod(name)
	return convertSchema(m), ok
}

func searchAdapter(query string, limit int) []MethodSchema {
	found := td.SearchMethods(query, limit)
	out := make([]MethodSchema, 0, len(found))
	for _, m := range found {
		out = append(out, convertSchema(m))
	}
	return out
}

// callAdapter adapts Client.DoRaw to CallFunc, converting *td.Error into
// the stdlib-only *CallError so server.go never imports go-tdlib. DoRaw
// injects the envelope "@type"/"@extra" itself; params are passed through
// untouched so nested "@type" fields survive.
func callAdapter(client *tdlib.Client) CallFunc {
	return func(ctx context.Context, method string, params map[string]any) (json.RawMessage, error) {
		if params == nil {
			params = map[string]any{}
		}
		raw, err := client.DoRaw(ctx, method, params)
		if err != nil {
			var tdErr *td.Error
			if errors.As(err, &tdErr) {
				return nil, &CallError{Code: tdErr.Code, Message: tdErr.Message}
			}
			return nil, err
		}
		return raw, nil
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
