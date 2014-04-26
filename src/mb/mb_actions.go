package mb

import (
	"fmt"
	"strings"
)

type Action struct {
	Spec       *ActionSpec
	Target     interface{}
	ActualCost int
}

func (a Action) String() string {
	return fmt.Sprintf("%s: %s", a.Spec.Description, a.Target)
}

type TargetType int

const (
	NoTarget TargetType = iota
	WarpathTarget
	LandTarget
	EnemyTarget
)

type PeacePipeAction int
type IncorporateAction int
type BuildAction int
type FortifyAction int
type AttackAction int
type RepairAction int
type PowwowAction int
type QuitAction int

type ActionCost int

const (
	ZeroCost ActionCost = iota
	OneCost
	TwoCost
	ChiefdomValueCost
	PalisadeValueCost
)

var actions = []ActionSpec{
	ActionSpec{"ppa", "Unopposed Peace Pipe advance", PeacePipeAction(0), WarpathTarget, OneCost},
	ActionSpec{"inc", "Incorporate a Chiefdom", IncorporateAction(0), WarpathTarget, OneCost},
	ActionSpec{"mnd", "Build a Mound", BuildAction(0), LandTarget, ChiefdomValueCost},
	ActionSpec{"frt", "Fortify Cahokia", FortifyAction(0), NoTarget, TwoCost},
	ActionSpec{"att", "Attack Hostile Army", AttackAction(0), EnemyTarget, OneCost},
	ActionSpec{"rep", "Repair Breach", RepairAction(0), NoTarget, PalisadeValueCost},
	ActionSpec{"pow", "Powwow", PowwowAction(0), WarpathTarget, TwoCost},
	ActionSpec{"qui", "Quit", QuitAction(0), NoTarget, ZeroCost},
}

type ActionSpec struct {
	Name        string
	Description string
	Type        state
	Target      TargetType
	Cost        ActionCost
}

// finder is a function that finds a unique game object given its prefix.
type finder func(string, *Game) (interface{}, error)

var findFunction = map[TargetType]finder{
	NoTarget:      findNothing,
	WarpathTarget: findWarpath,
	LandTarget:    findLand,
	EnemyTarget:   findEnemy,
}

func (g *Game) parseAction(c string) (*Action, error) {
	tokens := strings.Split(c, " ")
	as, err := findActionSpec(tokens[0])
	if err != nil {
		return nil, err
	}
	if len(tokens) < 2 && as.Target != NoTarget {
		return nil, fmt.Errorf("The %s action requires a target.", as.Description)
	}
	t := ""
	if len(tokens) > 1 {
		t = strings.ToLower(tokens[1])
	}
	target, err := findFunction[as.Target](t, g)
	if err != nil {
		return nil, err
	}
	return &Action{Spec: &as, Target: target}, nil
}

func findActionSpec(token string) (ActionSpec, error) {
	t := strings.ToLower(token)
	for _, as := range actions {
		if as.Name == t {
			return as, nil
		}
	}
	return ActionSpec{}, fmt.Errorf("Unknown action: %q", t)
}

func findNothing(string, *Game) (interface{}, error) {
	return "", nil
}

// findWarpath finds one of the five Indian tribes.
func findWarpath(t string, g *Game) (interface{}, error) {
	tribe, err := findEnemy(t, g)
	if err != nil {
		return nil, err
	}
	if tribe.(Tribe) > Caddo {
		return nil, fmt.Errorf("%q doesn't match a warpath.", t)
	}
	return tribe, nil
}

// findLand finds a land, given its abbreviation.
func findLand(t string, g *Game) (interface{}, error) {
	found := make([]Land, 0)
	for _, land := range g.Board.Lands {
		if strings.HasPrefix(strings.ToLower(land.Name), t) {
			found = append(found, land)
		}
	}
	switch {
	case len(found) == 0:
		return nil, fmt.Errorf("%q doesn't match a land.", t)
	case len(found) > 1:
		return nil, fmt.Errorf("%q matches more than one land.", t)
	}
	return found[0], nil
}

// findEnemy finds an enemy - either a tribe or the Spanish.
func findEnemy(t string, _ *Game) (interface{}, error) {
	found := make([]Tribe, 0)
	for k, v := range tribeNameLookup {
		if v > SpanishTribe {
			continue
		}
		if strings.HasPrefix(strings.ToLower(k), t) {
			found = append(found, v)
		}
	}
	switch {
	case len(found) == 0:
		return nil, fmt.Errorf("%q doesn't match an enemy.", t)
	case len(found) > 1:
		return nil, fmt.Errorf("%q matches more than one tribe.", t)
	case found[0] > Caddo && found[0] != SpanishTribe:
		return nil, fmt.Errorf("%q doesn't match an enemy.")
	}
	return found[0], nil
}

// executedAction is called whenever an action is legally performed (even if it didn't)
// succeed.
func (g *Game) executedAction() {
	g.Board.ActionPoints -= g.Action.ActualCost
}

