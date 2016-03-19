package model

import (
	"encoding/json"
	"io"
)

type AddonWebpanel struct {
	Id       string `json:"Id"`
	AddonId  string `json:"addon_id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Location string `json:"location"`
	URL      string `json:"url"`
}

func (w *AddonWebpanel) ToJson() string {
	b, err := json.Marshal(w)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonWebpanelFromJson(data io.Reader) *AddonWebpanel {
	decoder := json.NewDecoder(data)
	var o AddonWebpanel
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func AddonWebpanelListToJson(l []*AddonWebpanel) string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonWebpanelListFromJson(data io.Reader) []*AddonWebpanel {
	decoder := json.NewDecoder(data)
	var o []*AddonWebpanel
	err := decoder.Decode(&o)
	if err == nil {
		return o
	} else {
		return nil
	}
}

func (w *AddonWebpanel) IsValid() *AppError {

	if len(w.Id) != 26 {
		return NewLocAppError("AddonWebpanel.IsValid", "model.addon_webpanel.invalid_id.app_error", nil, "")
	}

	if len(w.AddonId) != 26 {
		return NewLocAppError("AddonWebpanel.IsValid", "model.addon_webpanel.addon_id.app_error", nil, "")
	}

	if len(w.Name) == 0 || len(w.Name) > 64 {
		return NewLocAppError("AddonWebpanel.IsValid", "model.addon_webpanel.invalid_name.app_error", nil, "id="+w.Id)
	}

	if len(w.Key) > 64 || len(w.Key) == 0 {
		return NewLocAppError("AddonWebpanel.IsValid", "model.addon_webpanel.invalid_key.app_error", nil, "addon_id="+w.Id)
	}

	if len(w.URL) == 0 || !IsValidHttpUrl(w.URL) {
		return NewLocAppError("AddonWebpanel.IsValid", "model.addon_webpanel.invalid_url.app_error", nil, "")
	}

	return nil
}

func (w *AddonWebpanel) PreSave() {
	if w.Id == "" {
		w.Id = NewId()
	}

	if len(w.Location) == 0 {
		w.Location = "sidebar"
	}
}
