package entities

type ChatType struct {
	Type         string `json:"@type"`
	SupergroupId int64  `json:"supergroup_id"`
	IsChannel    bool   `json:"is_channel"`
}
