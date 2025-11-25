package email

import (
	"fmt"
	"html/template"

<<<<<<<< HEAD:identity/pkg/email/util.go
	i18nPkg "smap-api/pkg/i18n"
========
	i18nPkg "smap-collector/pkg/i18n"
>>>>>>>> 9c65a15b02994a6cc9940a129c9a3c4f61fd0697:collector/pkg/email/utils.go

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Return raw template for email
func getEmailTemplate(lang string, templateType string) (*template.Template, error) {
	tmplFile := fmt.Sprintf("%s-%s.tmpl", templateType, lang)
	tmplPath := fmt.Sprintf("templates/%s", tmplFile)
	return template.Must(template.New(tmplFile).ParseFS(emailTemplates, tmplPath)), nil
}

// Translate, collect data to email template
func translateData(lang string, templateType string, data interface{}, translareData *map[string]interface{}) {
	localizer := i18nPkg.NewLocalizer(lang)
	switch templateType {
	case EmailVerificationTemplate:
		d := data.(EmailVerification)
		(*translareData)["Name"] = d.Name
		(*translareData)["OTP"] = d.OTP
		(*translareData)["OTPExpireMin"] = d.OTPExpireMin
		(*translareData)["Source"] = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "email_verification.title"})
		(*translareData)["SupportMail"] = localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "email_verification.support_email"})
		(*translareData)["Email"] = d.Email
	}
}

// Return email subject
func getEmailSubject(lang string, from string) string {
	localizer := i18nPkg.NewLocalizer(lang)
	switch from {
	case EmailVerificationTemplate:
		return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "email_verification.email_subject"})
	}
	return ""
}
