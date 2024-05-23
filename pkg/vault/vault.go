package vault

import (
	"os"
	"path/filepath"

	"github.com/Owbird/SVault-Engine/internal/database"
	"github.com/Owbird/SVault-Engine/pkg/models"
)

var db = database.NewDatabase()

type Vault models.Vault

func NewVault() *Vault {
	return &Vault{}
}

func (v *Vault) Create(name, password string) error {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	vaultDir := filepath.Join(userDir, ".svault", name)

	err = os.MkdirAll(vaultDir, 0777)
	if err != nil {
		return err
	}

	db.SaveVault(models.Vault{
		Name:     name,
		Password: password,
	})

	return nil
}

func (v *Vault) List() ([]models.Vault, error) {
	return db.ListVaults()
}
