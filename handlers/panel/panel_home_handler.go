package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func PanelHomeHandler(c *fiber.Ctx) error {
	return c.Render("panel/home/panel_home", fiber.Map{
		"Title": "Aracı Ana Sayfa",
	}, "layouts/panel_layout")
}
