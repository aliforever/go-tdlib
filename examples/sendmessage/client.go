package main

import (
	"encoding/json"
	"fmt"
	"github.com/aliforever/go-tdlib"
	"github.com/aliforever/go-tdlib/config"
	"github.com/aliforever/go-tdlib/entities"
	"github.com/aliforever/go-tdlib/incomingevents"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	client *tdlib.TDLib

	title string
}

func NewClient(title string) *Client {
	return &Client{title: title}
}

// Start starts the client
func (c *Client) Start() error {
	cfg := config.New().
		SetLogPath("./tdlib/logs.txt").
		SetFilesDirectory("./tdlib/tdlib-files").
		SetDatabaseDirectory("./tdlib/tdlib-db").
		IgnoreFileNames()

	h := tdlib.NewHandlers().
		SetOnUpdateAuthorizationStateEventHandler(c.onAuthorizationStateChange).
		SetRawIncomingEventHandler(c.onRawIncomingEvent).
		SetErrorHandler(c.onError)

	c.client = tdlib.NewClient(27625832, "e79990d433dd59de1ac4d06014cb5f32", h, cfg, nil)

	return c.client.ReceiveUpdates()
}

func (c *Client) receiveCodeFromChannel() {
	fmt.Printf("Enter Code: \n")
	code, err := c.receiveInputFromQasedlu("Send Received Code...")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		time.Sleep(time.Second * 5)
		c.receiveCodeFromChannel()
		return
	}

	if code == "STOP" {
		os.Exit(0)
		return
	}

	err = c.client.CheckAuthenticationCode(code)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		time.Sleep(time.Second * 5)
		c.receiveCodeFromChannel()
		return
	}
}

func (c *Client) receivePasswordFromChannel() {
	fmt.Printf("Enter Password: \n")
	password, err := c.receiveInputFromQasedlu("Send Password...")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		time.Sleep(time.Second * 5)
		c.receiveCodeFromChannel()
		return
	}

	err = c.client.CheckAuthenticationPassword(password)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		time.Sleep(time.Second * 5)
		c.receiveCodeFromChannel()
		return
	}
}

func (c *Client) onAuthorizationStateChange(newState entities.AuthorizationStateType) {
	fmt.Printf("Received Authorization State Change: %s\n", newState)
	switch newState {
	case entities.AuthorizationStateTypeAwaitingTdlibParameters:
		err := c.client.SetTdlibParameters()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
	case entities.AuthorizationStateTypeAwaitingPhoneNumber:
		c.receivePhoneNumberFromChannel()
	case entities.AuthorizationStateTypeAwaitingCode:
		c.receiveCodeFromChannel()
	case entities.AuthorizationStateTypeAwaitingPassword:
		c.receivePasswordFromChannel()
	case entities.AuthorizationStateTypeAwaitingRegistration:
		err := c.client.RegisterUser("Ley", "Johnson")
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
	case entities.AuthorizationStateTypeReady:
		c.logToFile("Authorization State is Ready\n")

		err := c.client.LoadChats(nil, 10)
		if err != nil {
			panic(err)
		}

		me, err := c.client.GetMe()
		if err != nil {
			panic(err)
		}

		result, err := c.client.SendMessage(
			me.Id,
			0,
			nil,
			&entities.InputMessageText{
				Text: &entities.FormattedText{Text: "Hello World"},
			},
			nil)
		if err != nil {
			panic(err)
		}

		j, _ := json.Marshal(result)

		c.logToFile(fmt.Sprintf("Message Sent: %s\n", string(j)))

		fmt.Printf("Loading Chats...\n")
	default:
		fmt.Printf("Unhandled Authorization State: %s\n", newState)
	}
}

func (c *Client) logToFile(data string) {
	f, err := os.OpenFile("./tdlib/inside_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.WriteString(f, fmt.Sprintf("%s\n", data))
	if err != nil {
		panic(err)
	}
}

func (c *Client) receivePhoneNumberFromChannel() {
	fmt.Println("Submit Phone Number")
	phone, err := c.receiveInputFromQasedlu("Send Phone Number...")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		time.Sleep(time.Second * 5)
		c.receivePhoneNumberFromChannel()
		return
	}

	if phone == "STOP" {
		os.Exit(0)
		return
	}

	flashCall := false
	if strings.HasSuffix(phone, "_flash") {
		phone = phone[:strings.Index(phone, "_flash")]
		flashCall = true
	}

	err = c.client.SetAuthenticationPhoneNumber(phone, &entities.PhoneNumberAuthenticationSettings{AllowFlashCall: flashCall})
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		time.Sleep(time.Second * 5)
		c.receivePhoneNumberFromChannel()
		return
	}
}

func (c *Client) receiveInputFromQasedlu(message string) (string, error) {
	form := url.Values{}
	form.Set("user_id", "81997375")
	form.Set("message", fmt.Sprintf("From: %s\n%s", c.title, message))

	resp, err := http.Post("https://qasedlu.quantom.org/send_message", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (c *Client) onRawIncomingEvent(bytes []byte) {
	return
	// write to ./tdlib/updates.txt
	f, err := os.OpenFile("./tdlib/updates.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	_, err = f.Write(bytes)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	err = f.Close()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
}

func (c *Client) onError(errr incomingevents.ErrorEvent) {
	f, err := os.OpenFile("./tdlib/errors.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	_, err = f.Write(errr.Message)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	err = f.Close()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
}
