package auth

import (
	"fmt"
	"time"

	"github.com/nguyentantai21042004/smap-api/internal/models"
)

// Register
type RegisterInput struct {
	Email    string
	Password string
}

type RegisterOutput struct {
	User models.User
}

// Send OTP
type SendOTPInput struct {
	Email    string
	Password string
}

// Verify OTP
type VerifyOTPInput struct {
	Email string
	OTP   string
}

// Login
type LoginInput struct {
	Email      string
	Password   string
	Remember   bool
	UserAgent  string
	IPAddress  string
	DeviceName string
}

type LoginOutput struct {
	User  models.User
	Role  models.Role
	Token TokenOutput
}

type TokenOutput struct {
	AccessToken  string
	RefreshToken string
	SessionID    string
	TokenType    string
	ExpiresAt    time.Time
}

// Producer
type PubSendEmailMsgInput struct {
	Subject     string
	Recipient   string
	Body        string
	ReplyTo     string
	CcAddresses []string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Social Login
var SocialProviders = []string{
	Facebook,
	Google,
	Gitlab,
}

const (
	Web      = "web"
	Facebook = "facebook"
	Google   = "google"
	Gitlab   = "gitlab"
)

type GenerateTokenAndSessionInput struct {
	UserID   string
	Username string
	Remember bool
	Scope    models.Scope
}

type SocialLoginInput struct {
	Provider string
	Redirect bool
}

type SocialLoginOutput struct {
	URL string
}

// Social Callback
type SocialCallbackInput struct {
	Provider string
	Code     string
}

type SocialCallbackOutput struct {
	User  models.User
	Role  models.Role
	Token TokenOutput
}

type GetUserInfoInput struct {
	Provider string
	Code     string
}

type GetUserInfoOutput struct {
	SocialUserInfo SocialUserInfo
}

// Social User Info
type SocialUserInfo struct {
	Provider  string // "google", "facebook", "gitlab"
	ID        string
	Email     string
	Name      string
	AvatarURL string
	Username  string
}

type GitlabUserInfo struct {
	AvatarURL                      string      `json:"avatar_url"`
	Bio                            string      `json:"bio"`
	Bot                            bool        `json:"bot"`
	CanCreateGroup                 bool        `json:"can_create_group"`
	CanCreateProject               bool        `json:"can_create_project"`
	ColorSchemeID                  int         `json:"color_scheme_id"`
	CommitEmail                    string      `json:"commit_email"`
	ConfirmedAt                    string      `json:"confirmed_at"`
	CreatedAt                      string      `json:"created_at"`
	CurrentSignInAt                string      `json:"current_sign_in_at"`
	Discord                        string      `json:"discord"`
	Email                          string      `json:"email"`
	External                       bool        `json:"external"`
	ExtraSharedRunnersMinutesLimit interface{} `json:"extra_shared_runners_minutes_limit"`
	ID                             int         `json:"id"`
	Identities                     []struct {
		ExternUID      string      `json:"extern_uid"`
		Provider       string      `json:"provider"`
		SamlProviderID interface{} `json:"saml_provider_id"`
	} `json:"identities"`
	JobTitle                  string        `json:"job_title"`
	LastActivityOn            string        `json:"last_activity_on"`
	LastSignInAt              string        `json:"last_sign_in_at"`
	LinkedIn                  string        `json:"linkedin"`
	LocalTime                 string        `json:"local_time"`
	Location                  string        `json:"location"`
	Locked                    bool          `json:"locked"`
	Name                      string        `json:"name"`
	Organization              string        `json:"organization"`
	PrivateProfile            bool          `json:"private_profile"`
	ProjectsLimit             int           `json:"projects_limit"`
	Pronouns                  string        `json:"pronouns"`
	PublicEmail               string        `json:"public_email"`
	ScimIdentities            []interface{} `json:"scim_identities"`
	SharedRunnersMinutesLimit interface{}   `json:"shared_runners_minutes_limit"`
	Skype                     string        `json:"skype"`
	State                     string        `json:"state"`
	ThemeID                   int           `json:"theme_id"`
	Twitter                   string        `json:"twitter"`
	TwoFactorEnabled          bool          `json:"two_factor_enabled"`
	Username                  string        `json:"username"`
	WebURL                    string        `json:"web_url"`
	WebsiteURL                string        `json:"website_url"`
	WorkInformation           interface{}   `json:"work_information"`
}

func (r GitlabUserInfo) ToSocialUserInfo() SocialUserInfo {
	return SocialUserInfo{
		Provider:  Gitlab,
		ID:        fmt.Sprintf("%d", r.ID),
		Email:     r.Email,
		Name:      r.Name,
		AvatarURL: r.AvatarURL,
		Username:  r.Username,
	}
}

type GoogleUserInfo struct {
	Email         string `json:"email"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func (g GoogleUserInfo) ToSocialUserInfo() SocialUserInfo {
	return SocialUserInfo{
		Provider:  Google,
		ID:        g.ID,
		Email:     g.Email,
		Name:      g.Name,
		AvatarURL: g.Picture,
		Username:  "",
	}
}

type FacebookUserInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			Height       int    `json:"height"`
			IsSilhouette bool   `json:"is_silhouette"`
			URL          string `json:"url"`
			Width        int    `json:"width"`
		} `json:"data"`
	} `json:"picture"`
	Email string `json:"email"`
}

func (f FacebookUserInfo) ToSocialUserInfo() SocialUserInfo {
	return SocialUserInfo{
		Provider:  Facebook,
		ID:        f.ID,
		Email:     f.Email,
		Name:      f.Name,
		AvatarURL: f.Picture.Data.URL,
		Username:  "",
	}
}
