package ai

import (
	"candycrush/engine"
	"fmt"
)

type AI struct {
	InnerEngine *engine.Engine
}

func (ai *AI) FindBestMove(state engine.State) engine.Action {
	validMoves := ai.InnerEngine.FindValidMoves(state)

	// display the valid moves in the console
	for _, move := range validMoves {
		println(fmt.Sprintf("Valid move: %v -> %v", move.From, move.To))
	}

	if len(validMoves) == 0 {
		panic("No valid moves")
	}

	// return the first move
	return validMoves[0]
}

// ScoreAction Score an action, higher is better.
// Applies the action and returns the number of cells destroyed.
// Does refill the board after exploding and falling to avoid any randomness.
func (ai *AI) ScoreAction(state engine.State, action engine.Action) int {
	return ai.ApplyActionAndResolve(state, action)
}

func (ai *AI) ApplyActionAndResolve(state engine.State, action engine.Action) int {
	//newState := ai.InnerEngine.Swap(action)
	//ai.InnerEngine.ExplodeAndFallUntilStableSync(newState)
	//return ai.InnerEngine.State.Score
	return -1
}
