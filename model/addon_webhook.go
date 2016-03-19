package model

import (
	"encoding/json"
	"fmt"
	"io"
)

type AddonWebhook struct {
	Id           string      `json:"id"`
	Key          string      `json:"key"`
	AddonId      string      `json:"addon_id"`
	Triggers     StringArray `json:"triggers"`
	CallbackURLs StringArray `json:"callback_urls"`
}

func (o *AddonWebhook) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonWebhookFromJson(data io.Reader) *AddonWebhook {
	decoder := json.NewDecoder(data)
	var o AddonWebhook
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func AddonWebhookListToJson(l []*AddonWebhook) string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonWebhookListFromJson(data io.Reader) []*AddonWebhook {
	decoder := json.NewDecoder(data)
	var o []*AddonWebhook
	err := decoder.Decode(&o)
	if err == nil {
		return o
	} else {
		return nil
	}
}

func (o *AddonWebhook) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewLocAppError("AddonWebhook.IsValid", "model.addon_webhook.invalid_id.app_error", nil, "")
	}

	if len(o.AddonId) != 26 {
		return NewLocAppError("AddonWebhook.IsValid", "model.addon_webhook.addon_id.app_error", nil, "")
	}

	if len(o.Key) > 64 || len(o.Key) == 0 {
		return NewLocAppError("AddonWebhook.IsValid", "model.addon_webhook.invalid_key.app_error", nil, "addon_id="+o.AddonId)
	}

	if len(fmt.Sprintf("%s", o.Triggers)) > 1024 {
		return NewLocAppError("AddonWebhook.IsValid", "model.addon_webhook.trigger_words.app_error", nil, "")
	}

	if len(o.CallbackURLs) == 0 || len(fmt.Sprintf("%s", o.CallbackURLs)) > 1024 {
		return NewLocAppError("AddonWebhook.IsValid", "model.addon_webhook.callbacks.app_error", nil, "")
	}

	for _, callback := range o.CallbackURLs {
		if !IsValidHttpUrl(callback) {
			return NewLocAppError("AddonWebhook.IsValid", "model.addon_webhook.callbacks_url.app_error", nil, "")
		}
	}

	return nil
}

func (o *AddonWebhook) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}
}
