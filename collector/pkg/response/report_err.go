package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"smap-collector/pkg/discord"

	"github.com/gin-gonic/gin"
)

func sendDiscordMesssageAsync(c *gin.Context, d *discord.Discord, message string) {
	go func() {
		splitMessages := splitMessageForDiscord(message)
		for _, message := range splitMessages {
			err := d.ReportBug(c, message)
			if err != nil {
				log.Printf("pkg.response.sendDiscordMesssageAsync.ReportBug: %v\n", err)
			}
		}
	}()
}

func splitMessageForDiscord(message string) []string {
	const maxLen = 5000
	var chunks []string
	var current string
	lines := strings.Split(message, "\n")

	for _, line := range lines {
		line += "\n"
		if len(current)+len(line) > maxLen {
			if current != "" {
				chunks = append(chunks, strings.TrimSuffix(current, "\n"))
				current = ""
			}
			for len(line) > maxLen {
				chunks = append(chunks, line[:maxLen])
				line = line[maxLen:]
			}
		}
		current += line
	}
	if current != "" {
		chunks = append(chunks, strings.TrimSuffix(current, "\n"))
	}
	return chunks
}

func buildInternalServerErrorDataForReportBug(c *gin.Context, errString string, backtrace []string) string {
	url := c.Request.URL.String()
	method := c.Request.Method
	params := c.Request.URL.Query().Encode()

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	body := string(bodyBytes)

	var sb strings.Builder
	sb.WriteString("================ Smap SERVICE ERROR ================\n")
	sb.WriteString(fmt.Sprintf("Route   : %s\n", url))
	sb.WriteString(fmt.Sprintf("Method  : %s\n", method))
	sb.WriteString("----------------------------------------------------\n")

	if len(c.Request.Header) > 0 {
		sb.WriteString("Headers :\n")
		for key, values := range c.Request.Header {
			sb.WriteString(fmt.Sprintf("    %s: %s\n", key, strings.Join(values, ", ")))
		}
		sb.WriteString("----------------------------------------------------\n")
	}

	if params != "" {
		sb.WriteString(fmt.Sprintf("Params  : %s\n", params))
	}

	if body != "" {
		sb.WriteString("Body    :\n")
		// Pretty print JSON if possible
		var prettyBody bytes.Buffer
		if err := json.Indent(&prettyBody, bodyBytes, "    ", "  "); err == nil {
			sb.WriteString(prettyBody.String() + "\n")
		} else {
			sb.WriteString("    " + body + "\n")
		}
		sb.WriteString("----------------------------------------------------\n")
	}

	sb.WriteString(fmt.Sprintf("Error   : %s\n", errString))

	if len(backtrace) > 0 {
		sb.WriteString("\nBacktrace:\n")
		for i, line := range backtrace {
			sb.WriteString(fmt.Sprintf("[%d]: %s\n", i, line))
		}
	}

	sb.WriteString("====================================================\n")
	return sb.String()
}
