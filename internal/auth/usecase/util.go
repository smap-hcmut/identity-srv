package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/nguyentantai21042004/smap-api/internal/appconfig/oauth"
	"github.com/nguyentantai21042004/smap-api/internal/auth"
	"github.com/nguyentantai21042004/smap-api/internal/auth/delivery/rabbitmq"
	"github.com/nguyentantai21042004/smap-api/internal/session"
	"github.com/nguyentantai21042004/smap-api/pkg/curl"
	"github.com/nguyentantai21042004/smap-api/pkg/scope"
)

// Specialize data type for Rabbitmqs
func (uc implUsecase) toSendEmailMsg(ip auth.PubSendEmailMsgInput) rabbitmq.SendEmailMsg {
	atchs := make([]rabbitmq.Attachment, 0)
	for _, a := range ip.Attachments {
		atchs = append(atchs, rabbitmq.Attachment{
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Data:        a.Data,
		})
	}
	return rabbitmq.SendEmailMsg{
		Subject:     ip.Subject,
		Recipient:   ip.Recipient,
		Body:        ip.Body,
		ReplyTo:     ip.ReplyTo,
		CcAddresses: ip.CcAddresses,
		Attachments: atchs,
	}
}

// Get OAuth config by provider
func (uc implUsecase) getOAuthConfig(ctx context.Context, provider string) (oauth.Oauth2Config, error) {
	cfgs := map[string]oauth.Oauth2Config{
		auth.Facebook: uc.oauth.Facebook,
		auth.Google:   uc.oauth.Google,
		auth.Gitlab:   uc.oauth.Gitlab,
	}

	cfg, ok := cfgs[provider]
	if !ok {
		uc.l.Errorf(ctx, "auth.usecase.getOAuthConfig: %v", auth.ErrInvalidProvider)
		return oauth.Oauth2Config{}, auth.ErrInvalidProvider
	}
	return cfg, nil
}

// Get user info from provider callback
func (uc implUsecase) getUserInfo(ctx context.Context, ip auth.GetUserInfoInput) (auth.SocialUserInfo, error) {
	oauthCfg, err := uc.getOAuthConfig(ctx, ip.Provider)
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.getUserInfo.getOAuthConfig: %v", err)
		return auth.SocialUserInfo{}, err
	}

	tk, err := oauthCfg.Config.Exchange(ctx, ip.Code)
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.getUserInfo.Exchange: %v", err)
		return auth.SocialUserInfo{}, err
	}

	resp, err := curl.Get(oauthCfg.UserInfoURL, map[string]string{
		"Authorization": "Bearer " + tk.AccessToken,
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.getUserInfo.curl.Get: %v", err)
		return auth.SocialUserInfo{}, err
	}

	uInfoTypes := map[string]interface{}{
		auth.Facebook: &auth.FacebookUserInfo{},
		auth.Google:   &auth.GoogleUserInfo{},
		auth.Gitlab:   &auth.GitlabUserInfo{},
	}

	uInfo, ok := uInfoTypes[ip.Provider]
	if !ok {
		uc.l.Errorf(ctx, "auth.usecase.getUserInfo: %v", auth.ErrInvalidProvider)
		return auth.SocialUserInfo{}, auth.ErrInvalidProvider
	}

	if err := json.Unmarshal([]byte(resp), uInfo); err != nil {
		uc.l.Errorf(ctx, "auth.usecase.getUserInfo.Decode: %v", err)
		return auth.SocialUserInfo{}, err
	}

	var suInfo auth.SocialUserInfo
	switch v := uInfo.(type) {
	case *auth.FacebookUserInfo:
		suInfo = v.ToSocialUserInfo()
	case *auth.GoogleUserInfo:
		suInfo = v.ToSocialUserInfo()
	case *auth.GitlabUserInfo:
		suInfo = v.ToSocialUserInfo()
	}

	return suInfo, nil
}

// Generate token and session
func (uc implUsecase) generateTokenAndSession(ctx context.Context, ip auth.GenerateTokenAndSessionInput) (auth.TokenOutput, error) {
	now := uc.clock()
	accExp := time.Hour
	refreshExp := 7 * 24 * time.Hour
	if ip.Remember {
		refreshExp = 30 * 24 * time.Hour
	}

	// Create access token with standard claims
	clms := jwt.StandardClaims{
		Audience:  "authenticate-api",
		ExpiresAt: now.Add(accExp).Unix(),
		IssuedAt:  now.Unix(),
		Issuer:    "authenticate-api",
		NotBefore: now.Unix(),
		Subject:   ip.UserID,
	}

	acTok, err := uc.scope.CreateToken(scope.Payload{
		StandardClaims: clms,
		UserID:         ip.UserID,
		Username:       ip.Username,
		Type:           "access",
		Refresh:        false,
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.generateTokenAndSession.scope.CreateToken: %v", err)
		return auth.TokenOutput{}, err
	}

	// Create session with refresh token
	rfTok := primitive.NewObjectID().Hex()
	so, err := uc.sessionUC.Create(ctx, ip.Scope, session.CreateSessionInput{
		UserID:       ip.UserID,
		AccessToken:  acTok,
		RefreshToken: rfTok,
		ExpiresAt:    now.Add(refreshExp),
	})
	if err != nil {
		uc.l.Errorf(ctx, "auth.usecase.generateTokenAndSession.sessionUC.Create: %v", err)
		return auth.TokenOutput{}, err
	}

	return auth.TokenOutput{
		AccessToken:  acTok,
		RefreshToken: rfTok,
		ExpiresAt:    so.Session.ExpiresAt,
		SessionID:    so.Session.ID.Hex(),
		TokenType:    "Bearer",
	}, nil
}
