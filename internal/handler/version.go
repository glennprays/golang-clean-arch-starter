package handler

import (
	"github.com/glennprays/golang-clean-arch-starter/internal/httperror"
	"github.com/glennprays/golang-clean-arch-starter/pkg/buildinfo"
	"github.com/gofiber/fiber/v2"
)

type VersionHandler struct{}

func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// Get returns the build metadata embedded at link time. Useful for
// correlating a running deploy with a specific commit.
func (h *VersionHandler) Get(c *fiber.Ctx) error {
	return httperror.OK(c, buildinfo.Get())
}
