package vault

import (
	"fmt"
	"log"
	"os"

	"github.com/Owbird/SVault-Engine/internal/crypto"
	"github.com/Owbird/SVault-Engine/internal/database"
	"github.com/Owbird/SVault-Engine/pkg/models"
)

type Vault struct {
	models.Vault
	db *database.Database
}

func NewVault() *Vault {
	return &Vault{
		db: database.NewDatabase(),
	}
}

func (v *Vault) Create(name, password string) error {
	err := v.db.SaveVault(models.Vault{
		Name:     name,
		Password: password,
	})
	if err != nil {
		return err
	}

	return nil
}

func (v *Vault) List() ([]models.Vault, error) {
	return v.db.ListVaults()
}

func (v *Vault) Auth(name, pwd string) (bool, error) {
	vault, err := v.db.GetVault(name)
	if err != nil {
		return false, err
	}

	if vault.Name == "" {
		return false, fmt.Errorf("'%v' vault does not exist", name)
	}

	return vault.Password == pwd, nil
}

func (v *Vault) Add(file, vault, password string) error {
	pwdMatch, err := v.Auth(vault, password)
	if err != nil {
		log.Fatalf("Failed to auth vault: %v", err)
	}

	if !pwdMatch {
		log.Fatal("Passwords do not match")
	}

	buffer, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	crypto := crypto.NewCrypto()

	encBuffer, err := crypto.Encrypt(buffer, password)
	if err != nil {
		return err
	}

	stat, err := os.Stat(file)
	if err != nil {
		return err
	}

	newFile := models.File{
		Name:    file,
		Data:    encBuffer,
		Size:    stat.Size(),
		Mode:    uint32(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = v.db.AddToVault(newFile)
	if err != nil {
		return err
	}

	return nil
}
