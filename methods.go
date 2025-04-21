package tdlib

import (
	"encoding/json"

	"github.com/aliforever/go-tdlib/entities"
	"github.com/aliforever/go-tdlib/incomingevents"
	"github.com/aliforever/go-tdlib/outgoingevents"
)

func (t *TDLib) CustomRequest(
	requestType string,
	parameters map[string]interface{},
	options ...*SendOptions,
) (result *json.RawMessage, err error) {

	resp, err := sendMap[json.RawMessage](t, requestType, parameters, options...)
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

func (t *TDLib) AddProxyHttp(
	server string,
	port int64,
	enable bool,
	username string,
	password string,
	httpOnly bool,
) (*entities.Proxy, error) {
	return send[entities.Proxy](t, outgoingevents.AddProxy{
		Server: server,
		Port:   port,
		Enable: enable,

		Proxy: &entities.ProxyHttp{
			Username: username,
			Password: password,
			HttpOnly: httpOnly,
		},
	})
}

func (t *TDLib) AddProxyMtProto(
	server string,
	port int64,
	enable bool,
	secret string,
) (*entities.Proxy, error) {
	return send[entities.Proxy](t, outgoingevents.AddProxy{
		Server: server,
		Port:   port,
		Enable: enable,

		Proxy: &entities.ProxyMtproto{
			Secret: secret,
		},
	})
}

func (t *TDLib) AddProxySocks5(
	server string,
	port int64,
	enable bool,
	username string,
	password string,
) (*entities.Proxy, error) {
	return send[entities.Proxy](t, outgoingevents.AddProxy{
		Server: server,
		Port:   port,
		Enable: enable,

		Proxy: &entities.ProxySocks5{
			Username: username,
			Password: password,
		},
	})
}

func (t *TDLib) EditProxyHttp(
	id int64,
	server string,
	port int64,
	enable bool,
	username string,
	password string,
	httpOnly bool,
) (*entities.Proxy, error) {
	return send[entities.Proxy](t, outgoingevents.EditProxy{
		ProxyID: id,
		Server:  server,
		Port:    port,
		Enable:  enable,

		Proxy: &entities.ProxyHttp{
			Username: username,
			Password: password,
			HttpOnly: httpOnly,
		},
	})
}

func (t *TDLib) EditProxyMtProto(
	id int64,
	server string,
	port int64,
	enable bool,
	secret string,
) (*entities.Proxy, error) {
	return send[entities.Proxy](t, outgoingevents.EditProxy{
		ProxyID: id,
		Server:  server,
		Port:    port,
		Enable:  enable,

		Proxy: &entities.ProxyMtproto{
			Secret: secret,
		},
	})
}

func (t *TDLib) EditProxySocks5(
	id int64,
	server string,
	port int64,
	enable bool,
	username string,
	password string,
) (*entities.Proxy, error) {
	return send[entities.Proxy](t, outgoingevents.EditProxy{
		ProxyID: id,
		Server:  server,
		Port:    port,
		Enable:  enable,

		Proxy: &entities.ProxySocks5{
			Username: username,
			Password: password,
		},
	})
}

func (t *TDLib) RemoveProxy(id int64) error {
	_, err := send[any](t, outgoingevents.RemoveProxy{
		ProxyID: id,
	})

	return err
}

func (t *TDLib) EnableProxy(id int64) error {
	_, err := send[any](t, outgoingevents.RemoveProxy{
		ProxyID: id,
	})

	return err
}

func (t *TDLib) GetProxyLink(id int64) (*entities.HttpUrl, error) {
	return send[entities.HttpUrl](t, outgoingevents.GetProxyLink{
		ProxyID: id,
	})
}

func (t *TDLib) PingProxy(id int64) (*entities.Seconds, error) {
	return send[entities.Seconds](t, outgoingevents.PingProxy{
		ProxyID: id,
	})
}

func (t *TDLib) SetAuthenticationPhoneNumber(
	phoneNumber string,
	settings *entities.PhoneNumberAuthenticationSettings,
) error {
	_, err := send[map[string]interface{}](t, outgoingevents.SetAuthenticationPhoneNumber{
		PhoneNumber: phoneNumber,
		Settings:    settings,
	})

	return err
}

func (t *TDLib) CheckAuthenticationBotToken(token string) error {
	_, err := send[map[string]interface{}](t, outgoingevents.CheckAuthenticationBotToken{
		Token: token,
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

func (t *TDLib) RegisterUser(firstName, lastName string) error {
	_, err := send[map[string]interface{}](t, outgoingevents.RegisterUser{
		FirstName: firstName,
		LastName:  lastName,
	})

	return err
}

func (t *TDLib) AcceptTermsOfService(termsOfServiceID string) error {
	_, err := send[map[string]interface{}](t, outgoingevents.AcceptTermsOfService{
		TermsOfServiceID: termsOfServiceID,
	})

	return err
}

func (t *TDLib) GetMe() (*entities.User, error) {
	return send[entities.User](t, outgoingevents.GetMe{})
}

func (t *TDLib) GetChat(chatID int64) (*incomingevents.GetChatResponse, error) {
	return send[incomingevents.GetChatResponse](t, outgoingevents.GetChat{
		ChatID: chatID,
	})
}

func (t *TDLib) GetChats(inboxType *entities.ChatList, limit int32) (*incomingevents.GetChatsResponse, error) {
	return send[incomingevents.GetChatsResponse](t, outgoingevents.GetChats{
		ChatList: inboxType,
		Limit:    limit,
	})
}

func (t *TDLib) LoadChats(inboxType *entities.ChatList, limit int32) error {
	_, err := send[map[string]interface{}](t, outgoingevents.LoadChats{
		ChatList: inboxType,
		Limit:    limit,
	})

	return err
}

func (t *TDLib) DownloadFile(
	fileID,
	priority,
	offset,
	limit int64,
	synchronous bool,
	options ...*SendOptions,
) (*incomingevents.DownloadFileResponse, error) {
	return send[incomingevents.DownloadFileResponse](t, outgoingevents.DownloadFile{
		FileID:      fileID,
		Priority:    priority,
		Offset:      offset,
		Limit:       limit,
		Synchronous: synchronous,
	}, options...)
}

func (t *TDLib) DeleteFile(
	fileID int64,
	options ...*SendOptions,
) error {
	_, err := send[map[string]interface{}](t, outgoingevents.DeleteFile{
		FileID: fileID,
	}, options...)

	return err
}

func (t *TDLib) GetUser(
	userID int64,
	options ...*SendOptions,
) (*entities.User, error) {
	return send[entities.User](t, outgoingevents.GetUser{
		UserID: userID,
	}, options...)
}

func (t *TDLib) SetMessageSenderBlockListMain(
	senderID entities.SenderID,
	options ...*SendOptions,
) (*entities.User, error) {
	return send[entities.User](t, outgoingevents.SetMessageSenderBlockList{
		SenderId:  senderID,
		BlockList: entities.BlockListMain(),
	}, options...)
}

func (t *TDLib) SetMessageSenderBlockListStories(
	senderID entities.SenderID,
	options ...*SendOptions,
) (*entities.User, error) {
	return send[entities.User](t, outgoingevents.SetMessageSenderBlockList{
		SenderId:  senderID,
		BlockList: entities.BlockListStories(),
	}, options...)
}

func (t *TDLib) ReadFilePart(fileID, offset, count int64) (*incomingevents.ReadFilePart, error) {
	return send[incomingevents.ReadFilePart](t, outgoingevents.ReadFilePart{
		FileID: fileID,
		Offset: offset,
		Count:  count,
	})
}

func (t *TDLib) GetRemoteFile(remoteFileID string, fileType *entities.FileType) (*entities.File, error) {
	return send[entities.File](t, outgoingevents.GetRemoteFile{
		RemoteFileID: remoteFileID,
		FileType:     fileType,
	})
}

func (t *TDLib) GetMessage(chatID, messageID int64) (*entities.Message, error) {
	return send[entities.Message](t, outgoingevents.GetMessage{
		ChatID:    chatID,
		MessageID: messageID,
	})
}

func (t *TDLib) GetChatEventLog(
	chatID int64,
	searchQuery string,
	fromEventID int64,
	limit int64,
	filters *entities.ChatEventLogFilters,
	userIDs []int64,
) (*entities.ChatEvents, error) {
	return send[entities.ChatEvents](t, outgoingevents.GetChatEventLog{
		ChatID:    chatID,
		Query:     searchQuery,
		FromEvent: fromEventID,
		Limit:     limit,
		Filters:   filters,
		UserIDs:   userIDs,
	})
}

func (t *TDLib) GetChatHistory(
	chatID,
	fromMessageID,
	offset int64,
	limit uint64,
	onlyLocal bool,
) (*incomingevents.GetChatHistoryResponse, error) {
	return send[incomingevents.GetChatHistoryResponse](t, outgoingevents.GetChatHistory{
		ChatID:        chatID,
		FromMessageID: fromMessageID,
		Offset:        offset,
		Limit:         limit,
		OnlyLocal:     onlyLocal,
	})
}

func (t *TDLib) SendMessage(
	chatID int64,
	replyToMessageID int64,
	replyMarkup entities.ReplyMarkup,
	inputMessageContent entities.InputMessageContent,
	options *outgoingevents.SendMessageOptions,
) (*incomingevents.SendMessage, error) {
	return send[incomingevents.SendMessage](t, outgoingevents.SendMessage{
		ChatID:              chatID,
		ReplyToMessageID:    replyToMessageID,
		Options:             options,
		ReplyMarkup:         replyMarkup,
		InputMessageContent: inputMessageContent,
	})
}

func (t *TDLib) DeleteMessages(chatID int64, messageIDs []int64, revoke bool) error {
	_, err := send[map[string]interface{}](t, outgoingevents.DeleteMessages{
		ChatID:     chatID,
		MessageIDs: messageIDs,
		Revoke:     revoke,
	})

	return err
}

func (t *TDLib) GetInlineQueryResults(
	botUserID int64,
	chatID int64,
	userLocation interface{},
	query string,
	offset string,
) (*entities.InlineQueryResults, error) {
	return send[entities.InlineQueryResults](t, outgoingevents.GetInlineQueryResults{
		BotUserID:    botUserID,
		ChatID:       chatID,
		UserLocation: userLocation,
		Query:        query,
		Offset:       offset,
	})
}

func (t *TDLib) SendInlineQueryResultMessage(
	chatID int64,
	messageThreadID int64,
	replyToMessageID int64,
	options *outgoingevents.SendMessageOptions,
	queryID string,
	resultID string,
	hideViaBot bool,
) (*entities.Message, error) {
	return send[entities.Message](t, outgoingevents.SendInlineQueryResultMessage{
		ChatID:           chatID,
		MessageThreadID:  messageThreadID,
		ReplyToMessageID: replyToMessageID,
		Options:          options,
		QueryID:          queryID,
		ResultID:         resultID,
		HideViaBot:       hideViaBot,
	})
}

func (t *TDLib) GetCallbackQueryAnswer(
	chatID int64,
	messageID int64,
	payload *entities.CallbackQueryPayload,
) (*entities.CallbackQueryAnswer, error) {
	return send[entities.CallbackQueryAnswer](t, outgoingevents.GetCallbackQueryAnswer{
		ChatID:    chatID,
		MessageID: messageID,
		Payload:   payload,
	})
}

func (t *TDLib) OpenChat(chatID int64) error {
	_, err := send[map[string]interface{}](t, outgoingevents.OpenChat{
		ChatID: chatID,
	})

	return err
}

func (t *TDLib) CloseChat(chatID int64) error {
	_, err := send[map[string]interface{}](t, outgoingevents.CloseChat{
		ChatID: chatID,
	})

	return err
}

func (t *TDLib) GetVideoChatAvailableParticipants(chatID int64) (*entities.MessageSenders, error) {
	return send[entities.MessageSenders](t, outgoingevents.GetVideoChatAvailableParticipants{
		ChatID: chatID,
	})
}

func (t *TDLib) GetGroupCall(groupCallID int64) (*entities.GroupCall, error) {
	return send[entities.GroupCall](t, outgoingevents.GetGroupCall{
		GroupCallID: groupCallID,
	})
}

func (t *TDLib) LoadGroupCallParticipants(groupCallID int64, limit int64) error {
	_, err := send[map[string]interface{}](t, outgoingevents.LoadGroupCallParticipants{
		GroupCallID: groupCallID,
		Limit:       limit,
	})

	return err
}

func (t *TDLib) JoinGroupCall(
	groupCallID int64,
	payload string,
	joinAs *entities.MessageSender, //  pass null to join as self
	audioSourceID int64,
	isMuted bool,
	isMyVideo bool,
	inviteHash string,
) (*entities.RawText, error) {
	return send[entities.RawText](t, outgoingevents.JoinGroupCall{
		GroupCallID:   groupCallID,
		ParticipantID: joinAs,
		AudioSourceID: audioSourceID,
		Payload:       payload,
		IsMuted:       isMuted,
		IsMyVideo:     isMyVideo,
		InviteHash:    inviteHash,
	})
}

func (t *TDLib) SearchPublicChat(username string) (*incomingevents.GetChatResponse, error) {
	return send[incomingevents.GetChatResponse](t, outgoingevents.SearchPublicChat{
		Username: username,
	})
}

func (t *TDLib) GetMessageLinkInfo(address string) (*entities.MessageLinkInfo, error) {
	return send[entities.MessageLinkInfo](t, outgoingevents.GetMessageLinkInfo{
		URL: address,
	})
}
