package rabbitmq

type SendEmailMsg struct {
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
