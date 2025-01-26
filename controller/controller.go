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

	uiInst := ui.BuildUI()

	engine.HandleChangedAfterExplode = func(changed bool, exploded [][]bool) {
		if changed {
			uiInst.SetAnimStep(ui.Explode)
			uiInst.SetAnimStart()
			uiInst.Destroyed = exploded
			uiInst.SetState(engine.State)

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
		uiInst.SetState(engine.State)
	}

	engine.HandleExplodeFinishedNoChange = func() {
		uiInst.SetAnimStep(ui.Idle)
		uiInst.SetState(engine.State)
	}

	engine.HandleFallFinished = func(newFilled [][]bool) {
		uiInst.Filled = newFilled
		uiInst.SetAnimStart()
		uiInst.SetState(engine.State)
	}

	engine.HandleAddMissingCandies = func() {
		uiInst.SetAnimStep(ui.Refill)
		uiInst.SetState(engine.State)
	}

	engine.OnScoreUpdated = func(score int) {
		uiInst.SetScore(score)
	}

	engine.Delay = func() {
		uiInst.Delay()
	}

	uiInst.OnSwap = func(fromX, fromY, toX, toY int) {
		engine.State = engine.Swap(engine.State, fromX, fromY, toX, toY)
		uiInst.SetState(engine.State)
	}

	uiInst.OnSwapFinished = func() {
		println("Swap finished")
		engine.ExplodeAndFallUntilStable()
	}

	uiInst.SetState(engine.State)

	return &Controller{
		engine: &engine,
		ai:     ai.AI{},
		ui:     uiInst,
	}
}

func (c *Controller) Run() {
	ui.RunUI(c.ui)
}
