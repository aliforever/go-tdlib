# Hello-World Example
This is example will send hello world to authroized user (saved messages) 

## How to run
1. Message https://t.me/qasedlubot on Telegram
2. Replace `apiID` and `apiHash` in `client.go`
3. Replace your telegramID with `81997375` in `client.go`
4. Run ```docker build -t hello-world-tdlib .```
5. Run ```docker run hello-world-tdlib```
6. Check and reply to messages from qsedlubot on Telegram regarding authorization