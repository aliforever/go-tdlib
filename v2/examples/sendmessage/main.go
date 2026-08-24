//go:build tdjson

// Command sendmessage is the v2 counterpart of the top-level
// examples/sendmessage: it authorizes a bot or user account, sends a
// message to the "Saved Messages" chat and echoes back every incoming text
// message.
//
// It links against libtdjson through cgo, so it is excluded from default
// builds. Run it with:
//
//	cd v2/examples/sendmessage
//	TDLIB_API_ID=12345 TDLIB_API_HASH=... TDLIB_BOT_TOKEN=... go run -tags tdjson .
//
// Configuration (environment variables):
//
//	TDLIB_API_ID, TDLIB_API_HASH  required, from https://my.telegram.org
//	TDLIB_BOT_TOKEN               bot flow; mutually exclusive with the phone flow
//	TDLIB_PHONE_NUMBER            user flow; the login code is read from stdin
//	TDLIB_PASSWORD                optional 2FA password for the user flow
//	TDLIB_DATABASE_DIR            optional, default ./tdlib-db
//	TDLIB_FILES_DIR               optional, default ./tdlib-files
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	manager := tdlib.NewManager()
	client := manager.NewClient() // allocates a TDLib client ID via libtdjson

	app := &app{
		client:      client,
		apiID:       int32(apiID),
		apiHash:     apiHash,
		botToken:    os.Getenv("TDLIB_BOT_TOKEN"),
		phoneNumber: os.Getenv("TDLIB_PHONE_NUMBER"),
		password:    os.Getenv("TDLIB_PASSWORD"),
	}

	// Drive authorization from updateAuthorizationState updates.
	client.On[*td.UpdateAuthorizationState](app.onAuthorizationState)

	// Echo bot: reply to every incoming text message with the same text.
	client.Messages().WithContent[*td.MessageText]().Handle(app.echo)

	// One process-wide receive loop; ends when ctx is cancelled.
	return manager.RunTDJSON(ctx)
}

type app struct {
	client *tdlib.Client

	apiID       int32
	apiHash     string
	botToken    string
	phoneNumber string
	password    string
}

// onAuthorizationState advances the TDLib authorization state machine.
// AuthorizationState is a generated union container; Match dispatches on
// its @type. Unknown states added by newer TDLib versions land in the
// Unknown arm and are only logged.
func (a *app) onAuthorizationState(ctx context.Context, u *td.UpdateAuthorizationState) {
	log.Printf("authorization state: %s", u.AuthorizationState.Kind())

	err := u.AuthorizationState.Match(td.AuthorizationStateCases[error]{
		AuthorizationStateWaitTdlibParameters: func(*td.AuthorizationStateWaitTdlibParameters) error {
			dbDir := envOr("TDLIB_DATABASE_DIR", "./tdlib-db")
			filesDir := envOr("TDLIB_FILES_DIR", "./tdlib-files")
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
				DeviceModel:         "go-tdlib v2 example",
				ApplicationVersion:  "1.0",
			})
			return err
		},
		AuthorizationStateWaitPhoneNumber: func(*td.AuthorizationStateWaitPhoneNumber) error {
			// Bot flow and user flow share this state.
			if a.botToken != "" {
				_, err := a.client.CheckAuthenticationBotToken(ctx, a.botToken)
				return err
			}
			if a.phoneNumber == "" {
				return fmt.Errorf("set TDLIB_BOT_TOKEN or TDLIB_PHONE_NUMBER")
			}
			_, err := a.client.SetAuthenticationPhoneNumber(ctx, a.phoneNumber, nil)
			return err
		},
		AuthorizationStateWaitCode: func(*td.AuthorizationStateWaitCode) error {
			var code string
			fmt.Print("Enter the code you received: ")
			if _, err := fmt.Scanln(&code); err != nil {
				return err
			}
			_, err := a.client.CheckAuthenticationCode(ctx, code)
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
			return fmt.Errorf("account is not registered; registration is not implemented in this example")
		},
		AuthorizationStateReady: func(*td.AuthorizationStateReady) error {
			go a.onReady(ctx)
			return nil
		},
		Unknown: func(kind string, _ json.RawMessage) error {
			log.Printf("unknown authorization state %q; ignoring", kind)
			return nil
		},
	})
	if err != nil {
		log.Printf("authorization step failed: %v", err)
	}
}

// onReady demonstrates the three request styles once authorization is done:
// the generated facade, Do with a typed request struct, and DoAsync with a
// mapped Future.
func (a *app) onReady(ctx context.Context) {
	// Facade method: one generated line per TDLib function.
	me, err := a.client.GetMe(ctx)
	if err != nil {
		log.Printf("GetMe: %v", err)
		return
	}
	log.Printf("authorized as %s (id %d)", me.FirstName, me.Id)

	content := td.NewInputMessageContent(&td.InputMessageText{
		Text: &td.FormattedText{Text: "Hello from go-tdlib v2"},
	})

	// Do with a typed request struct: the response type is pinned by the
	// request, so sent is *td.Message with no type assertion.
	sent, err := a.client.Do(ctx, td.SendMessage{
		ChatId:              me.Id,
		InputMessageContent: content,
	})
	if err != nil {
		log.Printf("SendMessage: %v", err)
		return
	}
	log.Printf("sent message %d via Do", sent.Id)

	// The facade equivalent of the same call (empty unions marshal as null).
	if _, err := a.client.SendMessage(ctx, me.Id, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{}, content); err != nil {
		log.Printf("SendMessage facade: %v", err)
	}

	// DoAsync + Future.Map: fire the request, derive the display name lazily.
	name, err := a.client.DoAsync(ctx, td.GetMe{}).
		Map(func(u *td.User) (string, error) { return u.FirstName, nil }).
		Await(ctx)
	if err != nil {
		log.Printf("DoAsync GetMe: %v", err)
		return
	}
	log.Printf("DoAsync/Map/Await resolved first name: %s", name)
}

// echo replies to incoming text messages with the same text. Outgoing
// messages are skipped so the bot does not echo itself.
func (a *app) echo(ctx context.Context, m *td.Message, c *td.MessageText) {
	if m.IsOutgoing {
		return
	}
	text := "echo: " + c.Text.Text
	_, err := a.client.SendMessage(ctx, m.ChatId, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{},
		td.NewInputMessageContent(&td.InputMessageText{
			Text: &td.FormattedText{Text: text},
		}))
	if err != nil {
		log.Printf("echo to chat %d: %v", m.ChatId, err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
