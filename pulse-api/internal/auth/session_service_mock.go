package auth

import "context"

// MockSessionService is a mock for SessionService interface
type MockSessionService struct {
	CreateSessionFunc  func(context.Context, string, string) (string, error)
	DeleteSessionFunc  func(context.Context, string) error
	GetSessionFunc       func(context.Context, string) (string, string, error)
}

func (m *MockSessionService) CreateSession(ctx context.Context, userID, role string) (string, error) {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, userID, role)
	}
	return "", nil
}

func (m *MockSessionService) DeleteSession(ctx context.Context, sessionID string) error {
	if m.DeleteSessionFunc != nil {
		return m.DeleteSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockSessionService) GetSession(ctx context.Context, sessionID string) (string, string, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc(ctx, sessionID)
	}
	return "", "", nil
}
