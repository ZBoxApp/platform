// Copyright (c) 2016 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

package api

import (
	"fmt"
	"net/http"

	l4g "github.com/alecthomas/log4go"
	"github.com/gorilla/mux"

	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/store"
	"github.com/mattermost/platform/utils"
	"strings"
)

func InitGuest(r *mux.Router) {
	l4g.Debug(utils.T("api.guest.init.debug"))

	sr := r.PathPrefix("/guest").Subrouter()
	sr.Handle("/create", ApiUserRequired(createInvitation)).Methods("POST")
	sr.Handle("/new", ApiAppHandler(createGuest)).Methods("POST")
	sr.Handle("/login", ApiAppHandler(loginGuest)).Methods("POST")
	sr.Handle("/{id:[A-Za-z0-9]+}/", ApiUserRequiredActivity(getInvitation, false)).Methods("GET")
	sr.Handle("/{id:[A-Za-z0-9]+}/remove", ApiUserRequired(removeInvitation)).Methods("POST")
}

func getInvitation(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	gchan := Srv.Store.Guest().GetByChannelId(id)

	if result := <-gchan; result.Err != nil {
		c.Err = result.Err
		c.Err.StatusCode = http.StatusNoContent
		return
	} else {
		data := result.Data.(*model.Guest)

		if HandleEtag(data.Etag(), w, r) {
			return
		} else {
			w.Header().Set(model.HEADER_ETAG_SERVER, data.Etag())
			w.Header().Set("Expires", "-1")
			w.Write([]byte(data.ToJson()))
		}
	}
}

func createInvitation(c *Context, w http.ResponseWriter, r *http.Request) {
	guest := model.GuestFromJson(r.Body)

	if guest == nil {
		c.SetInvalidParam("createInvitation", "guest")
		return
	}

	cchan := Srv.Store.Channel().CheckPermissionsTo(c.Session.TeamId, guest.ChannelId, c.Session.UserId)
	sc := Srv.Store.Channel().Get(guest.ChannelId)

	if !c.HasPermissionsToChannel(cchan, "createInvitation") {
		return
	}

	var channel *model.Channel

	if result := <-sc; result.Err != nil {
		c.Err = result.Err
		return
	} else {
		channel = result.Data.(*model.Channel)
	}

	if channel.Type == model.CHANNEL_DIRECT {
		c.Err = model.NewLocAppError("createInvitation", "api.guest.create_invitation.direct_channel.app_error", nil, "")
		return
	}

	scm := Srv.Store.Channel().GetMember(guest.ChannelId, c.Session.UserId)
	if scmresult := <-scm; scmresult.Err != nil {
		c.Err = scmresult.Err
		return
	} else {
		channelMember := scmresult.Data.(model.ChannelMember)

		if !c.IsTeamAdmin() && !strings.Contains(channelMember.Roles, model.CHANNEL_ROLE_ADMIN) {
			c.Err = model.NewLocAppError("createInvitation", "api.guest.create_invitation.permissions.app_error", nil, "")
			c.Err.StatusCode = http.StatusForbidden
			return
		}

		if sg, err := CreateInvitation(c, guest); err != nil {
			c.Err = err
			return
		} else {
			if result := <-Srv.Store.User().Get(c.Session.UserId); result.Err == nil {
				user := result.Data.(*model.User)
				PostUserAddRemoveMessageAndForget(c, channel.Id, fmt.Sprintf(c.T("api.guest.create_guest_invitation.post_and_forget"), user.Username))
			}
			w.Write([]byte(sg.ToJson()))
		}
	}
}

