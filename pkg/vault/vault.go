package vault

import (
	"io/fs"
	"os"
	"path"

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

	bankDir := path.Join(userDir, ".svault")

	stat, err := os.Stat(bankDir)
	if err != nil {
		return err
	}

	if stat == nil {
		os.Mkdir(bankDir, fs.FileMode(0777))
	}

	vaultDir := path.Join(bankDir, name)

	err = os.Mkdir(vaultDir, 0777)

	if err != nil {
		return err
	}

	db.SaveVault(models.Vault{
		Name:     name,
		Password: password,
	})

	return nil

}
