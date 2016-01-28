// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api

import (
	"bytes"
	l4g "code.google.com/p/log4go"
	b64 "encoding/base64"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/gorilla/mux"
	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/store"
	"github.com/mattermost/platform/utils"
	"github.com/mattermost/platform/i18n"
	goi18n "github.com/nicksnyder/go-i18n/i18n"
	"github.com/mssola/user_agent"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"encoding/json"
	"html/template"
)

func InitUser(r *mux.Router) {
	l4g.Debug(T("Initializing user api routes"))

	sr := r.PathPrefix("/users").Subrouter()
	sr.Handle("/create", ApiAppHandler(createUser)).Methods("POST")
	sr.Handle("/update", ApiUserRequired(updateUser)).Methods("POST")
	sr.Handle("/update_roles", ApiUserRequired(updateRoles)).Methods("POST")
	sr.Handle("/update_active", ApiUserRequired(updateActive)).Methods("POST")
	sr.Handle("/update_notify", ApiUserRequired(updateUserNotify)).Methods("POST")
	sr.Handle("/newpassword", ApiUserRequired(updatePassword)).Methods("POST")
	sr.Handle("/send_password_reset", ApiAppHandler(sendPasswordReset)).Methods("POST")
	sr.Handle("/reset_password", ApiAppHandler(resetPassword)).Methods("POST")
	sr.Handle("/login", ApiAppHandler(login)).Methods("POST")
	sr.Handle("/logout", ApiUserRequired(logout)).Methods("POST")
	sr.Handle("/revoke_session", ApiUserRequired(revokeSession)).Methods("POST")

	sr.Handle("/newimage", ApiUserRequired(uploadProfileImage)).Methods("POST")

	sr.Handle("/me", ApiAppHandler(getMe)).Methods("GET")
	sr.Handle("/status", ApiUserRequiredActivity(getStatuses, false)).Methods("POST")
	sr.Handle("/status_info", ApiAppHandler(getUserStatusInfo)).Methods("POST")
	sr.Handle("/profiles", ApiUserRequired(getProfiles)).Methods("GET")
	sr.Handle("/profiles/{id:[A-Za-z0-9]+}", ApiUserRequired(getProfiles)).Methods("GET")
	sr.Handle("/{id:[A-Za-z0-9]+}", ApiUserRequired(getUser)).Methods("GET")
	sr.Handle("/{id:[A-Za-z0-9]+}/sessions", ApiUserRequired(getSessions)).Methods("GET")
	sr.Handle("/{id:[A-Za-z0-9]+}/audits", ApiUserRequired(getAudits)).Methods("GET")
	sr.Handle("/{id:[A-Za-z0-9]+}/image", ApiUserRequired(getProfileImage)).Methods("GET")
}

func createUser(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	if !utils.Cfg.EmailSettings.EnableSignUpWithEmail {
		c.Err = model.NewAppError("signupTeam", T("User sign-up with email is disabled."), "")
		c.Err.StatusCode = http.StatusNotImplemented
		return
	}

	user := model.UserFromJson(r.Body)

	if user == nil {
		c.SetInvalidParam("createUser", "user")
		return
	}

	// the user's username is checked to be valid when they are saved to the database

	user.EmailVerified = false

	var team *model.Team

	if result := <-Srv.Store.Team().Get(user.TeamId, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		team = result.Data.(*model.Team)
	}

	hash := r.URL.Query().Get("h")

	sendWelcomeEmail := true

	if IsVerifyHashRequired(user, team, hash) {
		data := r.URL.Query().Get("d")
		props := model.MapFromJson(strings.NewReader(data))

		if !model.ComparePassword(hash, fmt.Sprintf("%v:%v", data, utils.Cfg.EmailSettings.InviteSalt)) {
			c.Err = model.NewAppError("createUser", T("The signup link does not appear to be valid"), "")
			return
		}

		t, err := strconv.ParseInt(props["time"], 10, 64)
		if err != nil || model.GetMillis()-t > 1000*60*60*48 { // 48 hours
			c.Err = model.NewAppError("createUser", T("The signup link has expired"), "")
			return
		}

		if user.TeamId != props["id"] {
			c.Err = model.NewAppError("createUser", T("Invalid team name"), data)
			return
		}

		user.Email = props["email"]
		user.EmailVerified = true
		sendWelcomeEmail = false
	}

	if user.IsSSOUser() {
		user.EmailVerified = true
	}

	ruser := CreateUser(c, team, user, T)
	if c.Err != nil {
		return
	}

	if sendWelcomeEmail {
		sendWelcomeEmailAndForget(ruser.Id, ruser.Email, team.Name, team.DisplayName, c.GetSiteURL(), c.GetTeamURLFromTeam(team), ruser.EmailVerified, T)
	}

	w.Write([]byte(ruser.ToJson()))

}

func IsVerifyHashRequired(user *model.User, team *model.Team, hash string) bool {
	shouldVerifyHash := true

	if team.Type == model.TEAM_INVITE && len(team.AllowedDomains) > 0 && len(hash) == 0 && user != nil {
		domains := strings.Fields(strings.TrimSpace(strings.ToLower(strings.Replace(strings.Replace(team.AllowedDomains, "@", " ", -1), ",", " ", -1))))

		matched := false
		for _, d := range domains {
			if strings.HasSuffix(user.Email, "@"+d) {
				matched = true
				break
			}
		}

		if matched {
			shouldVerifyHash = false
		} else {
			return true
		}
	}

	if team.Type == model.TEAM_OPEN {
		shouldVerifyHash = false
	}

	if len(hash) > 0 {
		shouldVerifyHash = true
	}

	return shouldVerifyHash
}

func CreateUser(c *Context, team *model.Team, user *model.User, T goi18n.TranslateFunc) *model.User {

	if !utils.Cfg.TeamSettings.EnableUserCreation {
		c.Err = model.NewAppError("CreateUser", T("User creation has been disabled. Please ask your systems administrator for details."), "")
		return nil
	}

	channelRole := ""
	if team.Email == user.Email {
		user.Roles = model.ROLE_TEAM_ADMIN
		channelRole = model.CHANNEL_ROLE_ADMIN

		// Below is a speical case where the first user in the entire
		// system is granted the system_admin role instead of admin
		if result := <-Srv.Store.User().GetTotalUsersCount(T); result.Err != nil {
			c.Err = result.Err
			return nil
		} else {
			count := result.Data.(int64)
			if count <= 0 {
				user.Roles = model.ROLE_SYSTEM_ADMIN
			}
		}

	} else {
		user.Roles = ""
	}

	user.MakeNonNil()

	if result := <-Srv.Store.User().Save(user, T); result.Err != nil {
		c.Err = result.Err
		l4g.Error(T("Couldn't save the user err=%v"), result.Err)
		return nil
	} else {
		ruser := result.Data.(*model.User)

		// Soft error if there is an issue joining the default channels
		if err := JoinDefaultChannels(ruser, channelRole, T); err != nil {
			l4g.Error(T("Encountered an issue joining default channels user_id=%s, team_id=%s, err=%v"), ruser.Id, ruser.TeamId, err)
		}

		addDirectChannelsAndForget(ruser, T)

		if user.EmailVerified {
			if cresult := <-Srv.Store.User().VerifyEmail(ruser.Id, T); cresult.Err != nil {
				l4g.Error(T("Failed to set email verified err=%v"), cresult.Err)
			}
		}

		pref := model.Preference{UserId: ruser.Id, Category: model.PREFERENCE_CATEGORY_TUTORIAL_STEPS, Name: ruser.Id, Value: "0"}
		embed_preview := model.Preference{UserId: ruser.Id, Category: model.PREFERENCE_CATEGORY_ADVANCED_SETTINGS, Name: "feature_enabled_embed_preview", Value: "true"}
		mdown_preview := model.Preference{UserId: ruser.Id, Category: model.PREFERENCE_CATEGORY_ADVANCED_SETTINGS, Name: "feature_enabled_markdown_preview", Value: "true"}
		if presult := <-Srv.Store.Preference().Save(&model.Preferences{pref, embed_preview, mdown_preview}, T); presult.Err != nil {
			l4g.Error(T("Encountered error saving tutorial preference, err=%v"), presult.Err.Message)
		}

		ruser.Sanitize(map[string]bool{})

		// This message goes to every channel, so the channelId is irrelevant
		message := model.NewMessage(team.Id, "", ruser.Id, model.ACTION_NEW_USER)

		PublishAndForget(message)

		return ruser
	}
}

