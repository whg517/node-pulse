package notify

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// buildMessage assembles a MIME multipart/mixed message: a text/plain part for
// the body, plus one part per attachment (base64-encoded). Returns RFC 5322
// bytes ready to write to the SMTP DATA stream.
func buildMessage(from, to, subject, body string, attachments []Attachment) ([]byte, error) {
	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      encodeHeader(subject),
		"Date":         time.Now().Format(time.RFC1123Z),
		"MIME-Version": "1.0",
	}

	var buf bytes.Buffer

	if len(attachments) == 0 {
		// Simple single-part text message.
		headers["Content-Type"] = "text/plain; charset=UTF-8"
		headers["Content-Transfer-Encoding"] = "8bit"
		writeHeaders(&buf, headers)
		buf.WriteString(body)
	} else {
		boundary := fmt.Sprintf("----------notify%016x", time.Now().UnixNano())
		headers["Content-Type"] = "multipart/mixed; boundary=\"" + boundary + "\""
		writeHeaders(&buf, headers)

		// Body part.
		fmt.Fprintf(&buf, "\r\n--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: 8bit\r\n\r\n")
		buf.WriteString(body)

		// Attachment parts.
		for _, att := range attachments {
			ct := att.ContentType
			if ct == "" {
				ct = "application/octet-stream"
			}
			fmt.Fprintf(&buf, "\r\n--%s\r\n", boundary)
			fmt.Fprintf(&buf, "Content-Type: %s\r\n", ct)
			fmt.Fprintf(&buf, "Content-Transfer-Encoding: base64\r\n")
			fmt.Fprintf(&buf, "Content-Disposition: attachment; filename=\"%s\"\r\n\r\n", att.Filename)
			b := base64.StdEncoding.EncodeToString(att.Content)
			// Wrap at 76 chars per RFC 2045.
			for i := 0; i < len(b); i += 76 {
				end := i + 76
				if end > len(b) {
					end = len(b)
				}
				buf.WriteString(b[i:end])
				buf.WriteString("\r\n")
			}
		}
		fmt.Fprintf(&buf, "\r\n--%s--\r\n", boundary)
	}

	return buf.Bytes(), nil
}

func writeHeaders(buf *bytes.Buffer, headers map[string]string) {
	// Stable-ish order: From, To, Subject, Date, then the rest.
	order := []string{"From", "To", "Subject", "Date", "MIME-Version", "Content-Type", "Content-Transfer-Encoding"}
	written := make(map[string]bool)
	for _, k := range order {
		if v, ok := headers[k]; ok {
			fmt.Fprintf(buf, "%s: %s\r\n", k, v)
			written[k] = true
		}
	}
	for k, v := range headers {
		if written[k] {
			continue
		}
		fmt.Fprintf(buf, "%s: %s\r\n", k, v)
	}
	buf.WriteString("\r\n")
}

// encodeHeader encodes a header value per RFC 2047 if it contains non-ASCII.
// For pure ASCII it returns the value unchanged (the common case).
func encodeHeader(s string) string {
	if isASCII(s) {
		return s
	}
	// Minimal RFC 2047 B-encoding for UTF-8.
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > '~' { // 0x7E
			return false
		}
	}
	return true
}

var _ = strings.TrimSpace // reserved for future header-folding helpers
