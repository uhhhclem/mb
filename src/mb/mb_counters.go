package mb

import (
	"encoding/csv"
	"strconv"
	"strings"
)

var counterData = `Good,PlainValue,PlainGreenBird,MoundedValue,MoundedGreenBird
Hides,2,,4,TRUE
Hides,2,TRUE,3,TRUE
Hides,3,,4,
Hides,3,,3,TRUE
Hides,4,,2,TRUE
Chert,2,,4,TRUE
Chert,3,,3,TRUE
Chert,4,,2,TRUE
Chert,4,,4,
Copper,2,TRUE,3,TRUE
Copper,3,,3,TRUE
Copper,4,,2,TRUE
Mica,2,,4,TRUE
Mica,3,,3,TRUE
Mica,4,,3,
Feathers,2,TRUE,2,TRUE
Feathers,2,,4,TRUE
Feathers,3,,3,TRUE
Feathers,4,,2,TRUE
Pipestone,3,,3,TRUE
Pipestone,4,,2,TRUE
Chalcedony,2,,4,TRUE
Chalcedony,3,TRUE,2,TRUE
Seashells,3,,4,
Obsidian,4,,2,TRUE`

const (
	colGood = iota
	colPlainValue
	colPlainGreenBird
	colMoundedValue
	colMoundedGreenBird
)

// parse parses CSV data from a string into a slice of slices.
func parse(data string) [][]string {
	r := strings.NewReader(data)
	c := csv.NewReader(r)
	c.TrailingComma = true
	result, err := c.ReadAll()
	if err != nil {
		panic(err)
	}
	return result
}

func makeCup() Cup {
	cup := makeChiefdomCounters()
	shuffleCup(cup)
	return cup
}

func shuffleCup(cup Cup) {
	for i, _ := range cup {
		j := i + rng.Intn(len(cup)-i)
		cup[i], cup[j] = cup[j], cup[i]
	}
}

func drawFromCup(cup Cup) (*ChiefdomCounter, Cup) {
	if len(cup) == 0 {
		return nil, cup
	}
	c := cup[0]
	cup = cup[1:]
	return c, cup
}

func makeChiefdomCounters() Cup {
	var err error
	data := parse(counterData)
	counters := make(Cup, len(data)-1)
	for i, r := range data {
		if i == 0 {
			// skip the header row
			continue
		}
		c := &ChiefdomCounter{
			Good: tradeGoodNameLookup[r[colGood]],
			Plain: ChiefdomCounterFace{
				IsGreenBird: r[colPlainGreenBird] == "TRUE",
			},
			Mounded: ChiefdomCounterFace{
				IsGreenBird: r[colMoundedGreenBird] == "TRUE",
			},
		}
		if c.Plain.Value, err = strconv.Atoi(r[colPlainValue]); err != nil {
			panic(err)
		}
		if c.Mounded.Value, err = strconv.Atoi(r[colMoundedValue]); err != nil {
			panic(err)
		}
		counters[i-1] = c
	}
	return counters
}
