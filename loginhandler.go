package tdlib

import (
	"fmt"
	"github.com/aliforever/go-tdlib/entities"
	"github.com/aliforever/go-tdlib/incomingevents"
)

type AuthorizationHandler interface {
	Process(client *TDLib, update entities.AuthorizationStateType)
}

func isAuthorizationEvent(event incomingevents.Event) bool {
	return event.Type == "updateAuthorizationState" && event.AuthorizationState != nil
}

type UserAuthorizationHandler struct {
	PhoneNumber string
	Password    string
	onCode      func() (string, error)
	onSuccess   func()
	onError     func(stateType entities.AuthorizationStateType, err error)
}

func NewUserAuthorizationHandler(
	phoneNumber,
	password string,
	onCode func() (string, error),
	onSuccess func(),
	onError func(stateType entities.AuthorizationStateType, err error),
) *UserAuthorizationHandler {

	return &UserAuthorizationHandler{
		PhoneNumber: phoneNumber,
		Password:    password,
		onCode:      onCode,
		onSuccess:   onSuccess,
		onError:     onError,
	}
}

func (ulh UserAuthorizationHandler) Process(client *TDLib, update entities.AuthorizationStateType) {
	switch update {
	case entities.AuthorizationStateTypeAwaitingTdlibParameters:
		if err := client.SetTdlibParameters(); err != nil && ulh.onError != nil {
			ulh.onError(update, err)
		}
	case entities.AuthorizationStateTypeAwaitingPhoneNumber:
		if err := client.SetAuthenticationPhoneNumber(ulh.PhoneNumber, nil); err != nil && ulh.onError != nil {
			ulh.onError(update, err)
		}
	case entities.AuthorizationStateTypeAwaitingCode:
		code, err := ulh.onCode()
		if err != nil && ulh.onError != nil {
			ulh.onError(update, err)
			return
		}

		if err := client.CheckAuthenticationCode(code); err != nil && ulh.onError != nil {
			ulh.onError(update, err)
		}
	case entities.AuthorizationStateTypeAwaitingPassword:
		if err := client.CheckAuthenticationPassword(ulh.Password); err != nil && ulh.onError != nil {
			ulh.onError(update, err)
		}
	case entities.AuthorizationStateTypeAwaitingRegistration:
		// TODO: Implement registration
		return
	case entities.AuthorizationStateTypeReady:
		ulh.onSuccess()
		return
	default:
		if ulh.onError != nil {
			ulh.onError(update, fmt.Errorf("unexpected authorization state: %s", update))
		}
	}
}

type BotAuthorizationHandler struct {
	Token     string
	onSuccess func()
	onError   func(stateType entities.AuthorizationStateType, err error)
}

func NewBotAuthorizationHandler(
	token string,
	onSuccess func(),
	onError func(stateType entities.AuthorizationStateType, err error),
) *BotAuthorizationHandler {
	return &BotAuthorizationHandler{
		Token:     token,
		onSuccess: onSuccess,
		onError:   onError,
	}
}

func (blh BotAuthorizationHandler) Process(client *TDLib, update entities.AuthorizationStateType) {
	switch update {
	case entities.AuthorizationStateTypeAwaitingTdlibParameters:
		if err := client.SetTdlibParameters(); err != nil && blh.onError != nil {
			blh.onError(update, err)
		}
	case entities.AuthorizationStateTypeAwaitingPhoneNumber:
		if err := client.CheckAuthenticationBotToken(blh.Token); err != nil && blh.onError != nil {
			blh.onError(update, err)
		}
	case entities.AuthorizationStateTypeReady:
		blh.onSuccess()
	default:
		if blh.onError != nil {
			blh.onError(update, fmt.Errorf("unexpected authorization state: %s", update))
		}
	}
}
