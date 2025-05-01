package main

import (
	"github.com/ultrazg/xyz/service"
	"github.com/ultrazg/xyz/utils"
	"log"
)

func main() {
	utils.InitCache()
	err := service.Start()
	if err != nil {
		log.Fatal(err)
	}
}