package alert

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/notify"
)

// --- fakes for email_notifier_test ---

type fakePrefsRepo struct {
	subs []*models.NotificationPrefs
	err  error
}

func (f *fakePrefsRepo) GetByUserID(ctx context.Context, userID string) (*models.NotificationPrefs, error) {
	return &models.NotificationPrefs{UserID: userID, EmailEnabled: true, MinAlertLevel: "P1"}, nil
}
func (f *fakePrefsRepo) Upsert(ctx context.Context, userID string, req *models.UpdateNotificationPrefsRequest) (*models.NotificationPrefs, error) {
	return &models.NotificationPrefs{UserID: userID}, nil
}
func (f *fakePrefsRepo) ListSubscribersForLevel(ctx context.Context, level string) ([]*models.NotificationPrefs, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.subs, nil
}

type fakeUserQuerier struct {
	emails map[string]string // user_id → email
}

func (f *fakeUserQuerier) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUserQuerier) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	email := f.emails[userID.String()]
	return &models.User{UserID: userID.String(), Email: &email}, nil
}
func (f *fakeUserQuerier) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, nil
}
func (f *fakeUserQuerier) CreateUser(ctx context.Context, user *models.User, passwordHash string) error {
	return nil
}
func (f *fakeUserQuerier) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	return nil
}
func (f *fakeUserQuerier) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (f *fakeUserQuerier) CountAdmins(ctx context.Context) (int, error) { return 1, nil }

type fakeSender struct {
	mu       sync.Mutex
	emails   []sentEmail
	configd  bool
}

type sentEmail struct {
	to      string
	subject string
	body    string
}

func (f *fakeSender) Send(ctx context.Context, to, subject, body string, attachments ...notify.Attachment) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.emails = append(f.emails, sentEmail{to: to, subject: subject, body: body})
	return nil
}
func (f *fakeSender) Configured() bool { return f.configd }

// --- tests ---

func TestEmailNotifier_NoSubscribers_Noop(t *testing.T) {
	sender := &fakeSender{configd: true}
	n := NewEmailNotifier(
		&fakePrefsRepo{subs: nil},
		&fakeUserQuerier{},
		sender,
		"http://localhost:6532",
	)
	err := n.NotifyAlertSubscribers(context.Background(), &models.AlertEvent{
		ID: "evt-1", Level: "P1", NodeID: "n1", Metric: "latency",
	})
	require.NoError(t, err)
	assert.Empty(t, sender.emails)
}

func TestEmailNotifier_SendsToSubscribers(t *testing.T) {
	uid1 := uuid.New().String()
	uid2 := uuid.New().String()
	override := "custom@example.com"
	sender := &fakeSender{configd: true}
	n := NewEmailNotifier(
		&fakePrefsRepo{subs: []*models.NotificationPrefs{
			{UserID: uid1, EmailEnabled: true, MinAlertLevel: "P1"},
			{UserID: uid2, EmailEnabled: true, MinAlertLevel: "P0", NotifyEmail: &override},
		}},
		&fakeUserQuerier{emails: map[string]string{uid1: "user1@example.com"}},
		sender,
		"http://localhost:6532",
	)

	err := n.NotifyAlertSubscribers(context.Background(), &models.AlertEvent{
		ID: "evt-1", Level: "P1", NodeID: "n1", Metric: "latency",
		Threshold: 100, CurrentValue: 150, CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	sender.mu.Lock()
	defer sender.mu.Unlock()
	require.Len(t, sender.emails, 2)
	// uid1 got profile email, uid2 got override
	addrs := []string{sender.emails[0].to, sender.emails[1].to}
	assert.Contains(t, addrs, "user1@example.com")
	assert.Contains(t, addrs, "custom@example.com")
}

func TestEmailNotifier_SkipsWhenSenderNotConfigured(t *testing.T) {
	sender := &fakeSender{configd: false}
	n := NewEmailNotifier(
		&fakePrefsRepo{subs: []*models.NotificationPrefs{{UserID: "u1"}}},
		&fakeUserQuerier{},
		sender,
		"",
	)
	err := n.NotifyAlertSubscribers(context.Background(), &models.AlertEvent{Level: "P1"})
	require.NoError(t, err)
	assert.Empty(t, sender.emails)
}

func TestEmailNotifier_SkipsUserWithNoEmail(t *testing.T) {
	uid := uuid.New().String()
	sender := &fakeSender{configd: true}
	n := NewEmailNotifier(
		&fakePrefsRepo{subs: []*models.NotificationPrefs{{UserID: uid}}},
		&fakeUserQuerier{emails: map[string]string{uid: ""}}, // empty email
		sender,
		"",
	)
	err := n.NotifyAlertSubscribers(context.Background(), &models.AlertEvent{Level: "P1"})
	require.NoError(t, err)
	assert.Empty(t, sender.emails)
}