func removeInvitation(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	var channel *model.Channel

	if sgresult := <-Srv.Store.Guest().GetByInviteId(id); sgresult.Err != nil {
		c.Err = sgresult.Err
		return
	} else {
		guest := sgresult.Data.(*model.Guest)

		sc := Srv.Store.Channel().Get(guest.ChannelId)
		if result := <-sc; result.Err != nil {
			c.Err = result.Err
			return
		} else {
			channel = result.Data.(*model.Channel)
		}

		cchan := Srv.Store.Channel().CheckPermissionsTo(c.Session.TeamId, guest.ChannelId, c.Session.UserId)
		if !c.HasPermissionsToChannel(cchan, "removeInvitation") {
			return
		}

		scm := Srv.Store.Channel().GetMember(guest.ChannelId, c.Session.UserId)
		if scmresult := <-scm; scmresult.Err != nil {
			c.Err = scmresult.Err
			return
		} else {
			channelMember := scmresult.Data.(model.ChannelMember)

			if !c.IsTeamAdmin() && !strings.Contains(channelMember.Roles, model.CHANNEL_ROLE_ADMIN) {
				c.Err = model.NewLocAppError("deleteChannel", "api.guest.create_invitation.permissions.app_error", nil, "")
				c.Err.StatusCode = http.StatusForbidden
				return
			}

			if guest.DeleteAt > 0 {
				c.Err = model.NewLocAppError("deleteChannel", "api.guest.remove_invitation.deleted.app_error", nil, "")
				c.Err.StatusCode = http.StatusBadRequest
				return
			}

			guestMembers := Srv.Store.Guest().GetMembers(guest.ChannelId)

			if gmresults := <-guestMembers; gmresults.Err != nil {
				c.Err = gmresults.Err
				return
			} else {
				members := gmresults.Data.([]*model.GuestMember)
				userIds := make([]string, 0)

				for _, member := range members {
					if model.IsInRole(member.Roles, model.ROLE_GUEST_USER) {
						RemoveUserFromChannel(member.UserId, c.Session.UserId, channel)
						RevokeAllSession(c, member.UserId)
						if member.ChannelsCount == 1 {
							userIds = append(userIds, member.UserId)
							<-Srv.Store.Preference().PermanentDeleteByUser(member.UserId)
						}
					}
				}

				if len(userIds) > 0 {
					<-Srv.Store.Guest().RemoveUsers(userIds)
				}

				if result := <-Srv.Store.Guest().Delete(guest.ChannelId, model.GetMillis()); result.Err != nil {
					c.Err = result.Err
					return
				} else {
					if result := <-Srv.Store.User().Get(c.Session.UserId); result.Err == nil {
						user := result.Data.(*model.User)
						PostUserAddRemoveMessageAndForget(c, channel.Id, fmt.Sprintf(c.T("api.guest.remove_guest_invitation.post_and_forget"), user.Username))
					}
					result := make(map[string]string)
					result["id"] = guest.InviteId
					w.Write([]byte(model.MapToJson(result)))
				}
			}
		}
	}

}

func CreateInvitation(c *Context, guest *model.Guest) (*model.Guest, *model.AppError) {
	if result := <-Srv.Store.Guest().Save(guest); result.Err != nil {
		return nil, result.Err
	} else {
		sg := result.Data.(*model.Guest)
		return sg, nil
	}
}

func createGuest(c *Context, w http.ResponseWriter, r *http.Request) {
	user := model.UserFromJson(r.Body)

	if user == nil {
		c.SetInvalidParam("createGuest", "user")
		return
	}

	// the user's username is checked to be valid when they are saved to the database

	data := r.URL.Query().Get("d")
	props := model.MapFromJson(strings.NewReader(data))

	var guest *model.Guest
	var channel *model.Channel

	// validar que la invitación siga vigente
	if result := <-Srv.Store.Guest().GetByInviteId(props["invite_id"]); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		guest = result.Data.(*model.Guest)
	}

	if result := <-Srv.Store.Channel().Get(props["channel_id"]); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		channel = result.Data.(*model.Channel)
	}

	if user.TeamId != props["id"] {
		c.Err = model.NewLocAppError("createGuest", "api.user.create_user.team_name.app_error", nil, data)
		return
	}

	user.EmailVerified = true
	user.Password = model.ROLE_GUEST_USER
	user.Roles = model.ROLE_GUEST_USER

	ruser, err := CreateGuest(c, guest, channel, user)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(ruser.ToJson()))

}

