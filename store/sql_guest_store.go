// Copyright (c) 2016 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

package store

import (
	"github.com/mattermost/platform/model"
	"strconv"
	"strings"
)

type SqlGuestStore struct {
	*SqlStore
}

func NewSqlGuestStore(sqlStore *SqlStore) GuestStore {
	s := &SqlGuestStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Guest{}, "Guests").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("ChannelId").SetMaxSize(26)
		table.ColMap("InviteId").SetMaxSize(32)
	}

	return s
}

func (s SqlGuestStore) UpgradeSchemaIfNeeded() {
}

func (s SqlGuestStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_guests_channels_id", "Guests", "ChannelId")
	s.CreateIndexIfNotExists("idx_guests_invite_id", "Guests", "InviteId")
}

func (s SqlGuestStore) Save(guest *model.Guest) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		var result StoreResult

		if len(guest.Id) > 0 {
			result.Err = model.NewLocAppError("SqlGuestStore.Save",
				"store.sql_guest.save.existing.app_error", nil, "id="+guest.Id)
			storeChannel <- result
			close(storeChannel)
			return
		}

		guest.PreSave()

		if result.Err = guest.IsValid(); result.Err != nil {
			storeChannel <- result
			close(storeChannel)
			return
		}

		if exists := <-s.GetByChannelId(guest.ChannelId); exists.Err == nil {
			result.Err = model.NewLocAppError("SqlGuestStore.Save",
				"store.sql_guest.save.already.app_error", nil, "id="+guest.Id)
			storeChannel <- result
			close(storeChannel)
			return
		}

		if err := s.GetMaster().Insert(guest); err != nil {
			result.Err = model.NewLocAppError("SqlGuestStore.Save", "store.sql_guest.save.app_error", nil, "id="+guest.Id+", "+err.Error())
		} else {
			result.Data = guest
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlGuestStore) Update(guest *model.Guest) StoreChannel {

	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		guest.PreUpdate()

		if oldResult, err := s.GetMaster().Get(model.Guest{}, guest.Id); err != nil {
			result.Err = model.NewLocAppError("SqlGuestStore.Update", "store.sql_guest.update.finding.app_error", nil, "id="+guest.Id+", "+err.Error())
		} else if oldResult == nil {
			result.Err = model.NewLocAppError("SqlGuestStore.Update", "store.sql_guest.update.find.app_error", nil, "id="+guest.Id)
		} else {
			oldGuest := oldResult.(*model.Guest)
			guest.CreateAt = oldGuest.CreateAt
			guest.UpdateAt = model.GetMillis()
			guest.TotalUsedCount += 1

			if count, err := s.GetMaster().Update(guest); err != nil {
				result.Err = model.NewLocAppError("SqlGuestStore.Update", "store.sql_guest.update.updating.app_error", nil, "id="+guest.Id+", "+err.Error())
			} else if count != 1 {
				result.Err = model.NewLocAppError("SqlGuestStore.Update", "store.sql_guest.update.app_error", nil, "id="+guest.Id)
			} else {
				result.Data = guest
			}
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlGuestStore) GetByChannelId(channelId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		guest := model.Guest{}

		if err := s.GetReplica().SelectOne(&guest, "SELECT * FROM Guests WHERE ChannelId = :ChannelId AND DeleteAt = 0", map[string]interface{}{"ChannelId": channelId}); err != nil {
			result.Err = model.NewLocAppError("SqlGuestStore.GetByInviteId", "store.sql_guest.get_by_channel_id.finding.app_error", nil, "channelId="+channelId+", "+err.Error())

		}

		result.Data = &guest
		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlGuestStore) GetByInviteId(inviteId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		guest := model.Guest{}

		if err := s.GetReplica().SelectOne(&guest, "SELECT * FROM Guests WHERE InviteId = :InviteId AND DeleteAt = 0", map[string]interface{}{"InviteId": inviteId}); err != nil {
			result.Err = model.NewLocAppError("SqlGuestStore.GetByInviteId", "store.sql_guest.get_by_invite_id.finding.app_error", nil, "inviteId="+inviteId+", "+err.Error())
		}

		if guest.DeleteAt != 0 {
			result.Err = model.NewLocAppError("SqlGuestStore.GetByInviteId", "store.sql_guest.get_by_invite_id.expired.app_error", nil, "inviteId="+inviteId)
		}

		result.Data = &guest

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlGuestStore) Delete(channelId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update Guests SET DeleteAt = :Time, UpdateAt = :Time WHERE ChannelId = :ChannelId", map[string]interface{}{"Time": time, "ChannelId": channelId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlGuestStore.Delete", "store.sql_guest.delete.invite.app_error", nil, "id="+channelId+", err="+err.Error())
		}

		_, err = s.GetMaster().Exec("Update Channels SET ExtraUpdateAt = :Time WHERE Id = :ChannelId", map[string]interface{}{"Time": time, "ChannelId": channelId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlGuestStore.Delete", "store.sql_guest.delete.invite.app_error", nil, "id="+channelId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlGuestStore) GetMembers(channelId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var members []*model.GuestMember
		_, err := s.GetReplica().Select(&members, `SELECT cm.ChannelId, cm.UserId, u.Username, u.Roles, COUNT(m.UserId) as ChannelsCount
		FROM ChannelMembers cm
		INNER JOIN Users u ON u.Roles LIKE '%guest%' AND u.Id = cm.UserId AND cm.ChannelId = :ChannelId
		LEFT OUTER JOIN ChannelMembers m ON cm.UserId=m.UserId
		GROUP BY cm.UserId, cm.ChannelId, u.Username, u.Roles
		`, map[string]interface{}{"ChannelId": channelId})

		if err != nil {
			result.Err = model.NewLocAppError("SqlGuestStore.GetMembers", "store.sql_guest.get_members.app_error", nil, "channel_id="+channelId+err.Error())
		} else {
			result.Data = members
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlGuestStore) RemoveMembers(channelId string, members []*model.GuestMember) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		queryParams := map[string]interface{}{
			"ChannelId": channelId,
		}

		var args []interface{}
		for _, member := range members {
			args = append(args, member.UserId)
		}

		if len(args) > 0 {
			inClause := ":Member0"
			queryParams["Member0"] = args[0]

			for i := 1; i < len(args); i++ {
				paramName := "Member" + strconv.FormatInt(int64(i), 10)
				inClause += ", :" + paramName
				queryParams[paramName] = args[i]
			}
			query := strings.Replace("DELETE FROM ChannelMembers WHERE ChannelId = :ChannelId AND UserId IN (MEMBERS_FILTER)", "MEMBERS_FILTER", inClause, 1)
			if _, err := s.GetMaster().Exec(query, queryParams); err != nil {
				result.Err = model.NewLocAppError("SqlUserStore.RemoveMembers", "store.sql_guest.remove_members.app_error", nil, "channelId="+channelId+", "+err.Error())
			}
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlGuestStore) RemoveUsers(usersIds []string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		queryParams := map[string]interface{}{
			"Time": time,
		}

		if len(usersIds) > 0 {
			inClause := ":User0"
			queryParams["User0"] = usersIds[0]

			for i := 1; i < len(usersIds); i++ {
				paramName := "User" + strconv.FormatInt(int64(i), 10)
				inClause += ", :" + paramName
				queryParams[paramName] = usersIds[i]
			}
			query := strings.Replace("UPDATE Users SET DeleteAt = :Time, UpdateAt = :Time WHERE Id IN (USERS_FILTER)", "USERS_FILTER", inClause, 1)
			if _, err := s.GetMaster().Exec(query, queryParams); err != nil {
				result.Err = model.NewLocAppError("SqlUserStore.RemoveMembers", "store.sql_guest.remove_users.app_error", nil, err.Error())
			}
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}
