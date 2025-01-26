package controller

import (
	"candycrush/ai"
	"candycrush/engine"
	"candycrush/ui"
)

type Controller struct {
	engine engine.Engine
	ai     ai.AI
	ui     ui.UI
}
