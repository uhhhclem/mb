package mb

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// important package-wide variables go here
var (
	rng *rand.Rand
)

func init() {
	rng = rand.New(rand.NewSource(time.Now().Unix()))
}

func die() int {
	return rng.Intn(6) + 1
}

type Game struct {
	Board       Board
	HistoryDeck Pile
	Cup         Cup
	State       state
	Request		Request
	Response	*Response
	ResponseReady bool
	AdvancingArmies []Tribe
	RevoltingTribe Tribe
	Error error
}

type Request struct {
	Input Input
}

type Response struct {
	Prompt Prompt
	Error error
}

// NewGame initializes a new Game.
func NewGame() *Game {
	return &Game{
		HistoryDeck: makeHistoryDeck(),
		Board:       makeBoard(),
		Cup:         makeCup(),
	}
}

func (g *Game) StartGame() {
	g.State = stateStartOfGame{}
	g.HandleRequest(Request{})
}

// HandleRequest handles the next request pending for the game.
func (g *Game) HandleRequest(q Request) {
	g.Request = q
	g.Response = nil
	g.Error = nil
	for {
		switch g.State.(type) {
		case stateEndProgram:
			return
		default:
			g.State = g.State.handle(g)
			// if there's a response set, send it and wait for the next request.
			if g.Response != nil {
				return
			}
		}
	}
}

// drawChiefdomCounter draws the next ChiefdomCounter from the cup.
func (g *Game) drawChiefdomCounter() *ChiefdomCounter {
	var c *ChiefdomCounter
	c, g.Cup = drawFromCup(g.Cup)
	return c
}

// drawHistoryCard draws the next HistoryCard from the deck.  If Board.Card is nil, game over.
func (g *Game) drawHistoryCard() {
	c, p := drawFromPile(g.HistoryDeck)
	g.HistoryDeck = p
	g.Board.Card = c
}

func (*Game) logPhase(f string, args ...interface{}) {
	fmt.Printf("\n\n"+f, args...)	
}

func (*Game) logEvent(f string, args ...interface{}) {
	fmt.Printf("\n  "+f, args...)
}

func (g *Game) respond(p string, err error) {
	g.Response = &Response{Prompt: Prompt(p), Error: err}
}

type state interface {
	handle(g *Game) state
}

type stateBlackBannerEvent struct{}

func (stateBlackBannerEvent) handle(*Game) state {
	panic("stateBlackBannerEvent")
}

type stateEconomicPhase struct{}

func (stateEconomicPhase) handle(g *Game) state {

	landsWithGood := func(t TradeGood) int {
		var result int
		for i := 0; i < LandCount; i++ {
			c := g.Board.Chiefdoms[i]
			if c != nil && c.Counter.Good == t {
				result += 1
			}
		}
		return result
	}

	g.logPhase("Economic Phase:")
	c := g.Board.Card
	if c.IsWhite {
		g.Board.ActionPoints += c.ActionPoints
		g.logEvent("White AP number, APs added: %d", c.ActionPoints)
	} else {
		ap := g.Board.TradeGoods - c.ActionPoints
		if ap <= 0 {
			ap = 1
		}
		g.logEvent("Black AP number: %d, trade goods: %d, APs added: %d")
		for _, rb := range c.ResourceBonus {
			bp := landsWithGood(rb)
			if bp > 0 {
				g.logEvent("Resource bonus: %d AP for %s", bp, rb)
				ap += bp
			}
		}
		g.Board.ActionPoints += ap
		g.logEvent("Total APs added: %d", ap)
	}
	return stateHostilesPhase{}
}

type stateEndOfGame struct{}

func (stateEndOfGame) handle(g *Game) state {
	return stateEndProgram{}
}

type stateEndProgram struct{}

func (stateEndProgram) handle(g *Game) state {
	return stateEndProgram{}
}

type stateHistoryPhase struct{}

func (stateHistoryPhase) handle(g *Game) state {
	g.logPhase("History Phase:")
	g.drawHistoryCard()
	g.logEvent("Drew %s", g.Board.Card)
	if g.Board.Card == nil {
		return stateEndOfGame{}
	}
	if g.Board.Card.IsAvaricia {
		return stateBlackBannerEvent{}
	}
	if g.Board.Card.IsSpanish {
		return stateSpanishEvent{}
	}
	return stateEconomicPhase{}
}

type stateHostilesPhase struct{}

func (stateHostilesPhase) handle(g *Game) state {
	g.logPhase("Hostiles Phase:")
	c := g.Board.Card
	drm := 0
	if c.Modifier != None {
		if c.IsAscendant {
			drm = 1
		} else {
			drm = -1
		}
	}
	g.Board.WarpathStatus = WarpathStatus{Warpath: g.Board.Card.Modifier, DRM: drm}
	g.logEvent("Warpath status is %s", g.Board.WarpathStatus)

	g.AdvancingArmies = make([]Tribe, len(c.AdvancingArmies))
	copy(g.AdvancingArmies, c.AdvancingArmies)
	g.RevoltingTribe = c.Revolt

	return stateAdvanceHostile{}
	
}

