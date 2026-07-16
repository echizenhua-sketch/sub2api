package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type betaPolicyPersistRepoStub struct {
	values map[string]string
}

func (s *betaPolicyPersistRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *betaPolicyPersistRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if v, ok := s.values[key]; ok {
		return v, nil
	}
	return "", ErrSettingNotFound
}

func (s *betaPolicyPersistRepoStub) Set(ctx context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *betaPolicyPersistRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *betaPolicyPersistRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *betaPolicyPersistRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *betaPolicyPersistRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSetBetaPolicySettings_AllowsInjectAction(t *testing.T) {
	repo := &betaPolicyPersistRepoStub{values: map[string]string{}}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.SetBetaPolicySettings(context.Background(), &BetaPolicySettings{
		Rules: []BetaPolicyRule{{
			BetaToken: "context-management-2025-06-27",
			Action:    BetaPolicyActionInject,
			Scope:     BetaPolicyScopeAll,
		}},
	})
	require.NoError(t, err)

	got, err := svc.GetBetaPolicySettings(context.Background())
	require.NoError(t, err)
	require.Len(t, got.Rules, 1)
	require.Equal(t, BetaPolicyActionInject, got.Rules[0].Action)
	require.Equal(t, "context-management-2025-06-27", got.Rules[0].BetaToken)
}
