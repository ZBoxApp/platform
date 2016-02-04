package store

import (
	"github.com/go-gorp/gorp"
	"github.com/mattermost/platform/model"
)

type SqlAddonStore struct {
	*SqlStore
}

func NewSqlAddonStore(sqlStore *SqlStore) AddonStore {
	as := &SqlAddonStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Addon{}, "Addons").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Key").SetMaxSize(64)
		table.ColMap("Name").SetMaxSize(64).SetUnique(true)
		table.ColMap("IconURL").SetMaxSize(255)
		table.ColMap("Description").SetMaxSize(1024)
		table.ColMap("PublishedURL").SetMaxSize(255)
		table.ColMap("DescriptorURL").SetMaxSize(255)
		table.ColMap("HomepageURL").SetMaxSize(255)
		table.ColMap("AllowGlobal").SetMaxSize(1)
		table.ColMap("AllowChannel").SetMaxSize(1)
		table.ColMap("InstalledURL").SetMaxSize(255)
		table.ColMap("UninstalledURL").SetMaxSize(255)
		table.ColMap("ConfigURL").SetMaxSize(255)
		table.ColMap("EnableIncomingWebhook").SetMaxSize(1)
		table.ColMap("Price")
		table.ColMap("Currency").SetMaxSize(3)
		table.ColMap("Enabled").SetMaxSize(1)

		// AddonWebhooks
		tableh := db.AddTableWithName(model.AddonWebhook{}, "AddonWebhooks").SetKeys(false, "Id")
		tableh.ColMap("Id").SetMaxSize(26)
		tableh.ColMap("Key").SetMaxSize(64)
		tableh.ColMap("AddonId").SetMaxSize(26)
		tableh.ColMap("Triggers").SetMaxSize(1024)
		tableh.ColMap("CallbackURLs").SetMaxSize(1024)

		// AddonCompacts
		tablec := db.AddTableWithName(model.AddonCompact{}, "AddonCompacts").SetKeys(false, "Id")
		tablec.ColMap("Id").SetMaxSize(26)
		tablec.ColMap("AddonId").SetMaxSize(26)
		tablec.ColMap("Name").SetMaxSize(64)
		tablec.ColMap("QueryURL").SetMaxSize(255)
		tablec.ColMap("Key").SetMaxSize(64)
		tablec.ColMap("Target").SetMaxSize(64)
		tablec.ColMap("IconURL").SetMaxSize(255)
		tablec.ColMap("IconURL2x").SetMaxSize(255)

		// AddonWebpanels
		tablew := db.AddTableWithName(model.AddonWebpanel{}, "AddonWebpanels").SetKeys(false, "Id")
		tablew.ColMap("Id").SetMaxSize(26)
		tablew.ColMap("AddonId").SetMaxSize(26)
		tablew.ColMap("Name").SetMaxSize(64)
		tablew.ColMap("Key").SetMaxSize(64)
		tablew.ColMap("Location").SetMaxSize(32)
		tablew.ColMap("URL").SetMaxSize(255)

		// AddonDialogs
		tabled := db.AddTableWithName(model.AddonDialog{}, "AddonDialogs").SetKeys(false, "Id")
		tabled.ColMap("Id").SetMaxSize(26)
		tabled.ColMap("AddonId").SetMaxSize(26)
		tabled.ColMap("Title").SetMaxSize(32)
		tabled.ColMap("Key").SetMaxSize(64)
		tabled.ColMap("URL").SetMaxSize(255)
		tabled.ColMap("PrimaryActionName").SetMaxSize(32)
		tabled.ColMap("PrimaryActionKey").SetMaxSize(64)
		tabled.ColMap("PrimaryActionEnabled").SetMaxSize(1)
		tabled.ColMap("SecondaryActionName").SetMaxSize(32)
		tabled.ColMap("SecondaryActionKey").SetMaxSize(64)
		tabled.ColMap("Size").SetMaxSize(16)

		// AddonActions
		tablea := db.AddTableWithName(model.AddonAction{}, "AddonActions").SetKeys(false, "Id")
		tablea.ColMap("Id").SetMaxSize(26)
		tablea.ColMap("AddonId").SetMaxSize(26)
		tablea.ColMap("Name").SetMaxSize(64)
		tablea.ColMap("Key").SetMaxSize(64)
		tablea.ColMap("Location").SetMaxSize(32)
		tablea.ColMap("Target").SetMaxSize(64)

		// TeamAddon
		tablet := db.AddTableWithName(model.TeamAddon{}, "TeamAddons").SetKeys(false, "Id")
		tablet.ColMap("Id").SetMaxSize(26)
		tablet.ColMap("AddonId").SetMaxSize(26)
		tablet.ColMap("TeamId").SetMaxSize(26)
		tablet.ColMap("Enabled").SetMaxSize(1)
		tablet.SetUniqueTogether("AddonId", "TeamId")
	}

	return as
}

