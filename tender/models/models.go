package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;size:100;default:uuid_generate_v4()" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	FirstName string    `gorm:"size:50" json:"firstName"`
	LastName  string    `gorm:"size:50" json:"lastName"`
	CreatedAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

type OrganizationType string

const (
	IE  OrganizationType = "IE"
	LLC OrganizationType = "LLC"
	JSC OrganizationType = "JSC"
)

type Organization struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey;size:100;default:uuid_generate_v4()" json:"id"`
	Name        string           `gorm:"not null;size:100" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	Type        OrganizationType `gorm:"type:organization_type" json:"type"`
	CreatedAt   time.Time        `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time        `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

type OrganizationResponsible struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey;size:100;default:uuid_generate_v4()" json:"id"`
	OrganizationID uuid.UUID    `gorm:"not null" json:"organizationId"`
	UserID         uuid.UUID    `gorm:"not null" json:"userId"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE" json:"organization"`
	User           Employee     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user"`
}

type TenderStatus string

const (
	TENDER_CREATED   TenderStatus = "CREATED"
	TENDER_PUBLISHED TenderStatus = "PUBLISHED"
	TENDER_CLOSED    TenderStatus = "CLOSED"
)

type TenderServiceType string

const (
	CONSTRUCTION TenderServiceType = "CONSTRUCTION"
	DELIVERY     TenderServiceType = "DELIVERY"
	MANUFACTURE  TenderServiceType = "MANUFACTURE"
)

type Tender struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey;size:100;default:uuid_generate_v4()" json:"id"`
	Name           string            `gorm:"not null;size:100" json:"name"`
	Description    string            `gorm:"not null;size:500" json:"description"`
	Status         TenderStatus      `gorm:"type:tender_status;default:'CREATED'" json:"status"`
	ServiceType    TenderServiceType `gorm:"type:tender_service_type" json:"serviceType"`
	OrganizationID string            `gorm:"not null;size:100" json:"organizationId"`
	Version        uint              `gorm:"default:1;not null" json:"version"`
	CreatedAt      time.Time         `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

type NewTenderRequest struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	ServiceType     TenderServiceType `json:"serviceType"`
	OrganizationID  string            `json:"organizationId"`
	CreatorUsername string            `json:"creatorUsername"`
}

type TenderResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      TenderStatus      `json:"status"`
	ServiceType TenderServiceType `json:"serviceType"`
	Version     uint              `json:"version"`
	CreatedAt   time.Time         `json:"createdAt"`
}

type TenderVersion struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey;size:100;default:uuid_generate_v4()" json:"id"`
	TenderID       string            `gorm:"not null" json:"tenderId"`
	Name           string            `gorm:"not null;size:100" json:"name"`
	Description    string            `gorm:"not null;size:500" json:"description"`
	Status         TenderStatus      `gorm:"type:tender_status;default:'CREATED'" json:"status"`
	ServiceType    TenderServiceType `gorm:"type:tender_service_type;not null" json:"serviceType"`
	OrganizationID string            `gorm:"not null" json:"organizationId"`
	Version        uint              `gorm:"not null" json:"version"`
	CreatedAt      time.Time         `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

type BidStatus string

const (
	BID_CREATED   BidStatus = "CREATED"
	BID_PUBLISHED BidStatus = "PUBLISHED"
	BID_CANCELED  BidStatus = "CANCELED"
)

type BidAuthorType string

const (
	AUTHOR_ORGANIZATION BidAuthorType = "ORGANIZATION"
	AUTHOR_USER         BidAuthorType = "USER"
)

type Bid struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;size:100;default:uuid_generate_v4()" json:"id"`
	Name        string        `gorm:"not null;size:100" json:"name"`
	Description string        `gorm:"not null;size:500" json:"description"`
	Status      BidStatus     `gorm:"type:bid_status;default:'CREATED'" json:"status"`
	TenderID    string        `gorm:"not null;size:100" json:"tenderId"`
	AuthorType  BidAuthorType `gorm:"type:bid_author_type; not null" json:"authorType"`
	AuthorID    string        `gorm:"not null;size:100;default:uuid_generate_v4()" json:"authorId"`
	Version     uint          `gorm:"default:1;not null" json:"version"`
	CreatedAt   time.Time     `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

type NewBidRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	TenderID    string        `json:"tenderId"`
	AuthorType  BidAuthorType `json:"authorType"`
	AuthorID    string        `json:"authorId"`
}

type BidResponse struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Status     BidStatus     `json:"status"`
	AuthorType BidAuthorType `json:"authorType"`
	AuthorID   string        `json:"authorId"`
	Version    uint          `json:"version"`
	CreatedAt  time.Time     `json:"createdAt"`
}

type BidVersion struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;size:100;default:uuid_generate_v4()" json:"id"`
	BidID       string        `gorm:"not null" json:"bidId"`
	Name        string        `gorm:"not null;size:100" json:"name"`
	Description string        `gorm:"not null;size:500" json:"description"`
	Status      BidStatus     `gorm:"type:bid_status;default:'CREATED'" json:"status"`
	TenderID    string        `gorm:"not null;size:100" json:"tenderId"`
	AuthorType  BidAuthorType `gorm:"type:bid_author_type; not null" json:"authorType"`
	AuthorID    string        `gorm:"not null;size:100" json:"authorId"`
	Version     uint          `gorm:"default:1;not null" json:"version"`
	CreatedAt   time.Time     `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

type ErrorResponse struct {
	Reason string `json:"reason"`
}

func NewErrorResponse(reason string) *ErrorResponse {
	if len(reason) < 5 {
		reason = "Описание ошибки должно быть не менее 5 символов."
	}
	return &ErrorResponse{
		Reason: reason,
	}
}
