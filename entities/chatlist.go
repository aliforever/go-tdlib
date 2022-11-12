package entities

var (
	ChatListMain = &ChatList{
		Type: "chatListMain",
	}
	ChatListArchive = &ChatList{
		Type: "chatListArchive",
	}
	ChatListFilter = &ChatList{
		Type: "chatListFilter",
	}
)

type ChatList struct {
	Type string `json:"@type"`
}
