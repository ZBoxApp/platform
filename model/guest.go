// Copyright (c) 2016 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

type Guest struct {
	Id              string `json:"id"`
	CreateAt        int64  `json:"create_at,omitempty"`
	UpdateAt        int64  `json:"update_at,omitempty"`
	DeleteAt        int64  `json:"delete_at"`
	ChannelId       string `json:"channel_id"`
	InviteId        string `json:"invite_id"`
	TotalGuestCount int64  `json:"total_guest_count"`
	TotalUsedCount  int64  `json:"total_used_count"`
}

// IsValid validates the user and returns an error if it isn't configured
// correctly.
func (g *Guest) IsValid() *AppError {

	if len(g.Id) != 26 {
		return NewLocAppError("Guest.IsValid", "model.guest.is_valid.id.app_error", nil, "")
	}

	if g.CreateAt == 0 {
		return NewLocAppError("Guest.IsValid", "model.guest.is_valid.create_at.app_error", nil, "user_id="+g.Id)
	}

	if g.UpdateAt == 0 {
		return NewLocAppError("Guest.IsValid", "model.guest.is_valid.update_at.app_error", nil, "user_id="+g.Id)
	}

	if len(g.ChannelId) != 26 {
		return NewLocAppError("Guest.IsValid", "model.guest.is_valid.channel_id.app_error", nil, "")
	}

	return nil
}

// PreSave will set the Id and Invite Id if missing.  It will also fill
// in the CreateAt, UpdateAt times.  It should
// be run before saving the guest to the db.
func (g *Guest) PreSave() {
	if g.Id == "" {
		g.Id = NewId()
	}

	if g.InviteId == "" {
		g.InviteId = NewId()
	}

	g.CreateAt = GetMillis()
	g.UpdateAt = g.CreateAt
	g.DeleteAt = 0
}

// PreUpdate should be run before updating the guest in the db.
func (g *Guest) PreUpdate() {
	g.UpdateAt = GetMillis()
}

// ToJson convert a Guest to a json string
func (g *Guest) ToJson() string {
	b, err := json.Marshal(g)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func GuestFromJson(data io.Reader) *Guest {
	decoder := json.NewDecoder(data)
	var o Guest
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

// Generate a valid strong etag so the browser can cache the results
func (g *Guest) Etag() string {
	return Etag(g.Id, g.UpdateAt)
}