type stateAdvanceHostile struct{}

func(stateAdvanceHostile) handle(g *Game) state {
	if len(g.AdvancingArmies) == 0 {
		g.logEvent("No advancing armies.")
		return stateRevoltPhase{}
	}
	// TODO:  once we can get out of Hopewell
	//a := g.AdvancingArmies[0]
	g.AdvancingArmies = g.AdvancingArmies[1:]
	//h := g.Board.Hostiles[int(a)]
	return stateAdvanceHostile{}
}

type stateRevoltPhase struct{}

func(stateRevoltPhase) handle(g *Game) (s state) {
	s = stateActionPhase{}
	tribe := g.RevoltingTribe
	if tribe == None {
		return
	}
	g.logEvent("%s tribe is revolting.", tribe)
	roll := die()
	land := g.Board.findLand(tribe, roll)
	g.logEvent("%d rolled, land = %s", roll, land)
	if land.IsWilderness {
		g.logEvent("No revolt in wilderness.")
		return
	}
	if land.Space == 6 {
		g.logEvent("No revolt in tribal homeland.")
		return
	}
	if !land.IsControlled {
		if g.Board.CurrentEra == Hopewell {
			g.logEvent("Land is uncontrolled; revolt has no effect.")
			return
		}
		g.logEvent("Retreating tribe %s", tribe)
		// TODO:  implement retreat; must also handle Spanish.
		return
	}
	c := g.Board.findChiefdom(tribe, roll)
	if c != nil && c.IsGreenBirdman() {
		g.logEvent("Green Birdman people love you and do not revolt.")
		return
	}
	g.logEvent("Retreating peace pipe.")
	// TODO:  implement retreat/remove peace pipe.
	g.logEvent("Advancing army")
	// TODO:  implement advancing army

	return
}

type stateActionPhase struct{}

func(stateActionPhase) handle(g *Game) state {
	g.logPhase("Action Phase:")
	return stateGetNextAction{}
}

type stateGetNextAction struct{}

func(stateGetNextAction) handle(g *Game) state {
	if g.Board.ActionPoints < 1 {
		return stateTest{}
	}
	// the previous state can override this prompt by setting g.Response
	g.respond("Enter action", g.Error)
	return stateProcessAction{}
}

type stateProcessAction struct{}

func(stateProcessAction) handle(g *Game) state {
	action, err := g.parseAction(string(g.Request.Input))
	if err == nil {
		err = g.prepareAction(action)
	}
	if err != nil {
		g.Error = err
		return stateGetNextAction{}
	}
	g.Board.ActionPoints = g.Board.ActionPoints - action.ActualCost
	g.logEvent("Got action %q, %d APs remaining.", action, g.Board.ActionPoints)

	switch action.Spec.Type {
	case QuitAction:
		return stateQuitGameAction{}
	}
	return stateGetNextAction{}
}

// prepareAction updates the Action with values that the action logic will need.
// It returns an error if the action is invalid for any reason.
func (g *Game) prepareAction(a *Action) error {
	// can we afford the action?
	var cost int
	switch a.Spec.Cost {
	case ZeroCost, OneCost, TwoCost:
		cost = int(a.Spec.Cost)
	case ChiefdomValueCost:
		// TODO
	case PalisadeValueCost:
		cost = g.Board.palisade().Value
	}
	if avail := g.Board.ActionPoints; cost > avail {
		return fmt.Errorf("This action costs %dAPs, but you only have %d.", cost, avail)
	}
	a.ActualCost = cost
	return nil
}

type stateQuitGameAction struct{}

func (stateQuitGameAction) handle(g *Game) state {
	g.respond("Do you really want to quit (Y/N)?",  nil)
	return stateVerifyQuitGame{}
}

type stateVerifyQuitGame struct{}

func (stateVerifyQuitGame) handle(g *Game) state {
	if strings.ToLower(string(g.Request.Input)) == "y" {
		return stateEndOfGame{}
	}
	return stateGetNextAction{}
}

type stateTest struct{}

func (stateTest) handle(g *Game) state {
	g.logEvent("stateTest reached.")
	return stateEndOfGame{}
}

type stateSpanishEvent struct{}

func (stateSpanishEvent) handle(g *Game) state {
	panic("stateSpanishEvent")
}

type stateStartOfGame struct{}

func (stateStartOfGame) handle(g *Game) state {

	g.logPhase("Setup:")

	for t := HoChunk; t <= Caddo; t++ {
		c := g.drawChiefdomCounter()
		i := toLandIndex(t, 1)
		g.Board.Chiefdoms[i] = &Chiefdom{Counter: *c}
		g.logEvent("Land %s: %s", g.Board.Lands[i], g.Board.Chiefdoms[i])
	}

	return stateHistoryPhase{}
}
