package mb

import (
	"fmt"
	"strings"
)

type Action struct {
	Spec   *ActionSpec
	Target interface{}
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

type ActionType int

const (
	PeacePipeAction ActionType = iota
	IncorporateAction
	BuildAction
	FortifyAction
	AttackAction
	RepairAction
	PowwowAction
	QuitAction
)

type ActionCost int

const (
	ZeroCost ActionCost = iota
	OneCost
	TwoCost 
	ChiefdomValueCost
	PalisadeValueCost
)

var actions = []ActionSpec{
	ActionSpec{"ppa", "Unopposed Peace Pipe advance", PeacePipeAction, WarpathTarget, OneCost},
	ActionSpec{"inc", "Incorporate a Chiefdom", IncorporateAction, WarpathTarget, OneCost},
	ActionSpec{"mnd", "Build a Mound", BuildAction, LandTarget, ChiefdomValueCost},
	ActionSpec{"frt", "Fortify Cahokia", FortifyAction, NoTarget, TwoCost},
	ActionSpec{"att", "Attack Hostile Army", AttackAction, EnemyTarget, OneCost},
	ActionSpec{"rep", "Repair Breach", RepairAction, NoTarget, PalisadeValueCost},
	ActionSpec{"pow", "Powwow", PowwowAction, WarpathTarget, TwoCost},
	ActionSpec{"qui", "Quit", QuitAction, NoTarget, ZeroCost},
}

type ActionSpec struct {
	Name        string
	Description string
	Type 		ActionType
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
