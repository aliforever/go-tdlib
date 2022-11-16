package tdlib

import (
	"encoding/json"
	"github.com/aliforever/go-tdlib/entities"
	"github.com/aliforever/go-tdlib/incomingevents"
	"github.com/aliforever/go-tdlib/outgoingevents"
)

func (t *TDLib) CustomRequest(requestType string, parameters map[string]interface{}) (result *json.RawMessage, err error) {
	resp, err := sendMap[json.RawMessage](t, requestType, parameters)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (t *TDLib) GetAuthorizationState() (state *entities.AuthorizationStateType, err error) {
	resp, err := send[entities.AuthorizationStateType](t, outgoingevents.GetAuthorizationState{})
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (t *TDLib) SetTdlibParameters() error {
	_, err := send[map[string]interface{}](t, outgoingevents.SetTdlibParameters{
		UseTestDc:              t.cfg.UseTestDC,
		DatabaseDirectory:      t.cfg.DatabaseDirectory,
		FilesDirectory:         t.cfg.FilesDirectory,
		DatabaseEncryptionKey:  t.cfg.DatabaseEncryptionKey,
		UseFileDatabase:        t.cfg.UseFileDatabase,
		UseChatInfoDatabase:    t.cfg.UseChatInfoDatabase,
		UseMessageDatabase:     t.cfg.UseMessageDatabase,
		UseSecretChats:         t.cfg.UseSecretChats,
		ApiId:                  t.apiID,
		ApiHash:                t.apiHash,
		SystemLanguageCode:     t.cfg.SystemLanguageCode,
		DeviceModel:            t.cfg.DeviceModel,
		SystemVersion:          t.cfg.SystemVersion,
		ApplicationVersion:     t.cfg.ApplicationVersion,
		EnableStorageOptimizer: t.cfg.StorageOptimizer,
		IgnoreFileNames:        t.cfg.FilenamesIgnored,
	})

	return err
}

func (t *TDLib) SetAuthenticationPhoneNumber(phoneNumber string, settings *entities.PhoneNumberAuthenticationSettings) error {
	_, err := send[map[string]interface{}](t, outgoingevents.SetAuthenticationPhoneNumber{
		PhoneNumber: phoneNumber,
		Settings:    settings,
	})

	return err
}

func (t *TDLib) CheckAuthenticationCode(code string) error {
	_, err := send[map[string]interface{}](t, outgoingevents.CheckAuthenticationCode{
		Code: code,
	})

	return err
}

func (t *TDLib) CheckAuthenticationPassword(password string) error {
	_, err := send[map[string]interface{}](t, outgoingevents.CheckAuthenticationPassword{
		Password: password,
	})

	return err
}

func (t *TDLib) GetChat(chatID int64) (*incomingevents.GetChatResponse, error) {
	resp, err := send[incomingevents.GetChatResponse](t, outgoingevents.GetChat{
		ChatID: chatID,
	})

	return resp, err
}

func (t *TDLib) GetChats(inboxType *entities.ChatList, limit int32) (*incomingevents.GetChatsResponse, error) {
	resp, err := send[incomingevents.GetChatsResponse](t, outgoingevents.GetChats{
		ChatList: inboxType,
		Limit:    limit,
	})

	return resp, err
}

func (t *TDLib) LoadChats(inboxType *entities.ChatList, limit int32) error {
	_, err := send[map[string]interface{}](t, outgoingevents.LoadChats{
		ChatList: inboxType,
		Limit:    limit,
	})

	return err
}

func (t *TDLib) DownloadFile(fileID, priority, offset, limit int64, synchronous bool) (*incomingevents.DownloadFileResponse, error) {
	resp, err := send[incomingevents.DownloadFileResponse](t, outgoingevents.DownloadFile{
		FileID:      fileID,
		Priority:    priority,
		Offset:      offset,
		Limit:       limit,
		Synchronous: synchronous,
	})

	return resp, err
}

func (t *TDLib) GetMessage(chatID, messageID int64) (*incomingevents.GetMessageResponse, error) {
	resp, err := send[incomingevents.GetMessageResponse](t, outgoingevents.GetMessage{
		ChatID:    chatID,
		MessageID: messageID,
	})

	return resp, err
}

func (t *TDLib) GetChatHistory(chatID, fromMessageID, offset int64, limit uint64, onlyLocal bool) (*incomingevents.GetChatHistoryResponse, error) {
	resp, err := send[incomingevents.GetChatHistoryResponse](t, outgoingevents.GetChatHistory{
		ChatID:        chatID,
		FromMessageID: fromMessageID,
		Offset:        offset,
		Limit:         limit,
		OnlyLocal:     onlyLocal,
	})

	return resp, err
}
