package oauth

import (
	"github.com/nguyentantai21042004/smap-api/config"
	"golang.org/x/oauth2"
)

type Oauth2Config struct {
	Config      oauth2.Config
	UserInfoURL string
}

type OauthConfig struct {
	Google   Oauth2Config
	Facebook Oauth2Config
	Gitlab   Oauth2Config
}

func NewOauthConfig(cfg config.OauthConfig) OauthConfig {
	googleConfig := Oauth2Config{
		Config: oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Scopes:       cfg.Google.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.Google.AuthURL,
				TokenURL: cfg.Google.TokenURL,
			},
		},
		UserInfoURL: cfg.Google.UserInfoURL,
	}

	facebookConfig := Oauth2Config{
		Config: oauth2.Config{
			ClientID:     cfg.Facebook.ClientID,
			ClientSecret: cfg.Facebook.ClientSecret,
			RedirectURL:  cfg.Facebook.RedirectURL,
			Scopes:       cfg.Facebook.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.Facebook.AuthURL,
				TokenURL: cfg.Facebook.TokenURL,
			},
		},
		UserInfoURL: cfg.Facebook.UserInfoURL,
	}

	gitlabConfig := Oauth2Config{
		Config: oauth2.Config{
			ClientID:     cfg.Gitlab.ClientID,
			ClientSecret: cfg.Gitlab.ClientSecret,
			RedirectURL:  cfg.Gitlab.RedirectURL,
			Scopes:       cfg.Gitlab.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.Gitlab.AuthURL,
				TokenURL: cfg.Gitlab.TokenURL,
			},
		},
		UserInfoURL: cfg.Gitlab.UserInfoURL,
	}

	return OauthConfig{
		Google:   googleConfig,
		Facebook: facebookConfig,
		Gitlab:   gitlabConfig,
	}
}
