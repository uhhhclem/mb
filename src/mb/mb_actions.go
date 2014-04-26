package mb

import (
	"fmt"
	"strings"
)

type Action struct {
	Spec   *ActionSpec
	Target interface{}
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
	Type 		state
	Target      TargetType
	Cost 		ActionCost
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


func (PeacePipeAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func(IncorporateAction) handle(g *Game) state {
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
	g.respond("Do you really want to quit (Y/N)?",  nil)
	return stateVerifyQuitGame(0)
}

type stateVerifyQuitGame int

func (stateVerifyQuitGame) handle(g *Game) state {
	if strings.ToLower(string(g.Request.Input)) == "y" {
		return stateEndOfGame{}
	}
	return stateGetNextAction{}
}
