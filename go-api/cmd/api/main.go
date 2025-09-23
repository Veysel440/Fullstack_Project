package main

import (
	"log"
	"net/http"

	"fullstack-oracle/go-api/internal/config"
	"fullstack-oracle/go-api/internal/db"
	hh "fullstack-oracle/go-api/internal/http"
	"fullstack-oracle/go-api/internal/repo"
	"fullstack-oracle/go-api/internal/service"
)

func main() {
	cfg := config.FromEnv()

	d, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	rp := repo.NewItemRepo(d)
	sv := service.NewItemService(rp)
	h := &hh.Handlers{S: sv}
	app := hh.Router(h, hh.CORS(cfg.CORSOrigins))

	addr := ":" + cfg.Port
	log.Printf("api %s", addr)
	log.Fatal(http.ListenAndServe(addr, app))
}
