package model

import (
	"encoding/json"
	"io"
)

type TeamAddon struct {
	Id          string `json:"id"`
	InstalledAt int64  `json:"installed_at"`
	TeamId      string `json:"team_id"`
	AddonId     string `json:"addon_id"`
	Enabled     bool   `json:"enabled"`
}

func (a *TeamAddon) IsValid() *AppError {

	if len(a.Id) != 26 {
		return NewLocAppError("TeamAddon.IsValid", "model.team_addon.team_addon_id.app_error", nil, "")
	}

	if a.InstalledAt == 0 {
		return NewLocAppError("TeamAddon.IsValid", "model.team_addon.installed.app_error", nil, "team_addon_id="+a.Id)
	}

	if len(a.AddonId) != 26 {
		return NewLocAppError("TeamAddon.IsValid", "model.team_addon.addon_id.app_error", nil, "")
	}

	if len(a.TeamId) != 26 {
		return NewLocAppError("TeamAddon.IsValid", "model.team_addon.team_id.app_error", nil, "")
	}

	return nil
}

func (a *TeamAddon) PreSave() {
	if a.Id == "" {
		a.Id = NewId()
	}

	a.Enabled = true
	a.InstalledAt = GetMillis()
}

func (a *TeamAddon) ToJson() string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func TeamAddonFromJson(data io.Reader) *TeamAddon {
	decoder := json.NewDecoder(data)
	var addon TeamAddon
	err := decoder.Decode(&addon)
	if err == nil {
		return &addon
	} else {
		return nil
	}
}

func TeamAddonMapToJson(a map[string]*TeamAddon) string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func TeamAddonMapFromJson(data io.Reader) map[string]*TeamAddon {
	decoder := json.NewDecoder(data)
	var addons map[string]*TeamAddon
	err := decoder.Decode(&addons)
	if err == nil {
		return addons
	} else {
		return nil
	}
}
