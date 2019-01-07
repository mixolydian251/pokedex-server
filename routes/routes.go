package routes

import (
	"database/sql"
	"encoding/json"
	"github.com/gchaincl/dotsql"
	"github.com/gin-gonic/gin"
	"github.com/mixolydian251/pokedex-server/utils"
	"net/http"
	"sort"
	"strconv"
	"sync"
)

var wg sync.WaitGroup

// Pokemon is the abridged data structure that gets sent as JSON to client on range load.
type Pokemon struct {
	URL    string   `json:"url"`
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Sprite string   `json:"sprite"`
	Types  []string `json:"types"`
}

// FlavorText is the structure of a pokemon's description.
type FlavorText struct {
	Version int
	Text    string
}

// DetailResponse is the in depth data structure that gets sent about a specific pokemon
type DetailResponse struct {
	Text    []FlavorText
	Sprites struct {
		Front string `json:"front_default"`
		Back  string `json:"back_default"`
	} `json:"sprites"`
	Stats []struct {
		Value int `json:"base_stat"`
		Stat  struct {
			Name string `json:"name"`
		} `json:"stat"`
	}
	Types []struct {
		PokemonType struct {
			Name string `json:"name"`
		} `json:"type"`
	}
}

// SpriteResponse is the JSON form that is returned from PokeAPI
type SpriteResponse struct {
	URL   string `json:"url"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Sprites struct {
		Front string `json:"front_default"`
	}
}

// GetSprite grabs data from PokeAPI needed for range load
func GetSprite(url string, arr *[]Pokemon) {
	res, err := http.Get(url)
	utils.CheckError(err, "getting sprite")

	defer res.Body.Close()
	dc := json.NewDecoder(res.Body)
	data := SpriteResponse{}
	dc.Decode(&data)

	var types []string
	for _, v := range data.Types {
		types = append(types, v.Type.Name)
	}

	p := Pokemon{
		url,
		data.ID,
		data.Name,
		data.Sprites.Front,
		types,
	}

	*arr = append(*arr, p)
	wg.Done()
}

/*
GetPokemonRange is a route that takes in 2 query parameters, "start" and "end".
these parameters are integers that represent a range of pokemon based on their ID.
This function returns an array of type Pokemon as JSON.
*/
func GetPokemonRange(c *gin.Context) {
	start := c.Query("start")
	end := c.Query("end")
	s, _ := strconv.Atoi(start)
	e, _ := strconv.Atoi(end)

	var pokemonArr []Pokemon

	wg.Add(e + 1 - s)
	for i := s; i <= e; i++ {
		go GetSprite("https://pokeapi.co/api/v2/pokemon/"+strconv.Itoa(i), &pokemonArr)
	}
	wg.Wait()
	sort.Slice(pokemonArr, func(i, j int) bool {
		return pokemonArr[i].ID < pokemonArr[j].ID
	})
	c.JSON(200, pokemonArr)
}

/*
GetPokemonDetails is a route endpoint at "/pokemon/:name", and returns
details as JSON about a specific pokemon based on their name or ID.
*/
func GetPokemonDetails(c *gin.Context) {
	data := DetailResponse{}
	name := c.Param("name")

	wg.Add(1)
	go FlavorTextLookup(name, &data)

	res, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + name)
	utils.CheckError(err, "getting pokemon")

	defer res.Body.Close()
	dc := json.NewDecoder(res.Body)
	dc.Decode(&data)

	wg.Wait()
	c.JSON(200, data)
}

/*
FlavorTextLookup is a function used in GetPokemonDetails. It queries the Postgres
DB to return the flavor text of a pokemon and the version associated with that text.
*/
func FlavorTextLookup(n string, data *DetailResponse) {
	connStr := "host=localhost user=Jordy dbname=testing sslmode=disable"

	dot, err := dotsql.LoadFromFile("queries.sql")
	utils.CheckError(err, "Locating sql file and query")
	db, err := sql.Open("postgres", connStr)
	utils.CheckError(err, "connecting to SQL")
	defer db.Close()

	var xt []FlavorText

	rows, err := dot.Query(db, "flavor-text", n)
	utils.CheckError(err, "in the SQL Query")

	for rows.Next() {
		t := FlavorText{}
		err := rows.Scan(&t.Text, &t.Version)
		utils.CheckError(err, "Scanning SQL response rows")
		xt = append(xt, t)
	}

	data.Text = xt
	wg.Done()
}

/*
LookupPokemon is a route at endpoint "/search/:name" that runs a SQL query
based on a user string, and returns any matching results for auto-completion
returns name, weight and height.
*/
func LookupPokemon(c *gin.Context) {
	type pokemon struct {
		Name   string
		Height int
		Weight int
	}

	chars := c.Param("name") + "%"

	connStr := "host=localhost user=Jordy dbname=testing sslmode=disable"

	dot, err := dotsql.LoadFromFile("queries.sql")
	utils.CheckError(err, "Locating sql file and query")
	db, err := sql.Open("postgres", connStr)
	utils.CheckError(err, "connecting to SQL")
	defer db.Close()

	var pkmArr []pokemon

	rows, err := dot.Query(db, "search-bar", chars)
	utils.CheckError(err, "in the SQL Query")

	for rows.Next() {
		p := pokemon{}
		err := rows.Scan(&p.Name, &p.Weight, &p.Height)
		utils.CheckError(err, "Scanning SQL response rows")
		pkmArr = append(pkmArr, p)
	}

	c.JSON(200, pkmArr)
}