// findPeacePipeLands finds the Land on a tribe's warpath that currently contains the
// peace pipe, if any, and the next land to which the peace pipe can be moved,
// if any.
func (g *Game) findPeacePipeLands(t Tribe) (Land, Land) {
	var oldLand, newLand Land
	for i := 1; i < 6; i++ {
		idx := toLandIndex(t, i)
		if g.Board.PeacePipes[idx] {
			oldLand = g.Board.Lands[idx]
			break
		}
	}

	if (oldLand == Land{}) {
		newLand = g.Board.Lands[toLandIndex(t, 1)]
	} else {
		if oldLand.Space < 5 {
			newLand = g.Board.Lands[oldLand.Index+1]
		}
	}

	return oldLand, newLand
}

// advancePeacePipe advances a PeacePipe from an old to a new land, drawing a new
// chiefdom counter if the next land out isn't wilderness.
func (g *Game) advancePeacePipe(oldLand, newLand Land) {
	if (oldLand == Land{}) {
		g.logEvent("Placed new Peace Pipe on %s.", newLand)
	} else {
		g.logEvent("Advanced Peace Pipe from %s to %s.", oldLand, newLand)
		g.Board.PeacePipes[oldLand.Index] = false
	}
	g.Board.PeacePipes[newLand.Index] = true

	// discovery only happens during the Hopewell era
	if g.Board.CurrentEra != Hopewell {
		return
	}
	if newLand.Space >= 5 {
		return
	}
	newLand = g.Board.Lands[newLand.Index+1]
	if !newLand.IsWilderness {
		g.drawChiefdomCounter(newLand)
		g.logEvent("Placed new chiefdom (%s) in %s.", g.Board.Chiefdoms[newLand.Index], newLand)
	}
}

func (PeacePipeAction) handle(g *Game) state {
	if g.Board.CurrentEra != Hopewell {
		g.Error = fmt.Errorf("This action is only allowed during the Hopewell era.")
		return stateGetNextAction{}
	}

	t := g.Action.Target.(Tribe)
	oldLand, newLand := g.findPeacePipeLands(t)

	switch {
	case newLand == Land{}:
		g.Error = fmt.Errorf("Peace Pipe on %s cannot be advanced.", oldLand)
	case newLand.IsWilderness:
		g.advancePeacePipe(oldLand, newLand)
		g.executedAction()
	case g.Board.Chiefdoms[newLand.Index].IsMounded:
		g.advancePeacePipe(oldLand, newLand)
		g.executedAction()
	default:
		g.Error = fmt.Errorf("Cannot advance Peace Pipe; chiefdom in %s must be incorporated first.", newLand)
	}
	return stateGetNextAction{}
}

func (IncorporateAction) handle(g *Game) state {
	if g.Board.CurrentEra != Hopewell {
		g.Error = fmt.Errorf("This action is only allowed during the Hopewell era.")
		return stateGetNextAction{}
	}

	t := g.Action.Target.(Tribe)
	oldLand, newLand := g.findPeacePipeLands(t)

	switch {
	case newLand == Land{}:
		g.Error = fmt.Errorf("Cannot advance Peace Pipe beyond %s.", oldLand)
	case newLand.IsWilderness:
		g.Error = fmt.Errorf("%s cannot contain a chiefdom.", newLand)
	case g.Board.Chiefdoms[newLand.Index] == nil:
		g.Error = fmt.Errorf("%s does not contain a chiefdom.", newLand)
	default:
		var r int
		if (oldLand != Land{}) {
			r1, r2 := die(), die()
			r = r2
			if r1 > r2 {
				r = r1
			}
			g.logEvent("Busk roll on %s warpath: %d and %d, choosing %d.", t, r1, r2, r)
		} else {
			r = die()
			g.logEvent("Diplomacy roll on %s warpath : %d.", t, r)
		}
		oldChiefdom := g.Board.Chiefdoms[newLand.Index]
		v := oldChiefdom.getValue()
		if g.Board.WarpathStatus.Warpath == t {
			g.logEvent("%s status modifies chiefdom's value of %d.", g.Board.WarpathStatus, v)
			v += g.Board.WarpathStatus.Modifier
		}
		if r > v {
			g.logEvent("%d exceeded value of %d; incorporation succeeded.", r, v)
			oldChiefdom.IsControlled = true
			g.Board.PeacePipes[newLand.Index] = true
			if (oldLand != Land{}) {
				g.Board.PeacePipes[oldLand.Index] = false
			}
			g.advancePeacePipe(oldLand, newLand)
		} else {
			g.logEvent("%d didn't exceed value of %d; incorporation failed.", r, v)
		}
		g.executedAction()
	}

	return stateGetNextAction{}
}

func (BuildAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (FortifyAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (AttackAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (RepairAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (PowwowAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (QuitAction) handle(g *Game) state {
	g.respond("Do you really want to quit (Y/N)?", nil)
	return stateVerifyQuitGame(0)
}

type stateVerifyQuitGame int

func (stateVerifyQuitGame) handle(g *Game) state {
	if strings.ToLower(string(g.Request.Input)) == "y" {
		return stateEndOfGame{}
	}
	return stateGetNextAction{}
}
