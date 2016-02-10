// Copyright (c) 2016 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	l4g "github.com/alecthomas/log4go"
	"github.com/gorilla/mux"

	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/utils"
)

func InitAddon(r *mux.Router) {
	l4g.Debug(utils.T("Initializing addons api routes"))

	sr := r.PathPrefix("/addon").Subrouter()
	sr.Handle("/publish", ApiUserRequired(publishAddon)).Methods("POST")
	sr.Handle("/publish_descriptor", ApiUserRequired(publishAddonWithDescriptor)).Methods("POST")
	sr.Handle("/available", ApiUserRequired(availableAddon)).Methods("GET")
	sr.Handle("/install", ApiUserRequired(installAddon)).Methods("POST")
	sr.Handle("/uninstall", ApiUserRequired(uninstallAddon)).Methods("POST")
}

func publishAddon(c *Context, w http.ResponseWriter, r *http.Request) {
	if !utils.Cfg.ServiceSettings.EnableOAuthServiceProvider {
		c.Err = model.NewLocAppError("publishAddon", "web.get_access_token.disabled.app_error", nil, "")
		c.Err.StatusCode = http.StatusNotImplemented
		return
	}

	if !c.IsTeamAdmin() {
		c.Err = model.NewLocAppError("publishAddon", "api.context.permissions.app_error", nil, "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	m := model.MapFromJson(r.Body)
	url := m["url"]
	category := m["category"]

	if len(url) == 0 || !model.IsValidHttpUrl(url) {
		c.Err = model.NewLocAppError("publishAddon", "model.addon.descriptor_url.app_error", nil, "")
		return
	} else {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Accept", "application/json")
		if resp, err := client.Do(req); err != nil {
			l4g.Error(utils.T("Event GET failed, err=%s"), err.Error())
			c.Err = model.NewLocAppError("publishAddon", "api.addon.publish.descriptor_url.app_error", nil, err.Error())
			return
		} else {
			descriptor := model.AddonDescriptorFromJson(resp.Body)
			descriptor.Category = category
			handlePublish(c, w, descriptor)
		}
	}
}

func publishAddonWithDescriptor(c *Context, w http.ResponseWriter, r *http.Request) {
	if !utils.Cfg.ServiceSettings.EnableOAuthServiceProvider {
		c.Err = model.NewLocAppError("publishAddon", "web.get_access_token.disabled.app_error", nil, "")
		c.Err.StatusCode = http.StatusNotImplemented
		return
	}

	if !c.IsTeamAdmin() {
		c.Err = model.NewLocAppError("publishAddon", "api.context.permissions.app_error", nil, "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	descriptor := model.AddonDescriptorFromJson(r.Body)
	handlePublish(c, w, descriptor)
}

func availableAddon(c *Context, w http.ResponseWriter, r *http.Request) {
	if !utils.Cfg.ServiceSettings.EnableOAuthServiceProvider {
		c.Err = model.NewLocAppError("publishAddon", "web.get_access_token.disabled.app_error", nil, "")
		c.Err.StatusCode = http.StatusNotImplemented
		return
	}

	if !c.IsTeamAdmin() {
		c.Err = model.NewLocAppError("publishAddon", "api.context.permissions.app_error", nil, "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	// Get all the Team Addons
	var data_team []model.TeamAddon
	installed := make(map[string]int64)
	if tchan := <-Srv.Store.Addon().GetInstalled(c.Session.TeamId); tchan.Err != nil {
		c.Err = tchan.Err
		return
	} else {
		data_team = tchan.Data.([]model.TeamAddon)
		for k := range data_team {
			t := data_team[k]
			installed[t.AddonId] = t.InstalledAt
		}
	}

	// Get all the Enable Addons
	var data []model.Addon
	if result := <-Srv.Store.Addon().GetAddons(); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		data = result.Data.([]model.Addon)
		available := make(model.AddonList, len(data))
		for i := range data {
			a := data[i]
			status := false
			if _, ok := installed[a.Id]; ok {
				status = true
			}
			available[i] = &model.AvailableAddon{
				Id: a.Id, Name: a.Name, Description: a.Description, IconURL: a.IconURL, Category: a.Category,
				Currency: a.Currency, Price: a.Price, Installed: status, ConfigURL: a.ConfigURL}
		}

		w.Write([]byte(available.ToJson()))
		return
	}
}

type InstalledAddonIncomingHook struct {
	Token string `json:"token"`
}

type InstalledAddonOutgoingHook struct {
	Key   string `json:"key"`
	Token string `json:"token"`
}

type InstalledAddonWebhook struct {
	Incoming InstalledAddonIncomingHook   `json:"incoming"`
	Outgoing []InstalledAddonOutgoingHook `json:"outgoing"`
}

type InstalledAddon struct {
	TeamId   string                `json:"team_id"`
	TeamName string                `json:"team_id"`
	Webooks  InstalledAddonWebhook `json:"webhook"`
}

func installAddon(c *Context, w http.ResponseWriter, r *http.Request) {
	if !utils.Cfg.ServiceSettings.EnableOAuthServiceProvider {
		c.Err = model.NewLocAppError("publishAddon", "web.get_access_token.disabled.app_error", nil, "")
		c.Err.StatusCode = http.StatusNotImplemented
		return
	}

	if !c.IsTeamAdmin() {
		c.Err = model.NewLocAppError("publishAddon", "api.context.permissions.app_error", nil, "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	m := model.MapFromJson(r.Body)
	addonId := m["addon_id"]

	// check if the addon is already installed for this team
	if already := <-Srv.Store.Addon().IsInstalled(c.Session.TeamId, addonId); already.Err != nil {
		c.Err = already.Err
		return
	}

	// get the System Bot user for the Team
	var bot *model.User
	if uchan := <-Srv.Store.User().GetByUsername(c.Session.TeamId, model.SYSTEM_BOT_NAME); uchan.Err != nil {
		c.Err = uchan.Err
		return
	} else {
		bot = uchan.Data.(*model.User)
	}

	// get the default channel
	var channel *model.Channel
	if cchan := <-Srv.Store.Channel().GetByName(bot.TeamId, model.DEFAULT_CHANNEL); cchan.Err != nil {
		c.Err = cchan.Err
		return
	} else {
		channel = cchan.Data.(*model.Channel)
	}

	// get the Addon
	var addon model.Addon
	if achan := <-Srv.Store.Addon().GetAddonById(addonId); achan.Err != nil {
		c.Err = achan.Err
		return
	} else {
		addon = achan.Data.(model.Addon)
	}

	// get the addon webhooks
	var addonHooks []model.AddonWebhook
	if awchan := <-Srv.Store.Addon().GetAddonWebooks(addon.Id); awchan.Err != nil {
		c.Err = awchan.Err
		return
	} else {
		addonHooks = awchan.Data.([]model.AddonWebhook)
	}

	var team *model.Team
	if tchan := <-Srv.Store.Team().Get(bot.TeamId); tchan.Err != nil {
		c.Err = tchan.Err
		return
	} else {
		team = tchan.Data.(*model.Team)
	}

	var installedIncoming InstalledAddonIncomingHook
	installed := &InstalledAddon{TeamId: bot.TeamId, TeamName: team.Name, Webooks: InstalledAddonWebhook{}}

	// Add the incoming Webhook
	var incoming *model.IncomingWebhook
	if addon.EnableIncomingWebhook {
		incoming = &model.IncomingWebhook{UserId: bot.Id, TeamId: bot.TeamId, ChannelId: channel.Id, AddonId: addon.Id}
		if ichan := <-Srv.Store.Webhook().SaveIncoming(incoming); ichan.Err != nil {
			c.Err = ichan.Err
			return
		} else {
			incoming = ichan.Data.(*model.IncomingWebhook)
			installedIncoming = InstalledAddonIncomingHook{Token: incoming.Id}
			installed.Webooks.Incoming = installedIncoming
		}
	}

	// Add the outgoing webhooks
	hooks := make([]InstalledAddonOutgoingHook, 0)
	for _, h := range addonHooks {
		hook := &model.OutgoingWebhook{CreatorId: bot.Id, TeamId: bot.TeamId, TriggerWords: h.Triggers, CallbackURLs: h.CallbackURLs, AddonWebhookId: h.Id}
		if wchan := <-Srv.Store.Webhook().SaveOutgoing(hook); wchan.Err != nil {
			c.Err = wchan.Err
			<-Srv.Store.Webhook().PermanentDeleteIncomingByAddon(addon.Id)
			<-Srv.Store.Webhook().PermanentDeleteOutgoingByAddon(addon.Id)
			return
		} else {
			wh := wchan.Data.(*model.OutgoingWebhook)
			hooks = append(hooks, InstalledAddonOutgoingHook{Key: h.Key, Token: wh.Token})
		}
	}
	installed.Webooks.Outgoing = hooks

	// pending Addon -> Compacts, WebPanels, Dialogs and Actions

	// Set the addon as installed
	team_addon := &model.TeamAddon{AddonId: addon.Id, TeamId: bot.TeamId}
	if tachan := <-Srv.Store.Addon().InstallAddon(team_addon); tachan.Err != nil {
		c.Err = tachan.Err
		<-Srv.Store.Webhook().PermanentDeleteIncomingByAddon(addon.Id)
		<-Srv.Store.Webhook().PermanentDeleteOutgoingByAddon(addon.Id)
		return
	}

	handleInstalledAndForget(addon.InstalledURL, installed)

	rm := make(map[string]string)
	rm["SUCCESS"] = "true"
	w.Write([]byte(model.MapToJson(rm)))
}

func uninstallAddon(c *Context, w http.ResponseWriter, r *http.Request) {

	if !utils.Cfg.ServiceSettings.EnableOAuthServiceProvider {
		c.Err = model.NewLocAppError("publishAddon", "web.get_access_token.disabled.app_error", nil, "")
		c.Err.StatusCode = http.StatusNotImplemented
		return
	}

	if !c.IsTeamAdmin() {
		c.Err = model.NewLocAppError("publishAddon", "api.context.permissions.app_error", nil, "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	m := model.MapFromJson(r.Body)
	addonId := m["addon_id"]
	teamId := c.Session.TeamId

	// get the systembot user
	var systembot *model.User
	if uchan := <-Srv.Store.User().GetByUsername(teamId, model.SYSTEM_BOT_NAME); uchan.Err != nil {
		c.Err = model.NewLocAppError("uninstallAddon", "store.sql_addon.uninstall.app_error", nil, uchan.Err.DetailedError)
		return
	} else {
		systembot = uchan.Data.(*model.User)
	}

	if schan := <-Srv.Store.Session().PermanentDeleteSessionsByUser(systembot.Id); schan.Err != nil {
		c.Err = model.NewLocAppError("uninstallAddon", "store.sql_addon.uninstall.app_error", nil, schan.Err.DetailedError)
		return
	}

	// get the addon
	var addon model.Addon
	if achan := <-Srv.Store.Addon().GetAddonById(addonId); achan.Err != nil {
		c.Err = achan.Err
		return
	} else {
		addon = achan.Data.(model.Addon)
	}

	// check if the addon is installed for this team
	if already := <-Srv.Store.Addon().IsInstalled(teamId, addonId); already.Err == nil {
		c.Err = model.NewLocAppError("uninstallAddon", "api.addon.uninstall.no_installed.app_error", nil, "")
		return
	}

	// get the incoming webhook for this addon in case of rollback
	var incoming []*model.IncomingWebhook
	if iwchan := <-Srv.Store.Webhook().GetIncomingByTeamAddon(teamId, addonId); iwchan.Err != nil {
		c.Err = model.NewLocAppError("uninstallAddon", "store.sql_addon.uninstall.app_error", nil, iwchan.Err.DetailedError)
		return
	} else {
		incoming = iwchan.Data.([]*model.IncomingWebhook)
	}

	// get the outgoing webhooks for this addon in case of rollback
	var outgoing []*model.OutgoingWebhook
	if owchan := <-Srv.Store.Webhook().GetOutgoingByTeamAddon(teamId, addonId); owchan.Err != nil {
		c.Err = model.NewLocAppError("uninstallAddon", "store.sql_addon.uninstall.app_error", nil, owchan.Err.DetailedError)
		return
	} else {
		outgoing = owchan.Data.([]*model.OutgoingWebhook)
	}

	// pending Addon -> Compacts, WebPanels, Dialogs and Actions

	//remove the incoming hooks (it should be only one)
	if riwchan := <-Srv.Store.Webhook().PermanentDeleteIncomingByTeamAddon(teamId, addonId); riwchan.Err != nil {
		c.Err = model.NewLocAppError("uninstallAddon", "store.sql_addon.uninstall.app_error", nil, riwchan.Err.DetailedError)
		return
	}

	// remove the outgoing webhooks
	if rowchan := <-Srv.Store.Webhook().PermanentDeleteOutgoingByTeamAddon(teamId, addonId); rowchan.Err != nil {
		c.Err = model.NewLocAppError("uninstallAddon", "store.sql_addon.uninstall.app_error", nil, rowchan.Err.DetailedError)
		for _, in := range incoming {
			<-Srv.Store.Webhook().SaveIncoming(in)
		}
		return
	}

	// Remove the Team Addon
	if rtachan := <-Srv.Store.Addon().UninstallAddon(&model.TeamAddon{TeamId: teamId, AddonId: addonId}); rtachan.Err != nil {
		c.Err = model.NewLocAppError("uninstallAddon", "store.sql_addon.uninstall.app_error", nil, rtachan.Err.DetailedError)
		for _, in := range incoming {
			<-Srv.Store.Webhook().SaveIncoming(in)
		}

		for _, out := range outgoing {
			<-Srv.Store.Webhook().SaveOutgoing(out)
		}

		return
	}

	handleUninstalledAndForget(addon.UninstalledURL, teamId)

	rm := make(map[string]string)
	rm["SUCCESS"] = "true"
	w.Write([]byte(model.MapToJson(rm)))
}

func handlePublish(c *Context, w http.ResponseWriter, descriptor *model.AddonDescriptor) {
	if descriptor == nil {
		c.SetInvalidParam("publishAddon", "descriptor")
		return
	}

	if result := <-Srv.Store.Addon().Save(descriptor); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		sa := result.Data.(*model.Addon)
		secret := model.NewId()
		app := &model.OAuthApp{CreatorId: sa.Id, ClientSecret: secret, Name: sa.Name, Description: sa.Description, Homepage: sa.HomepageURL, CallbackUrls: []string{sa.HomepageURL}, Type: model.OAUTH_APP_ADDON}
		if result := <-Srv.Store.OAuth().SaveApp(app); result.Err != nil {
			c.Err = result.Err
			<-Srv.Store.Addon().PermanentDeleteAddon(sa.Id)
			return
		} else {
			app = result.Data.(*model.OAuthApp)
			app.ClientSecret = secret
			handlePublishedEventAndForget(sa.PublishedURL, app)

			rm := make(map[string]string)
			rm["AddonId"] = sa.Id
			rm["ClientId"] = app.Id
			rm["ClientSecret"] = secret
			rm["SUCCESS"] = "true"
			w.Write([]byte(model.MapToJson(rm)))
		}
	}
}

func handlePublishedEventAndForget(publishedUrl string, app *model.OAuthApp) {
	go func() {
		p := make(map[string]string)
		p["client_id"] = app.Id
		p["client_secret"] = app.ClientSecret

		jsonStr := []byte(model.MapToJson(p))

		client := &http.Client{}

		go func(url string) {
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			if _, err := client.Do(req); err != nil {
				l4g.Error(utils.T("Event POST failed, err=%s"), err.Error())
			}
		}(publishedUrl)
	}()
}

func handleInstalledAndForget(installedUrl string, installedInfo *InstalledAddon) {
	go func() {
		jsonStr, _ := json.Marshal(installedInfo)
		client := &http.Client{}

		go func(url string) {
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			if _, err := client.Do(req); err != nil {
				l4g.Error(utils.T("Event POST failed, err=%s"), err.Error())
			}
		}(installedUrl)
	}()
}

func handleUninstalledAndForget(uninstalledUrl, teamId string) {
	go func() {
		p := make(map[string]string)
		p["team_id"] = teamId

		jsonStr := []byte(model.MapToJson(p))
		client := &http.Client{}

		go func(url string) {
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			if _, err := client.Do(req); err != nil {
				l4g.Error(utils.T("Event POST failed, err=%s"), err.Error())
			}
		}(uninstalledUrl)
	}()
}
