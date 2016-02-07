// Copyright (c) 2016 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package store

import (
	l4g "github.com/alecthomas/log4go"
	"github.com/mattermost/platform/model"
	"time"
)

type StoreResult struct {
	Data interface{}
	Err  *model.AppError
}

type StoreChannel chan StoreResult

func Must(sc StoreChannel) interface{} {
	r := <-sc
	if r.Err != nil {
		l4g.Close()
		time.Sleep(time.Second)
		panic(r.Err)
	}

	return r.Data
}

type Store interface {
	Team() TeamStore
	Channel() ChannelStore
	Post() PostStore
	User() UserStore
	Audit() AuditStore
	Session() SessionStore
	OAuth() OAuthStore
	System() SystemStore
	Webhook() WebhookStore
	Command() CommandStore
	Preference() PreferenceStore
	License() LicenseStore
	MarkSystemRanUnitTests()
	Addon() AddonStore
	Close()
}

type TeamStore interface {
	Save(team *model.Team) StoreChannel
	Update(team *model.Team) StoreChannel
	UpdateDisplayName(name string, teamId string) StoreChannel
	Get(id string) StoreChannel
	GetByName(name string) StoreChannel
	GetTeamsForEmail(domain string) StoreChannel
	GetAll() StoreChannel
	GetAllTeamListing() StoreChannel
	GetByInviteId(inviteId string) StoreChannel
	PermanentDelete(teamId string) StoreChannel
}

type ChannelStore interface {
	Save(channel *model.Channel) StoreChannel
	SaveDirectChannel(channel *model.Channel, member1 *model.ChannelMember, member2 *model.ChannelMember) StoreChannel
	Update(channel *model.Channel) StoreChannel
	Get(id string) StoreChannel
	GetFromMaster(id string) StoreChannel
	Delete(channelId string, time int64) StoreChannel
	PermanentDeleteByTeam(teamId string) StoreChannel
	GetByName(team_id string, domain string) StoreChannel
	GetChannels(teamId string, userId string) StoreChannel
	GetMoreChannels(teamId string, userId string) StoreChannel
	GetChannelCounts(teamId string, userId string) StoreChannel
	GetForExport(teamId string) StoreChannel

	SaveMember(member *model.ChannelMember) StoreChannel
	UpdateMember(member *model.ChannelMember) StoreChannel
	GetMembers(channelId string) StoreChannel
	GetMember(channelId string, userId string) StoreChannel
	GetMemberCount(channelId string) StoreChannel
	RemoveMember(channelId string, userId string) StoreChannel
	PermanentDeleteMembersByUser(userId string) StoreChannel
	GetExtraMembers(channelId string, limit int) StoreChannel
	CheckPermissionsTo(teamId string, channelId string, userId string) StoreChannel
	CheckOpenChannelPermissions(teamId string, channelId string) StoreChannel
	CheckPermissionsToByName(teamId string, channelName string, userId string) StoreChannel
	UpdateLastViewedAt(channelId string, userId string) StoreChannel
	IncrementMentionCount(channelId string, userId string) StoreChannel
	AnalyticsTypeCount(teamId string, channelType string) StoreChannel

	GetAllChannels(teamId string) StoreChannel
}

type PostStore interface {
	Save(post *model.Post) StoreChannel
	Update(post *model.Post, newMessage string, newHashtags string) StoreChannel
	Get(id string) StoreChannel
	Delete(postId string, time int64) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
	GetPosts(channelId string, offset int, limit int) StoreChannel
	GetPostsBefore(channelId string, postId string, numPosts int, offset int) StoreChannel
	GetPostsAfter(channelId string, postId string, numPosts int, offset int) StoreChannel
	GetPostsSince(channelId string, time int64) StoreChannel
	GetEtag(channelId string) StoreChannel
	Search(teamId string, userId string, params *model.SearchParams) StoreChannel
	GetForExport(channelId string) StoreChannel
	AnalyticsUserCountsWithPostsByDay(teamId string) StoreChannel
	AnalyticsPostCountsByDay(teamId string) StoreChannel
	AnalyticsPostCount(teamId string, mustHaveFile bool, mustHaveHashtag bool) StoreChannel
}

type UserStore interface {
	Save(user *model.User) StoreChannel
	Update(user *model.User, allowRoleUpdate bool) StoreChannel
	UpdateLastPictureUpdate(userId string) StoreChannel
	UpdateLastPingAt(userId string, time int64) StoreChannel
	UpdateLastActivityAt(userId string, time int64) StoreChannel
	UpdateUserAndSessionActivity(userId string, sessionId string, time int64) StoreChannel
	UpdatePassword(userId, newPassword string) StoreChannel
	UpdateAuthData(userId, service, authData, email string) StoreChannel
	Get(id string) StoreChannel
	GetProfiles(teamId string) StoreChannel
	GetByEmail(teamId string, email string) StoreChannel
	GetByAuth(teamId string, authData string, authService string) StoreChannel
	GetByUsername(teamId string, username string) StoreChannel
	VerifyEmail(userId string) StoreChannel
	GetEtagForProfiles(teamId string) StoreChannel
	UpdateFailedPasswordAttempts(userId string, attempts int) StoreChannel
	GetForExport(teamId string) StoreChannel
	GetTotalUsersCount() StoreChannel
	GetTotalActiveUsersCount() StoreChannel
	GetSystemAdminProfiles() StoreChannel
	PermanentDelete(userId string) StoreChannel
	AnalyticsUniqueUserCount(teamId string) StoreChannel
}

