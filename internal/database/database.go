package database

import (
	"log"

	"github.com/Owbird/SVault-Engine/pkg/models"
	c "github.com/ostafen/clover"
)

type Database struct {
	Store *c.DB
}

func NewDatabase() *Database {
	db, err := c.Open(".svault")
	if err != nil {
		log.Fatalln(err)
	}

	err = db.CreateCollection("vaults")
	if err != nil {
		if err.Error() != "collection already exist" {
			log.Fatalln(err)
		}
	}

	return &Database{
		Store: db,
	}
}

func (db *Database) SaveVault(vault models.Vault) {
	doc := c.NewDocument()
	doc.Set(vault.Name, vault)

	db.Store.InsertOne("vaults", doc)
}
