// Copyright (c) 2016 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

package api

import (
	"net/http"

	l4g "github.com/alecthomas/log4go"
	"github.com/gorilla/mux"

	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/utils"
	"strings"
)

func InitGuest(r *mux.Router) {
	l4g.Debug(utils.T("api.guest.init.debug"))

	sr := r.PathPrefix("/guest").Subrouter()
	sr.Handle("/{id:[A-Za-z0-9]+}/", ApiUserRequiredActivity(getInvitation, false)).Methods("GET")
	sr.Handle("/create", ApiUserRequired(createInvitation)).Methods("POST")
	sr.Handle("/{id:[A-Za-z0-9]+}/remove", ApiUserRequired(removeInvitation)).Methods("POST")
}

func getInvitation(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	gchan := Srv.Store.Guest().GetByChannelId(id)

	if result := <-gchan; result.Err != nil {
		c.Err = result.Err
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
		c.Err = model.NewLocAppError("createDirectChannel", "api.guest.create_invitation.direct_channel.app_error", nil, "")
		return
	}

	scm := Srv.Store.Channel().GetMember(guest.ChannelId, c.Session.UserId)
	if scmresult := <-scm; scmresult.Err != nil {
		c.Err = scmresult.Err
		return
	} else {
		channelMember := scmresult.Data.(model.ChannelMember)

		if !c.IsTeamAdmin() || !strings.Contains(channelMember.Roles, model.CHANNEL_ROLE_ADMIN) {
			c.Err = model.NewLocAppError("deleteChannel", "api.guest.create_invitation.permissions.app_error", nil, "")
			c.Err.StatusCode = http.StatusForbidden
			return
		}

		if sg, err := CreateInvitation(c, guest); err != nil {
			c.Err = err
			return
		} else {
			w.Write([]byte(sg.ToJson()))
		}
	}
}

func removeInvitation(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	if sgresult := <-Srv.Store.Guest().GetByInviteId(id); sgresult.Err != nil {
		c.Err = sgresult.Err
		return
	} else {
		guest := sgresult.Data.(*model.Guest)

		cchan := Srv.Store.Channel().CheckPermissionsTo(c.Session.TeamId, guest.ChannelId, c.Session.UserId)
		if !c.HasPermissionsToChannel(cchan, "removeInvitation") {
			return
		}

		scm := Srv.Store.Channel().GetMember(id, c.Session.UserId)
		if scmresult := <-scm; scmresult.Err != nil {
			c.Err = scmresult.Err
			return
		} else {
			channelMember := scmresult.Data.(model.ChannelMember)

			if !c.IsTeamAdmin() || !strings.Contains(channelMember.Roles, model.CHANNEL_ROLE_ADMIN) {
				c.Err = model.NewLocAppError("deleteChannel", "api.guest.create_invitation.permissions.app_error", nil, "")
				c.Err.StatusCode = http.StatusForbidden
				return
			}

			if guest.DeleteAt > 0 {
				c.Err = model.NewLocAppError("deleteChannel", "api.guest.remove_invitation.deleted.app_error", nil, "")
				c.Err.StatusCode = http.StatusBadRequest
				return
			}

			guestMembers := Srv.Store.Guest().GetMembers(id)

			if gmresults := <-guestMembers; gmresults.Err != nil {
				c.Err = gmresults.Err
				return
			} else {
				members := gmresults.Data.(*[]model.GuestMember)
				<-Srv.Store.Guest().RemoveMembers(id, members)

				userIds := make([]string, len(*members))

				for _, member := range *members {
					if model.IsInRole(member.Roles, model.ROLE_GUEST_USER) {
						userIds = append(userIds, member.UserId)
					}
				}
				<-Srv.Store.Guest().RemoveUsers(userIds)

				if result := <-Srv.Store.Guest().Delete(id, model.GetMillis()); result.Err != nil {
					c.Err = result.Err
					return
				} else {
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
