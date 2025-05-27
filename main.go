package main

import (
	"bufio"
	"fmt"
	"io"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	config := &commandConfig {
		Next: "https://pokeapi.co/api/v2/location-area",
		Previous: "",
	}
    commands := map[string]cliCommand{
        "help": {
            name:        "help",
            description: "Displays a help message",
            callback:    commandHelp,
        },
        "exit": {
            name:        "exit",
            description: "Exit the Pokedex",
            callback:    commandExit,
        },
		"map": {
			name:        "map",
			description: "Display the next 20 locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "map back",
			description: "Display the previous 20 locations",
			callback:    commandMapBack,
		},
    }

    for {
        fmt.Print("Pokedex > ")
        scanner.Scan()
        input := scanner.Text()

        cleaned := cleanInput(input)
        if len(cleaned) == 0 {
            continue
        }

        commandName := cleaned[0]
        
        command, exists := commands[commandName]
        if exists {
			fmt.Println("")
            err := command.callback(config)
            if err != nil {
                fmt.Printf("Error: %s\n", err)
            }
        } else {
            fmt.Println("Unknown command")
        }
    }
}

type cliCommand struct {
	name        string
	description string
	callback    func(*commandConfig) error
}

type commandConfig struct {
	Next        string
	Previous    string
}

type locations struct {
	Next        string
	Previous    string
	Results     []locationResults
}

type locationResults struct {
	Name        string
	Url         string
}

func commandHelp(config *commandConfig) error {
    fmt.Println("Welcome to the Pokedex!")
    fmt.Println("Usage:")
    fmt.Println()
    fmt.Println("help: Displays a help message")
    fmt.Println("exit: Exit the Pokedex")
    return nil
}

func commandExit(config *commandConfig) error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}

func commandMap(config *commandConfig) error {
	res, err := http.Get(config.Next)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and \nbody: %s\n", res.StatusCode, body)
	}

	var locations locations
	err = json.Unmarshal(body, &locations)
	if err != nil {
		log.Fatal(err)
	}

	config.Next = locations.Next
	config.Previous = locations.Previous

	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}
	return nil
}

func commandMapBack(config *commandConfig) error {
	if config.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	res, err := http.Get(config.Previous)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and \nbody: %s\n", res.StatusCode, body)
	}

	var locations locations
	err = json.Unmarshal(body, &locations)
	if err != nil {
		log.Fatal(err)
	}

	config.Next = locations.Next
	config.Previous = locations.Previous

	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}
	return nil
}

func cleanInput(str string) []string {
    lowered := strings.ToLower(strings.TrimSpace(str))
    words := strings.Fields(lowered)
    return words
}
