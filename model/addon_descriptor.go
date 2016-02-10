package model

import (
	"encoding/json"
	"io"
)

type AddonDescriptorVendor struct {
	Name    string `json:"name"`
	IconURL string `json:"icon_url"`
}

type AddonDescriptorPrice struct {
	Currency string  `json:"currency,omitempty"`
	Value    float64 `json:"value,omitempty"`
}

type AddonDescriptorLinks struct {
	PublishedURL  string `json:"published_url"`
	DescriptorURL string `json:"descriptor_url"`
	HomepageURL   string `json:"homepage_url"`
}

type AddonDescriptorInstallable struct {
	InstalledURL   string `json:"installed_url"`
	UninstalledURL string `json:"uninstalled_url"`
	ConfigURL      string `json:"config_url"`
	AllowGlobal    bool   `json:"allow_global,omitempty"`
	AllowChannel   bool   `json:"allow_channel,omitempty"`
}

type AddonDescriptorOutgoinWebhook struct {
	Key          string   `json:"key"`
	CallbackURLs []string `json:"callback_urls"`
	Triggers     []string `json:"triggers"`
}

type AddonDescriptorWebhook struct {
	EnableIncoming bool                            `json:"enable_incoming,omitempty"`
	Outgoing       []AddonDescriptorOutgoinWebhook `json:"outgoing"`
}

type AddonDescriptorCompactIcon struct {
	URL   string `json:"url"`
	URL2x string `json:"url@2x"`
}

type AddonDescriptorCompact struct {
	Name     string                     `json:"name"`
	QueryURL string                     `json:"query_url"`
	Key      string                     `json:"key"`
	Target   string                     `json:"target"`
	Icon     AddonDescriptorCompactIcon `json:"icon"`
}

type AddonDescriptorWebPanel struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Location string `json:"location"`
	URL      string `json:"url"`
}

type AddonDescriptorDialogOptionsAction struct {
	Name    string `json:"name"`
	Key     string `json:"key"`
	Enabled bool   `json:"enabled,omitempty"`
}

type AddonDescriptorDialogOptions struct {
	Style           string                             `json:"style"`
	PrimaryAction   AddonDescriptorDialogOptionsAction `json:"primaryAction"`
	SecondaryAction AddonDescriptorDialogOptionsAction `json:"secondaryAction"`
	Size            string                             `json:"size"`
}

type AddonDescriptorDialog struct {
	Key     string                       `json:"key"`
	Title   string                       `json:"title"`
	URL     string                       `json:"url"`
	Options AddonDescriptorDialogOptions `json:"options"`
}

type AddonDescriptorAction struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Target   string `json:"target"`
	Location string `json:"location"`
}

type AddonDescriptor struct {
	Key         string                     `json:"key"`
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	Category    string                     `json:"category"`
	Vendor      AddonDescriptorVendor      `json:"vendor"`
	Price       AddonDescriptorPrice       `json:"price"`
	Links       AddonDescriptorLinks       `json:"links"`
	Installable AddonDescriptorInstallable `json:"installable"`
	Webhook     AddonDescriptorWebhook     `json:"webhook"`
	Compact     []AddonDescriptorCompact   `json:"compact"`
	Webpanel    []AddonDescriptorWebPanel  `json:"webPanel"`
	Dialog      []AddonDescriptorDialog    `json:"dialog"`
	Action      []AddonDescriptorAction    `json:"action"`
}

func (a *AddonDescriptor) ToJson() string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonDescriptorFromJson(data io.Reader) *AddonDescriptor {
	decoder := json.NewDecoder(data)
	var addon AddonDescriptor
	err := decoder.Decode(&addon)
	if err == nil {
		return &addon
	} else {
		return nil
	}
}

func AddonDescriptorMapToJson(a map[string]*AddonDescriptor) string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func AddonDescriptorMapFromJson(data io.Reader) map[string]*AddonDescriptor {
	decoder := json.NewDecoder(data)
	var addons map[string]*AddonDescriptor
	err := decoder.Decode(&addons)
	if err == nil {
		return addons
	} else {
		return nil
	}
}
