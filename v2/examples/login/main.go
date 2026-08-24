//go:build tdjson

// Command login is a minimal end-to-end test harness for go-tdlib v2:
// it authorizes a user account and prints the result of GetMe.
//
// It is designed to run inside Docker without an interactive terminal:
//
//	TDLIB_API_ID, TDLIB_API_HASH  required, from https://my.telegram.org
//	TDLIB_PHONE_NUMBER            required, user account phone number
//	TDLIB_PASSWORD                optional 2FA password
//	TDLIB_CODE                    optional login code; if unset, the program
//	                              polls /data/code.txt until it appears
//	TDLIB_DATA_DIR                optional, default /data
//
// Run with (named volume: TDLib's atomic binlog rename fails on
// Windows-backed bind mounts such as WSL /mnt/c):
//
//	docker run --env-file .env -v tdlib-v2-data:/data go-tdlib-v2-test
//	docker cp code.txt <container>:/data/code.txt   # inject the login code
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tdlib "github.com/aliforever/go-tdlib/v2"
	"github.com/aliforever/go-tdlib/v2/td"
)

func main() {
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Keep TDLib's own logging to errors; debug verbosity echoes request
	// parameters (including credentials) into the logs. td_execute is the
	// reliable path — the log-callback registration does not suppress
	// TDLib's default stderr stream in this TDLib build.
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
		done:     make(chan error, 1),
		closedCh: make(chan struct{}),
	}

	client.On[*td.UpdateAuthorizationState](a.onAuthorizationState)

	go func() {
		<-ctx.Done()
		select {
		case a.done <- ctx.Err():
		default:
		}
	}()

	go func() {
		if err := manager.RunTDJSON(ctx); err != nil {
			a.done <- err
		}
	}()

	return <-a.done
}

type app struct {
	client *tdlib.Client

	apiID    int32
	apiHash  string
	phone    string
	password string
	dataDir  string

	done     chan error
	closedCh chan struct{}
	closeOnce sync.Once
}

func (a *app) onAuthorizationState(ctx context.Context, u *td.UpdateAuthorizationState) {
	log.Printf("authorization state: %s", u.AuthorizationState.Kind())

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
				DeviceModel:         "go-tdlib v2 docker test",
				ApplicationVersion:  "1.0",
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
			return fmt.Errorf("account is not registered; registration is not supported by this test")
		},
		AuthorizationStateReady: func(*td.AuthorizationStateReady) error {
			go a.onReady(ctx)
			return nil
		},
		AuthorizationStateClosed: func(*td.AuthorizationStateClosed) error {
			// Reached after our own td.Close during graceful shutdown.
			a.closeOnce.Do(func() { close(a.closedCh) })
			return nil
		},
		Unknown: func(kind string, _ json.RawMessage) error {
			log.Printf("unknown authorization state %q; ignoring", kind)
			return nil
		},
	})
	if err != nil {
		log.Printf("authorization step failed: %v", err)
		select {
		case a.done <- err:
		default:
		}
	}
}

func (a *app) onReady(ctx context.Context) {
	me, err := a.client.GetMe(ctx)
	if err != nil {
		a.done <- fmt.Errorf("GetMe: %w", err)
		return
	}
	log.Printf("TEST PASSED: authorized as %q %q (id %d, phone %s)",
		me.FirstName, me.LastName, me.Id, me.PhoneNumber)

	// Close TDLib gracefully so the session is flushed to the database and
	// the next run resumes without a new login code.
	if _, err := a.client.Do(ctx, td.Close{}); err != nil {
		log.Printf("close: %v", err)
	}
	select {
	case <-a.closedCh:
	case <-time.After(10 * time.Second):
		log.Printf("timed out waiting for authorizationStateClosed")
	}
	a.done <- nil
}

// waitForCode returns TDLIB_CODE if set, otherwise polls
// <dataDir>/code.txt until it appears or ctx is cancelled. The file is
// removed after reading so a stale code is never reused.
func (a *app) waitForCode(ctx context.Context) (string, error) {
	if code := os.Getenv("TDLIB_CODE"); code != "" {
		return code, nil
	}
	path := a.dataDir + "/code.txt"
	log.Printf("waiting for login code: write it to %s on the host", path)
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(2 * time.Second):
		}
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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
