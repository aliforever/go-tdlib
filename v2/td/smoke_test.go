package td_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aliforever/go-tdlib/v2/td"
)

func TestMessageContentUnion(t *testing.T) {
	data := []byte(`{"@type":"messageText","text":{"@type":"formattedText","text":"hi","entities":[]}}`)

	var c td.MessageContent
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := c.Kind(); got != "messageText" {
		t.Fatalf("Kind() = %q, want messageText", got)
	}
	if !c.Known() {
		t.Fatal("Known() = false for a registered variant")
	}

	mt, ok := c.As[*td.MessageText]()
	if !ok {
		t.Fatalf("As[*MessageText] failed, value is %T", c.Value())
	}
	if mt.Text == nil || mt.Text.Text != "hi" {
		t.Fatalf("Text = %+v, want \"hi\"", mt.Text)
	}

	got := c.Match(td.MessageContentCases[string]{
		MessageText: func(v *td.MessageText) string { return v.Text.Text },
		Unknown:     func(kind string, raw json.RawMessage) string { return "unknown:" + kind },
	})
	if got != "hi" {
		t.Fatalf("Match = %q, want hi", got)
	}

	// Round-trip: marshaling the container re-injects @type.
	out, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"@type":"messageText"`) {
		t.Fatalf("marshal missing @type: %s", out)
	}
}

func TestMessageContentUnknownVariant(t *testing.T) {
	data := []byte(`{"@type":"messageHologram","text":"future"}`)
	var c td.MessageContent
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.Known() {
		t.Fatal("Known() = true for an unregistered variant")
	}
	if c.Kind() != "messageHologram" {
		t.Fatalf("Kind() = %q", c.Kind())
	}
	got := c.Match(td.MessageContentCases[string]{
		Unknown: func(kind string, raw json.RawMessage) string { return kind },
	})
	if got != "messageHologram" {
		t.Fatalf("Match unknown = %q", got)
	}
	// Unknown variants marshal back to their retained raw JSON.
	out, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != string(data) {
		t.Fatalf("marshal = %s, want %s", out, data)
	}
}

func TestMarshalRequest(t *testing.T) {
	b, err := td.MarshalRequest(td.GetMe{}, "req-1")
	if err != nil {
		t.Fatalf("MarshalRequest: %v", err)
	}
	if got := string(b); got != `{"@type":"getMe","@extra":"req-1"}` {
		t.Fatalf("got %s", got)
	}

	b, err = td.MarshalRequest(td.SendMessage{
		ChatId: 42,
		InputMessageContent: td.NewInputMessageContent(&td.InputMessageText{
			Text: &td.FormattedText{Text: "hello"},
		}),
	}, "")
	if err != nil {
		t.Fatalf("MarshalRequest sendMessage: %v", err)
	}
	s := string(b)
	for _, want := range []string{
		`"@type":"sendMessage"`, `"chat_id":42`,
		`"@type":"inputMessageText"`,                // union variant via container
		`"@type":"formattedText"`, `"text":"hello"`, // nested plain struct
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("sendMessage payload missing %s: %s", want, s)
		}
	}
}

func TestUpdateVariantTypes(t *testing.T) {
	types := td.UpdateVariantTypes()
	if len(types) == 0 {
		t.Fatal("UpdateVariantTypes is empty")
	}
	for i := 1; i < len(types); i++ {
		if types[i-1] >= types[i] {
			t.Fatalf("UpdateVariantTypes not sorted at %d: %q >= %q", i, types[i-1], types[i])
		}
	}
	found := false
	for _, typ := range types {
		if typ == "updateNewMessage" {
			found = true
		}
	}
	if !found {
		t.Fatal("updateNewMessage not in UpdateVariantTypes")
	}
}

// TDLib serializes TL int64 fields as JSON strings; int53 fields (also Go
// int64) stay plain numbers.
func TestInt64StringEncoding(t *testing.T) {
	// Scalar int64 arrives quoted; scalar int53 (User.Id) arrives as a number.
	data := []byte(`{"@type":"user","id":5748123456,"first_name":"A","last_name":"","usernames":null,"phone_number":"","status":{"@type":"userStatusEmpty"},"profile_photo":null,"accent_color_id":0,"background_custom_emoji_id":"5748384012345678901","upgraded_gift_colors":null,"profile_accent_color_id":-1,"profile_background_custom_emoji_id":"0","emoji_status":null,"is_contact":false,"is_mutual_contact":false,"is_close_friend":false,"verification_status":null,"is_premium":false,"type":{"@type":"userTypeRegular"},"language_code":"","added_to_attachment_menu":false}`)
	var u td.User
	if err := json.Unmarshal(data, &u); err != nil {
		t.Fatalf("unmarshal user: %v", err)
	}
	if u.Id != 5748123456 {
		t.Fatalf("Id = %d, want 5748123456 (int53, plain number)", u.Id)
	}
	if u.BackgroundCustomEmojiId != 5748384012345678901 {
		t.Fatalf("BackgroundCustomEmojiId = %d, want 5748384012345678901 (int64, quoted)", u.BackgroundCustomEmojiId)
	}
	if u.ProfileBackgroundCustomEmojiId != 0 {
		t.Fatalf("ProfileBackgroundCustomEmojiId = %d, want 0", u.ProfileBackgroundCustomEmojiId)
	}

	// vector<int64> arrives as an array of quoted values.
	var es td.EmojiStatusCustomEmojis
	if err := json.Unmarshal([]byte(`{"@type":"emojiStatusCustomEmojis","custom_emoji_ids":["5748","9000"]}`), &es); err != nil {
		t.Fatalf("unmarshal emojiStatusCustomEmojis: %v", err)
	}
	if len(es.CustomEmojiIds) != 2 || es.CustomEmojiIds[0] != 5748 || es.CustomEmojiIds[1] != 9000 {
		t.Fatalf("CustomEmojiIds = %v, want [5748 9000]", es.CustomEmojiIds)
	}

	// Outbound: scalar int64 and vector<int64> are both quoted.
	b, err := td.MarshalRequest(td.CreateTextCompositionStyle{CustomEmojiId: 5748384012345678901}, "")
	if err != nil {
		t.Fatalf("MarshalRequest: %v", err)
	}
	if !strings.Contains(string(b), `"custom_emoji_id":"5748384012345678901"`) {
		t.Fatalf("int64 field not quoted: %s", b)
	}

	b, err = td.MarshalRequest(td.ViewTrendingStickerSets{StickerSetIds: td.Int64List{1, 2}}, "")
	if err != nil {
		t.Fatalf("MarshalRequest: %v", err)
	}
	if !strings.Contains(string(b), `"sticker_set_ids":["1","2"]`) {
		t.Fatalf("vector<int64> elements not quoted: %s", b)
	}
}

// TDLib sends bytes fields base64-encoded, which matches encoding/json's
// default []byte handling.
func TestBytesBase64(t *testing.T) {
	b, err := td.MarshalRequest(td.SetTdlibParameters{DatabaseEncryptionKey: []byte{1, 2, 3}}, "")
	if err != nil {
		t.Fatalf("MarshalRequest: %v", err)
	}
	if !strings.Contains(string(b), `"database_encryption_key":"AQID"`) {
		t.Fatalf("bytes not base64: %s", b)
	}
	var req td.SetTdlibParameters
	if err := json.Unmarshal([]byte(`{"database_encryption_key":"AQID"}`), &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(req.DatabaseEncryptionKey) != 3 || req.DatabaseEncryptionKey[0] != 1 {
		t.Fatalf("DatabaseEncryptionKey = %v, want [1 2 3]", req.DatabaseEncryptionKey)
	}
}
