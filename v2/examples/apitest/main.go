//go:build tdjson

// Command apitest is a broad end-to-end API test battery for go-tdlib v2:
// it authorizes a user account (reusing the login example's bootstrap) and
// then exercises the facade, Do, DoAsync/Future, the update router
// (On/Next/Updates), the message stream builder, file upload/download and
// the error path against the "Saved Messages" chat (chat ID == own user ID).
//
// Environment contract is identical to examples/login:
//
//	TDLIB_API_ID, TDLIB_API_HASH  required, from https://my.telegram.org
//	TDLIB_PHONE_NUMBER            required, user account phone number
//	TDLIB_PASSWORD                optional 2FA password
//	TDLIB_CODE                    optional login code; if unset, the program
//	                              polls <TDLIB_DATA_DIR>/code.txt until it appears
//	TDLIB_DATA_DIR                optional, default /data
//
// Every check logs one line: "PASS <name> <detail>" or "FAIL <name> <err>".
// The run ends with "SUMMARY: N passed, M failed". Exit code is 0 iff no
// critical check failed, 1 if any critical check failed, 2 if the 4-minute
// watchdog had to kill a hung battery.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
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

const (
	checkTimeout   = 12 * time.Second
	fileTimeout    = 60 * time.Second
	watchdogPeriod = 4 * time.Minute
)

