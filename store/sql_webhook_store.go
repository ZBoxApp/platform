// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package store

import (
	"github.com/mattermost/platform/model"
)

type SqlWebhookStore struct {
	*SqlStore
}

func NewSqlWebhookStore(sqlStore *SqlStore) WebhookStore {
	s := &SqlWebhookStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.IncomingWebhook{}, "IncomingWebhooks").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("ChannelId").SetMaxSize(26)
		table.ColMap("TeamId").SetMaxSize(26)
		table.ColMap("AddonId").SetMaxSize(26)

		tableo := db.AddTableWithName(model.OutgoingWebhook{}, "OutgoingWebhooks").SetKeys(false, "Id")
		tableo.ColMap("Id").SetMaxSize(26)
		tableo.ColMap("Token").SetMaxSize(26)
		tableo.ColMap("CreatorId").SetMaxSize(26)
		tableo.ColMap("ChannelId").SetMaxSize(26)
		tableo.ColMap("TeamId").SetMaxSize(26)
		tableo.ColMap("TriggerWords").SetMaxSize(1024)
		tableo.ColMap("CallbackURLs").SetMaxSize(1024)
		tableo.ColMap("AddonWebhookId").SetMaxSize(26)
	}

	return s
}

func (s SqlWebhookStore) UpgradeSchemaIfNeeded() {
}

func (s SqlWebhookStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_incoming_webhook_user_id", "IncomingWebhooks", "UserId")
	s.CreateIndexIfNotExists("idx_incoming_webhook_team_id", "IncomingWebhooks", "TeamId")
	s.CreateIndexIfNotExists("idx_outgoing_webhook_team_id", "OutgoingWebhooks", "TeamId")
}

