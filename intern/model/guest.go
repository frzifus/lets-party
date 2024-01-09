package model

import (
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DietaryCategory int

const (
	DietaryCategoryUnknown DietaryCategory = iota
	DietaryCategoryVegan
	DietaryCategoryVegetarian
	DietaryCatagoryOmnivore
)

type InvitationStatus int

const (
	InvitationStatusUnknown InvitationStatus = iota
	InvitationStatusAccepted
	InvitationStatusRejected
	InvitationStatusNotAnswered
)

type GuestAgeCategory int

const (
	GuestAgeCategoryUnknown InvitationStatus = iota
	GuestAgeCategoryBaby
	GuestAgeCategorTeenager
	GuestAgeCategoryAdult
)

type Guest struct {
	ID               uuid.UUID        `json:"id" form:"-"`
	Deleteable       bool             `json:"deleteable" form:"-"`
	CreatedAt        *time.Time       `json:"created_at" form:"-"`
	UpdatedAt        *time.Time       `json:"updated_at" form:"-"`
	Firstname        string           `json:"firstname" form:"firstname"`
	Lastname         string           `json:"lastname" form:"lastname"`
	AgeCategory      GuestAgeCategory `json:"guest_age_category" form:"guest_age_category"`
	DietaryCategory  DietaryCategory  `json:"dietary_category" form:"dietary_category"`
	InvitationStatus InvitationStatus `json:"invitation_status" form:"invitation_status"`
}

func (g *Guest) Parse(input url.Values) {
	guestType := reflect.TypeOf(*g)
	for i := 0; i < guestType.NumField(); i++ {
		field := guestType.Field(i)
		fieldName := field.Tag.Get("form")

		if fieldName != "" {
			value, exists := input[fieldName]
			if exists && len(value) > 0 {
				// NOTE: Take only the first value
				fieldValue := value[0]

				switch field.Type.Kind() {
				case reflect.String:
					reflect.ValueOf(g).Elem().Field(i).SetString(fieldValue)
				case reflect.Bool:
					boolValue := strings.ToLower(fieldValue) == "true"
					reflect.ValueOf(g).Elem().Field(i).SetBool(boolValue)
				}
			}
		}
	}
}
