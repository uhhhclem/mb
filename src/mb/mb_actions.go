package mb

import (
	"errors"
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

// Actions implement mb.state. Actions that can be taken on a warpath
// implement mb.warpathAction.
type PeacePipeAction int
type IncorporateAction int
type BuildAction int
type FortifyAction int
type AttackAction int
type RepairAction int
type PowwowAction int
type QuitAction int

type warpathAction interface {
	isEnabledOnWarpath(g *Game, t Tribe) bool
}

type ActionCost int

const (
	ZeroCost ActionCost = iota
	OneCost
	TwoCost
	ChiefdomValueCost
	PalisadeValueCost
)

var actions = []ActionSpec{
	ActionSpec{"ppa", "Peace Pipe", "Unopposed Peace Pipe advance", PeacePipeAction(0), WarpathTarget, OneCost},
	ActionSpec{"inc", "Incorporate", "Incorporate a Chiefdom", IncorporateAction(0), WarpathTarget, OneCost},
	ActionSpec{"mnd", "Build", "Build a Mound", BuildAction(0), LandTarget, ChiefdomValueCost},
	ActionSpec{"frt", "Fortify", "Fortify Cahokia", FortifyAction(0), NoTarget, TwoCost},
	ActionSpec{"att", "Attack", "Attack Hostile Army", AttackAction(0), EnemyTarget, OneCost},
	ActionSpec{"rep", "Repair", "Repair Breach", RepairAction(0), NoTarget, PalisadeValueCost},
	ActionSpec{"pow", "Powwow", "Powwow", PowwowAction(0), WarpathTarget, TwoCost},
	ActionSpec{"qui", "Quit", "Quit the Game", QuitAction(0), NoTarget, ZeroCost},
}

type ActionSpec struct {
	Name        string
	Abbr 		string
	Description string
	Type        state
	Target      TargetType
	Cost        ActionCost
}

type FrontEndAction struct {
	ActionSpec
	IsAvailable bool
	ActualCost int
}

func (g *Game) availableWarpathActions() map[string][]FrontEndAction {
	result := make(map[string][]FrontEndAction)
	for _, t := range tribes {
		for _, s := range actions {
			if at, ok := s.Type.(warpathAction); ok {
					f := FrontEndAction{s, at.isEnabledOnWarpath(g, t), 0}
					switch s.Cost {
					case ChiefdomValueCost:
						// TODO
						f.ActualCost = 99
					case PalisadeValueCost:
						// TODO
						f.ActualCost = 99
					default:
						f.ActualCost = int(s.Cost)						
					}
					result[tribeNames[t]] = append(result[tribeNames[t]], f)
				}
			}
		}
	return result
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

func (a PeacePipeAction) handle(g *Game) state {
	t := g.Action.Target.(Tribe)
	s, err := a.perform(g, t, true)
	if err != nil {
		g.Error = err
	}
	return s
}

func (a PeacePipeAction) perform(g *Game, t Tribe, mutate bool) (state, error) {
	var err error

	if g.Board.CurrentEra != Hopewell {
		err = fmt.Errorf("This action is only allowed during the Hopewell era.")
		return stateGetNextAction{}, err
	}

	oldLand, newLand := g.findPeacePipeLands(t)

	switch {
	case g.Board.ActionPoints < 1:
		err = errors.New("Not enough APs remaining; 1 required.")
	case newLand == Land{}:
		err = fmt.Errorf("Peace Pipe on %s cannot be advanced.", oldLand)
	case newLand.IsWilderness:
		if mutate {
				g.advancePeacePipe(oldLand, newLand)
				g.executedAction()
		}
	case g.Board.Chiefdoms[newLand.Index].IsMounded:
		if mutate {
				g.advancePeacePipe(oldLand, newLand)
				g.executedAction()
		}
	default:
		err = fmt.Errorf("Cannot advance Peace Pipe; chiefdom in %s must be incorporated first.", newLand)
	}
	return stateGetNextAction{}, err
}

func (a PeacePipeAction) isEnabledOnWarpath(g *Game, t Tribe) bool {
	_, err := a.perform(g, t, false)
	return err == nil
}

func (a IncorporateAction) handle(g *Game) state {
	t := g.Action.Target.(Tribe)
	s, err := a.perform(g, t, true)
	if err != nil {
		g.Error = err
	}
	return s
}

func (IncorporateAction) perform(g *Game, t Tribe, mutate bool) (state, error) {
	var err error

	if g.Board.CurrentEra != Hopewell {
		err = fmt.Errorf("This action is only allowed during the Hopewell era.")
		return stateGetNextAction{}, err
	}

	oldLand, newLand := g.findPeacePipeLands(t)

	switch {
	case g.Board.ActionPoints < 1:
		err = errors.New("Not enough APs remaining; 1 required.")
	case newLand == Land{}:
		err = fmt.Errorf("Cannot advance Peace Pipe beyond %s.", oldLand)
	case newLand.IsWilderness:
		err = fmt.Errorf("%s cannot contain a chiefdom.", newLand)
	case g.Board.Chiefdoms[newLand.Index] == nil:
		err = fmt.Errorf("%s does not contain a chiefdom.", newLand)
	case !mutate:
		break
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

	return stateGetNextAction{}, err
}

func (a IncorporateAction) isEnabledOnWarpath(g *Game, t Tribe) bool {
	_, err := a.perform(g, t, false)
	if err != nil {
		fmt.Println(err)
	}
	return err == nil
}

func (BuildAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (BuildAction) isEnabledOnWarpath(g *Game, t Tribe) bool {
	return false
}

func (FortifyAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (AttackAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (AttackAction) isEnabledOnWarpath(g *Game, t Tribe) bool {
	return false
}

func (RepairAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (PowwowAction) handle(g *Game) state {
	return stateGetNextAction{}
}

func (PowwowAction) isEnabledOnWarpath(g *Game, t Tribe) bool {
	return false
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
