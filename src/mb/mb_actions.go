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
	TribeTarget
)

var actions = []ActionSpec{
	ActionSpec{"ppa", "Unopposed Peace Pipe advance", WarpathTarget},
	ActionSpec{"inc", "Incorporate a Chiefdom", WarpathTarget},
	ActionSpec{"mnd", "Build a Mound", LandTarget},
	ActionSpec{"frt", "Fortify Cahokia", NoTarget},
	ActionSpec{"att", "Attack Hostile Army", TribeTarget},
	ActionSpec{"rep", "Repair Breach", NoTarget},
	ActionSpec{"pow", "Powwow", WarpathTarget},
}

type ActionSpec struct {
	Name        string
	Description string
	Target      TargetType
}

// finder is a function that finds a unique game object given its prefix.
type finder func(string, *Game) (interface{}, error)

var findFunction = map[TargetType]finder{
	NoTarget:      findNothing,
	WarpathTarget: findWarpath,
	LandTarget:    findLand,
	TribeTarget:   findTribe,
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
	tribe, err := findTribe(t, g)
	if err != nil {
		return nil, err
	}
	if tribe.(Tribe) > Caddo {
		return nil, fmt.Errorf("%q doesn't match a warpath.", t)
	}
	return tribe, nil
}

// findLand finds
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

// findTribe finds an enemy tribe - this can be the Spanish.
func findTribe(t string, _ *Game) (interface{}, error) {
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