func sendWelcomeEmailAndForget(userId, email, teamName, teamDisplayName, siteURL, teamURL string, verified bool, T goi18n.TranslateFunc) {
	go func() {

		subjectPage := NewServerTemplatePage("welcome_subject")
		subjectPage.Props["TeamDisplayName"] = teamDisplayName
		subjectPage.Props["Subject"] = T("You joined")
		bodyPage := NewServerTemplatePage("welcome_body")
		bodyPage.Props["SiteURL"] = siteURL
		bodyPage.Props["TeamURL"] = teamURL
		bodyPage.Props["VerifyTitle"] = T("You've been invited")
		bodyPage.Props["VerifyInfo"] = T("Please verify your email address by clicking below.")
		bodyPage.Props["VerifyButton"] = T("Verify Email")
		bodyPage.Props["Info"] = T("You can sign-in to your new team from the web address:")
		bodyPage.Props["Extra"] = T("Mattermost lets you share messages and files from your PC or phone, with instant search and archiving.")
		bodyPage.Html["Footer"] = template.HTML(T("footer"))
		bodyPage.Html["EmailInfo"] = template.HTML(fmt.Sprintf(T("email_info"), utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["SiteName"]))

		if !verified {
			link := fmt.Sprintf("%s/verify_email?uid=%s&hid=%s&teamname=%s&email=%s", siteURL, userId, model.HashPassword(userId), teamName, email)
			bodyPage.Props["VerifyUrl"] = link
		}

		if err := utils.SendMail(email, subjectPage.Render(), bodyPage.Render(), T); err != nil {
			l4g.Error(T("Failed to send welcome email successfully err=%v"), err)
		}
	}()
}

func addDirectChannelsAndForget(user *model.User, T goi18n.TranslateFunc) {
	go func() {
		var profiles map[string]*model.User
		if result := <-Srv.Store.User().GetProfiles(user.TeamId, T); result.Err != nil {
			l4g.Error(T("Failed to add direct channel preferences for user user_id=%s, team_id=%s, err=%v"), user.Id, user.TeamId, result.Err.Error())
			return
		} else {
			profiles = result.Data.(map[string]*model.User)
		}

		var preferences model.Preferences

		for id := range profiles {
			if id == user.Id {
				continue
			}

			profile := profiles[id]

			preference := model.Preference{
				UserId:   user.Id,
				Category: model.PREFERENCE_CATEGORY_DIRECT_CHANNEL_SHOW,
				Name:     profile.Id,
				Value:    "true",
			}

			preferences = append(preferences, preference)

			if len(preferences) >= 10 {
				break
			}
		}

		if result := <-Srv.Store.Preference().Save(&preferences, T); result.Err != nil {
			l4g.Error(T("Failed to add direct channel preferences for new user user_id=%s, eam_id=%s, err=%v"), user.Id, user.TeamId, result.Err.Error())
		}
	}()
}

func SendVerifyEmailAndForget(userId, userEmail, teamName, teamDisplayName, siteURL, teamURL string, T goi18n.TranslateFunc) {
	go func() {

		link := fmt.Sprintf("%s/verify_email?uid=%s&hid=%s&teamname=%s&email=%s", siteURL, userId, model.HashPassword(userId), teamName, userEmail)

		subjectPage := NewServerTemplatePage("verify_subject")
		subjectPage.Props["SiteURL"] = siteURL
		subjectPage.Props["TeamDisplayName"] = teamDisplayName
		subjectPage.Props["Subject"] = T("Email Verification")
		bodyPage := NewServerTemplatePage("verify_body")
		bodyPage.Props["SiteURL"] = siteURL
		bodyPage.Props["Title"] = T("You've been invited")
		bodyPage.Props["Info"] = T("Please verify your email address by clicking below.")
		bodyPage.Props["Button"] = T("Verify Email")
		bodyPage.Props["TeamDisplayName"] = teamDisplayName
		bodyPage.Props["VerifyUrl"] = link
		bodyPage.Html["Footer"] = template.HTML(T("footer"))
		bodyPage.Html["EmailInfo"] = template.HTML(fmt.Sprintf(T("email_info"), utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["SiteName"]))

		if err := utils.SendMail(userEmail, subjectPage.Render(), bodyPage.Render(), T); err != nil {
			l4g.Error(T("Failed to send verification email successfully err=%v"), err)
		}
	}()
}

func LoginById(c *Context, w http.ResponseWriter, r *http.Request, userId, password, deviceId string) *model.User {
	T := i18n.Language(w, r)
	if result := <-Srv.Store.User().Get(userId, T); result.Err != nil {
		c.Err = result.Err
		return nil
	} else {
		user := result.Data.(*model.User)
		if checkUserPassword(c, user, password, T) {
			Login(c, w, r, user, deviceId)
			return user
		}
	}

	return nil
}

func LoginByEmail(c *Context, w http.ResponseWriter, r *http.Request, email, name, password, deviceId string) *model.User {
	var team *model.Team
	T := i18n.Language(w, r)

	if result := <-Srv.Store.Team().GetByName(name, T); result.Err != nil {
		c.Err = result.Err
		return nil
	} else {
		team = result.Data.(*model.Team)
	}

	if result := <-Srv.Store.User().GetByEmail(team.Id, email, T); result.Err != nil {
		c.Err = result.Err
		c.Err.StatusCode = http.StatusForbidden
		return nil
	} else {
		user := result.Data.(*model.User)

		if checkUserPassword(c, user, password, T) {
			Login(c, w, r, user, deviceId)
			return user
		}
	}

	return nil
}

func LoginByZBox(c *Context, w http.ResponseWriter, r *http.Request, email, password, deviceId string) *model.User {
	var team *model.Team
	T := i18n.Language(w, r)

	if body, err := zboxAuth(email, password, T); err != nil {
		c.Err = err
		return nil
	} else {
		authData := ""
		teamName := ""
		var user *model.User
		zbu := model.ZboxUserFromJson(body)
		teamName = zbu.Team
		authData = zbu.Email
		user = model.UserFromZBoxUser(zbu)

		if len(authData) == 0 {
			c.Err = model.NewAppError("zboxOAuth", T("Could not parse auth data out of zbox user object"), "")
			return nil
		}

		if len(teamName) == 0 {
			c.Err = model.NewAppError("zboxOAuth", T("Invalid team name"), "team_name="+teamName)
			c.Err.StatusCode = http.StatusBadRequest
			return nil
		}

		if !zbu.Enabled || !zbu.ChatEnabled {
			c.Err = model.NewAppError("zboxOAuth", T("Your user is not authorized to use the Chat app"), "")
			c.Err.StatusCode = http.StatusForbidden
			return nil
		}

		if result := <-Srv.Store.Team().GetByName(teamName, T); result.Err != nil {
			c.Err = result.Err
			return nil
		} else {
			team = result.Data.(*model.Team)
		}

		if result := <-Srv.Store.User().GetByAuth(team.Id, authData, "zbox", T); result.Err != nil {
			if user != nil {
				SignUpAndLogin(c, w, r, team, user, deviceId, "zbox", T)
				if c.Err != nil {
					return nil
				}

				return user
			} else {
				c.Err = result.Err
				return nil
			}
		} else {
			user = result.Data.(*model.User)
			Login(c, w, r, user, deviceId)

			if c.Err != nil {
				return nil
			}

			return user
		}

		return nil
	}
}

