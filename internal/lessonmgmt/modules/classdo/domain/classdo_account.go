package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type ClassDoAction string

const (
	ActionUpsert ClassDoAction = "upsert"
	ActionDelete ClassDoAction = "delete"
)

const ClassDoIDLabel = "classdo_id"
const ClassDoEmailLabel = "classdo_email"
const ClassDoAPIKeyLabel = "classdo_api_key"
const ClassDoActionLabel = "action"

type ClassDoAccount struct {
	ClassDoID     string
	ClassDoEmail  string
	ClassDoAPIKey string
	Action        ClassDoAction
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func (c *ClassDoAccount) IsValid() error {
	if len(c.ClassDoEmail) == 0 {
		return fmt.Errorf("email cannot be empty")
	}
	if len(c.ClassDoAPIKey) == 0 {
		return fmt.Errorf("api key cannot be empty")
	}
	if len(c.Action) == 0 {
		return fmt.Errorf("action cannot be empty")
	}

	return nil
}

func (c *ClassDoAccount) DecryptAPIKey(decryptKey string) {
	generatedAPIDecrypted, err := crypt.Decrypt(c.ClassDoAPIKey, decryptKey)
	if err != nil {
		c.ClassDoAPIKey = ""
	} else {
		c.ClassDoAPIKey = generatedAPIDecrypted
	}
}

type ClassDoAccounts []*ClassDoAccount

type ClassDoAccountBuilder struct {
	classDoAccount *ClassDoAccount
}

func NewClassDoAccountBuilder() *ClassDoAccountBuilder {
	return &ClassDoAccountBuilder{
		classDoAccount: &ClassDoAccount{},
	}
}

func (c *ClassDoAccountBuilder) WithClassDoID(id string) *ClassDoAccountBuilder {
	c.classDoAccount.ClassDoID = id
	if len(id) == 0 {
		c.classDoAccount.ClassDoID = idutil.ULIDNow()
	}
	return c
}

func (c *ClassDoAccountBuilder) WithClassDoEmail(email string) *ClassDoAccountBuilder {
	c.classDoAccount.ClassDoEmail = email
	return c
}

func (c *ClassDoAccountBuilder) WithClassDoAPIKey(apiKey string) *ClassDoAccountBuilder {
	c.classDoAccount.ClassDoAPIKey = apiKey
	return c
}

func (c *ClassDoAccountBuilder) WithAction(action string) *ClassDoAccountBuilder {
	c.classDoAccount.Action = ClassDoAction(strings.ToLower(action))
	now := time.Now()

	if c.classDoAccount.Action == ActionDelete {
		c.classDoAccount.DeletedAt = &now
	}
	c.classDoAccount.CreatedAt = now
	c.classDoAccount.UpdatedAt = now

	return c
}

func (c *ClassDoAccountBuilder) EncryptAPIKey(encryptKey string) *ClassDoAccountBuilder {
	generatedAPIEncrypted, err := crypt.Encrypt(c.classDoAccount.ClassDoAPIKey, encryptKey)
	if err != nil {
		c.classDoAccount.ClassDoAPIKey = ""
	} else {
		c.classDoAccount.ClassDoAPIKey = generatedAPIEncrypted
	}
	return c
}

func (c *ClassDoAccountBuilder) Build() (*ClassDoAccount, error) {
	if err := c.classDoAccount.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid ClassDo account detail: %w", err)
	}

	return c.classDoAccount, nil
}