func CreateGuest(c *Context, guest *model.Guest, channel *model.Channel, user *model.User) (*model.User, *model.AppError) {

	user.MakeNonNil()

	var ruser *model.User

	if result := <-Srv.Store.User().GetByEmail(user.TeamId, user.Email); result.Err != nil {
		if result := <-Srv.Store.User().Save(user); result.Err != nil {
			l4g.Error(utils.T("api.user.create_user.save.error"), result.Err)
			return nil, result.Err
		} else {
			UpdateGuestCounts(guest, guest.TotalGuestCount+1)

			ruser = result.Data.(*model.User)

			mark := model.Preference{UserId: ruser.Id, Category: model.PREFERENCE_CATEGORY_ADVANCED_SETTINGS, Name: store.FEATURE_TOGGLE_PREFIX + "markdown_preview", Value: "true"}
			embed := model.Preference{UserId: ruser.Id, Category: model.PREFERENCE_CATEGORY_ADVANCED_SETTINGS, Name: store.FEATURE_TOGGLE_PREFIX + "embed_preview", Value: "true"}
			if presult := <-Srv.Store.Preference().Save(&model.Preferences{mark, embed}); presult.Err != nil {
				l4g.Error(utils.T("api.user.create_user.tutorial.error"), presult.Err.Message)
			}

			if _, err := AddUserToChannel(ruser, channel); err != nil {
				return nil, err
			}
			mockSession := model.Session{UserId: ruser.Id, TeamId: ruser.TeamId, IsOAuth: false}
			newContext := &Context{mockSession, model.NewId(), "", c.Path, nil, c.teamURLValid, c.teamURL, c.siteURL, 0, c.T, model.DEFAULT_LOCALE}
			PostUserAddRemoveMessageAndForget(newContext, channel.Id, fmt.Sprintf(newContext.T("api.guest.join_channel.post_and_forget"), ruser.Username))
		}
	} else {
		UpdateGuestCounts(guest, guest.TotalGuestCount)

		ruser = result.Data.(*model.User)
		if result := <-Srv.Store.Channel().GetMember(channel.Id, ruser.Id); result.Err != nil {
			if _, err := AddUserToChannel(ruser, channel); err != nil {
				return nil, err
			}
			mockSession := model.Session{UserId: ruser.Id, TeamId: ruser.TeamId, IsOAuth: false}
			newContext := &Context{mockSession, model.NewId(), "", c.Path, nil, c.teamURLValid, c.teamURL, c.siteURL, 0, c.T, model.DEFAULT_LOCALE}
			PostUserAddRemoveMessageAndForget(newContext, channel.Id, fmt.Sprintf(newContext.T("api.guest.join_channel.post_and_forget"), ruser.Username))
		}
	}

	ruser.Sanitize(map[string]bool{})
	return ruser, nil
}

func loginGuest(c *Context, w http.ResponseWriter, r *http.Request) {
	props := model.MapFromJson(r.Body)

	if len(props["inviteId"]) == 0 {
		c.Err = model.NewLocAppError("login", "api.user.login.blank_invite.app_error", nil, "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	var user *model.User
	if len(props["email"]) != 0 && len(props["name"]) != 0 {
		user = LoginGuest(c, w, r, props["email"], props["name"], props["inviteId"], "")
	} else {
		c.Err = model.NewLocAppError("login", "api.user.login.not_provided.app_error", nil, "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	if c.Err != nil {
		return
	}

	if user != nil {
		user.Sanitize(map[string]bool{})
	} else {
		user = &model.User{}
	}
	w.Write([]byte(user.ToJson()))
}

func LoginGuest(c *Context, w http.ResponseWriter, r *http.Request, email, name, inviteId, deviceId string) *model.User {
	var team *model.Team

	if result := <-Srv.Store.Team().GetByName(name); result.Err != nil {
		c.Err = result.Err
		return nil
	} else {
		team = result.Data.(*model.Team)
	}

	if result := <-Srv.Store.User().GetByEmail(team.Id, email); result.Err != nil {
		c.Err = result.Err
		c.Err.StatusCode = http.StatusForbidden
		return nil
	} else {
		user := result.Data.(*model.User)

		if result := <-Srv.Store.Guest().GetByInviteId(inviteId); result.Err != nil {
			c.Err = result.Err
			c.Err.StatusCode = http.StatusForbidden
			return nil
		} else {
			guest := result.Data.(*model.Guest)

			if result := <-Srv.Store.Channel().GetMember(guest.ChannelId, user.Id); result.Err != nil {
				c.Err = result.Err
				c.Err.StatusCode = http.StatusForbidden
				return nil
			}
		}

		Login(c, w, r, user, deviceId)
		return user
	}

	return nil
}

func UpdateGuestCounts(guest *model.Guest, totalGuestCount int64) {
	guest.TotalGuestCount = totalGuestCount
	<-Srv.Store.Guest().Update(guest)
}
