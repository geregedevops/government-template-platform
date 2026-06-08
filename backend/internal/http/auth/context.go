// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package auth нь Fiber хүсэлтийн Locals-д агуулагдсан JWT claim-г
// handler-уудад хэрэгтэй битүүмжилсэн CurrentUser утга руу зохицуулна.
package auth

import (
	"errors"

	"geregetemplateai/internal/constants"
	"geregetemplateai/pkg/jwt"
	"github.com/gofiber/fiber/v3"
)

// CurrentUser нь баталгаажуулагдсан хүсэлтийн HTTP-давхаргын дүр төрх юм.
type CurrentUser struct {
	ID      string
	Email   string
	IsAdmin bool
	RoleID  int
	OrgID   string
	JTI     string
}

// ErrNotAuthenticated нь auth middleware хүлээгдэж буй Locals түлхүүрийг
// бөглөөгүй гэсэн үг (route дээр auth middleware суулгаагүй, эсвэл токен
// дээд урсгалд татгалзагдсан).
var ErrNotAuthenticated = errors.New("request is not authenticated")

// CurrentUserFromContext нь Fiber хүсэлтийн Locals-аас баталгаажуулагдсан
// хэрэглэгчийг гаргаж авна. Locals дотор танигдах claim байхгүй үед
// ErrNotAuthenticated-г буцаана; тийм тохиолдолд handler-ууд 401-ээр
// хариулах ёстой.
func CurrentUserFromContext(c fiber.Ctx) (CurrentUser, error) {
	raw := c.Locals(constants.CtxAuthenticatedUserKey)
	if raw == nil {
		return CurrentUser{}, ErrNotAuthenticated
	}
	claims, ok := raw.(jwt.JwtCustomClaim)
	if !ok {
		return CurrentUser{}, ErrNotAuthenticated
	}
	return CurrentUser{
		ID:      claims.UserID,
		Email:   claims.Email,
		IsAdmin: claims.IsAdmin,
		RoleID:  claims.RoleID,
		OrgID:   claims.OrgID,
		JTI:     claims.ID,
	}, nil
}
