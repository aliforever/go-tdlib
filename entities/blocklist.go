package entities

type BlockList struct {
	Type string `json:"@type"`
}

func BlockListMain() *BlockList {
	return &BlockList{Type: "blockListMain"}
}

func BlockListStories() *BlockList {
	return &BlockList{Type: "blockListStories"}
}
