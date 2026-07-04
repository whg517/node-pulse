package notify

import (
	"context"
	"strings"
	"testing"
)

func TestNoopSenderNeverErrors(t *testing.T) {
	n := &NoopSender{}
	if n.Configured() {
		t.Fatal("noop sender should report Configured() == false")
	}
	if err := n.Send(context.Background(), "a@b.com", "subj", "body"); err != nil {
		t.Fatalf("noop Send returned error: %v", err)
	}
}

func TestNewReturnsNoopWhenUnconfigured(t *testing.T) {
	s := New(SMTPConfig{}) // empty host
	if _, ok := s.(*NoopSender); !ok {
		t.Fatalf("expected NoopSender for empty config, got %T", s)
	}
}

func TestNewReturnsSMTPWhenConfigured(t *testing.T) {
	s := New(SMTPConfig{Host: "smtp.example.com", From: "n@example.com"})
	if _, ok := s.(*SMTPSender); !ok {
		t.Fatalf("expected SMTPSender for configured host, got %T", s)
	}
	if !s.Configured() {
		t.Fatal("SMTPSender should report Configured() == true")
	}
}

func TestBuildMessagePlainText(t *testing.T) {
	msg, err := buildMessage("from@x.com", "to@x.com", "Hello", "body line", nil)
	if err != nil {
		t.Fatalf("buildMessage: %v", err)
	}
	s := string(msg)
	if !strings.Contains(s, "From: from@x.com") {
		t.Errorf("missing From header: %s", s)
	}
	if !strings.Contains(s, "Subject: Hello") {
		t.Errorf("missing Subject header: %s", s)
	}
	if !strings.Contains(s, "Content-Type: text/plain") {
		t.Errorf("expected text/plain content type, got: %s", s)
	}
	if !strings.Contains(s, "body line") {
		t.Errorf("missing body: %s", s)
	}
}

func TestBuildMessageWithAttachmentIsMultipart(t *testing.T) {
	msg, err := buildMessage("from@x.com", "to@x.com", "Report", "see attached",
		[]Attachment{{Filename: "r.csv", Content: []byte("a,b\n1,2\n"), ContentType: "text/csv"}})
	if err != nil {
		t.Fatalf("buildMessage: %v", err)
	}
	s := string(msg)
	if !strings.Contains(s, "multipart/mixed") {
		t.Errorf("expected multipart/mixed, got: %s", s)
	}
	if !strings.Contains(s, `filename="r.csv"`) {
		t.Errorf("missing attachment filename header: %s", s)
	}
	// base64 of "a,b\n1,2\n"
	if !strings.Contains(s, "YSxiCjEsMgo=") {
		t.Errorf("missing base64 attachment content: %s", s)
	}
}

func TestEncodeHeaderASCII(t *testing.T) {
	if got := encodeHeader("plain subject"); got != "plain subject" {
		t.Errorf("ascii subject should be unchanged, got %q", got)
	}
}