func (s SqlAddonStore) UpgradeSchemaIfNeeded() {

}

func (s SqlAddonStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_addons_name", "Addons", "Name")
	s.CreateIndexIfNotExists("idx_addon_webhook_addon_id", "AddonWebhooks", "AddonId")
	s.CreateIndexIfNotExists("idx_addon_compact_addon_id", "AddonCompacts", "AddonId")
	s.CreateIndexIfNotExists("idx_addon_webpanel_addon_id", "AddonWebpanels", "AddonId")
	s.CreateIndexIfNotExists("idx_addon_dialog_addon_id", "AddonDialogs", "AddonId")
	s.CreateIndexIfNotExists("idx_addon_action_addon_id", "AddonActions", "AddonId")
}

func (s SqlAddonStore) Save(descriptor *model.AddonDescriptor) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		var result StoreResult
		var addon *model.Addon
		if transaction, err := s.GetMaster().Begin(); err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.open_transaction.app_error", nil, err.Error())
		} else {

			addon = &model.Addon{
				Key:                   descriptor.Key,
				Name:                  descriptor.Name,
				IconURL:               descriptor.Vendor.IconURL,
				Description:           descriptor.Description,
				DescriptorURL:         descriptor.Links.DescriptorURL,
				HomepageURL:           descriptor.Links.HomepageURL,
				PublishedURL:          descriptor.Links.PublishedURL,
				AllowGlobal:           descriptor.Installable.AllowGlobal,
				AllowChannel:          descriptor.Installable.AllowChannel,
				InstalledURL:          descriptor.Installable.InstalledURL,
				UninstalledURL:        descriptor.Installable.UninstalledURL,
				ConfigURL:             descriptor.Installable.ConfigURL,
				EnableIncomingWebhook: descriptor.Webhook.EnableIncoming,
				Price:    descriptor.Price.Value,
				Currency: descriptor.Price.Currency}

			result = s.saveAddonT(transaction, addon)
			if result.Err == nil {
				addon = result.Data.(*model.Addon)

				for _, hook := range descriptor.Webhook.Outgoing {
					aw := &model.AddonWebhook{AddonId: addon.Id, Key: hook.Key, Triggers: hook.Triggers, CallbackURLs: hook.CallbackURLs}
					if upsertResult := s.saveWebhookT(transaction, aw); upsertResult.Err != nil {
						result = upsertResult
						break
					}
				}

				for _, compact := range descriptor.Compact {
					ac := &model.AddonCompact{AddonId: addon.Id, Name: compact.Name, QueryURL: compact.QueryURL, Key: compact.Key, Target: compact.Target,
						IconURL: compact.Icon.URL, IconURL2x: compact.Icon.URL2x}
					if upsertResult := s.saveCompactT(transaction, ac); upsertResult.Err != nil {
						result = upsertResult
						break
					}
				}

				for _, panel := range descriptor.Webpanel {
					ap := &model.AddonWebpanel{AddonId: addon.Id, Name: panel.Name, Key: panel.Key, Location: panel.Location, URL: panel.URL}
					if upsertResult := s.saveWebpanelT(transaction, ap); upsertResult.Err != nil {
						result = upsertResult
						break
					}
				}

				for _, dialog := range descriptor.Dialog {
					ad := &model.AddonDialog{AddonId: addon.Id, Title: dialog.Title, Key: dialog.Key, URL: dialog.URL, Size: dialog.Options.Size,
						PrimaryActionName: dialog.Options.PrimaryAction.Name, PrimaryActionKey: dialog.Options.PrimaryAction.Key, PrimaryActionEnabled: dialog.Options.PrimaryAction.Enabled,
						SecondaryActionName: dialog.Options.SecondaryAction.Name, SecondaryActionKey: dialog.Options.SecondaryAction.Key}
					if upsertResult := s.saveDialogT(transaction, ad); upsertResult.Err != nil {
						result = upsertResult
						break
					}
				}

				for _, action := range descriptor.Action {
					aa := &model.AddonAction{AddonId: addon.Id, Name: action.Name, Key: action.Key, Location: action.Location, Target: action.Target}
					if upsertResult := s.saveActionT(transaction, aa); upsertResult.Err != nil {
						result = upsertResult
						break
					}
				}
			}

			if result.Err != nil {
				transaction.Rollback()
			} else {
				if err := transaction.Commit(); err != nil {
					result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.commit_transaction.app_error", nil, err.Error())
				} else {
					result.Data = addon
				}
			}

		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) saveAddonT(transaction *gorp.Transaction, addon *model.Addon) StoreResult {
	result := StoreResult{}

	if len(addon.Id) > 0 {
		result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.existing_addon.app_error", nil, "id="+addon.Id)
		return result
	}

	addon.PreSave()
	if result.Err = addon.IsValid(); result.Err == nil {
		if err := transaction.Insert(addon); err != nil {
			if IsUniqueConstraintError(err.Error(), "Name", "addons_name") {
				dupAddon := model.Addon{}
				s.GetMaster().SelectOne(&dupAddon, "SELECT * FROM Addons WHERE Name = :Name AND DeleteAt > 0", map[string]interface{}{"Name": addon.Name})
				if dupAddon.DeleteAt > 0 {
					result.Err = model.NewLocAppError("SqlAddonStore.Update", "store.sql_addon.save.previously.app_error", nil, "id="+addon.Id+", "+err.Error())
				} else {
					result.Err = model.NewLocAppError("SqlAddonStore.Update", "store.sql_addon.save.already_exists.app_error", nil, "id="+addon.Id+", "+err.Error())
				}
			} else {
				result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.app_error", nil, "id="+addon.Id+", "+err.Error())
			}
		} else {
			result.Data = addon
		}
	}

	return result
}