type SessionStore interface {
	Save(session *model.Session) StoreChannel
	Get(sessionIdOrToken string) StoreChannel
	GetSessions(userId string) StoreChannel
	Remove(sessionIdOrToken string) StoreChannel
	RemoveAllSessionsForTeam(teamId string) StoreChannel
	PermanentDeleteSessionsByUser(teamId string) StoreChannel
	UpdateLastActivityAt(sessionId string, time int64) StoreChannel
	UpdateRoles(userId string, roles string) StoreChannel
	UpdateDeviceId(id string, deviceId string) StoreChannel
}

type AuditStore interface {
	Save(audit *model.Audit) StoreChannel
	Get(user_id string, limit int) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
}

type OAuthStore interface {
	SaveApp(app *model.OAuthApp) StoreChannel
	UpdateApp(app *model.OAuthApp) StoreChannel
	GetApp(id string) StoreChannel
	GetAppByUser(userId string) StoreChannel
	SaveAuthData(authData *model.AuthData) StoreChannel
	GetAuthData(code string) StoreChannel
	RemoveAuthData(code string) StoreChannel
	PermanentDeleteAuthDataByUser(userId string) StoreChannel
	SaveAccessData(accessData *model.AccessData) StoreChannel
	GetAccessData(token string) StoreChannel
	GetAccessDataByAuthCode(authCode string) StoreChannel
	RemoveAccessData(token string) StoreChannel
}

type SystemStore interface {
	Save(system *model.System) StoreChannel
	SaveOrUpdate(system *model.System) StoreChannel
	Update(system *model.System) StoreChannel
	Get() StoreChannel
}

type WebhookStore interface {
	SaveIncoming(webhook *model.IncomingWebhook) StoreChannel
	GetIncoming(id string) StoreChannel
	GetIncomingByTeam(teamId string) StoreChannel
	GetIncomingByChannel(channelId string) StoreChannel
	DeleteIncoming(webhookId string, time int64) StoreChannel
	PermanentDeleteIncomingByUser(userId string) StoreChannel
	SaveOutgoing(webhook *model.OutgoingWebhook) StoreChannel
	GetOutgoing(id string) StoreChannel
	GetOutgoingByChannel(channelId string) StoreChannel
	GetOutgoingByTeam(teamId string) StoreChannel
	DeleteOutgoing(webhookId string, time int64) StoreChannel
	PermanentDeleteOutgoingByUser(userId string) StoreChannel
	UpdateOutgoing(hook *model.OutgoingWebhook) StoreChannel
	AnalyticsIncomingCount(teamId string) StoreChannel
	AnalyticsOutgoingCount(teamId string) StoreChannel

	DeleteIncomingByAddon(addonId string, time int64) StoreChannel
	DeleteIncomingByTeamAddon(teamId, addonId string, time int64) StoreChannel
	EnableIncomingByAddon(addonId string, time int64) StoreChannel
	PermanentDeleteIncomingByAddon(addonId string) StoreChannel
	PermanentDeleteIncomingByTeamAddon(teamId, addonId string) StoreChannel
	GetIncomingByAddon(addonId string) StoreChannel
	GetIncomingByTeamAddon(teamId, addonId string) StoreChannel
	GetOutgoingByAddon(addonId string) StoreChannel
	GetOutgoingByTeamAddon(teamId, addonId string) StoreChannel
	DeleteOutgoingByAddon(addonId string, time int64) StoreChannel
	DeleteOutgoingByTeamAddon(teamId, addonId string, time int64) StoreChannel
	EnableOutgoingByAddon(addonId string, time int64) StoreChannel
	PermanentDeleteOutgoingByAddon(addonId string) StoreChannel
	PermanentDeleteOutgoingByTeamAddon(teamId, addonId string) StoreChannel
}

type CommandStore interface {
	Save(webhook *model.Command) StoreChannel
	Get(id string) StoreChannel
	GetByTeam(teamId string) StoreChannel
	Delete(commandId string, time int64) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
	Update(hook *model.Command) StoreChannel
}

type PreferenceStore interface {
	Save(preferences *model.Preferences) StoreChannel
	Get(userId string, category string, name string) StoreChannel
	GetCategory(userId string, category string) StoreChannel
	GetAll(userId string) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
	IsFeatureEnabled(feature, userId string) StoreChannel
}

type LicenseStore interface {
	Save(license *model.LicenseRecord) StoreChannel
	Get(id string) StoreChannel
}

type AddonStore interface {
	Save(descriptor *model.AddonDescriptor) StoreChannel
	GetAddons() StoreChannel
	GetAddonById(addonId string) StoreChannel
	GetAddonByName(name string) StoreChannel
	GetAddonWebooks(addonId string) StoreChannel
	GetAddonCompacts(addonId string) StoreChannel
	GetAddonWebpanels(addonId string) StoreChannel
	GetAddonDialogs(addonId string) StoreChannel
	GetAddonActions(addonId string) StoreChannel
	GetDisabledAddons() StoreChannel
	GetInstalled(teamId string) StoreChannel
	IsInstalled(teamId, addonId string) StoreChannel
	InstallAddon(teamAddon *model.TeamAddon) StoreChannel
	UninstallAddon(teamAddon *model.TeamAddon) StoreChannel
	EnableAddon(addonId string, time int64) StoreChannel
	DisableAddon(addonId string, time int64) StoreChannel
	DeleteAddon(addonId string, time int64) StoreChannel
	PermanentDeleteAddon(addonId string) StoreChannel
}
