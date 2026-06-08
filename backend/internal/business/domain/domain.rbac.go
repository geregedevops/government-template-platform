// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package domain

import "time"

// Эрхийн (permission) түлхүүрүүд — migration 13-ийн seed-тэй таарна. Код доторх
// шалгалтууд эдгээр тогтмолыг ашиглана (мөр шууд бичихгүй).
const (
	PermDashboardView   = "dashboard.view"
	PermSettingsManage  = "settings.manage"
	PermAIChat          = "ai.chat"
	PermKnowledgeManage = "knowledge.manage"
	PermVoiceTranslate  = "voice.translate"
	PermBPMManage       = "bpm.manage"
	PermUsersManage     = "users.manage"
	PermRolesManage     = "roles.manage"
	PermPersonalView    = "personal.view"
	PermManagerView     = "manager.view"
	PermOrgManage       = "org.manage"
	PermFedManage       = "fed.manage"
)

// Role нь динамик эрх (RBAC). is_system эрхүүдийг (admin/user) устгаж/түлхүүрийг
// нь өөрчилж болохгүй.
type Role struct {
	ID          int
	Key         string
	Name        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// Permission нь эрхийн каталогийн нэг бичлэг (код дотор тодорхойлогдсон, зөвхөн
// role-д онооно).
type Permission struct {
	Key      string
	Label    string
	Category string
}
