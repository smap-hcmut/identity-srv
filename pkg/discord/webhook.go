package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// sendWithRetry gửi request với retry mechanism
func (d *Discord) sendWithRetry(ctx context.Context, payload *WebhookPayload) error {
	var lastErr error

	for attempt := 0; attempt <= d.config.RetryCount; attempt++ {
		if attempt > 0 {
			d.l.Infof(ctx, "pkg.discord.webhook.sendWithRetry: retrying attempt %d/%d", attempt, d.config.RetryCount)
			time.Sleep(d.config.RetryDelay)
		}

		err := d.sendRequest(ctx, payload)
		if err == nil {
			return nil
		}

		lastErr = err
		d.l.Warnf(ctx, "pkg.discord.webhook.sendWithRetry: attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("failed after %d attempts, last error: %w", d.config.RetryCount+1, lastErr)
}

// sendRequest gửi request đến Discord webhook
func (d *Discord) sendRequest(ctx context.Context, payload *WebhookPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := d.GetWebhookURL()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Smap-Bot/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// validateMessageLength kiểm tra độ dài message
func (d *Discord) validateMessageLength(content string) error {
	if len(content) > MaxMessageLength {
		return fmt.Errorf("message too long: %d characters (max: %d)", len(content), MaxMessageLength)
	}
	return nil
}

// validateEmbedLength kiểm tra độ dài embed
func (d *Discord) validateEmbedLength(embed *Embed) error {
	totalLength := len(embed.Name) + len(embed.Description)

	for _, field := range embed.Fields {
		totalLength += len(field.Name) + len(field.Value)
	}

	if totalLength > MaxEmbedLength {
		return fmt.Errorf("embed too long: %d characters (max: %d)", totalLength, MaxEmbedLength)
	}

	return nil
}

// getColorForType trả về màu cho message type
func (d *Discord) getColorForType(msgType MessageType) int {
	switch msgType {
	case MessageTypeInfo:
		return ColorInfo
	case MessageTypeSuccess:
		return ColorSuccess
	case MessageTypeWarning:
		return ColorWarning
	case MessageTypeError:
		return ColorError
	default:
		return ColorInfo
	}
}

// formatTimestamp format timestamp cho Discord
func (d *Discord) formatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

// truncateString cắt chuỗi nếu quá dài
func (d *Discord) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// SendMessage gửi message đơn giản
func (d *Discord) SendMessage(ctx context.Context, content string) error {
	if err := d.validateMessageLength(content); err != nil {
		return err
	}

	payload := &WebhookPayload{
		Content:   content,
		Username:  d.config.DefaultUsername,
		AvatarURL: d.config.DefaultAvatarURL,
	}

	return d.sendWithRetry(ctx, payload)
}

// SendEmbed gửi embed message với options
func (d *Discord) SendEmbed(ctx context.Context, options MessageOptions) error {
	embed := &Embed{
		Name:        d.truncateString(options.Name, 256),
		Description: d.truncateString(options.Description, 4096),
		Color:       d.getColorForType(options.Type),
		Fields:      options.Fields,
		Footer:      options.Footer,
		Author:      options.Author,
		Thumbnail:   options.Thumbnail,
		Image:       options.Image,
	}

	if !options.Timestamp.IsZero() {
		embed.Timestamp = d.formatTimestamp(options.Timestamp)
	}

	if err := d.validateEmbedLength(embed); err != nil {
		return err
	}

	payload := &WebhookPayload{
		Embeds:    []Embed{*embed},
		Username:  options.Username,
		AvatarURL: options.AvatarURL,
	}

	if payload.Username == "" {
		payload.Username = d.config.DefaultUsername
	}
	if payload.AvatarURL == "" {
		payload.AvatarURL = d.config.DefaultAvatarURL
	}

	return d.sendWithRetry(ctx, payload)
}

// SendError gửi error message
func (d *Discord) SendError(ctx context.Context, Name, description string, err error) error {
	fields := []EmbedField{}
	if err != nil {
		fields = append(fields, EmbedField{
			Name:   "Error",
			Value:  d.truncateString(err.Error(), 1024),
			Inline: false,
		})
	}

	options := MessageOptions{
		Type:        MessageTypeError,
		Level:       LevelHigh,
		Name:        Name,
		Description: description,
		Fields:      fields,
		Timestamp:   time.Now(),
	}

	return d.SendEmbed(ctx, options)
}

// SendSuccess gửi success message
func (d *Discord) SendSuccess(ctx context.Context, Name, description string) error {
	options := MessageOptions{
		Type:        MessageTypeSuccess,
		Level:       LevelNormal,
		Name:        Name,
		Description: description,
		Timestamp:   time.Now(),
	}

	return d.SendEmbed(ctx, options)
}

// SendWarning gửi warning message
func (d *Discord) SendWarning(ctx context.Context, Name, description string) error {
	options := MessageOptions{
		Type:        MessageTypeWarning,
		Level:       LevelNormal,
		Name:        Name,
		Description: description,
		Timestamp:   time.Now(),
	}

	return d.SendEmbed(ctx, options)
}

// SendInfo gửi info message
func (d *Discord) SendInfo(ctx context.Context, Name, description string) error {
	options := MessageOptions{
		Type:        MessageTypeInfo,
		Level:       LevelNormal,
		Name:        Name,
		Description: description,
		Timestamp:   time.Now(),
	}

	return d.SendEmbed(ctx, options)
}

// ReportBug gửi bug report (backward compatibility)
func (d *Discord) ReportBug(ctx context.Context, message string) error {
	// Truncate message if too long
	if len(message) > 4096 {
		message = message[:4093] + "..."
	}

	options := MessageOptions{
		Type:        MessageTypeError,
		Level:       LevelUrgent,
		Name:        "SMAP Service Error Report",
		Description: fmt.Sprintf("```%s```", message),
		Timestamp:   time.Now(),
	}

	return d.SendEmbed(ctx, options)
}

// SendNotification gửi notification với custom fields
func (d *Discord) SendNotification(ctx context.Context, Name, description string, fields map[string]string) error {
	var embedFields []EmbedField
	for name, value := range fields {
		embedFields = append(embedFields, EmbedField{
			Name:   d.truncateString(name, 256),
			Value:  d.truncateString(value, 1024),
			Inline: true,
		})
	}

	options := MessageOptions{
		Type:        MessageTypeInfo,
		Level:       LevelNormal,
		Name:        Name,
		Description: description,
		Fields:      embedFields,
		Timestamp:   time.Now(),
	}

	return d.SendEmbed(ctx, options)
}

// SendActivityLog gửi activity log
func (d *Discord) SendActivityLog(ctx context.Context, action, user, details string) error {
	fields := []EmbedField{
		{
			Name:   "Action",
			Value:  action,
			Inline: true,
		},
		{
			Name:   "User",
			Value:  user,
			Inline: true,
		},
	}

	if details != "" {
		fields = append(fields, EmbedField{
			Name:   "Details",
			Value:  details,
			Inline: false,
		})
	}

	options := MessageOptions{
		Type:        MessageTypeInfo,
		Level:       LevelLow,
		Name:        "Activity Log",
		Description: fmt.Sprintf("**%s** performed **%s**", user, action),
		Fields:      fields,
		Timestamp:   time.Now(),
	}

	return d.SendEmbed(ctx, options)
}