func (s SqlAddonStore) saveWebhookT(transaction *gorp.Transaction, hook *model.AddonWebhook) StoreResult {
	result := StoreResult{}

	if len(hook.Id) > 0 {
		result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.existing_addon_hook.app_error", nil, "id="+hook.Id)
		return result
	}

	hook.PreSave()
	if result.Err = hook.IsValid(); result.Err == nil {
		if err := transaction.Insert(hook); err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.addon_hook.app_error", nil, "id="+hook.Id+", "+err.Error())
		}
	}

	return result
}

func (s SqlAddonStore) saveCompactT(transaction *gorp.Transaction, compact *model.AddonCompact) StoreResult {
	result := StoreResult{}

	if len(compact.Id) > 0 {
		result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.existing_addon_compact.app_error", nil, "id="+compact.Id)
		return result
	}

	compact.PreSave()
	if result.Err = compact.IsValid(); result.Err == nil {
		if err := transaction.Insert(compact); err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.addon_compact.app_error", nil, "id="+compact.Id+", "+err.Error())
		}
	}

	return result
}

func (s SqlAddonStore) saveWebpanelT(transaction *gorp.Transaction, panel *model.AddonWebpanel) StoreResult {
	result := StoreResult{}

	if len(panel.Id) > 0 {
		result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.existing_addon_webpanel.app_error", nil, "id="+panel.Id)
		return result
	}

	panel.PreSave()
	if result.Err = panel.IsValid(); result.Err == nil {
		if err := transaction.Insert(panel); err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.addon_webpanel.app_error", nil, "id="+panel.Id+", "+err.Error())
		}
	}

	return result
}

