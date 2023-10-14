package domain

import (
	"time"
)

type PartnerFormConfig struct {
	FormConfigID   string
	PartnerID      int
	FeatureName    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      time.Time
	FormConfigData []byte
}
type PartnerFormConfigBuilder struct {
	partnerFormConfig *PartnerFormConfig
}

func NewPartnerFormConfigBuilder() *PartnerFormConfigBuilder {
	return &PartnerFormConfigBuilder{
		partnerFormConfig: &PartnerFormConfig{},
	}
}

func (p *PartnerFormConfigBuilder) Build() (*PartnerFormConfig, error) {
	return p.partnerFormConfig, nil
}

func (p *PartnerFormConfigBuilder) WithFormConfigID(id string) *PartnerFormConfigBuilder {
	p.partnerFormConfig.FormConfigID = id
	return p
}

func (p *PartnerFormConfigBuilder) WithPartnerID(id int) *PartnerFormConfigBuilder {
	p.partnerFormConfig.PartnerID = id
	return p
}

func (p *PartnerFormConfigBuilder) WithFeatureName(name string) *PartnerFormConfigBuilder {
	p.partnerFormConfig.FeatureName = name
	return p
}

func (p *PartnerFormConfigBuilder) WithFormConfigData(data []byte) *PartnerFormConfigBuilder {
	p.partnerFormConfig.FormConfigData = data
	return p
}

func (p *PartnerFormConfigBuilder) WithModificationTime(createdAt, updatedAt time.Time) *PartnerFormConfigBuilder {
	p.partnerFormConfig.CreatedAt = createdAt
	p.partnerFormConfig.UpdatedAt = updatedAt
	return p
}