func main() {
	code, err := run()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

func run() (int, error) {
	apiID, err := strconv.ParseInt(os.Getenv("TDLIB_API_ID"), 10, 32)
	if err != nil {
		return 1, fmt.Errorf("TDLIB_API_ID: %w", err)
	}
	apiHash := os.Getenv("TDLIB_API_HASH")
	if apiHash == "" {
		return 1, fmt.Errorf("TDLIB_API_HASH is required")
	}
	phone := os.Getenv("TDLIB_PHONE_NUMBER")
	if phone == "" {
		return 1, fmt.Errorf("TDLIB_PHONE_NUMBER is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// See examples/login: td_execute is the reliable way to keep TDLib's own
	// logging at error level.
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

	if err := <-a.done; err != nil {
		return 1, err
	}
	return a.exitCode, nil
}

type app struct {
	client *tdlib.Client

	apiID    int32
	apiHash  string
	phone    string
	password string
	dataDir  string

	done      chan error
	closedCh  chan struct{}
	closeOnce sync.Once
	exitCode  int
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
				DeviceModel:         "go-tdlib v2 apitest",
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

// onReady runs the battery under a watchdog, then closes TDLib gracefully so
// the session is flushed and the next run resumes without a new login code.
func (a *app) onReady(ctx context.Context) {
	batteryDone := make(chan int, 1)
	go func() {
		batteryDone <- a.runBattery(ctx)
	}()

	var code int
	select {
	case code = <-batteryDone:
	case <-time.After(watchdogPeriod):
		log.Printf("watchdog: battery exceeded %s; aborting", watchdogPeriod)
		os.Exit(2)
	}

	if _, err := a.client.Do(ctx, td.Close{}); err != nil {
		log.Printf("close: %v", err)
	}
	select {
	case <-a.closedCh:
	case <-time.After(10 * time.Second):
		log.Printf("timed out waiting for authorizationStateClosed")
	}
	a.exitCode = code
	a.done <- nil
}

// battery accumulates results and shared state between checks.
type battery struct {
	client  *tdlib.Client
	dataDir string

	passed     int
	failed     int
	critFailed int

	savedID int64   // Saved Messages chat ID == own user ID
	msgID   int64   // message created by the update-pipeline check
	msgIDs  []int64 // everything created, deleted by the cleanup check

	// remap tracks temporary->final message ID replacement reported by
	// updateMessageSendSucceeded (the sendMessage response can carry a
	// pending message with a temporary ID).
	remap sync.Map // int64 -> int64
}

// resolveID waits briefly for a temporary message ID to be remapped to its
// final server ID, returning the best known ID.
func (b *battery) resolveID(ctx context.Context, id int64) int64 {
	deadline := time.Now().Add(5 * time.Second)
	for {
		if v, ok := b.remap.Load(id); ok {
			return v.(int64)
		}
		if time.Now().After(deadline) || ctx.Err() != nil {
			return id
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// check runs fn with a timeout and logs exactly one PASS/FAIL line.
// critical failures drive the exit code; optional failures are only logged.
func (b *battery) check(parent context.Context, name string, critical bool, timeout time.Duration, fn func(ctx context.Context) (string, error)) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	detail, err := fn(ctx)
	if err != nil {
		b.failed++
		if critical {
			b.critFailed++
			log.Printf("FAIL %s (critical) %v", name, err)
		} else {
			log.Printf("FAIL %s (optional) %v", name, err)
		}
		return
	}
	b.passed++
	log.Printf("PASS %s %s", name, detail)
}

func textContent(text string) td.InputMessageContent {
	return td.NewInputMessageContent(&td.InputMessageText{
		Text: &td.FormattedText{Text: text},
	})
}

func messageText(m *td.Message) (string, bool) {
	if m == nil {
		return "", false
	}
	mt, ok := m.Content.As[*td.MessageText]()
	if !ok || mt.Text == nil {
		return "", false
	}
	return mt.Text.Text, true
}

func (a *app) runBattery(ctx context.Context) int {
	b := &battery{client: a.client, dataDir: a.dataDir}

	// Diagnostic: log temporary->final message ID remapping to understand
	// edit/react failures on freshly sent messages.
	b.client.On[*td.UpdateMessageSendSucceeded](func(_ context.Context, u *td.UpdateMessageSendSucceeded) {
		log.Printf("diag: sendSucceeded old_message_id=%d new_id=%d chat=%d", u.OldMessageId, u.Message.Id, u.Message.ChatId)
		b.remap.Store(u.OldMessageId, u.Message.Id)
	})
	b.client.On[*td.UpdateMessageSendFailed](func(_ context.Context, u *td.UpdateMessageSendFailed) {
		log.Printf("diag: sendFailed old_message_id=%d error=%v", u.OldMessageId, u.Error)
	})

	// 1. GetMe in all three request styles.
	b.check(ctx, "GetMe/facade", true, checkTimeout, func(ctx context.Context) (string, error) {
		me, err := b.client.GetMe(ctx)
		if err != nil {
			return "", err
		}
		b.savedID = me.Id
		return fmt.Sprintf("id=%d first_name=%q", me.Id, me.FirstName), nil
	})
	b.check(ctx, "GetMe/Do", true, checkTimeout, func(ctx context.Context) (string, error) {
		me, err := b.client.Do(ctx, td.GetMe{})
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("id=%d", me.Id), nil
	})
	b.check(ctx, "GetMe/DoAsync", true, checkTimeout, func(ctx context.Context) (string, error) {
		name, err := b.client.DoAsync(ctx, td.GetMe{}).
			Map(func(u *td.User) (string, error) { return u.FirstName, nil }).
			Await(ctx)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("first_name=%q", name), nil
	})

	if b.savedID == 0 {
		log.Printf("SUMMARY: %d passed, %d failed", b.passed, b.failed)
		log.Printf("GetMe failed; cannot continue without a Saved Messages chat ID")
		return 1
	}

	// 2. LoadChats + GetChats on the main chat list.
	b.check(ctx, "LoadChats+GetChats", true, checkTimeout, func(ctx context.Context) (string, error) {
		list := td.NewChatList(&td.ChatListMain{})
		if _, err := b.client.LoadChats(ctx, list, 10); err != nil {
			// TDLib answers 404 when the list is already fully loaded.
			var tdErr *td.Error
			if !errors.As(err, &tdErr) || tdErr.Code != 404 {
				return "", fmt.Errorf("LoadChats: %w", err)
			}
		}
		chats, err := b.client.GetChats(ctx, list, 10)
		if err != nil {
			return "", fmt.Errorf("GetChats: %w", err)
		}
		return fmt.Sprintf("chat_ids=%d total=%d", len(chats.ChatIds), chats.TotalCount), nil
	})

	// 3. OpenChat / CloseChat.
	b.check(ctx, "OpenChat+CloseChat", true, checkTimeout, func(ctx context.Context) (string, error) {
		if _, err := b.client.OpenChat(ctx, b.savedID); err != nil {
			return "", fmt.Errorf("OpenChat: %w", err)
		}
		if _, err := b.client.CloseChat(ctx, b.savedID); err != nil {
			return "", fmt.Errorf("CloseChat: %w", err)
		}
		return fmt.Sprintf("chat_id=%d", b.savedID), nil
	})

	// 4. GetChat.
	b.check(ctx, "GetChat", true, checkTimeout, func(ctx context.Context) (string, error) {
		chat, err := b.client.GetChat(ctx, b.savedID)
		if err != nil {
			return "", err
		}
		if chat.Title == "" {
			return "", fmt.Errorf("chat %d has empty title", b.savedID)
		}
		return fmt.Sprintf("title=%q", chat.Title), nil
	})

	// 5. Update pipeline: On[UpdateNewMessage] + SendMessage + union decode.
	b.check(ctx, "On+SendMessage", true, checkTimeout, func(ctx context.Context) (string, error) {
		const text = "apitest: hello"
		updates := make(chan *td.UpdateNewMessage, 4)
		sub := b.client.On[*td.UpdateNewMessage](func(_ context.Context, u *td.UpdateNewMessage) {
			select {
			case updates <- u:
			default:
			}
		})
		defer sub.Close()

		sent, err := b.client.SendMessage(ctx, b.savedID, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{}, textContent(text))
		if err != nil {
			return "", fmt.Errorf("SendMessage: %w", err)
		}
		b.msgID = sent.Id
		b.msgIDs = append(b.msgIDs, sent.Id)

		for {
			select {
			case u := <-updates:
				if got, ok := messageText(u.Message); ok && u.Message.ChatId == b.savedID && got == text {
					return fmt.Sprintf("message_id=%d update_id=%d text=%q", sent.Id, u.Message.Id, got), nil
				}
			case <-ctx.Done():
				return "", fmt.Errorf("updateNewMessage not received: %w", ctx.Err())
			}
		}
	})

	// 6. EditMessageText + GetMessage.
	b.check(ctx, "EditMessageText+GetMessage", true, checkTimeout, func(ctx context.Context) (string, error) {
		if b.msgID == 0 {
			return "", fmt.Errorf("skipped: no message from On+SendMessage")
		}
		const text = "apitest: edited"
		editID := b.resolveID(ctx, b.msgID)
		if pre, perr := b.client.GetMessage(ctx, b.savedID, editID); perr != nil {
			return "", fmt.Errorf("pre-edit GetMessage: %w", perr)
		} else {
			log.Printf("diag: pre-edit message id=%d outgoing=%v sending=%q edit_date=%d",
				pre.Id, pre.IsOutgoing, pre.SendingState.Kind(), pre.EditDate)
		}
		if _, err := b.client.EditMessageText(ctx, b.savedID, editID, td.ReplyMarkup{}, textContent(text)); err != nil {
			return "", fmt.Errorf("EditMessageText: %w", err)
		}
		m, err := b.client.GetMessage(ctx, b.savedID, editID)
		if err != nil {
			return "", fmt.Errorf("GetMessage: %w", err)
		}
		got, ok := messageText(m)
		if !ok || got != text {
			return "", fmt.Errorf("GetMessage text = %q, want %q", got, text)
		}
		return fmt.Sprintf("message_id=%d text=%q", m.Id, got), nil
	})

	// 7. GetChatHistory.
	b.check(ctx, "GetChatHistory", true, checkTimeout, func(ctx context.Context) (string, error) {
		ms, err := b.client.GetChatHistory(ctx, b.savedID, 0, 0, 10, false)
		if err != nil {
			return "", err
		}
		if len(ms.Messages) == 0 {
			return "", fmt.Errorf("history is empty")
		}
		return fmt.Sprintf("messages=%d", len(ms.Messages)), nil
	})

	// 8. ForwardMessages Saved Messages -> Saved Messages.
	b.check(ctx, "ForwardMessages", true, checkTimeout, func(ctx context.Context) (string, error) {
		if b.msgID == 0 {
			return "", fmt.Errorf("skipped: no message from On+SendMessage")
		}
		ms, err := b.client.ForwardMessages(ctx, b.savedID, td.MessageTopic{}, b.savedID, []int64{b.msgID}, nil, false, false)
		if err != nil {
			return "", err
		}
		if len(ms.Messages) == 0 {
			return "", fmt.Errorf("forward returned no messages")
		}
		for _, m := range ms.Messages {
			if m != nil {
				b.msgIDs = append(b.msgIDs, m.Id)
			}
		}
		return fmt.Sprintf("forwarded=%d", len(ms.Messages)), nil
	})

	// 9. Next[UpdateNewMessage]: one-shot receive in a goroutine, then send.
	b.check(ctx, "Next", true, checkTimeout, func(ctx context.Context) (string, error) {
		const text = "apitest: next"
		type result struct {
			id  int64
			err error
		}
		got := make(chan result, 1)
		go func() {
			u, err := b.client.Next[*td.UpdateNewMessage](ctx, func(u *td.UpdateNewMessage) bool {
				t, ok := messageText(u.Message)
				return ok && u.Message.ChatId == b.savedID && t == text
			})
			if err != nil {
				got <- result{err: err}
				return
			}
			got <- result{id: u.Message.Id}
		}()

		// Give Next a moment to register before the update can arrive.
		time.Sleep(200 * time.Millisecond)
		sent, err := b.client.SendMessage(ctx, b.savedID, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{}, textContent(text))
		if err != nil {
			return "", fmt.Errorf("SendMessage: %w", err)
		}
		b.msgIDs = append(b.msgIDs, sent.Id)

		select {
		case r := <-got:
			if r.err != nil {
				return "", fmt.Errorf("Next: %w", r.err)
			}
			return fmt.Sprintf("message_id=%d", r.id), nil
		case <-ctx.Done():
			return "", fmt.Errorf("Next did not return: %w", ctx.Err())
		}
	})

	// 10. Updates[UpdateNewMessage]: channel form, then cancel.
	b.check(ctx, "Updates", true, checkTimeout, func(ctx context.Context) (string, error) {
		const text = "apitest: updates"
		ch, cancel := b.client.Updates[*td.UpdateNewMessage](ctx)
		defer cancel()

		sent, err := b.client.SendMessage(ctx, b.savedID, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{}, textContent(text))
		if err != nil {
			return "", fmt.Errorf("SendMessage: %w", err)
		}
		b.msgIDs = append(b.msgIDs, sent.Id)

		for {
			select {
			case u, ok := <-ch:
				if !ok {
					return "", fmt.Errorf("updates channel closed before delivery")
				}
				if t, match := messageText(u.Message); match && u.Message.ChatId == b.savedID && t == text {
					return fmt.Sprintf("message_id=%d", sent.Id), nil
				}
			case <-ctx.Done():
				return "", fmt.Errorf("update not received: %w", ctx.Err())
			}
		}
	})

	// 11. Message stream builder: Messages().InChat().WithContent().Handle().
	b.check(ctx, "Stream", true, checkTimeout, func(ctx context.Context) (string, error) {
		const text = "apitest: stream"
		got := make(chan int64, 1)
		sub := b.client.Messages().InChat(b.savedID).WithContent[*td.MessageText]().Handle(
			func(_ context.Context, m *td.Message, c *td.MessageText) {
				if c.Text != nil && c.Text.Text == text {
					select {
					case got <- m.Id:
					default:
					}
				}
			})
		defer sub.Close()

		time.Sleep(200 * time.Millisecond)
		sent, err := b.client.SendMessage(ctx, b.savedID, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{}, textContent(text))
		if err != nil {
			return "", fmt.Errorf("SendMessage: %w", err)
		}
		b.msgIDs = append(b.msgIDs, sent.Id)

		select {
		case id := <-got:
			return fmt.Sprintf("message_id=%d", id), nil
		case <-ctx.Done():
			return "", fmt.Errorf("stream handler did not fire: %w", ctx.Err())
		}
	})

	// 12. File round-trip: upload a document, download it back, compare bytes.
	b.check(ctx, "FileRoundTrip", true, fileTimeout, func(ctx context.Context) (string, error) {
		path := b.dataDir + "/apitest-doc.txt"
		payload := "apitest document payload\n"
		if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
			return "", err
		}

		doc, err := b.client.SendMessage(ctx, b.savedID, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{},
			td.NewInputMessageContent(&td.InputMessageDocument{
				Document:                    td.NewInputFile(&td.InputFileLocal{Path: path}),
				DisableContentTypeDetection: true,
				Caption:                     &td.FormattedText{Text: "apitest: doc"},
			}))
		if err != nil {
			return "", fmt.Errorf("SendMessage document: %w", err)
		}
		b.msgIDs = append(b.msgIDs, doc.Id)

		md, ok := doc.Content.As[*td.MessageDocument]()
		if !ok || md.Document == nil || md.Document.Document == nil {
			return "", fmt.Errorf("sent message content is not a usable messageDocument")
		}
		file := md.Document.Document

		dl, err := b.client.DownloadFile(ctx, file.Id, 1, 0, 0, true)
		if err != nil {
			return "", fmt.Errorf("DownloadFile: %w", err)
		}
		if dl.Local == nil || !dl.Local.IsDownloadingCompleted || dl.Local.Path == "" {
			return "", fmt.Errorf("download not completed (local=%+v)", dl.Local)
		}
		data, err := os.ReadFile(dl.Local.Path)
		if err != nil {
			return "", fmt.Errorf("read downloaded file: %w", err)
		}
		if string(data) != payload {
			return "", fmt.Errorf("downloaded content %q != %q", string(data), payload)
		}

		remoteNote := "no remote id"
		if file.Remote != nil && file.Remote.Id != "" {
			if _, err := b.client.GetRemoteFile(ctx, file.Remote.Id, td.FileType{}); err != nil {
				remoteNote = fmt.Sprintf("GetRemoteFile: %v", err)
			} else {
				remoteNote = "GetRemoteFile ok"
			}
		}
		return fmt.Sprintf("file_id=%d size=%d path=%s (%s)", file.Id, len(data), dl.Local.Path, remoteNote), nil
	})

	// 13. Error path: an invalid chat ID must surface as *td.Error.
	b.check(ctx, "ErrorPath", true, checkTimeout, func(ctx context.Context) (string, error) {
		_, err := b.client.GetChat(ctx, math.MaxInt64-1)
		if err == nil {
			return "", fmt.Errorf("GetChat with invalid ID unexpectedly succeeded")
		}
		var tdErr *td.Error
		if !errors.As(err, &tdErr) {
			return "", fmt.Errorf("error type is %T, want *td.Error", err)
		}
		if tdErr.Code == 0 {
			return "", fmt.Errorf("td.Error has code 0")
		}
		return fmt.Sprintf("code=%d message=%q", tdErr.Code, tdErr.Message), nil
	})

	// Optional checks run before cleanup so the edited message still exists.

	// 15. AddMessageReaction / RemoveMessageReaction.
	b.check(ctx, "AddMessageReaction", false, checkTimeout, func(ctx context.Context) (string, error) {
		if b.msgID == 0 {
			return "", fmt.Errorf("skipped: no message from On+SendMessage")
		}
		reaction := td.NewReactionType(&td.ReactionTypeEmoji{Emoji: "👍"})
		reactID := b.resolveID(ctx, b.msgID)
		if _, err := b.client.AddMessageReaction(ctx, b.savedID, reactID, reaction, false, false); err != nil {
			return "", err
		}
		if _, err := b.client.RemoveMessageReaction(ctx, b.savedID, reactID, reaction); err != nil {
			return "", fmt.Errorf("RemoveMessageReaction: %w", err)
		}
		return "added and removed 👍", nil
	})

	// 16. PinChatMessage / UnpinChatMessage.
	b.check(ctx, "PinChatMessage", false, checkTimeout, func(ctx context.Context) (string, error) {
		if b.msgID == 0 {
			return "", fmt.Errorf("skipped: no message from On+SendMessage")
		}
		pinID := b.resolveID(ctx, b.msgID)
		if _, err := b.client.PinChatMessage(ctx, b.savedID, pinID, true, false); err != nil {
			return "", err
		}
		if _, err := b.client.UnpinChatMessage(ctx, b.savedID, pinID); err != nil {
			return "", fmt.Errorf("UnpinChatMessage: %w", err)
		}
		return fmt.Sprintf("message_id=%d pinned and unpinned", pinID), nil
	})

	// 17. SendChatAction typing.
	b.check(ctx, "SendChatAction", false, checkTimeout, func(ctx context.Context) (string, error) {
		if _, err := b.client.SendChatAction(ctx, b.savedID, td.MessageTopic{}, "", td.NewChatAction(&td.ChatActionTyping{})); err != nil {
			return "", err
		}
		return "typing sent", nil
	})

	// 18. SearchPublicChat.
	b.check(ctx, "SearchPublicChat", false, checkTimeout, func(ctx context.Context) (string, error) {
		chat, err := b.client.SearchPublicChat(ctx, "telegram")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("chat_id=%d title=%q", chat.Id, chat.Title), nil
	})

	// 14. Cleanup: delete every message the battery created.
	b.check(ctx, "DeleteMessages", true, checkTimeout, func(ctx context.Context) (string, error) {
		if len(b.msgIDs) == 0 {
			return "nothing to delete", nil
		}
		ids := make([]int64, len(b.msgIDs))
		for i, id := range b.msgIDs {
			ids[i] = b.resolveID(ctx, id)
		}
		if _, err := b.client.DeleteMessages(ctx, b.savedID, ids, true); err != nil {
			return "", err
		}
		return fmt.Sprintf("deleted=%d", len(ids)), nil
	})

	log.Printf("SUMMARY: %d passed, %d failed", b.passed, b.failed)
	if b.critFailed > 0 {
		return 1
	}
	return 0
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