func (s SqlAddonStore) saveDialogT(transaction *gorp.Transaction, dialog *model.AddonDialog) StoreResult {
	result := StoreResult{}

	if len(dialog.Id) > 0 {
		result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.existing_addon_dialog.app_error", nil, "id="+dialog.Id)
		return result
	}

	dialog.PreSave()
	if result.Err = dialog.IsValid(); result.Err == nil {
		if err := transaction.Insert(dialog); err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.addon_dialog.app_error", nil, "id="+dialog.Id+", "+err.Error())
		}
	}

	return result
}

func (s SqlAddonStore) saveActionT(transaction *gorp.Transaction, action *model.AddonAction) StoreResult {
	result := StoreResult{}

	if len(action.Id) > 0 {
		result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.existing_addon_action.app_error", nil, "id="+action.Id)
		return result
	}

	action.PreSave()
	if result.Err = action.IsValid(); result.Err == nil {
		if err := transaction.Insert(action); err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.Save", "store.sql_addon.save.addon_action.app_error", nil, "id="+action.Id+", "+err.Error())
		}
	}

	return result
}

func (s SqlAddonStore) GetAddons() StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data []model.Addon

		_, err := s.GetReplica().Select(&data, "SELECT * FROM Addons WHERE DeleteAt = 0 AND Enabled = 1 ORDER BY Name")
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddons", "store.sql_addon.get.addons.app_error", nil, "err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetAddonById(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data model.Addon

		err := s.GetReplica().SelectOne(&data, "SELECT * FROM Addons WHERE Id = :AddonId ORDER BY Name", map[string]interface{}{"AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddons", "store.sql_addon.get.addon.app_error", nil, "id="+addonId+",err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetAddonByName(name string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data model.Addon

		err := s.GetReplica().SelectOne(&data, "SELECT * FROM Addons WHERE Name = :Name ORDER BY Name", map[string]interface{}{"Name": name})
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddons", "store.sql_addon.get.addon.app_error", nil, "addon_name="+name+",err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetAddonWebooks(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data []model.AddonWebhook

		_, err := s.GetReplica().Select(&data, "SELECT * FROM AddonWebhooks WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddonHooks", "store.sql_addon.get.addon_hooks.app_error", nil, "err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetAddonCompacts(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data []model.AddonCompact

		_, err := s.GetReplica().Select(&data, "SELECT * FROM AddonCompacts WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddonCompacts", "store.sql_addon.get.addon_compacts.app_error", nil, "err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetAddonWebpanels(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data []model.AddonWebpanel

		_, err := s.GetReplica().Select(&data, "SELECT * FROM AddonWebpanels WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddonWebpanels", "store.sql_addon.get.addon_webpanels.app_error", nil, "err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetAddonDialogs(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data []model.AddonDialog

		_, err := s.GetReplica().Select(&data, "SELECT * FROM AddonDialogs WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddonDialogs", "store.sql_addon.get.addon_dialogs.app_error", nil, "err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetAddonActions(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data []model.AddonAction

		_, err := s.GetReplica().Select(&data, "SELECT * FROM AddonActions WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddonActions", "store.sql_addon.get.addon_actions.app_error", nil, "err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetDisabledAddons() StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data []model.Addon

		_, err := s.GetReplica().Select(&data, "SELECT * FROM Addons WHERE DeleteAt = 0 AND Enabled = 0 ORDER BY Name")
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetAddons", "store.sql_addon.get.addons.app_error", nil, "err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) GetInstalled(teamId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		var data []model.TeamAddon

		_, err := s.GetReplica().Select(&data, "SELECT * FROM TeamAddons WHERE TeamId = :TeamId AND Enabled = 1", map[string]interface{}{"TeamId": teamId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetInstalled", "store.sql_addon.get.addons.app_error", nil, "err="+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) IsInstalled(teamId, addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}
		data := model.TeamAddon{}

		err := s.GetReplica().SelectOne(&data, "SELECT * FROM TeamAddons WHERE TeamId = :TeamId AND AddonId = :AddonId", map[string]interface{}{"TeamId": teamId, "AddonId": addonId})
		if err == nil {
			result.Err = model.NewLocAppError("SqlAddonStore.GetInstalled", "store.sql_addon.installed.already.app_error", nil, "")
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) InstallAddon(teamAddon *model.TeamAddon) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		if len(teamAddon.Id) > 0 {
			result.Err = model.NewLocAppError("SqlAddonStore.InstallAddon", "store.sql_addon.installed.already.app_error", nil, "id="+teamAddon.Id)
			storeChannel <- result
			close(storeChannel)
			return
		}

		teamAddon.PreSave()
		if result.Err = teamAddon.IsValid(); result.Err != nil {
			storeChannel <- result
			close(storeChannel)
			return
		}

		if err := s.GetMaster().Insert(teamAddon); err != nil {
			if IsUniqueConstraintError(err.Error(), "TeamId", "addon_teamid_key") {
				result.Err = model.NewLocAppError("SqlAddonStore.InstallAddon", "store.sql_addon.installed.already.app_error", nil, "addon_id="+teamAddon.AddonId+", "+err.Error())
			} else {
				result.Err = model.NewLocAppError("SqlAddonStore.InstallAddon", "store.sql_addon.install.app_error", nil, "addon_id="+teamAddon.AddonId+", "+err.Error())
			}
		} else {
			result.Data = teamAddon
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) UninstallAddon(teamAddon *model.TeamAddon) StoreChannel {

	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		if _, err := s.GetMaster().Exec("DELETE FROM TeamAddons WHERE AddonId = :AddonId AND TeamId = :TeamId", map[string]interface{}{"AddonId": teamAddon.AddonId, "TeamId": teamAddon.TeamId}); err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.UninstallAddon", "store.sql_addon.uninstall.app_error", nil, "team_id="+teamAddon.TeamId+", addon_id="+teamAddon.AddonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) EnableAddon(addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update Addons SET Enabled = 1, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"UpdateAt": time, "Id": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoing", "store.sql_addon.enable.app_error", nil, "id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) DisableAddon(addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update Addons SET Enabled = 0, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"UpdateAt": time, "Id": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoing", "store.sql_addon.disable.app_error", nil, "id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) DeleteAddon(addonId string, time int64) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		_, err := s.GetMaster().Exec("Update Addons SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": addonId})
		if err != nil {
			result.Err = model.NewLocAppError("SqlWebhookStore.DeleteOutgoing", "store.sql_addon.delete.addon.app_error", nil, "id="+addonId+", err="+err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) PermanentDeleteAddon(addonId string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		var result StoreResult
		if transaction, err := s.GetMaster().Begin(); err != nil {
			result.Err = model.NewLocAppError("SqlAddonStore.PermanentDeleteAddon", "store.sql_addon.save.open_transaction.app_error", nil, err.Error())
		} else {

			result = s.permanentDeleteAddonActionsT(transaction, addonId)

			if result.Err == nil {
				result = s.permanentDeleteAddonDialogsT(transaction, addonId)
			}

			if result.Err == nil {
				result = s.permanentDeleteAddonWebpanelsT(transaction, addonId)
			}

			if result.Err == nil {
				result = s.permanentDeleteAddonCompactsT(transaction, addonId)
			}

			if result.Err == nil {
				result = s.permanentDeleteAddonWebhooksT(transaction, addonId)
			}

			if result.Err == nil {
				result = s.permanentDeleteAddonT(transaction, addonId)
			}

			if result.Err != nil {
				transaction.Rollback()
			} else {
				if err := transaction.Commit(); err != nil {
					result.Err = model.NewLocAppError("SqlAddonStore.PermanentDeleteAddon", "store.sql_addon.save.commit_transaction.app_error", nil, err.Error())
				}
			}
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlAddonStore) permanentDeleteAddonT(transaction *gorp.Transaction, addonId string) StoreResult {
	result := StoreResult{}

	_, err := transaction.Exec("DELETE FROM Addons WHERE Id = :AddonId", map[string]interface{}{"AddonId": addonId})

	if err != nil {
		result.Err = model.NewLocAppError("SqlAddonStore.permanentDeleteAddonT", "store.sql_addon.delete.addon.app_error", nil, "id="+addonId+", err="+err.Error())
	}

	return result
}

func (s SqlAddonStore) permanentDeleteAddonWebhooksT(transaction *gorp.Transaction, addonId string) StoreResult {
	result := StoreResult{}

	_, err := transaction.Exec("DELETE FROM AddonWebhooks WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})

	if err != nil {
		result.Err = model.NewLocAppError("SqlAddonStore.permanentDeleteAddonWebhooksT", "store.sql_addon.delete.addon_hooks.app_error", nil, "addon_id="+addonId+", err="+err.Error())
	}

	return result
}

func (s SqlAddonStore) permanentDeleteAddonCompactsT(transaction *gorp.Transaction, addonId string) StoreResult {
	result := StoreResult{}

	_, err := transaction.Exec("DELETE FROM AddonCompacts WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})

	if err != nil {
		result.Err = model.NewLocAppError("SqlAddonStore.permanentDeleteAddonCompactsT", "store.sql_addon.delete.addon_compacts.app_error", nil, "addon_id="+addonId+", err="+err.Error())
	}

	return result
}

func (s SqlAddonStore) permanentDeleteAddonWebpanelsT(transaction *gorp.Transaction, addonId string) StoreResult {
	result := StoreResult{}

	_, err := transaction.Exec("DELETE FROM AddonWebpanels WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})

	if err != nil {
		result.Err = model.NewLocAppError("SqlAddonStore.permanentDeleteAddonWebpanelsT", "store.sql_addon.delete.addon_webpanels.app_error", nil, "addon_id="+addonId+", err="+err.Error())
	}

	return result
}

func (s SqlAddonStore) permanentDeleteAddonDialogsT(transaction *gorp.Transaction, addonId string) StoreResult {
	result := StoreResult{}

	_, err := transaction.Exec("DELETE FROM AddonDialogs WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})

	if err != nil {
		result.Err = model.NewLocAppError("SqlAddonStore.permanentDeleteAddonDialogsT", "store.sql_addon.delete.addon_dialogs.app_error", nil, "addon_id="+addonId+", err="+err.Error())
	}

	return result
}

func (s SqlAddonStore) permanentDeleteAddonActionsT(transaction *gorp.Transaction, addonId string) StoreResult {
	result := StoreResult{}

	_, err := transaction.Exec("DELETE FROM AddonActions WHERE AddonId = :AddonId", map[string]interface{}{"AddonId": addonId})

	if err != nil {
		result.Err = model.NewLocAppError("SqlAddonStore.permanentDeleteAddonActionsT", "store.sql_addon.delete.addon_actions.app_error", nil, "addon_id="+addonId+", err="+err.Error())
	}

	return result
}
