package database

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Owbird/SVault-Engine/internal/crypto"
	"github.com/Owbird/SVault-Engine/internal/utils"
	"github.com/Owbird/SVault-Engine/pkg/models"
	c "github.com/ostafen/clover"
)

type Database struct {
	*c.DB
	mu sync.Mutex
}

var (
	crypFunc = crypto.NewCrypto()
	instance *Database
)

func NewDatabase() *Database {
	if instance != nil {
		return instance
	}

	userDir, err := utils.GetSVaultDir()
	if err != nil {
		log.Fatalln(err)
	}

	store, err := c.Open(userDir)
	if err != nil {
		log.Fatalln(err)
	}

	collections := []string{"vaults", "vault_keys", "files"}

	for _, collection := range collections {
		err := store.CreateCollection(collection)
		if err != nil {
			if !errors.Is(c.ErrCollectionExist, err) {
				log.Fatalln(err)
			}
		}
	}

	instance = &Database{
		mu: sync.Mutex{},
		DB: store,
	}

	return instance
}

func (db *Database) SaveVault(vault models.Vault) error {
	existingVault, err := db.GetVault(vault.Name)
	if err != nil && err.Error() != "vault not found" {
		return err
	}

	if existingVault.Name != "" {
		return fmt.Errorf("vault already exists")
	}

	doc := c.NewDocument()
	doc.Set("Name", strings.ToLower(vault.Name))
	doc.Set("Password", crypFunc.Hash(vault.Password))
	doc.Set("CreatedAt", time.Now())

	_, err = db.InsertOne("vaults", doc)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ListVaults() ([]models.Vault, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	docs, err := db.Query("vaults").FindAll()
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
	db.mu.Lock()
	defer db.mu.Unlock()

	query := c.Field("Name").Eq(strings.ToLower(vault))

	doc, err := db.Query("vaults").Where(query).FindFirst()
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
	db.mu.Lock()
	defer db.mu.Unlock()

	doc := c.NewDocument()
	doc.Set("Vault", file.Vault)
	doc.Set("Name", filepath.Base(file.Name))
	doc.Set("Data", file.Data)
	doc.Set("Size", file.Size)
	doc.Set("Mode", file.Mode)
	doc.Set("ModTime", file.ModTime)

	_, err := db.InsertOne("files", doc)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ListVaultFiles(vault string) ([]models.File, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := c.Field("Vault").Eq(vault)

	docs, err := db.Query("files").Where(query).FindAll()
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

func (db *Database) GetVaultFile(vault, file string) (models.File, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := c.Field("Vault").Eq(vault).And(c.Field("Name").Eq(file))

	doc, err := db.Query("files").Where(query).FindFirst()
	if err != nil {
		return models.File{}, err
	}

	if doc == nil {
		return models.File{}, fmt.Errorf("file not found")
	}

	var foundFile models.File

	doc.Unmarshal(&foundFile)

	return foundFile, nil
}

func (db *Database) SaveVaultKey(key []byte, password, vault string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	doc := c.NewDocument()
	doc.Set("Password", crypFunc.Hash(password))
	doc.Set("VaultKey", key)
	doc.Set("Vault", vault)

	_, err := db.InsertOne("vault_keys", doc)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) GetVaultKey(vault, password string) ([]byte, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := c.Field("Vault").Eq(vault)

	doc, err := db.Query("vault_keys").Where(query).FindFirst()
	if err != nil {
		return []byte{}, err
	}

	if doc == nil {
		return []byte{}, nil
	}

	var v struct {
		VaultKey []byte
	}

	doc.Unmarshal(&v)

	return v.VaultKey, nil
}

func (db *Database) DeleteFromVault(name, vault string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := c.Field("Vault").Eq(vault).And(c.Field("Name").Eq(name))

	err := db.Query("files").Where(query).Delete()
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) DeleteVault(vault string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	query := c.Field("Name").Eq(vault)

	err := db.Query("vaults").Where(query).Delete()
	if err != nil {
		return err
	}

	return nil
}
