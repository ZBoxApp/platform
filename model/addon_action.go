package model

import (
	"encoding/json"
	"io"
)

type AddonAction struct {
	Id       string `json:"Id"`
	AddonId  string `json:"addon_id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Target   string `json:"target"`
	Location string `json:"location"`
}

func (a *AddonAction) ToJson() string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonActionFromJson(data io.Reader) *AddonAction {
	decoder := json.NewDecoder(data)
	var o AddonAction
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func AddonActionListToJson(l []*AddonAction) string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonActionListFromJson(data io.Reader) []*AddonAction {
	decoder := json.NewDecoder(data)
	var o []*AddonAction
	err := decoder.Decode(&o)
	if err == nil {
		return o
	} else {
		return nil
	}
}

func (a *AddonAction) IsValid() *AppError {

	if len(a.Id) != 26 {
		return NewLocAppError("AddonAction.IsValid", "model.addon_action.invalid_id.app_error", nil, "")
	}

	if len(a.AddonId) != 26 {
		return NewLocAppError("AddonAction.IsValid", "model.addon_action.addon_id.app_error", nil, "")
	}

	if len(a.Name) == 0 || len(a.Name) > 64 {
		return NewLocAppError("AddonAction.IsValid", "model.addon_action.invalid_name.app_error", nil, "id="+a.Id)
	}

	if len(a.Key) > 64 || len(a.Key) == 0 {
		return NewLocAppError("AddonAction.IsValid", "model.addon_action.invalid_key.app_error", nil, "addon_id="+a.Id)
	}

	if len(a.Target) == 0 {
		return NewLocAppError("AddonAction.IsValid", "model.addon_action.invalid_target.app_error", nil, "addon_id="+a.Id)
	}

	return nil
}

func (a *AddonAction) PreSave() {
	if a.Id == "" {
		a.Id = NewId()
	}

	if len(a.Location) == 0 {
		a.Location = "actions"
	}
}
