// Copyright (c) 2015 Spinpunch, Inc. All Rights Reserved.
// See License.txt for license information.

package zbox

import (
	"encoding/json"
	"github.com/mattermost/platform/einterfaces"
	"github.com/mattermost/platform/model"
	"io"
)

const (
	USER_AUTH_SERVICE_ZBOX = "zbox"
)

type ZBoxProvider struct {
}

type ZBoxUser struct {
	Id          int64  `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"firstname"`
	LastName    string `json:"lastname"`
	Team        string `json:"team"`
	Enabled     bool   `json:"isEnabled"`
	ChatEnabled bool   `json:"chatEnabled"`
}

func init() {
	provider := &ZBoxProvider{}
	einterfaces.RegisterOauthProvider(USER_AUTH_SERVICE_ZBOX, provider)
}

func userFromZBoxUser(zbu *ZBoxUser) *model.User {
	user := &model.User{}
	user.Username = model.CleanUsername(model.SetUsernameFromEmail(zbu.Email))
	user.FirstName = zbu.FirstName
	user.LastName = zbu.LastName
	user.Email = zbu.Email
	user.AuthData = zbu.Email
	user.AuthService = USER_AUTH_SERVICE_ZBOX

	return user
}

func ZBoxUserFromJson(data io.Reader) *ZBoxUser {
	decoder := json.NewDecoder(data)
	var zbu ZBoxUser
	err := decoder.Decode(&zbu)
	if err == nil {
		return &zbu
	} else {
		return nil
	}
}

func (zbu *ZBoxUser) getAuthData() string {
	return zbu.Email
}

func (m *ZBoxProvider) GetIdentifier() string {
	return USER_AUTH_SERVICE_ZBOX
}

func (m *ZBoxProvider) GetUserFromJson(data io.Reader) *model.User {
	return userFromZBoxUser(ZBoxUserFromJson(data))
}

func (m *ZBoxProvider) GetAuthDataFromJson(data io.Reader) string {
	return ZBoxUserFromJson(data).getAuthData()
}
