package views

import (
	"context"

	"github.com/invopop/ctxi18n/i18n"
)

// ToRoleI18N converts a role string to its internationalized representation
func ToRoleI18N(ctx context.Context, role string) string {
	if role == "fibber" {
		return i18n.T(ctx, "common.fibber")
	}
	return i18n.T(ctx, "common.normal")
}
