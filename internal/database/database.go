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
	doc.Set("Name", vault.Name)
	doc.Set("Password", vault.Password)

	db.Store.InsertOne("vaults", doc)
}

func (db *Database) ListVaults() ([]models.Vault, error) {
	docs, err := db.Store.Query("vaults").FindAll()
	if err != nil {
		return []models.Vault{}, err
	}

	vaults := []models.Vault{}

	for _, doc := range docs {
		var v models.Vault

		doc.Unmarshal(&v)

		vaults = append(vaults, v)
	}

	return vaults, nil
}

func (db *Database) GetVault(vault string) (models.Vault, error) {
	query := c.Field("Name").Eq(vault)

	doc, err := db.Store.Query("vaults").Where(query).FindFirst()
	if err != nil {
		return models.Vault{}, err
	}

	if doc == nil {
		return models.Vault{}, nil
	}

	var v models.Vault

	doc.Unmarshal(&v)

	return v, nil
}
