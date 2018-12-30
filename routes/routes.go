package routes

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/mixolydian251/pokedex-server/utils"
	"net/http"
	"sort"
	"strconv"
	"sync"
)

var wg sync.WaitGroup

type Pokemon struct {
	Url    string   `json:"url"`
	Id     int      `json:"id"`
	Name   string   `json:"name"`
	Sprite string   `json:"sprite"`
	Types  []string `json:"types"`
}

type FlavorText struct {
	Version int
	Text string
}

type DetailResponse struct {
	Text []FlavorText
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
	Types []struct{
		PokemonType struct{
			Name string `json:"name"`
		}`json:"type"`
	}
}

type SpriteResponse struct {
	Url   string `json:"url"`
	Id    int    `json:"id"`
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
		data.Id,
		data.Name,
		data.Sprites.Front,
		types,
	}

	*arr = append(*arr, p)
	wg.Done()
}

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
		return pokemonArr[i].Id < pokemonArr[j].Id
	})
	c.JSON(200, pokemonArr)
}

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

func LookupPokemon(c *gin.Context) {
	type pokemon struct {
		Name string
		Height int
		Weight int
	}

	chars := c.Param("name") + "%"

	connStr := "host=localhost user=Jordy dbname=testing sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	defer db.Close()
	utils.CheckError(err, "connecting to SQL")

	var pkmArr []pokemon

	rows, err := db.Query("SELECT \"Name\", weight, height FROM pokemon WHERE \"Name\" LIKE $1", chars)
	utils.CheckError(err, "in the SQL Query")

	for rows.Next() {
		p := pokemon{}
		err := rows.Scan(&p.Name, &p.Weight, &p.Height)
		utils.CheckError(err, "Scanning SQL response rows")
		pkmArr = append(pkmArr, p)
	}

	c.JSON(200, pkmArr)
}

func FlavorTextLookup(n string, data *DetailResponse){
	connStr := "host=localhost user=Jordy dbname=testing sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	defer db.Close()
	utils.CheckError(err, "connecting to SQL")

	var xt []FlavorText

	rows, err := db.Query("SELECT flavor_text, version_id From pokemon INNER JOIN flavor_text ON pokemon.species_id = flavor_text.species_id WHERE \"Name\"=$1 AND language_id=9 AND version_id > 1", n)
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