func SignUpAndLogin (c *Context, w http.ResponseWriter, r *http.Request, team *model.Team, user *model.User, deviceId, service string, T goi18n.TranslateFunc) {
	euchan := Srv.Store.User().GetByEmail(team.Id, user.Email, T)

	if result := <-euchan; result.Err == nil {
		c.Err = model.NewAppError("signupCompleteOAuth", T("Team ")+team.DisplayName+T(" already has a user with the email address attached to your ")+service+" account", "email="+user.Email)
		return
	}

	if team.Email == "" {
		team.Email = user.Email
		if result := <-Srv.Store.Team().Update(team, T); result.Err != nil {
			c.Err = result.Err
			return
		}
	} else {
		found := true
		count := 0
		for found {
			if found = IsUsernameTaken(user.Username, team.Id, T); c.Err != nil {
				return
			} else if found {
				user.Username = user.Username + strconv.Itoa(count)
				count += 1
			}
		}
	}
	user.TeamId = team.Id
	user.EmailVerified = true
	if cookie, err := r.Cookie(model.SESSION_LANGUAGE); err != nil {
		if cookie != nil {
			user.Language = cookie.Value
		} else {
			user.Language = model.DEFAULT_LANGUAGE
		}
	} else {
		user.Language = model.DEFAULT_LANGUAGE
	}

	ruser := CreateUser(c, team, user, T)
	if c.Err != nil {
		return
	}

	Login(c, w, r, ruser, deviceId)

	if c.Err != nil {
		return
	}
}

func checkUserPassword(c *Context, user *model.User, password string, T goi18n.TranslateFunc) bool {

	if user.FailedAttempts >= utils.Cfg.ServiceSettings.MaximumLoginAttempts {
		c.LogAuditWithUserId(user.Id, "fail", T)
		c.Err = model.NewAppError("checkUserPassword", T("Your account is locked because of too many failed password attempts. Please reset your password."), "user_id="+user.Id)
		c.Err.StatusCode = http.StatusForbidden
		return false
	}

	if !model.ComparePassword(user.Password, password) {
		c.LogAuditWithUserId(user.Id, "fail", T)
		c.Err = model.NewAppError("checkUserPassword", T("Login failed because of invalid password"), "user_id="+user.Id)
		c.Err.StatusCode = http.StatusForbidden

		if result := <-Srv.Store.User().UpdateFailedPasswordAttempts(user.Id, user.FailedAttempts+1, T); result.Err != nil {
			c.LogError(result.Err)
		}

		return false
	} else {
		if result := <-Srv.Store.User().UpdateFailedPasswordAttempts(user.Id, 0, T); result.Err != nil {
			c.LogError(result.Err)
		}

		return true
	}

}

