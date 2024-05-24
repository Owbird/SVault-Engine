package vault

import (
	"fmt"

	"github.com/Owbird/SVault-Engine/internal/database"
	"github.com/Owbird/SVault-Engine/pkg/models"
)

var db = database.NewDatabase()

type Vault models.Vault

func NewVault() *Vault {
	return &Vault{}
}

func (v *Vault) Create(name, password string) error {
	err := db.SaveVault(models.Vault{
		Name:     name,
		Password: password,
	})
	if err != nil {
		return err
	}

	return nil
}

func (v *Vault) List() ([]models.Vault, error) {
	return db.ListVaults()
}

func (v *Vault) Auth(name, pwd string) (bool, error) {
	vault, err := db.GetVault(name)
	if err != nil {
		return false, err
	}

	if vault.Name == "" {
		return false, fmt.Errorf("'%v' vault does not exist", name)
	}

	return vault.Password == pwd, nil
}
