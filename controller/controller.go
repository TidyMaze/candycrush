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
	myEngine := engine.Engine{}
	myEngine.InitRandom()

	uiInst := ui.BuildUI(&myEngine.State)

	myEngine.HandleChangedAfterExplode = func(changed bool, exploded [][]bool) {
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

	myEngine.HandleExplodeFinished = func(fallen [][]bool) {
		uiInst.SetAnimStep(ui.Fall)
		uiInst.SetAnimStart()
		uiInst.Fallen = fallen
	}

	myEngine.HandleExplodeFinishedNoChange = func() {
		uiInst.SetAnimStep(ui.Idle)
	}

	myEngine.HandleFallFinished = func(newFilled [][]bool) {
		uiInst.Filled = newFilled
		uiInst.SetAnimStart()
	}

	myEngine.HandleAddMissingCandies = func() {
		uiInst.SetAnimStep(ui.Refill)
	}

	myEngine.OnScoreUpdated = func(score int) {
		uiInst.SetScore(score)
	}

	myEngine.Delay = func() {
		uiInst.Delay()
	}

	uiInst.OnSwap = func(action engine.Action) {
		myEngine.State = myEngine.Swap(action)
	}

	uiInst.OnSwapFinished = func() {
		println("Swap finished")
		myEngine.ExplodeAndFallUntilStable()
	}

	return &Controller{
		engine: &myEngine,
		ai: ai.AI{
			InnerEngine: &myEngine,
		},
		ui: uiInst,
	}
}

func (c *Controller) Run() {
	ui.RunUI(c.ui)
}