// User MUST be validated before calling Login
func Login(c *Context, w http.ResponseWriter, r *http.Request, user *model.User, deviceId string) {
	i18n.SetLanguageCookie(w, user.Language)
	T := i18n.Language(w, r)
	c.LogAuditWithUserId(user.Id, "attempt", T)

	if !user.EmailVerified && utils.Cfg.EmailSettings.RequireEmailVerification {
		c.Err = model.NewAppError("Login", T("Login failed because email address has not been verified"), "user_id="+user.Id)
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	if user.DeleteAt > 0 {
		c.Err = model.NewAppError("Login", T("Login failed because your account has been set to inactive.  Please contact an administrator."), "user_id="+user.Id)
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	session := &model.Session{UserId: user.Id, TeamId: user.TeamId, Roles: user.Roles, DeviceId: deviceId, IsOAuth: false}

	maxAge := model.SESSION_TIME_WEB_IN_SECS

	if len(deviceId) > 0 {
		session.SetExpireInDays(model.SESSION_TIME_MOBILE_IN_DAYS)
		maxAge = model.SESSION_TIME_MOBILE_IN_SECS
	} else {
		session.SetExpireInDays(model.SESSION_TIME_WEB_IN_DAYS)
	}

	ua := user_agent.New(r.UserAgent())

	plat := ua.Platform()
	if plat == "" {
		plat = "unknown"
	}

	os := ua.OS()
	if os == "" {
		os = "unknown"
	}

	bname, bversion := ua.Browser()
	if bname == "" {
		bname = "unknown"
	}

	if bversion == "" {
		bversion = "0.0"
	}

	session.AddProp(model.SESSION_PROP_PLATFORM, plat)
	session.AddProp(model.SESSION_PROP_OS, os)
	session.AddProp(model.SESSION_PROP_BROWSER, fmt.Sprintf("%v/%v", bname, bversion))

	if result := <-Srv.Store.Session().Save(session, T); result.Err != nil {
		c.Err = result.Err
		c.Err.StatusCode = http.StatusForbidden
		return
	} else {
		session = result.Data.(*model.Session)
		AddSessionToCache(session)
	}

	w.Header().Set(model.HEADER_TOKEN, session.Token)

	tokens := GetMultiSessionCookieTokens(r)
	multiToken := ""
	seen := make(map[string]string)
	seen[session.TeamId] = session.TeamId
	for _, token := range tokens {
		s := GetSession(token, T)
		if s != nil && !s.IsExpired() && seen[s.TeamId] == "" {
			multiToken += " " + token
			seen[s.TeamId] = s.TeamId
		}
	}

	multiToken = strings.TrimSpace(multiToken + " " + session.Token)

	multiSessionCookie := &http.Cookie{
		Name:     model.SESSION_COOKIE_TOKEN,
		Value:    multiToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
	}

	http.SetCookie(w, multiSessionCookie)

	c.Session = *session
	c.LogAuditWithUserId(user.Id, "success", T)
}

func login(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	props := model.MapFromJson(r.Body)

	if len(props["password"]) == 0 {
		c.Err = model.NewAppError("login", T("Password field must not be blank"), "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	var user *model.User
	if len(props["id"]) != 0 {
		user = LoginById(c, w, r, props["id"], props["password"], props["device_id"])
	} else if len(props["email"]) != 0 && len(props["zbox"]) != 0 {
		user = LoginByZBox(c, w, r, props["email"], props["password"], props["device_id"])
	} else if len(props["email"]) != 0 && len(props["name"]) != 0 {
		user = LoginByEmail(c, w, r, props["email"], props["name"], props["password"], props["device_id"])
	} else {
		c.Err = model.NewAppError("login", T("Either user id or team name and user email must be provided"), "")
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

func revokeSession(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	props := model.MapFromJson(r.Body)
	id := props["id"]

	if result := <-Srv.Store.Session().Get(id, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		session := result.Data.(*model.Session)

		c.LogAudit("session_id=" + session.Id, T)

		if session.IsOAuth {
			RevokeAccessToken(session.Token, T)
		} else {
			sessionCache.Remove(session.Token)

			if result := <-Srv.Store.Session().Remove(session.Id, T); result.Err != nil {
				c.Err = result.Err
				return
			} else {
				w.Write([]byte(model.MapToJson(props)))
				return
			}
		}
	}
}

func RevokeAllSession(c *Context, userId string, T goi18n.TranslateFunc) {
	if result := <-Srv.Store.Session().GetSessions(userId, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		sessions := result.Data.([]*model.Session)

		for _, session := range sessions {
			c.LogAuditWithUserId(userId, "session_id="+session.Id, T)
			if session.IsOAuth {
				RevokeAccessToken(session.Token, T)
			} else {
				sessionCache.Remove(session.Token)
				if result := <-Srv.Store.Session().Remove(session.Id, T); result.Err != nil {
					c.Err = result.Err
					return
				}
			}
		}
	}
}

func getSessions(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	params := mux.Vars(r)
	id := params["id"]

	if !c.HasPermissionsToUser(id, "getSessions", T) {
		return
	}

	if result := <-Srv.Store.Session().GetSessions(id, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		sessions := result.Data.([]*model.Session)
		for _, session := range sessions {
			session.Sanitize()
		}

		w.Write([]byte(model.SessionsToJson(sessions)))
	}
}

func logout(c *Context, w http.ResponseWriter, r *http.Request) {
	data := make(map[string]string)
	data["user_id"] = c.Session.UserId

	Logout(c, w, r)
	if c.Err == nil {
		w.Write([]byte(model.MapToJson(data)))
	}
}

func Logout(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	c.LogAudit("", T)
	c.RemoveSessionCookie(w, r)
	c.RemoveTeamCookie(w, r)
	if result := <-Srv.Store.Session().Remove(c.Session.Id, T); result.Err != nil {
		c.Err = result.Err
		return
	}
}

func getMe(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)

	if len(c.Session.UserId) == 0 {
		return
	}

	if result := <-Srv.Store.User().Get(c.Session.UserId, T); result.Err != nil {
		c.Err = result.Err
		c.RemoveSessionCookie(w, r)
		l4g.Error(T("Error in getting users profile for id=%v forcing logout"), c.Session.UserId)
		return
	} else if HandleEtag(result.Data.(*model.User).Etag(), w, r) {
		return
	} else {
		result.Data.(*model.User).Sanitize(map[string]bool{})
		w.Header().Set(model.HEADER_ETAG_SERVER, result.Data.(*model.User).Etag())
		w.Header().Set("Expires", "-1")
		w.Write([]byte(result.Data.(*model.User).ToJson()))
		return
	}
}

func getUser(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	params := mux.Vars(r)
	id := params["id"]

	if !c.HasPermissionsToUser(id, "getUser", T) {
		return
	}

	if result := <-Srv.Store.User().Get(id, T); result.Err != nil {
		c.Err = result.Err
		return
	} else if HandleEtag(result.Data.(*model.User).Etag(), w, r) {
		return
	} else {
		result.Data.(*model.User).Sanitize(map[string]bool{})
		w.Header().Set(model.HEADER_ETAG_SERVER, result.Data.(*model.User).Etag())
		w.Write([]byte(result.Data.(*model.User).ToJson()))
		return
	}
}

func getProfiles(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	params := mux.Vars(r)
	id, ok := params["id"]
	if ok {
		// You must be system admin to access another team
		if id != c.Session.TeamId {
			if !c.HasSystemAdminPermissions("getProfiles", T) {
				return
			}
		}

	} else {
		id = c.Session.TeamId
	}

	etag := (<-Srv.Store.User().GetEtagForProfiles(id)).Data.(string)
	if HandleEtag(etag, w, r) {
		return
	}

	if result := <-Srv.Store.User().GetProfiles(id, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		profiles := result.Data.(map[string]*model.User)

		for k, p := range profiles {
			options := utils.Cfg.GetSanitizeOptions()
			options["passwordupdate"] = false

			if c.IsSystemAdmin() {
				options["fullname"] = true
				options["email"] = true
			}

			p.Sanitize(options)
			p.ClearNonProfileFields()
			profiles[k] = p
		}

		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
		w.Write([]byte(model.UserMapToJson(profiles)))
		return
	}
}

func getAudits(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	params := mux.Vars(r)
	id := params["id"]

	if !c.HasPermissionsToUser(id, "getAudits", T) {
		return
	}

	userChan := Srv.Store.User().Get(id, T)
	auditChan := Srv.Store.Audit().Get(id, 20, T)

	if c.Err = (<-userChan).Err; c.Err != nil {
		return
	}

	if result := <-auditChan; result.Err != nil {
		c.Err = result.Err
		return
	} else {
		audits := result.Data.(model.Audits)
		etag := audits.Etag()

		if HandleEtag(etag, w, r) {
			return
		}

		if len(etag) > 0 {
			w.Header().Set(model.HEADER_ETAG_SERVER, etag)
		}

		w.Write([]byte(audits.ToJson()))
		return
	}
}

func createProfileImage(username string, userId string, T goi18n.TranslateFunc) ([]byte, *model.AppError) {
	colors := []color.NRGBA{
		{197, 8, 126, 255},
		{227, 207, 18, 255},
		{28, 181, 105, 255},
		{35, 188, 224, 255},
		{116, 49, 196, 255},
		{197, 8, 126, 255},
		{197, 19, 19, 255},
		{250, 134, 6, 255},
		{227, 207, 18, 255},
		{123, 201, 71, 255},
		{28, 181, 105, 255},
		{35, 188, 224, 255},
		{116, 49, 196, 255},
		{197, 8, 126, 255},
		{197, 19, 19, 255},
		{250, 134, 6, 255},
		{227, 207, 18, 255},
		{123, 201, 71, 255},
		{28, 181, 105, 255},
		{35, 188, 224, 255},
		{116, 49, 196, 255},
		{197, 8, 126, 255},
		{197, 19, 19, 255},
		{250, 134, 6, 255},
		{227, 207, 18, 255},
		{123, 201, 71, 255},
	}

	h := fnv.New32a()
	h.Write([]byte(userId))
	seed := h.Sum32()

	initial := string(strings.ToUpper(username)[0])

	fontBytes, err := ioutil.ReadFile(utils.FindDir("web/static/fonts") + utils.Cfg.FileSettings.InitialFont)
	if err != nil {
		return nil, model.NewAppError("createProfileImage", T("Could not create default profile image font"), err.Error())
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, model.NewAppError("createProfileImage", T("Could not create default profile image font"), err.Error())
	}

	width := int(utils.Cfg.FileSettings.ProfileWidth)
	height := int(utils.Cfg.FileSettings.ProfileHeight)
	color := colors[int64(seed)%int64(len(colors))]
	dstImg := image.NewRGBA(image.Rect(0, 0, width, height))
	srcImg := image.White
	draw.Draw(dstImg, dstImg.Bounds(), &image.Uniform{color}, image.ZP, draw.Src)
	size := float64((width + height) / 4)

	c := freetype.NewContext()
	c.SetFont(font)
	c.SetFontSize(size)
	c.SetClip(dstImg.Bounds())
	c.SetDst(dstImg)
	c.SetSrc(srcImg)

	pt := freetype.Pt(width/6, height*2/3)
	_, err = c.DrawString(initial, pt)
	if err != nil {
		return nil, model.NewAppError("createProfileImage", T("Could not add user initial to default profile picture"), err.Error())
	}

	buf := new(bytes.Buffer)

	if imgErr := png.Encode(buf, dstImg); imgErr != nil {
		return nil, model.NewAppError("createProfileImage", T("Could not encode default profile image"), imgErr.Error())
	} else {
		return buf.Bytes(), nil
	}
}

func getProfileImage(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	params := mux.Vars(r)
	id := params["id"]

	if result := <-Srv.Store.User().Get(id, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		var img []byte

		if len(utils.Cfg.FileSettings.DriverName) == 0 {
			var err *model.AppError
			if img, err = createProfileImage(result.Data.(*model.User).Username, id, T); err != nil {
				c.Err = err
				return
			}
		} else {
			path := "teams/" + c.Session.TeamId + "/users/" + id + "/profile.png"

			if data, err := readFile(path, T); err != nil {

				if img, err = createProfileImage(result.Data.(*model.User).Username, id, T); err != nil {
					c.Err = err
					return
				}

				if err := writeFile(img, path, T); err != nil {
					c.Err = err
					return
				}

			} else {
				img = data
			}
		}

		if c.Session.UserId == id {
			w.Header().Set("Cache-Control", "max-age=300, public") // 5 mins
		} else {
			w.Header().Set("Cache-Control", "max-age=86400, public") // 24 hrs
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(img)
	}
}

func uploadProfileImage(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	if len(utils.Cfg.FileSettings.DriverName) == 0 {
		c.Err = model.NewAppError("uploadProfileImage", T("Unable to upload file. Image storage is not configured."), "")
		c.Err.StatusCode = http.StatusNotImplemented
		return
	}

	if err := r.ParseMultipartForm(10000000); err != nil {
		c.Err = model.NewAppError("uploadProfileImage", T("Could not parse multipart form"), "")
		return
	}

	m := r.MultipartForm

	imageArray, ok := m.File["image"]
	if !ok {
		c.Err = model.NewAppError("uploadProfileImage", T("No file under 'image' in request"), "")
		c.Err.StatusCode = http.StatusBadRequest
		return
	}

	if len(imageArray) <= 0 {
		c.Err = model.NewAppError("uploadProfileImage", T("Empty array under 'image' in request"), "")
		c.Err.StatusCode = http.StatusBadRequest
		return
	}

	imageData := imageArray[0]

	file, err := imageData.Open()
	defer file.Close()
	if err != nil {
		c.Err = model.NewAppError("uploadProfileImage", T("Could not open image file"), err.Error())
		return
	}

	// Decode image config first to check dimensions before loading the whole thing into memory later on
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		c.Err = model.NewAppError("uploadProfileFile", T("Could not decode profile image config."), err.Error())
		return
	} else if config.Width*config.Height > MaxImageSize {
		c.Err = model.NewAppError("uploadProfileFile", T("Unable to upload profile image. File is too large."), err.Error())
		return
	}

	file.Seek(0, 0)

	// Decode image into Image object
	img, _, err := image.Decode(file)
	if err != nil {
		c.Err = model.NewAppError("uploadProfileImage", T("Could not decode profile image"), err.Error())
		return
	}

	// Scale profile image
	img = imaging.Resize(img, utils.Cfg.FileSettings.ProfileWidth, utils.Cfg.FileSettings.ProfileHeight, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		c.Err = model.NewAppError("uploadProfileImage", T("Could not encode profile image"), err.Error())
		return
	}

	path := "teams/" + c.Session.TeamId + "/users/" + c.Session.UserId + "/profile.png"

	if err := writeFile(buf.Bytes(), path, T); err != nil {
		c.Err = model.NewAppError("uploadProfileImage", T("Couldn't upload profile image"), "")
		return
	}

	Srv.Store.User().UpdateLastPictureUpdate(c.Session.UserId, T)

	c.LogAudit("", T)

	// write something as the response since jQuery expects a json response
	w.Write([]byte("true"))
}

func updateUser(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	user := model.UserFromJson(r.Body)

	if user == nil {
		c.SetInvalidParam("updateUser", "user")
		return
	}

	if !c.HasPermissionsToUser(user.Id, "updateUser", T) {
		return
	}

	if result := <-Srv.Store.User().Update(user, false, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		c.LogAudit("", T)

		if(user.Language != "") {
			i18n.SetLanguageCookie(w, user.Language)
		}

		rusers := result.Data.([2]*model.User)

		if rusers[0].Email != rusers[1].Email {
			if tresult := <-Srv.Store.Team().Get(rusers[1].TeamId, T); tresult.Err != nil {
				l4g.Error(tresult.Err.Message)
			} else {
				team := tresult.Data.(*model.Team)
				sendEmailChangeEmailAndForget(rusers[1].Email, rusers[0].Email, team.DisplayName, c.GetTeamURLFromTeam(team), c.GetSiteURL(), T)

				if utils.Cfg.EmailSettings.RequireEmailVerification {
					SendEmailChangeVerifyEmailAndForget(rusers[0].Id, rusers[0].Email, team.Name, team.DisplayName, c.GetSiteURL(), c.GetTeamURLFromTeam(team), T)
				}
			}
		}

		rusers[0].Password = ""
		rusers[0].AuthData = ""
		w.Write([]byte(rusers[0].ToJson()))
	}
}

func updatePassword(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	c.LogAudit("attempted", T)

	props := model.MapFromJson(r.Body)
	userId := props["user_id"]
	if len(userId) != 26 {
		c.SetInvalidParam("updatePassword", "user_id")
		return
	}

	currentPassword := props["current_password"]
	if len(currentPassword) <= 0 {
		c.SetInvalidParam("updatePassword", "current_password")
		return
	}

	newPassword := props["new_password"]
	if len(newPassword) < 5 {
		c.SetInvalidParam("updatePassword", "new_password")
		return
	}

	if userId != c.Session.UserId {
		c.Err = model.NewAppError("updatePassword", T("Update password failed because context user_id did not match props user_id"), "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	var result store.StoreResult

	if result = <-Srv.Store.User().Get(userId, T); result.Err != nil {
		c.Err = result.Err
		return
	}

	if result.Data == nil {
		c.Err = model.NewAppError("updatePassword", T("Update password failed because we couldn't find a valid account"), "")
		c.Err.StatusCode = http.StatusBadRequest
		return
	}

	user := result.Data.(*model.User)

	tchan := Srv.Store.Team().Get(user.TeamId, T)

	if user.AuthData != "" {
		c.LogAudit(T("failed - tried to update user password who was logged in through oauth"), T)
		c.Err = model.NewAppError("updatePassword", T("Update password failed because the user is logged in through an OAuth service"), "auth_service="+user.AuthService)
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	if !model.ComparePassword(user.Password, currentPassword) {
		c.Err = model.NewAppError("updatePassword", T("The \"Current Password\" you entered is incorrect. Please check that Caps Lock is off and try again."), "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	if uresult := <-Srv.Store.User().UpdatePassword(c.Session.UserId, model.HashPassword(newPassword), T); uresult.Err != nil {
		c.Err = model.NewAppError("updatePassword", T("Update password failed"), uresult.Err.Error())
		c.Err.StatusCode = http.StatusForbidden
		return
	} else {
		c.LogAudit("completed", T)

		if tresult := <-tchan; tresult.Err != nil {
			l4g.Error(tresult.Err.Message)
		} else {
			team := tresult.Data.(*model.Team)
			sendPasswordChangeEmailAndForget(user.Email, team.DisplayName, c.GetTeamURLFromTeam(team), c.GetSiteURL(), T("using the settings menu"), T)
		}

		data := make(map[string]string)
		data["user_id"] = uresult.Data.(string)
		w.Write([]byte(model.MapToJson(data)))
	}
}

func updateRoles(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	props := model.MapFromJson(r.Body)

	user_id := props["user_id"]
	if len(user_id) != 26 {
		c.SetInvalidParam("updateRoles", "user_id")
		return
	}

	new_roles := props["new_roles"]
	if !model.IsValidRoles(new_roles) {
		c.SetInvalidParam("updateRoles", "new_roles")
		return
	}

	if model.IsInRole(new_roles, model.ROLE_SYSTEM_ADMIN) && !c.IsSystemAdmin() {
		c.Err = model.NewAppError("updateRoles", T("The system_admin role can only be set by another system admin"), "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	var user *model.User
	if result := <-Srv.Store.User().Get(user_id, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		user = result.Data.(*model.User)
	}

	if !c.HasPermissionsToTeam(user.TeamId, "updateRoles", T) {
		return
	}

	if !c.IsTeamAdmin() {
		c.Err = model.NewAppError("updateRoles", T("You do not have the appropriate permissions"), "userId="+user_id)
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	if user.IsInRole(model.ROLE_SYSTEM_ADMIN) && !c.IsSystemAdmin() {
		c.Err = model.NewAppError("updateRoles", T("The system admin role can only by modified by another system admin"), "")
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	ruser := UpdateRoles(c, user, new_roles, T)
	if c.Err != nil {
		return
	}

	uchan := Srv.Store.Session().UpdateRoles(user.Id, new_roles, T)
	gchan := Srv.Store.Session().GetSessions(user.Id, T)

	if result := <-uchan; result.Err != nil {
		// soft error since the user roles were still updated
		l4g.Error(result.Err)
	}

	if result := <-gchan; result.Err != nil {
		// soft error since the user roles were still updated
		l4g.Error(result.Err)
	} else {
		sessions := result.Data.([]*model.Session)
		for _, s := range sessions {
			sessionCache.Remove(s.Token)
		}
	}

	options := utils.Cfg.GetSanitizeOptions()
	options["passwordupdate"] = false
	ruser.Sanitize(options)
	w.Write([]byte(ruser.ToJson()))
}

func UpdateRoles(c *Context, user *model.User, roles string, T goi18n.TranslateFunc) *model.User {
	// make sure there is at least 1 other active admin

	if !model.IsInRole(roles, model.ROLE_SYSTEM_ADMIN) {
		if model.IsInRole(user.Roles, model.ROLE_TEAM_ADMIN) && !model.IsInRole(roles, model.ROLE_TEAM_ADMIN) {
			if result := <-Srv.Store.User().GetProfiles(user.TeamId, T); result.Err != nil {
				c.Err = result.Err
				return nil
			} else {
				activeAdmins := -1
				profileUsers := result.Data.(map[string]*model.User)
				for _, profileUser := range profileUsers {
					if profileUser.DeleteAt == 0 && model.IsInRole(profileUser.Roles, model.ROLE_TEAM_ADMIN) {
						activeAdmins = activeAdmins + 1
					}
				}

				if activeAdmins <= 0 {
					c.Err = model.NewAppError("updateRoles", T("There must be at least one active admin"), "")
					return nil
				}
			}
		}
	}

	user.Roles = roles

	var ruser *model.User
	if result := <-Srv.Store.User().Update(user, true, T); result.Err != nil {
		c.Err = result.Err
		return nil
	} else {
		c.LogAuditWithUserId(user.Id, "roles="+roles, T)
		ruser = result.Data.([2]*model.User)[0]
	}

	return ruser
}

func updateActive(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	props := model.MapFromJson(r.Body)

	user_id := props["user_id"]
	if len(user_id) != 26 {
		c.SetInvalidParam("updateActive", "user_id")
		return
	}

	active := props["active"] == "true"

	var user *model.User
	if result := <-Srv.Store.User().Get(user_id, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		user = result.Data.(*model.User)
	}

	if !c.HasPermissionsToTeam(user.TeamId, "updateActive", T) {
		return
	}

	if !c.IsTeamAdmin() {
		c.Err = model.NewAppError("updateActive", T("You do not have the appropriate permissions"), "userId="+user_id)
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	// make sure there is at least 1 other active admin
	if !active && model.IsInRole(user.Roles, model.ROLE_TEAM_ADMIN) {
		if result := <-Srv.Store.User().GetProfiles(user.TeamId, T); result.Err != nil {
			c.Err = result.Err
			return
		} else {
			activeAdmins := -1
			profileUsers := result.Data.(map[string]*model.User)
			for _, profileUser := range profileUsers {
				if profileUser.DeleteAt == 0 && model.IsInRole(profileUser.Roles, model.ROLE_TEAM_ADMIN) {
					activeAdmins = activeAdmins + 1
				}
			}

			if activeAdmins <= 0 {
				c.Err = model.NewAppError("updateRoles", T("There must be at least one active admin"), "userId="+user_id)
				return
			}
		}
	}

	ruser := UpdateActive(c, user, active)

	if c.Err == nil {
		w.Write([]byte(ruser.ToJson()))
	}
}

func UpdateActive(c *Context, user *model.User, active bool) *model.User {
	if active {
		user.DeleteAt = 0
	} else {
		user.DeleteAt = model.GetMillis()
	}

	if result := <-Srv.Store.User().Update(user, true, T); result.Err != nil {
		c.Err = result.Err
		return nil
	} else {
		c.LogAuditWithUserId(user.Id, fmt.Sprintf("active=%v", active), T)

		if user.DeleteAt > 0 {
			RevokeAllSession(c, user.Id, T)
		}

		ruser := result.Data.([2]*model.User)[0]
		options := utils.Cfg.GetSanitizeOptions()
		options["passwordupdate"] = false
		ruser.Sanitize(options)
		return ruser
	}
}

func PermanentDeleteUser(c *Context, user *model.User, T goi18n.TranslateFunc) *model.AppError {
	l4g.Warn(T("Attempting to permanently delete account %v id=%v"), user.Email, user.Id)
	c.Path = "/users/permanent_delete"
	c.LogAuditWithUserId(user.Id, fmt.Sprintf(T("attempt userId=%v"), user.Id), T)
	c.LogAuditWithUserId("", fmt.Sprintf(T("attempt userId=%v"), user.Id), T)
	if user.IsInRole(model.ROLE_SYSTEM_ADMIN) {
		l4g.Warn(T("You are deleting %v that is a system administrator.  You may need to set another account as the system administrator using the command line tools."), user.Email)
	}

	UpdateActive(c, user, false)

	if result := <-Srv.Store.Session().PermanentDeleteSessionsByUser(user.Id, T); result.Err != nil {
		return result.Err
	}

	if result := <-Srv.Store.OAuth().PermanentDeleteAuthDataByUser(user.Id, T); result.Err != nil {
		return result.Err
	}

	if result := <-Srv.Store.Webhook().PermanentDeleteIncomingByUser(user.Id, T); result.Err != nil {
		return result.Err
	}

	if result := <-Srv.Store.Webhook().PermanentDeleteOutgoingByUser(user.Id, T); result.Err != nil {
		return result.Err
	}

	if result := <-Srv.Store.Preference().PermanentDeleteByUser(user.Id, T); result.Err != nil {
		return result.Err
	}

	if result := <-Srv.Store.Channel().PermanentDeleteMembersByUser(user.Id, T); result.Err != nil {
		return result.Err
	}

	if result := <-Srv.Store.Post().PermanentDeleteByUser(user.Id, T); result.Err != nil {
		return result.Err
	}

	if result := <-Srv.Store.User().PermanentDelete(user.Id, T); result.Err != nil {
		return result.Err
	}

	if result := <-Srv.Store.Audit().PermanentDeleteByUser(user.Id, T); result.Err != nil {
		return result.Err
	}

	l4g.Warn(T("Permanently deleted account %v id=%v"), user.Email, user.Id)
	c.LogAuditWithUserId("", fmt.Sprintf("success userId=%v", user.Id), T)

	return nil
}

func sendPasswordReset(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	props := model.MapFromJson(r.Body)

	email := props["email"]
	if len(email) == 0 {
		c.SetInvalidParam("sendPasswordReset", "email")
		return
	}

	name := props["name"]
	if len(name) == 0 {
		c.SetInvalidParam("sendPasswordReset", "name")
		return
	}

	var team *model.Team
	if result := <-Srv.Store.Team().GetByName(name, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		team = result.Data.(*model.Team)
	}

	var user *model.User
	if result := <-Srv.Store.User().GetByEmail(team.Id, email, T); result.Err != nil {
		c.Err = model.NewAppError("sendPasswordReset", T("We couldn’t find an account with that address."), "email="+email+" team_id="+team.Id)
		return
	} else {
		user = result.Data.(*model.User)
	}

	if len(user.AuthData) != 0 {
		c.Err = model.NewAppError("sendPasswordReset", T("Cannot reset password for SSO accounts"), "userId="+user.Id+", teamId="+team.Id)
		return
	}

	newProps := make(map[string]string)
	newProps["user_id"] = user.Id
	newProps["time"] = fmt.Sprintf("%v", model.GetMillis())

	data := model.MapToJson(newProps)
	hash := model.HashPassword(fmt.Sprintf("%v:%v", data, utils.Cfg.EmailSettings.PasswordResetSalt))

	link := fmt.Sprintf("%s/reset_password?d=%s&h=%s", c.GetTeamURLFromTeam(team), url.QueryEscape(data), url.QueryEscape(hash))

	subjectPage := NewServerTemplatePage("reset_subject")
	subjectPage.Props["Subject"] = T("Reset your password")
	subjectPage.Props["SiteURL"] = c.GetSiteURL()

	bodyPage := NewServerTemplatePage("reset_body")
	bodyPage.Props["Title"] = T("You requested a password reset")
	bodyPage.Html["Info"] = template.HTML(T("To change your password, click \"Reset Password\" below.<br>If you did not mean to reset your password, please ignore this email and your password will remain the same."))
	bodyPage.Props["Button"] = T("Reset Password")
	bodyPage.Props["SiteURL"] = c.GetSiteURL()
	bodyPage.Props["ResetUrl"] = link
	bodyPage.Html["Footer"] = template.HTML(T("footer"))
	bodyPage.Html["EmailInfo"] = template.HTML(fmt.Sprintf(T("email_info"), utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["SiteName"]))

	if err := utils.SendMail(email, subjectPage.Render(), bodyPage.Render(), T); err != nil {
		c.Err = model.NewAppError("sendPasswordReset", T("Failed to send password reset email successfully"), "err="+err.Message)
		return
	}

	c.LogAuditWithUserId(user.Id, "sent="+email, T)

	w.Write([]byte(model.MapToJson(props)))
}

func resetPassword(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	props := model.MapFromJson(r.Body)

	newPassword := props["new_password"]
	if len(newPassword) < 5 {
		c.SetInvalidParam("resetPassword", "new_password")
		return
	}

	name := props["name"]
	if len(name) == 0 {
		c.SetInvalidParam("resetPassword", "name")
		return
	}

	userId := props["user_id"]
	hash := props["hash"]
	timeStr := ""

	if !c.IsSystemAdmin() {
		if len(hash) == 0 {
			c.SetInvalidParam("resetPassword", "hash")
			return
		}

		data := model.MapFromJson(strings.NewReader(props["data"]))

		userId = data["user_id"]

		timeStr = data["time"]
		if len(timeStr) == 0 {
			c.SetInvalidParam("resetPassword", "data:time")
			return
		}
	}

	if len(userId) != 26 {
		c.SetInvalidParam("resetPassword", "user_id")
		return
	}

	c.LogAuditWithUserId(userId, "attempt", T)

	var team *model.Team
	if result := <-Srv.Store.Team().GetByName(name, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		team = result.Data.(*model.Team)
	}

	var user *model.User
	if result := <-Srv.Store.User().Get(userId, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		user = result.Data.(*model.User)
	}

	if len(user.AuthData) != 0 {
		c.Err = model.NewAppError("resetPassword", T("Cannot reset password for SSO accounts"), "userId="+user.Id+", teamId="+team.Id)
		return
	}

	if user.TeamId != team.Id {
		c.Err = model.NewAppError("resetPassword", T("Trying to reset password for user on wrong team."), "userId="+user.Id+", teamId="+team.Id)
		c.Err.StatusCode = http.StatusForbidden
		return
	}

	if !c.IsSystemAdmin() {
		if !model.ComparePassword(hash, fmt.Sprintf("%v:%v", props["data"], utils.Cfg.EmailSettings.PasswordResetSalt)) {
			c.Err = model.NewAppError("resetPassword", T("The reset password link does not appear to be valid"), "")
			return
		}

		t, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil || model.GetMillis()-t > 1000*60*60 { // one hour
			c.Err = model.NewAppError("resetPassword", T("The reset link has expired"), "")
			return
		}
	}

	if result := <-Srv.Store.User().UpdatePassword(userId, model.HashPassword(newPassword), T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		c.LogAuditWithUserId(userId, "success", T)
	}

	sendPasswordChangeEmailAndForget(user.Email, team.DisplayName, c.GetTeamURLFromTeam(team), c.GetSiteURL(), T("using a reset password link"), T)

	props["new_password"] = ""
	w.Write([]byte(model.MapToJson(props)))
}

func sendPasswordChangeEmailAndForget(email, teamDisplayName, teamURL, siteURL, method string, T goi18n.TranslateFunc) {
	go func() {

		subjectPage := NewServerTemplatePage("password_change_subject")
		subjectPage.Props["SiteURL"] = siteURL
		subjectPage.Props["Subject"] = fmt.Sprintf(T("You updated your password for %v on"), teamDisplayName)
		bodyPage := NewServerTemplatePage(T("password_change_body"))
		bodyPage.Props["Title"] = T("You updated your password")
		bodyPage.Html["Info"] = template.HTML(fmt.Sprintf(T("You updated your password for %v on %v by %v.<br>If this change wasn't initiated by you, please contact your system administrator."), teamDisplayName, teamURL, method))
		bodyPage.Props["SiteURL"] = siteURL
		bodyPage.Html["Footer"] = template.HTML(T("footer"))
		bodyPage.Html["EmailInfo"] = template.HTML(fmt.Sprintf(T("email_info"), utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["SiteName"]))

		if err := utils.SendMail(email, subjectPage.Render(), bodyPage.Render(), T); err != nil {
			l4g.Error(T("Failed to send update password email successfully err=%v"), err)
		}

	}()
}

func sendEmailChangeEmailAndForget(oldEmail, newEmail, teamDisplayName, teamURL, siteURL string, T goi18n.TranslateFunc) {
	go func() {

		subjectPage := NewServerTemplatePage(T("email_change_subject"))
		subjectPage.Props["SiteURL"] = siteURL
		subjectPage.Props["Subject"] = fmt.Sprintf(T("Your email address has changed for %v"), teamDisplayName)
		bodyPage := NewServerTemplatePage("email_change_body")
		bodyPage.Props["SiteURL"] = siteURL
		bodyPage.Props["TeamURL"] = teamURL
		bodyPage.Props["Title"] = T("You updated your email")
		bodyPage.Html["Info"] = template.HTML(fmt.Sprintf("Your email address for %v has been changed to %v.<br>If you did not make this change, please contact the system administrator.", teamDisplayName, newEmail))
		bodyPage.Html["Footer"] = template.HTML(T("footer"))
		bodyPage.Html["EmailInfo"] = template.HTML(fmt.Sprintf(T("email_info"), utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["SiteName"]))

		if err := utils.SendMail(oldEmail, subjectPage.Render(), bodyPage.Render(), T); err != nil {
			l4g.Error(T("Failed to send email change notification email successfully err=%v"), err)
		}

	}()
}

func SendEmailChangeVerifyEmailAndForget(userId, newUserEmail, teamName, teamDisplayName, siteURL, teamURL string, T goi18n.TranslateFunc) {
	go func() {

		link := fmt.Sprintf("%s/verify_email?uid=%s&hid=%s&teamname=%s&email=%s", siteURL, userId, model.HashPassword(userId), teamName, newUserEmail)

		subjectPage := NewServerTemplatePage("email_change_verify_subject")
		subjectPage.Props["SiteURL"] = siteURL
		subjectPage.Props["Subject"] = fmt.Sprintf(T("Verify new email address for %v"), teamDisplayName)
		bodyPage := NewServerTemplatePage("email_change_verify_body")
		bodyPage.Props["SiteURL"] = siteURL
		bodyPage.Props["Title"] = T("You updated your email")
		bodyPage.Props["Info"] = fmt.Sprintf(T("To finish updating your email address for %v, please click the link below to confirm this is the right address."), teamDisplayName)
		bodyPage.Props["Button"] = T("Verify Email")
		bodyPage.Props["VerifyUrl"] = link
		bodyPage.Html["Footer"] = template.HTML(T("footer"))
		bodyPage.Html["EmailInfo"] = template.HTML(fmt.Sprintf(T("email_info"), utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["FeedbackEmail"], utils.ClientCfg["SiteName"]))

		if err := utils.SendMail(newUserEmail, subjectPage.Render(), bodyPage.Render(), T); err != nil {
			l4g.Error(T("Failed to send email change verification email successfully err=%v"), err)
		}
	}()
}

func updateUserNotify(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	props := model.MapFromJson(r.Body)

	user_id := props["user_id"]
	if len(user_id) != 26 {
		c.SetInvalidParam("updateUserNotify", "user_id")
		return
	}

	uchan := Srv.Store.User().Get(user_id, T)

	if !c.HasPermissionsToUser(user_id, "updateUserNotify", T) {
		return
	}

	delete(props, "user_id")

	email := props["email"]
	if len(email) == 0 {
		c.SetInvalidParam("updateUserNotify", "email")
		return
	}

	desktop_sound := props["desktop_sound"]
	if len(desktop_sound) == 0 {
		c.SetInvalidParam("updateUserNotify", "desktop_sound")
		return
	}

	desktop := props["desktop"]
	if len(desktop) == 0 {
		c.SetInvalidParam("updateUserNotify", "desktop")
		return
	}

	var user *model.User
	if result := <-uchan; result.Err != nil {
		c.Err = result.Err
		return
	} else {
		user = result.Data.(*model.User)
	}

	user.NotifyProps = props

	if result := <-Srv.Store.User().Update(user, false, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		c.LogAuditWithUserId(user.Id, "", T)

		ruser := result.Data.([2]*model.User)[0]
		options := utils.Cfg.GetSanitizeOptions()
		options["passwordupdate"] = false
		ruser.Sanitize(options)
		w.Write([]byte(ruser.ToJson()))
	}
}

func getStatuses(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	userIds := model.ArrayFromJson(r.Body)
	if len(userIds) == 0 {
		c.SetInvalidParam("getStatuses", "userIds")
		return
	}

	if result := <-Srv.Store.User().GetProfiles(c.Session.TeamId, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		profiles := result.Data.(map[string]*model.User)

		statuses := map[string]string{}
		for _, profile := range profiles {
			found := false
			for _, uid := range userIds {
				if uid == profile.Id {
					found = true
				}
			}

			if !found {
				continue
			}

			if profile.IsOffline() {
				statuses[profile.Id] = model.USER_OFFLINE
			} else if profile.IsAway() {
				statuses[profile.Id] = model.USER_AWAY
			} else {
				statuses[profile.Id] = model.USER_ONLINE
			}
		}

		//w.Header().Set("Cache-Control", "max-age=9, public") // 2 mins
		w.Write([]byte(model.MapToJson(statuses)))
		return
	}
}

func getUserStatusInfo(c *Context, w http.ResponseWriter, r *http.Request) {
	T := i18n.Language(w, r)
	props := model.ArrayFromJson(r.Body)
	if result := <-Srv.Store.User().GetUserStatusByEmails(props, T); result.Err != nil {
		c.Err = result.Err
		return
	} else {
		profiles := result.Data.(map[string]*model.User)

		type UserStatus struct {
			Id			string	`json:"id"`
			TeamName	string	`json:"team_name"`
			Status		string	`json:"status"`
		}

		statuses := map[string]UserStatus{}
		for _, profile := range profiles {
			u := UserStatus{Id: profile.Id, TeamName: profile.AuthData}
			if profile.IsOffline() {
				u.Status = model.USER_OFFLINE
			} else if profile.IsAway() {
				u.Status = model.USER_AWAY
			} else {
				u.Status = model.USER_ONLINE
			}
			statuses[profile.Email] = u
		}

		res2B, _ := json.Marshal(statuses)
		//w.Header().Set("Cache-Control", "max-age=9, public") // 2 mins
		w.Write([]byte(string(res2B)))
		return
	}
}

func GetAuthorizationCode(c *Context, w http.ResponseWriter, r *http.Request, teamName, service, redirectUri, loginHint string) {
	T := i18n.Language(w, r)

	sso := utils.Cfg.GetSSOService(service)
	if sso != nil && !sso.Enable {
		c.Err = model.NewAppError("GetAuthorizationCode", T("Unsupported OAuth service provider"), "service="+service)
		c.Err.StatusCode = http.StatusBadRequest
		return
	}

	clientId := sso.Id
	endpoint := sso.AuthEndpoint
	scope := sso.Scope

	stateProps := map[string]string{"team": teamName, "hash": model.HashPassword(clientId)}
	state := b64.StdEncoding.EncodeToString([]byte(model.MapToJson(stateProps)))

	authUrl := endpoint + "?response_type=code&client_id=" + clientId + "&redirect_uri=" + url.QueryEscape(redirectUri) + "&state=" + url.QueryEscape(state)

	if len(scope) > 0 {
		authUrl += "&scope=" + utils.UrlEncode(scope)
	}

	if len(loginHint) > 0 {
		authUrl += "&login_hint=" + utils.UrlEncode(loginHint)
	}

	http.Redirect(w, r, authUrl, http.StatusFound)
}

func AuthorizeOAuthUser(service, code, state, redirectUri string, T goi18n.TranslateFunc) (io.ReadCloser, *model.Team, *model.AppError) {
	sso := utils.Cfg.GetSSOService(service)
	if sso == nil || !sso.Enable {
		return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Unsupported OAuth service provider"), "service="+service)
	}

	stateStr := ""
	if b, err := b64.StdEncoding.DecodeString(state); err != nil {
		return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Invalid state"), err.Error())
	} else {
		stateStr = string(b)
	}

	stateProps := model.MapFromJson(strings.NewReader(stateStr))

	if !model.ComparePassword(stateProps["hash"], sso.Id) {
		return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Invalid state"), "")
	}

	teamName := stateProps["team"]
	if len(teamName) == 0 {
		return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Invalid state; missing team name"), "")
	}

	tchan := Srv.Store.Team().GetByName(teamName, T)

	p := url.Values{}
	p.Set("client_id", sso.Id)
	p.Set("client_secret", sso.Secret)
	p.Set("code", code)
	p.Set("grant_type", model.ACCESS_TOKEN_GRANT_TYPE)
	p.Set("redirect_uri", redirectUri)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", sso.TokenEndpoint, strings.NewReader(p.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	var ar *model.AccessResponse
	if resp, err := client.Do(req); err != nil {
		return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Token request failed"), err.Error())
	} else {
		ar = model.AccessResponseFromJson(resp.Body)
		if ar == nil {
			return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Bad response from token request"), "")
		}
	}

	if strings.ToLower(ar.TokenType) != model.ACCESS_TOKEN_TYPE {
		return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Bad token type"), "token_type="+ar.TokenType)
	}

	if len(ar.AccessToken) == 0 {
		return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Missing access token"), "")
	}

	p = url.Values{}
	p.Set("access_token", ar.AccessToken)
	req, _ = http.NewRequest("GET", sso.UserApiEndpoint, strings.NewReader(""))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+ar.AccessToken)

	if resp, err := client.Do(req); err != nil {
		return nil, nil, model.NewAppError("AuthorizeOAuthUser", T("Token request to ")+service+T(" failed"), err.Error())
	} else {
		if result := <-tchan; result.Err != nil {
			return nil, nil, result.Err
		} else {
			return resp.Body, result.Data.(*model.Team), nil
		}
	}

}

func IsUsernameTaken(name string, teamId string, T goi18n.TranslateFunc) bool {

	if !model.IsValidUsername(name) {
		return false
	}

	if result := <-Srv.Store.User().GetByUsername(teamId, name, T); result.Err != nil {
		return false
	} else {
		return true
	}

	return false
}

func ZimbraAuth(service string, email string, T goi18n.TranslateFunc) (io.ReadCloser, *model.AppError) {
	sso := utils.Cfg.GetSSOService(service)
	if sso == nil || !sso.Enable {
		return nil, model.NewAppError("AuthorizeOAuthUser", T("Unsupported OAuth service provider"), "service="+service)
	}

	p := url.Values{}
	p.Set("email", email)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", sso.UserApiEndpoint, strings.NewReader(p.Encode()))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	if resp, err := client.Do(req); err != nil {
		return nil, model.NewAppError("ZimbraAuth", T("Token request failed"), "service="+service)
	} else {
		if resp.StatusCode != 200 {
			return nil, model.NewAppError("ZimbraAuth", T("Unauthorized"), "service="+service)
		}
		return resp.Body, nil
	}
}

func zboxAuth(email, password string, T goi18n.TranslateFunc) (io.ReadCloser, *model.AppError) {
	sso := utils.Cfg.GetSSOService("zbox")
	if sso == nil || !sso.Enable {
		return nil, model.NewAppError("AuthorizeOAuthUser", T("Unsupported OAuth service provider"), "service=zbox")
	}

	var jsonStr = []byte(fmt.Sprintf(`{"username":"%s", "password":"%s", "zbox":"true"}`, email, password))

	client := &http.Client{}
	req, _ := http.NewRequest("POST", sso.LoginEndPoint, bytes.NewBuffer(jsonStr))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if resp, err := client.Do(req); err != nil {
		return nil, model.NewAppError("ZBoxAuth", err.Error(), "service=zbox")
	} else {
		if resp.StatusCode != 200 {
			return nil, model.NewAppError("ZBoxAuth", T("Unauthorized"), "service=zbox")
		}
		return resp.Body, nil
	}
}
