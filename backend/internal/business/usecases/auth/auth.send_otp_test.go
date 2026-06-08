// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth_test

import (
	"context"
	"errors"
	"testing"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/usecases/auth"
	"geregetemplateai/internal/business/usecases/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSendOTP(t *testing.T) {
	tests := []struct {
		name  string
		email string
		setup func(f *fixture)
		// wantErr / wantErrType-ийг хослуулсан, учир нь apperror.ErrTypeInternal
		// нь iota-гийн тэг — ганц sentinel нь тэр төрөлтэй мөргөлдөж, чимээгүйхэн
		// тэнцэх байсан.
		wantErr     bool
		wantErrType apperror.ErrorType
	}{
		{
			name:  "happy path calls verify.Send, persists request_id, resets attempt counter",
			email: "patrick@example.com",
			setup: func(f *fixture) {
				user := activeUser(t)
				user.Active = false // SendOTP нь зөвхөн идэвхгүй бүртгэлд хүчинтэй
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).Return(users.GetByEmailResponse{User: user}, nil).Once()
				// verify.gecloud.mn /send нь request_id буцаана; үүнийг Redis-д
				// OTPTTL-тэйгээр атомар хадгална.
				f.verify.On("Send", mock.Anything, "patrick@example.com").Return("clv_abc123", nil).Once()
				f.redis.On("SetWithTTL", mock.Anything, "user_otp:patrick@example.com", "clv_abc123", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.redis.On("Del", mock.Anything, "otp_attempts:patrick@example.com").Return(nil).Once()
			},
		},
		{
			// Enumeration-аас сэргийлж идэвхтэй данс гэдгийг ИЛЧЛЭХГҮЙ —
			// generic амжилт (nil), mailer / redis дуудлага байхгүй.
			name:  "already-active account returns generic success (no enumeration)",
			email: "patrick@example.com",
			setup: func(f *fixture) {
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).Return(users.GetByEmailResponse{User: activeUser(t)}, nil).Once()
			},
			wantErr: false,
		},
		{
			// Тодорхойгүй email мөн generic амжилт буцаана (байхгүйг илчлэхгүй).
			name:  "unknown email returns generic success (no enumeration)",
			email: "ghost@example.com",
			setup: func(f *fixture) {
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "ghost@example.com"}).
					Return(users.GetByEmailResponse{}, apperror.NotFound("email not found")).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			tt.setup(f)

			err := f.usecase.SendOTP(context.Background(), auth.SendOTPRequest{Email: tt.email})

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			var domErr *apperror.DomainError
			require.True(t, errors.As(err, &domErr))
			assert.Equal(t, tt.wantErrType, domErr.Type)
		})
	}
}
