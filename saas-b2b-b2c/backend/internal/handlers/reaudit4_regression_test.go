package handlers

import (
	"testing"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// REAUDIT-4: владение чеклистом — creator/assignee, а не "любой из tenant".
func TestChecklistOwnedByCaller_NotWholeTenant(t *testing.T) {
	tenantID := uuid.New()
	owner := &models.User{ID: uuid.New(), Role: models.RoleFranchisor, TenantID: &tenantID}
	colleague := &models.User{ID: uuid.New(), Role: models.RoleDealer, TenantID: &tenantID}
	chk := &models.Checklist{ID: uuid.New(), Title: "private", UserID: owner.ID, TenantID: &tenantID}

	require.True(t, checklistOwnedByCaller(owner, chk))
	require.False(t, checklistOwnedByCaller(colleague, chk), "коллега не владеет чужим чеклистом")
	require.True(t, checklistReadableByCaller(colleague, chk), "прочитать в своей сети можно")
}

// Исполнитель из другой сети не получает доступ даже по assigned_to.
func TestChecklistOwnedByCaller_ForeignAssigneeRejected(t *testing.T) {
	tenantA, tenantB := uuid.New(), uuid.New()
	foreign := uuid.New()
	owner := &models.User{ID: uuid.New(), Role: models.RoleFranchisor, TenantID: &tenantA}
	attacker := &models.User{ID: foreign, Role: models.RoleDealer, TenantID: &tenantB}
	chk := &models.Checklist{ID: uuid.New(), UserID: owner.ID, TenantID: &tenantA, AssignedTo: &foreign}

	require.False(t, checklistOwnedByCaller(attacker, chk), "исполнитель чужой сети не допускается")
	require.False(t, checklistReadableByCaller(attacker, chk))
}
