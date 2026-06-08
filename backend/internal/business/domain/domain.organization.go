// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package domain

import "time"

// Байгууллагын төрлүүд (kind) — федератив hierarchy-ийн түвшнүүд.
const (
	OrgKindRoot     = "root"
	OrgKindMinistry = "ministry"
	OrgKindAgency   = "agency"
	OrgKindSOE      = "soe" // төрийн өмчит үйлдвэрийн газар
)

// RootOrgID нь үндэс байгууллагын тогтмол ID (migration 17-д seed).
const RootOrgID = "00000000-0000-0000-0000-000000000001"

// Organization нь байгууллагын модны нэг зангилаа. Path нь материалжсан зам
// (ltree) — дэд модны асуулга, цар хүрээний шалгалтад ашиглана.
type Organization struct {
	ID        string
	ParentID  string // root үед хоосон
	Path      string // ltree материалжсан зам
	Name      string
	Kind      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
