package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"identity-srv/internal/model"
	"identity-srv/internal/user"
	"identity-srv/internal/user/repository"
)

type mockDeps struct {
	repo *repository.MockRepository
}

func initUseCase(t *testing.T) (user.UseCase, mockDeps) {
	t.Helper()

	repo := repository.NewMockRepository(t)

	return New(nil, nil, repo), mockDeps{
		repo: repo,
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("repo error")
	expectedUser := model.User{
		ID:    "user-id",
		Email: "user@example.com",
	}

	type mockRepoUpsert struct {
		isCalled bool
		input    repository.UpsertOptions
		output   model.User
		err      error
	}

	type mockRepo struct {
		upsert mockRepoUpsert
	}

	type mock struct {
		repo mockRepo
	}

	tcs := map[string]struct {
		input  user.CreateInput
		mock   mock
		output model.User
		err    error
	}{
		"success": {
			input: user.CreateInput{
				Email:     "user@example.com",
				Name:      "User Name",
				AvatarURL: "https://example.com/avatar.png",
			},
			mock: mock{
				repo: mockRepo{
					upsert: mockRepoUpsert{
						isCalled: true,
						input: repository.UpsertOptions{
							Email:     "user@example.com",
							Name:      "User Name",
							AvatarURL: "https://example.com/avatar.png",
						},
						output: expectedUser,
					},
				},
			},
			output: expectedUser,
		},
		"err repo Upsert": {
			input: user.CreateInput{
				Email:     "user@example.com",
				Name:      "User Name",
				AvatarURL: "https://example.com/avatar.png",
			},
			mock: mock{
				repo: mockRepo{
					upsert: mockRepoUpsert{
						isCalled: true,
						input: repository.UpsertOptions{
							Email:     "user@example.com",
							Name:      "User Name",
							AvatarURL: "https://example.com/avatar.png",
						},
						err: expectedErr,
					},
				},
			},
			err: expectedErr,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.repo.upsert.isCalled {
				deps.repo.EXPECT().Upsert(ctx, tc.mock.repo.upsert.input).
					Return(tc.mock.repo.upsert.output, tc.mock.repo.upsert.err)
			}

			res, err := uc.Create(ctx, tc.input)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("repo error")

	type mockRepoUpdate struct {
		isCalled bool
		input    repository.UpdateOptions
		err      error
	}

	type mockRepo struct {
		update mockRepoUpdate
	}

	type mock struct {
		repo mockRepo
	}

	tcs := map[string]struct {
		input user.UpdateInput
		mock  mock
		err   error
	}{
		"success": {
			input: user.UpdateInput{
				UserID: "user-id",
				Role:   model.RoleAdmin,
			},
			mock: mock{
				repo: mockRepo{
					update: mockRepoUpdate{
						isCalled: true,
						input: repository.UpdateOptions{
							UserID: "user-id",
							Role:   model.RoleAdmin,
						},
					},
				},
			},
		},
		"err repo Update": {
			input: user.UpdateInput{
				UserID: "user-id",
				Role:   model.RoleAdmin,
			},
			mock: mock{
				repo: mockRepo{
					update: mockRepoUpdate{
						isCalled: true,
						input: repository.UpdateOptions{
							UserID: "user-id",
							Role:   model.RoleAdmin,
						},
						err: expectedErr,
					},
				},
			},
			err: expectedErr,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.repo.update.isCalled {
				deps.repo.EXPECT().Update(ctx, tc.mock.repo.update.input).
					Return(tc.mock.repo.update.err)
			}

			err := uc.Update(ctx, tc.input)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDetail(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("repo error")
	expectedUser := model.User{
		ID:    "user-id",
		Email: "user@example.com",
	}

	type mockRepoDetail struct {
		isCalled bool
		input    repository.DetailOptions
		output   model.User
		err      error
	}

	type mockRepo struct {
		detail mockRepoDetail
	}

	type mock struct {
		repo mockRepo
	}

	tcs := map[string]struct {
		input  string
		mock   mock
		output model.User
		err    error
	}{
		"success": {
			input: "user-id",
			mock: mock{
				repo: mockRepo{
					detail: mockRepoDetail{
						isCalled: true,
						input: repository.DetailOptions{
							UserID: "user-id",
						},
						output: expectedUser,
					},
				},
			},
			output: expectedUser,
		},
		"err repo Detail": {
			input: "user-id",
			mock: mock{
				repo: mockRepo{
					detail: mockRepoDetail{
						isCalled: true,
						input: repository.DetailOptions{
							UserID: "user-id",
						},
						err: expectedErr,
					},
				},
			},
			err: expectedErr,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			uc, deps := initUseCase(t)

			if tc.mock.repo.detail.isCalled {
				deps.repo.EXPECT().Detail(ctx, tc.mock.repo.detail.input).
					Return(tc.mock.repo.detail.output, tc.mock.repo.detail.err)
			}

			res, err := uc.Detail(ctx, tc.input)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.output, res)
		})
	}
}
