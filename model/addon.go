package model

import (
	"encoding/json"
	"io"
	"strings"
)

type Addon struct {
	Id                    string  `json:"id"`
	CreateAt              int64   `json:"create_at"`
	UpdateAt              int64   `json:"update_at"`
	DeleteAt              int64   `json:"delete_at"`
	Key                   string  `json:"key"`
	Name                  string  `json:"name"`
	IconURL               string  `json:"icon_url"`
	Description           string  `json:"description"`
	DescriptorURL         string  `json:"descriptor_url"`
	HomepageURL           string  `json:"homepage_url"`
	AllowGlobal           bool    `json:"allow_global"`
	AllowChannel          bool    `json:"allow_channel"`
	PublishedURL          string  `json:"published_url"`
	InstalledURL          string  `json:"installed_url"`
	UninstalledURL        string  `json:"uninstalled_url"`
	ConfigURL             string  `json:"config_url"`
	EnableIncomingWebhook bool    `json:"enable_incoming_webhook"`
	Price                 float64 `json:"price"`
	Currency              string  `json:"currency"`
	Enabled               bool    `json:"enabled"`
}

func (a *Addon) IsValid() *AppError {

	if len(a.Id) != 26 {
		return NewLocAppError("Addon.IsValid", "model.addon.invalid_id.app_error", nil, "")
	}

	if a.CreateAt == 0 {
		return NewLocAppError("Addon.IsValid", "model.addon.invalid_create_at.app_error", nil, "addon_id="+a.Id)
	}

	if a.UpdateAt == 0 {
		return NewLocAppError("Addon.IsValid", "model.addon.invalid_update_at.app_error", nil, "addon_id="+a.Id)
	}

	if len(a.Key) > 64 || len(a.Key) == 0 {
		return NewLocAppError("Addon.IsValid", "model.addon.invalid_key.app_error", nil, "addon_id="+a.Id)
	}

	if len(a.Name) > 64 || len(a.Key) == 0 {
		return NewLocAppError("Addon.IsValid", "model.addon.invalid_name.app_error", nil, "addon_id="+a.Id)
	}

	if len(a.Description) > 1024 {
		return NewLocAppError("Addon.IsValid", "model.addon.invalid_description.app_error", nil, "addon_id="+a.Id)
	}

	if len(a.DescriptorURL) == 0 || !IsValidHttpUrl(a.DescriptorURL) {
		return NewLocAppError("Addon.IsValid", "model.addon.descriptor_url.app_error", nil, "")
	}

	if len(a.PublishedURL) == 0 || !IsValidHttpUrl(a.PublishedURL) {
		return NewLocAppError("Addon.IsValid", "model.addon.published_url.app_error", nil, "")
	}

	if len(a.IconURL) == 0 && !IsValidHttpUrl(a.IconURL) {
		return NewLocAppError("Addon.IsValid", "model.addon.icon_url.app_error", nil, "")
	}

	if len(a.InstalledURL) == 0 || !IsValidHttpUrl(a.InstalledURL) {
		return NewLocAppError("Addon.IsValid", "model.addon.installed_url.app_error", nil, "")
	}

	if len(a.UninstalledURL) != 0 && !IsValidHttpUrl(a.UninstalledURL) {
		return NewLocAppError("Addon.IsValid", "model.addon.uninstalled_url.app_error", nil, "")
	}

	if len(a.ConfigURL) != 0 && !IsValidHttpUrl(a.ConfigURL) {
		return NewLocAppError("Addon.IsValid", "model.addon.config_url.app_error", nil, "")
	}

	if len(a.Currency) > 0 && len(a.Currency) != 3 {
		return NewLocAppError("Addon.IsValid", "model.addon.currency.app_error", nil, "")
	}

	return nil
}

func (a *Addon) PreSave() {
	if a.Id == "" {
		a.Id = NewId()
	}

	if len(a.Currency) > 0 {
		a.Currency = strings.ToUpper(a.Currency)
	}

	a.Enabled = true
	a.CreateAt = GetMillis()
	a.UpdateAt = a.CreateAt
}

func (a *Addon) PreUpdate() {
	if len(a.Currency) > 0 {
		a.Currency = strings.ToUpper(a.Currency)
	}

	a.UpdateAt = GetMillis()
}

func (a *Addon) ToJson() string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonFromJson(data io.Reader) *Addon {
	decoder := json.NewDecoder(data)
	var addon Addon
	err := decoder.Decode(&addon)
	if err == nil {
		return &addon
	} else {
		return nil
	}
}

func AddonMapToJson(a map[string]*Addon) string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonMapFromJson(data io.Reader) map[string]*Addon {
	decoder := json.NewDecoder(data)
	var addons map[string]*Addon
	err := decoder.Decode(&addons)
	if err == nil {
		return addons
	} else {
		return nil
	}
}
