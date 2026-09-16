package requests

import (
	"net/http"

	"github.com/leorolland/sortir.in/pkg/version"
	"github.com/pocketbase/pocketbase/core"
)

func GetVersion(e *core.RequestEvent) error {
	return e.JSON(http.StatusOK, map[string]string{"version": version.Version})
}
