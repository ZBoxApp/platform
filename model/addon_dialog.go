package model

import (
	"encoding/json"
	"io"
)

type AddonDialog struct {
	Id                   string `json:"Id"`
	AddonId              string `json:"addon_id"`
	Key                  string `json:"key"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	PrimaryActionName    string `json:"primaryAction_name"`
	PrimaryActionKey     string `json:"primaryAction_key"`
	PrimaryActionEnabled bool   `json:"primaryAction_enabled,omitempty"`
	SecondaryActionName  string `json:"secondaryAction_name"`
	SecondaryActionKey   string `json:"secondaryAction_key"`
	Size                 string `json:"size"`
}

func (d *AddonDialog) ToJson() string {
	b, err := json.Marshal(d)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonDialogFromJson(data io.Reader) *AddonDialog {
	decoder := json.NewDecoder(data)
	var o AddonDialog
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func AddonDialogListToJson(l []*AddonDialog) string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonDialogListFromJson(data io.Reader) []*AddonDialog {
	decoder := json.NewDecoder(data)
	var o []*AddonDialog
	err := decoder.Decode(&o)
	if err == nil {
		return o
	} else {
		return nil
	}
}

func (d *AddonDialog) IsValid() *AppError {

	if len(d.Id) != 26 {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.invalid_id.app_error", nil, "")
	}

	if len(d.AddonId) != 26 {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.addon_id.app_error", nil, "")
	}

	if len(d.Title) == 0 || len(d.Title) > 32 {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.title.app_error", nil, "id="+d.Id)
	}

	if len(d.Key) > 64 || len(d.Key) == 0 {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.invalid_key.app_error", nil, "addon_id="+d.Id)
	}

	if len(d.URL) == 0 || !IsValidHttpUrl(d.URL) {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.invalid_url.app_error", nil, "")
	}

	if len(d.PrimaryActionName) > 32 || len(d.PrimaryActionName) == 0 {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.primary_name.app_error", nil, "addon_id="+d.Id)
	}

	if len(d.PrimaryActionKey) > 64 || len(d.PrimaryActionKey) == 0 {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.primary_key.app_error", nil, "addon_id="+d.Id)
	}

	if len(d.SecondaryActionName) != 0 && len(d.SecondaryActionName) > 32 {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.secondary_name.app_error", nil, "addon_id="+d.Id)
	}

	if len(d.SecondaryActionKey) != 0 && len(d.SecondaryActionKey) > 64 {
		return NewLocAppError("AddonDialog.IsValid", "model.addon_dialog.secondary_key.app_error", nil, "addon_id="+d.Id)
	}

	return nil
}

func (d *AddonDialog) PreSave() {
	if d.Id == "" {
		d.Id = NewId()
	}

	if len(d.Size) == 0 {
		d.Size = "large"
	}
}
