package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/mixolydian251/pokedex-server/routes"
	"github.com/mixolydian251/pokedex-server/utils"
	"runtime"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	r := gin.Default()
	r.Use(utils.CORSMiddleware())

	r.GET("/pokemon", routes.GetPokemonRange) //parameters (start, end)
	r.GET("/pokemon/:name", routes.GetPokemonDetails)
	r.GET("/search/:name", routes.LookupPokemon)
	r.Run(":8080")
}
