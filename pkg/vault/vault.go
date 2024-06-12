// Package vault provides a way to manage vaults.
// It allows users to create vaults, authenticate against them, and add encrypted files to those vaults.
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

// A new vault with the database initialized
func NewVault() *Vault {
	return &Vault{
		db: database.NewDatabase(),
	}
}

// Create creates a vault with a name and password
// The vault is saved to the database
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

// List retrives all the created vaults
func (v *Vault) List() ([]models.Vault, error) {
	return v.db.ListVaults()
}

// Auth authorizes vault access based on the
// name of the vault and password
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

// Add adds a file to the vault after a successful
// authentication
func (v *Vault) Add(file, vault, password string) error {
	pwdMatch, err := v.Auth(vault, password)
	if err != nil {
		return err
	}

	if !pwdMatch {
		return fmt.Errorf("passwords do not match")
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
		Vault:   vault,
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

// ListFileVaults returns a slice of added files to the
// specified vault
func (v *Vault) ListFileVaults(vault, password string) ([]models.File, error) {
	pwdMatch, err := v.Auth(vault, password)
	if err != nil {
		log.Fatalf("Failed to auth vault: %v", err)
	}

	if !pwdMatch {
		log.Fatal("Passwords do not match")
	}

	return v.db.ListVaultFiles(vault)
}
