package controller

import (
	"candycrush/ai"
	"candycrush/engine"
	"candycrush/ui"
	"fmt"
)

type Controller struct {
	engine *engine.Engine
	ai     ai.AI
	ui     *ui.UI
}

func NewController() *Controller {
	engine := engine.Engine{}
	engine.InitRandom()

	uiInst := ui.BuildUI(&engine.State)

	engine.HandleChangedAfterExplode = func(changed bool, exploded [][]bool) {
		if changed {
			uiInst.SetAnimStep(ui.Explode)
			uiInst.SetAnimStart()
			uiInst.Destroyed = exploded

			println(fmt.Sprintf("Setting destroying to true, animationSince: %s", uiInst.AnimationSince))
		} else {
			println("Explode and fall until stable finished")
			uiInst.SetAnimStep(ui.Idle)
		}
	}

	engine.HandleExplodeFinished = func(fallen [][]bool) {
		uiInst.SetAnimStep(ui.Fall)
		uiInst.SetAnimStart()
		uiInst.Fallen = fallen
	}

	engine.HandleExplodeFinishedNoChange = func() {
		uiInst.SetAnimStep(ui.Idle)
	}

	engine.HandleFallFinished = func(newFilled [][]bool) {
		uiInst.Filled = newFilled
		uiInst.SetAnimStart()
	}

	engine.HandleAddMissingCandies = func() {
		uiInst.SetAnimStep(ui.Refill)
	}

	engine.OnScoreUpdated = func(score int) {
		uiInst.SetScore(score)
	}

	engine.Delay = func() {
		uiInst.Delay()
	}

	uiInst.OnSwap = func(fromX, fromY, toX, toY int) {
		engine.State = engine.Swap(engine.State, fromX, fromY, toX, toY)
	}

	uiInst.OnSwapFinished = func() {
		println("Swap finished")
		engine.ExplodeAndFallUntilStable()
	}

	return &Controller{
		engine: &engine,
		ai:     ai.AI{},
		ui:     uiInst,
	}
}

func (c *Controller) Run() {
	ui.RunUI(c.ui)
}
