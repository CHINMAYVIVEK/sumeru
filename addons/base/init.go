package base

import (
	"log"
	
	_ "sumeru/addons/base/models"
)

func init() {
	log.Println("Sumeru Base Addon Loaded")
}