func (s SqlWebhookStore) SaveIncoming(webhook *model.IncomingWebhook) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		if len(webhook.Id) > 0 {
			result.Err = model.NewLocAppError("SqlWebhookStore.SaveIncoming",
				"store.sql_webhooks.save_incoming.existing.app_error", nil, "id="+webhook.Id)
			storeChannel <- result
			close(storeChannel)
			return
		}

		webhook.PreSave()
		if result.Err = webhook.IsValid(); result.Err != nil {
			storeChannel <- result
			close(storeChannel)
			return
		}

		if err := s.GetMaster().Insert(webhook); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.SaveIncoming", "store.sql_webhooks.save_incoming.app_error", nil, "id="+webhook.Id+", "+err.Error())
		} else {
			result.Data = webhook
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetIncoming(id string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhook model.IncomingWebhook

		if err := s.GetReplica().SelectOne(&webhook, "SELECT * FROM IncomingWebhooks WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetIncoming", "store.sql_webhooks.get_incoming.app_error", nil, "id="+id+", err="+err.Error())
		}

		result.Data = &webhook

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) DeleteIncoming(webhookId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update IncomingWebhooks SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": webhookId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteIncoming", "store.sql_webhooks.delete_incoming.app_error", nil, "id="+webhookId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) DeleteIncomingByAddon(addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update IncomingWebhooks SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE AddonId = :AddonId", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteIncomingByAddon", "store.sql_webhooks.delete_incoming.app_error", nil, "addon_id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) DeleteIncomingByTeamAddon(teamId, addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update IncomingWebhooks SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE AddonId = :AddonId AND TeamId = :TeamId", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "AddonId": addonId, "TeamId": teamId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteIncomingByAddon", "store.sql_webhooks.delete_incoming.app_error", nil, "addon_id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) EnableIncomingByAddon(addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update IncomingWebhooks SET DeleteAt = 0, UpdateAt = :UpdateAt WHERE AddonId = :AddonId", map[string]interface{}{"UpdateAt": time, "AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.EnableIncomingByAddon", "store.sql_webhooks.enable_incoming.app_error", nil, "addon_id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) PermanentDeleteIncomingByUser(userId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("DELETE FROM IncomingWebhooks WHERE UserId = :UserId", map[string]interface{}{"UserId": userId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteIncomingByUser", "store.sql_webhooks.permanent_delete_incoming_by_user.app_error", nil, "id="+userId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) PermanentDeleteIncomingByAddon(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("DELETE FROM IncomingWebhooks WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.PermanentDeleteIncomingByAddon", "store.sql_webhooks.permanent_delete_incoming_by_user.app_error", nil, "addon_id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) PermanentDeleteIncomingByTeamAddon(teamId, addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("DELETE FROM IncomingWebhooks WHERE AddonId = :AddonId AND TeamId = :TeamId", map[string]interface{}{"AddonId": addonId, "TeamId": teamId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.PermanentDeleteIncomingByAddon", "store.sql_webhooks.permanent_delete_incoming_by_user.app_error", nil, "addon_id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetIncomingByTeam(teamId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhooks []*model.IncomingWebhook

		if _, err := s.GetReplica().Select(&webhooks, "SELECT * FROM IncomingWebhooks WHERE TeamId = :TeamId AND DeleteAt = 0", map[string]interface{}{"TeamId": teamId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetIncomingByUser", "store.sql_webhooks.get_incoming_by_user.app_error", nil, "teamId="+teamId+", err="+err.Error())
		}

		result.Data = webhooks

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetIncomingByChannel(channelId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhooks []*model.IncomingWebhook

		if _, err := s.GetReplica().Select(&webhooks, "SELECT * FROM IncomingWebhooks WHERE ChannelId = :ChannelId AND DeleteAt = 0", map[string]interface{}{"ChannelId": channelId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetIncomingByChannel", "store.sql_webhooks.get_incoming_by_channel.app_error", nil, "channelId="+channelId+", err="+err.Error())
		}

		result.Data = webhooks

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetIncomingByAddon(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhooks []*model.IncomingWebhook

		if _, err := s.GetReplica().Select(&webhooks, "SELECT * FROM IncomingWebhooks WHERE AddonId = :AddonId AND DeleteAt = 0", map[string]interface{}{"AddonId": addonId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetIncomingByAddon", "store.sql_webhooks.get_incoming_by_channel.app_error", nil, "addonId="+addonId+", err="+err.Error())
		}

		result.Data = webhooks

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetIncomingByTeamAddon(teamId, addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhooks []*model.IncomingWebhook

		if _, err := s.GetReplica().Select(&webhooks, "SELECT * FROM IncomingWebhooks WHERE AddonId = :AddonId AND TeamId = :TeamId AND DeleteAt = 0", map[string]interface{}{"AddonId": addonId, "TeamId": teamId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetIncomingByAddon", "store.sql_webhooks.get_incoming_by_channel.app_error", nil, "addonId="+addonId+", err="+err.Error())
		}

		result.Data = webhooks

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) SaveOutgoing(webhook *model.OutgoingWebhook) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		if len(webhook.Id) > 0 {
			result.Err = model.NewLocAppError("SqlWebhookStore.SaveOutgoing",
				"store.sql_webhooks.save_outgoing.override.app_error", nil, "id="+webhook.Id)
			storeChannel <- result
			close(storeChannel)
			return
		}

		webhook.PreSave()
		if result.Err = webhook.IsValid(); result.Err != nil {
			storeChannel <- result
			close(storeChannel)
			return
		}

		if err := s.GetMaster().Insert(webhook); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.SaveOutgoing", "store.sql_webhooks.save_outgoing.app_error", nil, "id="+webhook.Id+", "+err.Error())
		} else {
			result.Data = webhook
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetOutgoing(id string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhook model.OutgoingWebhook

		if err := s.GetReplica().SelectOne(&webhook, "SELECT * FROM OutgoingWebhooks WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetOutgoing", "store.sql_webhooks.get_outgoing.app_error", nil, "id="+id+", err="+err.Error())
		}

		result.Data = &webhook

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetOutgoingByChannel(channelId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhooks []*model.OutgoingWebhook

		if _, err := s.GetReplica().Select(&webhooks, "SELECT * FROM OutgoingWebhooks WHERE ChannelId = :ChannelId AND DeleteAt = 0", map[string]interface{}{"ChannelId": channelId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetOutgoingByChannel", "store.sql_webhooks.get_outgoing_by_channel.app_error", nil, "channelId="+channelId+", err="+err.Error())
		}

		result.Data = webhooks

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetOutgoingByTeam(teamId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhooks []*model.OutgoingWebhook

		if _, err := s.GetReplica().Select(&webhooks, "SELECT * FROM OutgoingWebhooks WHERE TeamId = :TeamId AND DeleteAt = 0", map[string]interface{}{"TeamId": teamId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetOutgoingByTeam", "store.sql_webhooks.get_outgoing_by_team.app_error", nil, "teamId="+teamId+", err="+err.Error())
		}

		result.Data = webhooks

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetOutgoingByAddon(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhooks []*model.OutgoingWebhook

		if _, err := s.GetReplica().Select(&webhooks, "SELECT ow.* FROM OutgoingWebhooks ow INNER JOIN AddonWebhooks aw ON ow.AddonWebhookId = aw.Id AND aw.AddonId = :AddonId AND DeleteAt = 0", map[string]interface{}{"AddonId": addonId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetOutgoingByAddon", "store.sql_webhooks.get_outgoing_by_team.app_error", nil, "addonId="+addonId+", err="+err.Error())
		}

		result.Data = webhooks

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) GetOutgoingByTeamAddon(teamId, addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var webhooks []*model.OutgoingWebhook

		if _, err := s.GetReplica().Select(&webhooks, "SELECT ow.* FROM OutgoingWebhooks ow INNER JOIN AddonWebhooks aw ON ow.AddonWebhookId = aw.Id AND aw.AddonId = :AddonId AND ow.TeamId = :TeamId AND DeleteAt = 0", map[string]interface{}{"AddonId": addonId, "TeamId": teamId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.GetOutgoingByAddon", "store.sql_webhooks.get_outgoing_by_team.app_error", nil, "addonId="+addonId+", err="+err.Error())
		}

		result.Data = webhooks

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) DeleteOutgoing(webhookId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update OutgoingWebhooks SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": webhookId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoing", "store.sql_webhooks.delete_outgoing.app_error", nil, "id="+webhookId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) DeleteOutgoingByAddon(addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update OutgoingWebhooks INNER JOIN AddonWebhooks ON AddonWebhooks.Id = OutgoingWebhooks.AddonWebhookId AND AddonWebhooks.AddonId = :AddonId SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoingByAddon", "store.sql_webhooks.delete_outgoing.app_error", nil, "id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) DeleteOutgoingByTeamAddon(teamId, addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update OutgoingWebhooks INNER JOIN AddonWebhooks ON AddonWebhooks.Id = OutgoingWebhooks.AddonWebhookId AND AddonWebhooks.AddonId = :AddonId AND OutgoingWebhooks.TeamId = :TeamId SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "AddonId": addonId, "TeamId": teamId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoingByAddon", "store.sql_webhooks.delete_outgoing.app_error", nil, "id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) EnableOutgoingByAddon(addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update OutgoingWebhooks SET DeleteAt = 0, UpdateAt = :UpdateAt WHERE AddonId = :AddonId", map[string]interface{}{"UpdateAt": time, "AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.EnableOutgoingByAddon", "store.sql_webhooks.enable_outgoing.app_error", nil, "id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) PermanentDeleteOutgoingByUser(userId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("DELETE FROM OutgoingWebhooks WHERE CreatorId = :UserId", map[string]interface{}{"UserId": userId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoingByUser", "store.sql_webhooks.permanent_delete_outgoing_by_user.app_error", nil, "id="+userId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) PermanentDeleteOutgoingByAddon(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("DELETE ow FROM OutgoingWebhooks ow INNER JOIN AddonWebhooks aow ON ow.AddonWebhookId=aow.Id AND aow.AddonId= :AddonId", map[string]interface{}{"AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoingByUser", "store.sql_webhooks.permanent_delete_outgoing_by_user.app_error", nil, "addon_id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) PermanentDeleteOutgoingByTeamAddon(teamId, addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("DELETE ow FROM OutgoingWebhooks ow INNER JOIN AddonWebhooks aow ON ow.AddonWebhookId=aow.Id AND aow.AddonId= :AddonId AND ow.TeamId = :TeamId", map[string]interface{}{"AddonId": addonId, "TeamId": teamId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoingByUser", "store.sql_webhooks.permanent_delete_outgoing_by_user.app_error", nil, "addon_id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) UpdateOutgoing(hook *model.OutgoingWebhook) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		hook.UpdateAt = model.GetMillis()

		if _, err := s.GetMaster().Update(hook); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.UpdateOutgoing", "store.sql_webhooks.update_outgoing.app_error", nil, "id="+hook.Id+", "+err.Error())
		} else {
			result.Data = hook
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) AnalyticsIncomingCount(teamId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		query :=
			`SELECT 
			    COUNT(*)
			FROM
			    IncomingWebhooks
			WHERE
                DeleteAt = 0`

		if len(teamId) > 0 {
			query += " AND TeamId = :TeamId"
		}

		if v, err := s.GetReplica().SelectInt(query, map[string]interface{}{"TeamId": teamId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.AnalyticsIncomingCount", "store.sql_webhooks.analytics_incoming_count.app_error", nil, "team_id="+teamId+", err="+err.Error())
		} else {
			result.Data = v
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlWebhookStore) AnalyticsOutgoingCount(teamId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		query :=
			`SELECT 
			    COUNT(*)
			FROM
			    OutgoingWebhooks
			WHERE
                DeleteAt = 0`

		if len(teamId) > 0 {
			query += " AND TeamId = :TeamId"
		}

		if v, err := s.GetReplica().SelectInt(query, map[string]interface{}{"TeamId": teamId}); err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.AnalyticsOutgoingCount", "store.sql_webhooks.analytics_outgoing_count.app_error", nil, "team_id="+teamId+", err="+err.Error())
		} else {
			result.Data = v
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}
