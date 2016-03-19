package model

import (
	"encoding/json"
	"io"
)

type AddonCommand struct {
	Id               string `json:"Id"`
	AddonId          string `json:"addon_id"`
	Trigger          string `json:"trigger"`
	Method           string `json:"method"`
	Username         string `json:"username"`
	IconURL          string `json:"icon_url"`
	AutoComplete     bool   `json:"auto_complete"`
	AutoCompleteDesc string `json:"auto_complete_desc"`
	AutoCompleteHint string `json:"auto_complete_hint"`
	DisplayName      string `json:"display_name"`
	URL              string `json:"url"`
}

func (a *AddonCommand) ToJson() string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonCommandFromJson(data io.Reader) *AddonCommand {
	decoder := json.NewDecoder(data)
	var o AddonCommand
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func AddonCommandListToJson(l []*AddonCommand) string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonCommandListFromJson(data io.Reader) []*AddonCommand {
	decoder := json.NewDecoder(data)
	var o []*AddonCommand
	err := decoder.Decode(&o)
	if err == nil {
		return o
	} else {
		return nil
	}
}

func (a *AddonCommand) IsValid() *AppError {

	if len(a.Id) != 26 {
		return NewLocAppError("AddonCommand.IsValid", "model.addon_command.invalid_id.app_error", nil, "")
	}

	if len(a.AddonId) != 26 {
		return NewLocAppError("AddonCommand.IsValid", "model.addon_command.addon_id.app_error", nil, "")
	}

	if len(a.Trigger) > 1024 {
		return NewLocAppError("AddonCommand.IsValid", "model.command.is_valid.trigger.app_error", nil, "")
	}

	if len(a.URL) == 0 || len(a.URL) > 1024 {
		return NewLocAppError("AddonCommand.IsValid", "model.command.is_valid.url.app_error", nil, "")
	}

	if !IsValidHttpUrl(a.URL) {
		return NewLocAppError("AddonCommand.IsValid", "model.command.is_valid.url_http.app_error", nil, "")
	}

	if !(a.Method == COMMAND_METHOD_GET || a.Method == COMMAND_METHOD_POST) {
		return NewLocAppError("AddonCommand.IsValid", "model.command.is_valid.method.app_error", nil, "")
	}

	return nil
}

func (a *AddonCommand) PreSave() {
	if a.Id == "" {
		a.Id = NewId()
	}
}
