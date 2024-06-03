package database

import (
	"errors"
	"log"
	"time"

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
		if !errors.Is(c.ErrCollectionExist, err) {
			log.Fatalln(err)
		}
	}

	err = db.CreateCollection("files")
	if err != nil {
		if !errors.Is(c.ErrCollectionExist, err) {
			log.Fatalln(err)
		}
	}

	return &Database{
		Store: db,
	}
}

func (db *Database) SaveVault(vault models.Vault) error {
	doc := c.NewDocument()
	doc.Set("Name", vault.Name)
	doc.Set("Password", vault.Password)
	doc.Set("CreatedAt", time.Now())

	_, err := db.Store.InsertOne("vaults", doc)
	if err != nil {
		return err
	}

	return nil
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

func (db *Database) AddToVault(file models.File) error {
	doc := c.NewDocument()
	doc.Set("Vault", file.Vault)
	doc.Set("Name", file.Name)
	doc.Set("Data", file.Data)
	doc.Set("Size", file.Size)
	doc.Set("Mode", file.Mode)
	doc.Set("ModTime", file.ModTime)

	_, err := db.Store.InsertOne("files", doc)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ListVaultFiles(vault string) ([]models.File, error) {
	query := c.Field("Vault").Eq(vault)

	docs, err := db.Store.Query("files").Where(query).FindAll()
	if err != nil {
		return []models.File{}, err
	}

	files := []models.File{}

	for _, doc := range docs {

		var file models.File

		doc.Unmarshal(&file)

		files = append(files, file)
	}

	return files, nil
}
