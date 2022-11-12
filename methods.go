package tdlib

import (
	"github.com/aliforever/go-tdlib/entities"
	"github.com/aliforever/go-tdlib/incomingevents"
	"github.com/aliforever/go-tdlib/outgoingevents"
)

func (t *TDLib) SetTdlibParameters() error {
	_, err := t.send(outgoingevents.SetTdlibParameters{
		UseTestDc:              false,
		DatabaseDirectory:      "./tdlib-db",
		FilesDirectory:         "./tdlib-files",
		DatabaseEncryptionKey:  nil,
		UseFileDatabase:        false,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		UseSecretChats:         false,
		ApiId:                  t.apiID,
		ApiHash:                t.apiHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "Server",
		SystemVersion:          "1.0.0",
		ApplicationVersion:     "1.0.0",
		EnableStorageOptimizer: false,
		IgnoreFileNames:        false,
	})

	return err
}

func (t *TDLib) SetAuthenticationPhoneNumber(phoneNumber string, settings *entities.PhoneNumberAuthenticationSettings) error {
	_, err := t.send(outgoingevents.SetAuthenticationPhoneNumber{
		PhoneNumber: phoneNumber,
		Settings:    settings,
	})

	return err
}

func (t *TDLib) CheckAuthenticationCode(code string) error {
	_, err := t.send(outgoingevents.CheckAuthenticationCode{
		Code: code,
	})

	return err
}

func (t *TDLib) GetChat(chatID int64) (*incomingevents.GetChatResponse, error) {
	resp, err := t.send(outgoingevents.GetChat{
		ChatID: chatID,
	})

	return resp.GetChatResponse, err
}

func (t *TDLib) GetChats(inboxType *entities.ChatList, limit int32) (*incomingevents.GetChatsResponse, error) {
	resp, err := t.send(outgoingevents.GetChats{
		ChatList: inboxType,
		Limit:    limit,
	})

	return resp.GetChatsResponse, err
}

func (t *TDLib) LoadChats(inboxType *entities.ChatList, limit int32) error {
	_, err := t.send(outgoingevents.LoadChats{
		ChatList: inboxType,
		Limit:    limit,
	})

	return err
}

func (t *TDLib) DownloadFile(fileID, priority, offset, limit int32, synchronous bool) (*incomingevents.File, error) {
	resp, err := t.send(outgoingevents.DownloadFile{
		FileID:      fileID,
		Priority:    priority,
		Offset:      offset,
		Limit:       limit,
		Synchronous: synchronous,
	})

	return resp.File, err
}
