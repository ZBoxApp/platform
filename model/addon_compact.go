package model

import (
	"encoding/json"
	"io"
)

type AddonCompact struct {
	Id        string `json:"Id"`
	AddonId   string `json:"addon_id"`
	Name      string `json:"name"`
	QueryURL  string `json:"query_url"`
	Key       string `json:"key"`
	Target    string `json:"target"`
	IconURL   string `json:"icon_url"`
	IconURL2x string `json:"icon_url@2x"`
}

func (c *AddonCompact) ToJson() string {
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonCompactFromJson(data io.Reader) *AddonCompact {
	decoder := json.NewDecoder(data)
	var o AddonCompact
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func AddonCompactListToJson(l []*AddonCompact) string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonCompactListFromJson(data io.Reader) []*AddonCompact {
	decoder := json.NewDecoder(data)
	var o []*AddonCompact
	err := decoder.Decode(&o)
	if err == nil {
		return o
	} else {
		return nil
	}
}

func (c *AddonCompact) IsValid() *AppError {

	if len(c.Id) != 26 {
		return NewLocAppError("AddonCompact.IsValid", "model.addon_compact.invalid_id.app_error", nil, "")
	}

	if len(c.AddonId) != 26 {
		return NewLocAppError("AddonCompact.IsValid", "model.addon_compact.addon_id.app_error", nil, "")
	}

	if len(c.Name) == 0 || len(c.Name) > 64 {
		return NewLocAppError("AddonCompact.IsValid", "model.addon_compact.invalid_name.app_error", nil, "id="+c.Id)
	}

	if len(c.QueryURL) == 0 || !IsValidHttpUrl(c.QueryURL) {
		return NewLocAppError("AddonCompact.IsValid", "model.addon_compact.query_url.app_error", nil, "")
	}

	if len(c.Key) > 64 || len(c.Key) == 0 {
		return NewLocAppError("AddonCompact.IsValid", "model.addon_compact.invalid_key.app_error", nil, "addon_id="+c.Id)
	}

	if len(c.IconURL) == 0 || !IsValidHttpUrl(c.IconURL) {
		return NewLocAppError("AddonCompact.IsValid", "model.addon_compact.icon_url.app_error", nil, "")
	}

	if len(c.Target) == 0 {
		return NewLocAppError("AddonCompact.IsValid", "model.addon_compact.invalid_target.app_error", nil, "addon_id="+c.Id)
	}

	return nil
}

func (c *AddonCompact) PreSave() {
	if c.Id == "" {
		c.Id = NewId()
	}

	if len(c.IconURL2x) == 0 {
		c.IconURL2x = c.IconURL
	}
}
