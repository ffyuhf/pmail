package response

import (
	"encoding/json"

	"github.com/ffyuhf/pmail/models"
)

type EmailResponseData struct {
	models.Email `xorm:"extends"`
	IsRead       int8 `json:"is_read"`
	SerialNumber int  `json:"serial_number"`
	UeId         int  `json:"ue_id"`
}

type UserEmailUIDData struct {
	models.UserEmail `xorm:"extends"`
	SerialNumber     int `json:"serial_number"`
}

// MarshalJSON 覆盖 Email 的方法提升，复用 Email 的序列化结果并补充 EmailResponseData 独有字段。
// 修复日期: 20260516 — 根因：Email.MarshalJSON 被提升导致 UeId/IsRead/SerialNumber 在 JSON 输出中丢失
func (e EmailResponseData) MarshalJSON() ([]byte, error) {
	// 复用 Email 的自定义序列化（日期格式化、附件处理等）
	baseBytes, err := json.Marshal(&e.Email)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(baseBytes, &m); err != nil {
		return nil, err
	}
	// 补充 EmailResponseData 独有字段
	m["is_read"] = e.IsRead
	m["serial_number"] = e.SerialNumber
	m["ue_id"] = e.UeId
	return json.Marshal(m)
}
