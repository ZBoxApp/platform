// Copyright (c) 2016 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

package model

type GuestMember struct {
	ChannelId     string `json:"channel_id"`
	UserId        string `json:"user_id"`
	Roles         string `json:"roles"`
	ChannelsCount int64  `json:"channel_count"`
}
