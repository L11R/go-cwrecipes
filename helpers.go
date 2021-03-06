package main

import (
	"errors"
)

func copyMap(m map[string]int) map[string]int {
	newMap := make(map[string]int)

	for key, value := range m {
		newMap[key] = value
	}

	return newMap
}

func findResource(name string) (resource, error) {
	for _, r := range resources {
		if name == r.Name {
			return r, nil
		}
	}

	return resource{}, errors.New("not found")
}

func newBasic(name string, amount int) basic {
	return basic{
		name,
		amount,
		0,
		0,
	}
}

func RecurBasics(rec map[string]int) (basics []basic) {
	for name, amount := range rec {
		res, err := findResource(name)
		if err != nil {
			// if can't find (for example unknown element, recipe or frag)
			basics = append(basics, newBasic(name, amount))
			continue
		}

		if !res.Composite {
			// if it already basic
			basics = append(basics, newBasic(name, amount))
			continue
		}

		// copy (else we will change reference)
		newRec := copyMap(res.Recipe)

		// multiple amount in recipe
		for recipeName, recipeAmount := range newRec {
			newRec[recipeName] = recipeAmount * amount
		}

		basics = append(basics, RecurBasics(newRec)...)
	}

	return
}

func FoldBasics(basics []basic) []basic {
	newBasics := make([]basic, 0)

	for _, b := range basics {
		found := false

		for j, nb := range newBasics {
			if b.Name == nb.Name {
				found = true
				newBasics[j].Amount += b.Amount
			}
		}

		if !found {
			newBasics = append(newBasics, b)
		}
	}

	return newBasics
}

func newCommand(r resource, amount int) command {
	return command{
		r.ID,

		r.Name,
		amount,
		r.ManaCost * amount,
	}
}

func RecurCommands(rec map[string]int) (commands []command) {
	for name, amount := range rec {
		res, err := findResource(name)
		if err != nil {
			continue
		}

		if res.Composite {
			commands = append(commands, newCommand(res, amount))

			// copy (else we will change reference)
			recipe := copyMap(res.Recipe)

			// multiple amount in recipe
			for recipeName, recipeAmount := range recipe {
				recipe[recipeName] = recipeAmount * amount
			}

			commands = append(commands, RecurCommands(recipe)...)
		}
	}

	return
}

func FoldCommands(commands []command) []command {
	// don't forget to reverse array
	for i, j := 0, len(commands)-1; i < j; i, j = i+1, j-1 {
		commands[i], commands[j] = commands[j], commands[i]
	}

	newCommands := make([]command, 0)

	for _, c := range commands {
		// Skip zero amount for user commands
		if c.Amount == 0 {
			continue
		}

		found := false

		for j, nc := range newCommands {
			if c.Name == nc.Name {
				found = true

				newCommands[j].Amount += c.Amount
				newCommands[j].CommandManaCost += c.CommandManaCost
			}
		}

		if !found {
			newCommands = append(newCommands, c)
		}
	}

	return newCommands
}

func RecurPurchases(rec map[string]int, userStock map[string]int) (purchases []basic) {
	for name, amount := range rec {
		userAmount := 0
		requiredAmount := amount

		if ua, ok := userStock[name]; ok {
			userAmount = ua
			requiredAmount = amount - ua

			if requiredAmount < 0 {
				requiredAmount = 0
			}
		}

		res, err := findResource(name)
		if err != nil {
			// if can't find (for example unknown element, recipe or frag)
			purchases = append(purchases, basic{
				name,
				amount,
				userAmount,
				requiredAmount,
			})
			continue
		}

		if !res.Composite {
			// if it already basic
			purchases = append(purchases, basic{
				name,
				amount,
				userAmount,
				requiredAmount,
			})
			continue
		}

		// copy (else we will change reference)
		newRec := copyMap(res.Recipe)

		// multiple amount in recipe
		for recipeName, recipeAmount := range newRec {
			newRec[recipeName] = recipeAmount * requiredAmount
		}

		purchases = append(purchases, RecurPurchases(newRec, userStock)...)
	}

	return
}

func FoldPurchases(purchases []basic) []basic {
	newPurchases := make([]basic, 0)

	for _, p := range purchases {
		if p.RequiredAmount == 0 {
			continue
		}

		found := false

		for j, np := range newPurchases {
			if p.Name == np.Name {
				found = true
				newPurchases[j].RequiredAmount += p.RequiredAmount
			}
		}

		if !found {
			newPurchases = append(newPurchases, p)
		}
	}

	return newPurchases
}

func RecurUserCommands(rec map[string]int, userStock map[string]int) (commands []command) {
	for name, amount := range rec {
		requiredAmount := amount

		if ua, ok := userStock[name]; ok {
			requiredAmount = amount - ua

			if requiredAmount < 0 {
				requiredAmount = 0
			}
		}

		res, err := findResource(name)
		if err != nil {
			continue
		}

		if res.Composite {
			commands = append(commands, newCommand(res, requiredAmount))

			// copy (else we will change reference)
			recipe := copyMap(res.Recipe)

			// multiple amount in recipe
			for recipeName, recipeAmount := range recipe {
				recipe[recipeName] = recipeAmount * requiredAmount
			}

			commands = append(commands, RecurUserCommands(recipe, userStock)...)
		}
	}

	return
}
