package mail

import (
	"log"

	_ "sumeru/addons/mail/models"
)

func init() {
	log.Println("Sumeru Mail Addon Loaded")
}
