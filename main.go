package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"pokedexcli/internal/pokecache"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
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
		"explore": {
			name:        "explore",
			description: "Display pokemon at explored location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempt to catch a pokemon",
			callback:    commandCatch,
		},
    }
	config := &config{
		cache: pokecache.NewCache(5 * time.Minute),
		pagination: &locationPagination{
            Next:     "https://pokeapi.co/api/v2/location-area",
            Previous: "",
        },
		pokedex: make(map[string]Pokemon),
	}
    for {
        fmt.Print("Pokedex > ")
        scanner.Scan()
        input := scanner.Text()

        cleaned := cleanInput(input)
        if len(cleaned) == 0 {
            continue
        }

        config.input = cleaned
        commandName := cleaned[0]

        command, exists := commands[commandName]
        if exists {
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
	callback    func(config *config) error
}

type locationPagination struct {
	Next        string
	Previous    string
}

type config struct {
	cache      *pokecache.Cache
	pokedex    map[string]Pokemon
	pagination *locationPagination
	input      []string
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

type encounterResults struct {
    EncounterMethodRates []struct {
        EncounterMethod struct {
            Name string `json:"name"`
            URL  string `json:"url"`
        } `json:"encounter_method"`
        VersionDetails []struct {
            Rate    int `json:"rate"`
            Version struct {
                Name string `json:"name"`
                URL  string `json:"url"`
            } `json:"version"`
        } `json:"version_details"`
    } `json:"encounter_method_rates"`
    GameIndex int `json:"game_index"`
    ID        int `json:"id"`
    Location  struct {
        Name string `json:"name"`
        URL  string `json:"url"`
    } `json:"location"`
    Name  string `json:"name"`
    Names []struct {
        Language struct {
            Name string `json:"name"`
            URL  string `json:"url"`
        } `json:"language"`
        Name string `json:"name"`
    } `json:"names"`
    PokemonEncounters []struct {
        Pokemon struct {
            Name string `json:"name"`
            URL  string `json:"url"`
        } `json:"pokemon"`
        VersionDetails []struct {
            EncounterDetails []struct {
                Chance          int           `json:"chance"`
                ConditionValues []interface{} `json:"condition_values"`
                MaxLevel        int           `json:"max_level"`
                Method         struct {
                    Name string `json:"name"`
                    URL  string `json:"url"`
                } `json:"method"`
                MinLevel int `json:"min_level"`
            } `json:"encounter_details"`
            MaxChance int `json:"max_chance"`
            Version   struct {
                Name string `json:"name"`
                URL  string `json:"url"`
            } `json:"version"`
        } `json:"version_details"`
    } `json:"pokemon_encounters"`
}

type Pokemon struct {
    Abilities []struct {
        Ability struct {
            Name string `json:"name"`
            URL  string `json:"url"`
        } `json:"ability"`
        IsHidden bool `json:"is_hidden"`
        Slot     int  `json:"slot"`
    } `json:"abilities"`
    BaseExperience int `json:"base_experience"`
    Cries          struct {
        Latest string `json:"latest"`
        Legacy string `json:"legacy"`
    } `json:"cries"`
    Forms []struct {
        Name string `json:"name"`
        URL  string `json:"url"`
    } `json:"forms"`
    GameIndices []struct {
        GameIndex int `json:"game_index"`
        Version   struct {
            Name string `json:"name"`
            URL  string `json:"url"`
        } `json:"version"`
    } `json:"game_indices"`
    Height    int `json:"height"`
    HeldItems []struct {
        Item struct {
            Name string `json:"name"`
            URL  string `json:"url"`
        } `json:"item"`
        VersionDetails []struct {
            Rarity  int `json:"rarity"`
            Version struct {
                Name string `json:"name"`
                URL  string `json:"url"`
            } `json:"version"`
        } `json:"version_details"`
    } `json:"held_items"`
    ID                     int    `json:"id"`
    IsDefault             bool   `json:"is_default"`
    LocationAreaEncounters string `json:"location_area_encounters"`
    Moves                 []struct {
        Move struct {
            Name string `json:"name"`
            URL  string `json:"url"`
        } `json:"move"`
        VersionGroupDetails []struct {
            LevelLearnedAt  int `json:"level_learned_at"`
            MoveLearnMethod struct {
                Name string `json:"name"`
                URL  string `json:"url"`
            } `json:"move_learn_method"`
            VersionGroup struct {
                Name string `json:"name"`
                URL  string `json:"url"`
            } `json:"version_group"`
        } `json:"version_group_details"`
    } `json:"moves"`
}

func commandHelp(config *config) error {
    fmt.Println("Welcome to the Pokedex!")
    fmt.Println("Usage:")
    fmt.Println()
    fmt.Println("help: Displays a help message")
    fmt.Println("exit: Exit the Pokedex")
	fmt.Println("map: Display the next 20 locations")
	fmt.Println("mapb: Display the previous 20 locations")
	fmt.Println("explore <location>: Display pokemon at explored location")
    return nil
}

func commandExit(config *config) error {
    fmt.Println("Closing the Pokedex... Goodbye!")
    os.Exit(0)
    return nil
}

func commandMap(config *config) error {
	err := fetchData(config, true)
	return err
}

func commandMapBack(config *config) error {
	err := fetchData(config, false)
	return err
}

func commandExplore(config *config) error {
    url := "https://pokeapi.co/api/v2/location-area/" + config.input[1]
    
    if data, ok := config.cache.Get(url); ok {
        var results encounterResults
        err := json.Unmarshal(data, &results)
        if err != nil {
            return err
        }
        for _, pokemon := range results.PokemonEncounters {
            fmt.Println(pokemon.Pokemon.Name)
        }
        return nil
    }

    res, err := http.Get(url)
    if err != nil {
        return err
    }
    defer res.Body.Close()

    body, err := io.ReadAll(res.Body)
    if err != nil {
        return err
    }

    if res.StatusCode > 299 {
        return fmt.Errorf("response failed with status code: %d and \nbody: %s", res.StatusCode, body)
    }

    config.cache.Add(url, body)

    var results encounterResults
    err = json.Unmarshal(body, &results)
    if err != nil {
        return err
    }

    for _, pokemon := range results.PokemonEncounters {
        fmt.Println(pokemon.Pokemon.Name)
    }

    return nil
}

func commandCatch(config *config) error {
	target := config.input[1]
	url := "https://pokeapi.co/api/v2/pokemon/" + target
	var pokemon Pokemon
	if data, ok := config.cache.Get(url); ok {
        err := json.Unmarshal(data, &pokemon)
        if err != nil {
            return err
        }

        return nil
    } else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with status code: %d and \nbody: %s", res.StatusCode, body)
		}

		err = json.Unmarshal(body, &pokemon)
		if err != nil {
			return err
		}
	}
	fmt.Println()
	fmt.Println("Throwing a Pokeball at " + target + "...")

	catchRate := rand.Intn(201)
	if pokemon.BaseExperience > catchRate {
		fmt.Printf("%s escaped!\n", target)
	} else {
		fmt.Printf("%s was caught!\n", target)
		config.pokedex[target] = pokemon
	}

	return nil
}

func fetchData(config *config, next bool) error {
	var url string
	if next {
		url = config.pagination.Next
	} else {
		url = config.pagination.Previous
	}

    if cachedData, ok := config.cache.Get(url); ok {
        var locations locations
        err := json.Unmarshal(cachedData, &locations)
        if err != nil {
            return err
        }
        
        config.pagination.Next = locations.Next
        config.pagination.Previous = locations.Previous

        for _, location := range locations.Results {
            fmt.Println(location.Name)
        }
        return nil
    }

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode > 299 {
    	return fmt.Errorf("response failed with status code: %d and \nbody: %s", res.StatusCode, body)
	}

	var locations locations
	err = json.Unmarshal(body, &locations)
	if err != nil {
		return err
	}

	config.cache.Add(url, body)
	config.pagination.Next = locations.Next
	config.pagination.Previous = locations.Previous

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
