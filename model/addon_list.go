package model

import (
	"encoding/json"
)

type AvailableAddon struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	IconURL     string  `json:"icon_url"`
	ConfigURL   string  `json:"config_url"`
	Currency    string  `json:"currency"`
	Price       float64 `json:"price"`
	Installed   bool    `json:"installed"`
}

type AddonList []*AvailableAddon

func (o *AddonList) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}
