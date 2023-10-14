package queries

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"gotest.tools/assert"
)

func TestGenerateUnauthorizedLocationV2(t *testing.T) {
	t.Parallel()
	service := &GetLocationQueryHandler{}

	testCases := []struct {
		name              string
		locationTypes     map[string]*domain.LocationType
		locations         []*domain.Location
		expectedLocations map[string]domain.Location
		expectedError     error
	}{
		{
			name: "success case",
			locations: []*domain.Location{
				{LocationID: "B1", Name: "Brand 1", LocationType: "T1", ParentLocationID: "Org", AccessPath: "Org/B1", IsUnauthorized: false},
				{LocationID: "C2", Name: "Center 2", LocationType: "T3", ParentLocationID: "B2", AccessPath: "Org/B2/C2", IsUnauthorized: false},
				{LocationID: "C3", Name: "Center 3", LocationType: "T3", ParentLocationID: "Org", AccessPath: "Org/C3", IsUnauthorized: false},
			},
			expectedLocations: map[string]domain.Location{
				"Org": {LocationID: "Org", Name: "UnAuthorized", ParentLocationID: "", LocationType: "", AccessPath: "Org", IsUnauthorized: true},
				"B1":  {LocationID: "B1", Name: "Brand 1", ParentLocationID: "Org", LocationType: "T1", AccessPath: "Org/B1", IsUnauthorized: false},
				"B2":  {LocationID: "B2", Name: "UnAuthorized", ParentLocationID: "Org", LocationType: "", AccessPath: "Org/B2", IsUnauthorized: true},
				"C2":  {LocationID: "C2", Name: "Center 2", ParentLocationID: "B2", LocationType: "T3", AccessPath: "Org/B2/C2", IsUnauthorized: false},
				"C3":  {LocationID: "C3", Name: "Center 3", ParentLocationID: "Org", LocationType: "T3", AccessPath: "Org/C3", IsUnauthorized: false},
			},
		},
		{
			name: "just one center",
			locations: []*domain.Location{
				{LocationID: "P1", Name: "Place One 1", LocationType: "T4", ParentLocationID: "C1", AccessPath: "O/B1/A1/C1/P1", IsUnauthorized: false},
			},
			expectedLocations: map[string]domain.Location{
				"O":  {LocationID: "O", Name: "UnAuthorized", ParentLocationID: "", LocationType: "", AccessPath: "O", IsUnauthorized: true},
				"B1": {LocationID: "B1", Name: "UnAuthorized", ParentLocationID: "O", LocationType: "", AccessPath: "O/B1", IsUnauthorized: true},
				"A1": {LocationID: "A1", Name: "UnAuthorized", ParentLocationID: "B1", LocationType: "", AccessPath: "O/B1/A1", IsUnauthorized: true},
				"C1": {LocationID: "C1", Name: "UnAuthorized", ParentLocationID: "A1", LocationType: "", AccessPath: "O/B1/A1/C1", IsUnauthorized: true},
				"P1": {LocationID: "P1", Name: "Place One 1", ParentLocationID: "C1", LocationType: "T4", AccessPath: "O/B1/A1/C1/P1", IsUnauthorized: false},
			},
		},
		{
			name: "Two different branches",
			locations: []*domain.Location{
				{LocationID: "P1", Name: "Place One 1", LocationType: "T4", ParentLocationID: "C1", AccessPath: "O/B1/A1/C1/P1", IsUnauthorized: false},
				{LocationID: "P2", Name: "Place Two 2", LocationType: "T4", ParentLocationID: "C2", AccessPath: "O/B1/A2/C2/P2", IsUnauthorized: false},
			},
			expectedLocations: map[string]domain.Location{
				"O":  {LocationID: "O", Name: "UnAuthorized", ParentLocationID: "", LocationType: "", AccessPath: "O", IsUnauthorized: true},
				"B1": {LocationID: "B1", Name: "UnAuthorized", ParentLocationID: "O", LocationType: "", AccessPath: "O/B1", IsUnauthorized: true},
				// branch 1
				"A1": {LocationID: "A1", Name: "UnAuthorized", ParentLocationID: "B1", LocationType: "", AccessPath: "O/B1/A1", IsUnauthorized: true},
				"C1": {LocationID: "C1", Name: "UnAuthorized", ParentLocationID: "A1", LocationType: "", AccessPath: "O/B1/A1/C1", IsUnauthorized: true},
				"P1": {LocationID: "P1", Name: "Place One 1", ParentLocationID: "C1", LocationType: "T4", AccessPath: "O/B1/A1/C1/P1", IsUnauthorized: false},
				// branch 2
				"A2": {LocationID: "A2", Name: "UnAuthorized", ParentLocationID: "B1", LocationType: "", AccessPath: "O/B1/A2", IsUnauthorized: true},
				"C2": {LocationID: "C2", Name: "UnAuthorized", ParentLocationID: "A2", LocationType: "", AccessPath: "O/B1/A2/C2", IsUnauthorized: true},
				"P2": {LocationID: "P2", Name: "Place Two 2", ParentLocationID: "C2", LocationType: "T4", AccessPath: "O/B1/A2/C2/P2", IsUnauthorized: false},
			},
		},
		{
			name: "Two different branches with authorized Parent (put parent after)",
			locations: []*domain.Location{
				{LocationID: "P1", Name: "Place One 1", LocationType: "T4", ParentLocationID: "C1", AccessPath: "O/B1/A1/C1/P1", IsUnauthorized: false},
				{LocationID: "P2", Name: "Place Two 2", LocationType: "T4", ParentLocationID: "C2", AccessPath: "O/B1/A2/C2/P2", IsUnauthorized: false},
				{LocationID: "C2", Name: "Center Two 2", LocationType: "T3", ParentLocationID: "A2", AccessPath: "O/B1/A2/C2", IsUnauthorized: false},
			},
			expectedLocations: map[string]domain.Location{
				"O":  {LocationID: "O", Name: "UnAuthorized", ParentLocationID: "", LocationType: "", AccessPath: "O", IsUnauthorized: true},
				"B1": {LocationID: "B1", Name: "UnAuthorized", ParentLocationID: "O", LocationType: "", AccessPath: "O/B1", IsUnauthorized: true},
				// branch 1
				"A1": {LocationID: "A1", Name: "UnAuthorized", ParentLocationID: "B1", LocationType: "", AccessPath: "O/B1/A1", IsUnauthorized: true},
				"C1": {LocationID: "C1", Name: "UnAuthorized", ParentLocationID: "A1", LocationType: "", AccessPath: "O/B1/A1/C1", IsUnauthorized: true},
				"P1": {LocationID: "P1", Name: "Place One 1", ParentLocationID: "C1", LocationType: "T4", AccessPath: "O/B1/A1/C1/P1", IsUnauthorized: false},
				// branch 2
				"A2": {LocationID: "A2", Name: "UnAuthorized", ParentLocationID: "B1", LocationType: "", AccessPath: "O/B1/A2", IsUnauthorized: true},
				"C2": {LocationID: "C2", Name: "Center Two 2", ParentLocationID: "A2", LocationType: "T3", AccessPath: "O/B1/A2/C2", IsUnauthorized: false},
				"P2": {LocationID: "P2", Name: "Place Two 2", ParentLocationID: "C2", LocationType: "T4", AccessPath: "O/B1/A2/C2/P2", IsUnauthorized: false},
			},
		},
	}

	for _, tc := range testCases {
		t.Run("success", func(t *testing.T) {
			result, err := service.generateUnauthorizedLocationV2(tc.locations)
			fmt.Println(result)
			assert.Equal(t, len(tc.expectedLocations), len(result))
			assert.Equal(t, tc.expectedError, err)
			for _, l := range result {
				assert.Equal(t, tc.expectedLocations[l.LocationID], *l)
			}
		})
	}
}

func TestBuildLocationTree(t *testing.T) {
	t.Parallel()
	service := &GetLocationQueryHandler{}
	today := time.Now()
	yesterday := time.Now().AddDate(0, 0, -1)
	testCases := []struct {
		name         string
		locations    []*domain.Location
		locationTree *domain.TreeLocation
		locationType map[string]*domain.LocationType
		lowestType   string
		expectedErr  error
	}{
		{
			name: "normal case",
			locations: []*domain.Location{
				{LocationID: "O", Name: "UnAuthorized", PartnerInternalID: "p_O", ParentLocationID: "", LocationType: "", CreatedAt: yesterday, UpdatedAt: yesterday, AccessPath: "O", IsUnauthorized: true},
				{LocationID: "B1", Name: "UnAuthorized", PartnerInternalID: "p_B1", ParentLocationID: "O", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1", IsUnauthorized: true},
				{LocationID: "A1", Name: "UnAuthorized", PartnerInternalID: "p_A1", ParentLocationID: "B1", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A1", IsUnauthorized: true},
				{LocationID: "C1", Name: "UnAuthorized", PartnerInternalID: "p_C1", ParentLocationID: "A1", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A1/C1", IsUnauthorized: true},
				{LocationID: "P1", Name: "Place One 1", PartnerInternalID: "p_P1", ParentLocationID: "C1", LocationType: "T4", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A1/C1/P1", IsUnauthorized: false},
				// branch 2
				{LocationID: "A2", Name: "UnAuthorized", PartnerInternalID: "p_A2", ParentLocationID: "B1", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A2", IsUnauthorized: true},
				{LocationID: "C2", Name: "UnAuthorized", PartnerInternalID: "p_C2", ParentLocationID: "A2", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A2/C2", IsUnauthorized: true},
				{LocationID: "P2", Name: "Place Two 2", PartnerInternalID: "p_P2", ParentLocationID: "C2", LocationType: "T4", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A2/C2/P2", IsUnauthorized: false},
			},
			locationType: map[string]*domain.LocationType{
				"T0": {LocationTypeID: "TO", Level: 0, Name: "Org"},
				"T1": {LocationTypeID: "T1", Level: 1, Name: "Brand"},
				"T2": {LocationTypeID: "T2", Level: 2, Name: "Area"},
				"T3": {LocationTypeID: "T3", Level: 3, Name: "Center"},
				"T4": {LocationTypeID: "T4", Level: 4, Name: "Place"},
			},
			lowestType: "T4",
			locationTree: &domain.TreeLocation{
				LocationID:        "O",
				Name:              "UnAuthorized",
				PartnerInternalID: "p_O",
				ParentLocationID:  "",
				LocationType:      "",
				IsArchived:        false,
				AccessPath:        "O",
				IsUnauthorized:    true,
				IsLowestLevel:     false,
				CreatedAt:         yesterday,
				UpdatedAt:         yesterday,
				Children: []*domain.TreeLocation{
					{
						LocationID:        "B1",
						Name:              "UnAuthorized",
						PartnerInternalID: "p_B1",
						ParentLocationID:  "O",
						LocationType:      "",
						IsArchived:        false,
						AccessPath:        "O/B1",
						IsUnauthorized:    true,
						IsLowestLevel:     false,
						CreatedAt:         today,
						UpdatedAt:         today,
						Children: []*domain.TreeLocation{
							{
								LocationID:        "A1",
								Name:              "UnAuthorized",
								PartnerInternalID: "p_A1",
								ParentLocationID:  "B1",
								LocationType:      "",
								IsArchived:        false,
								AccessPath:        "O/B1/A1",
								IsUnauthorized:    true,
								IsLowestLevel:     false,
								CreatedAt:         today,
								UpdatedAt:         today,
								Children: []*domain.TreeLocation{
									{
										LocationID:        "C1",
										Name:              "UnAuthorized",
										PartnerInternalID: "p_C1",
										ParentLocationID:  "A1",
										LocationType:      "",
										IsArchived:        false,
										AccessPath:        "O/B1/A1/C1",
										IsUnauthorized:    true,
										IsLowestLevel:     false,
										CreatedAt:         today,
										UpdatedAt:         today,
										Children: []*domain.TreeLocation{
											{
												LocationID:        "P1",
												Name:              "Place One 1",
												PartnerInternalID: "p_P1",
												ParentLocationID:  "C1",
												LocationType:      "T4",
												IsArchived:        false,
												AccessPath:        "O/B1/A1/C1/P1",
												IsUnauthorized:    false,
												IsLowestLevel:     true,
												CreatedAt:         today,
												UpdatedAt:         today,
											},
										},
									},
								},
							},
							{
								LocationID:        "A2",
								Name:              "UnAuthorized",
								PartnerInternalID: "p_A2",
								ParentLocationID:  "B1",
								LocationType:      "",
								IsArchived:        false,
								AccessPath:        "O/B1/A2",
								IsUnauthorized:    true,
								IsLowestLevel:     false,
								CreatedAt:         today,
								UpdatedAt:         today,
								Children: []*domain.TreeLocation{
									{
										LocationID:        "C2",
										Name:              "UnAuthorized",
										PartnerInternalID: "p_C2",
										ParentLocationID:  "A2",
										LocationType:      "",
										IsArchived:        false,
										AccessPath:        "O/B1/A2/C2",
										IsUnauthorized:    true,
										IsLowestLevel:     false,
										CreatedAt:         today,
										UpdatedAt:         today,
										Children: []*domain.TreeLocation{
											{
												LocationID:        "P2",
												Name:              "Place Two 2",
												PartnerInternalID: "p_P2",
												ParentLocationID:  "C2",
												LocationType:      "T4",
												IsArchived:        false,
												AccessPath:        "O/B1/A2/C2/P2",
												IsUnauthorized:    false,
												IsLowestLevel:     true,
												CreatedAt:         today,
												UpdatedAt:         today,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "sort by updated at -> updated at ascending case",
			locations: []*domain.Location{
				{LocationID: "O", Name: "UnAuthorized", PartnerInternalID: "p_O", ParentLocationID: "", LocationType: "", CreatedAt: yesterday, UpdatedAt: yesterday, AccessPath: "O", IsUnauthorized: true},
				{LocationID: "B1", Name: "UnAuthorized", PartnerInternalID: "p_B1", ParentLocationID: "O", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1", IsUnauthorized: true},
				{LocationID: "A1", Name: "UnAuthorized", PartnerInternalID: "p_A1", ParentLocationID: "B1", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A1", IsUnauthorized: true},
				{LocationID: "C1", Name: "UnAuthorized", PartnerInternalID: "p_C1", ParentLocationID: "A1", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A1/C1", IsUnauthorized: true},
				{LocationID: "P1", Name: "Place One 1", PartnerInternalID: "p_P1", ParentLocationID: "C1", LocationType: "T4", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A1/C1/P1", IsUnauthorized: false},
				// branch 2
				{LocationID: "A2", Name: "UnAuthorized", PartnerInternalID: "p_A2", ParentLocationID: "B1", LocationType: "", CreatedAt: yesterday, UpdatedAt: today, AccessPath: "O/B1/A2", IsUnauthorized: true},
				{LocationID: "C2", Name: "UnAuthorized", PartnerInternalID: "p_C2", ParentLocationID: "A2", LocationType: "", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A2/C2", IsUnauthorized: true},
				{LocationID: "P2", Name: "Place Two 2", PartnerInternalID: "p_P2", ParentLocationID: "C2", LocationType: "T4", CreatedAt: today, UpdatedAt: today, AccessPath: "O/B1/A2/C2/P2", IsUnauthorized: false},
			},
			locationType: map[string]*domain.LocationType{
				"T0": {LocationTypeID: "TO", Level: 0, Name: "Org"},
				"T1": {LocationTypeID: "T1", Level: 1, Name: "Brand"},
				"T2": {LocationTypeID: "T2", Level: 2, Name: "Area"},
				"T3": {LocationTypeID: "T3", Level: 3, Name: "Center"},
				"T4": {LocationTypeID: "T4", Level: 4, Name: "Place"},
			},
			lowestType: "T4",
			locationTree: &domain.TreeLocation{
				LocationID:        "O",
				Name:              "UnAuthorized",
				PartnerInternalID: "p_O",
				ParentLocationID:  "",
				LocationType:      "",
				IsArchived:        false,
				AccessPath:        "O",
				IsUnauthorized:    true,
				IsLowestLevel:     false,
				CreatedAt:         yesterday,
				UpdatedAt:         yesterday,
				Children: []*domain.TreeLocation{
					{
						LocationID:        "B1",
						Name:              "UnAuthorized",
						PartnerInternalID: "p_B1",
						ParentLocationID:  "O",
						LocationType:      "",
						IsArchived:        false,
						AccessPath:        "O/B1",
						IsUnauthorized:    true,
						IsLowestLevel:     false,
						CreatedAt:         today,
						UpdatedAt:         today,
						Children: []*domain.TreeLocation{
							{
								LocationID:        "A2",
								Name:              "UnAuthorized",
								PartnerInternalID: "p_A2",
								ParentLocationID:  "B1",
								LocationType:      "",
								IsArchived:        false,
								AccessPath:        "O/B1/A2",
								IsUnauthorized:    true,
								IsLowestLevel:     false,
								CreatedAt:         yesterday,
								UpdatedAt:         today,
								Children: []*domain.TreeLocation{
									{
										LocationID:        "C2",
										Name:              "UnAuthorized",
										PartnerInternalID: "p_C2",
										ParentLocationID:  "A2",
										LocationType:      "",
										IsArchived:        false,
										AccessPath:        "O/B1/A2/C2",
										IsUnauthorized:    true,
										IsLowestLevel:     false,
										CreatedAt:         today,
										UpdatedAt:         today,
										Children: []*domain.TreeLocation{
											{
												LocationID:        "P2",
												Name:              "Place Two 2",
												PartnerInternalID: "p_P2",
												ParentLocationID:  "C2",
												LocationType:      "T4",
												IsArchived:        false,
												AccessPath:        "O/B1/A2/C2/P2",
												IsUnauthorized:    false,
												IsLowestLevel:     true,
												CreatedAt:         today,
												UpdatedAt:         today,
											},
										},
									},
								},
							},
							{
								LocationID:        "A1",
								Name:              "UnAuthorized",
								PartnerInternalID: "p_A1",
								ParentLocationID:  "B1",
								LocationType:      "",
								IsArchived:        false,
								AccessPath:        "O/B1/A1",
								IsUnauthorized:    true,
								IsLowestLevel:     false,
								CreatedAt:         today,
								UpdatedAt:         today,
								Children: []*domain.TreeLocation{
									{
										LocationID:        "C1",
										Name:              "UnAuthorized",
										PartnerInternalID: "p_C1",
										ParentLocationID:  "A1",
										LocationType:      "",
										IsArchived:        false,
										AccessPath:        "O/B1/A1/C1",
										IsUnauthorized:    true,
										IsLowestLevel:     false,
										CreatedAt:         today,
										UpdatedAt:         today,
										Children: []*domain.TreeLocation{
											{
												LocationID:        "P1",
												Name:              "Place One 1",
												PartnerInternalID: "p_P1",
												ParentLocationID:  "C1",
												LocationType:      "T4",
												IsArchived:        false,
												AccessPath:        "O/B1/A1/C1/P1",
												IsUnauthorized:    false,
												IsLowestLevel:     true,
												CreatedAt:         today,
												UpdatedAt:         today,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "no locations belong to lowest location type",
			locations: []*domain.Location{
				{LocationID: "O", Name: "Org", PartnerInternalID: "p_O", ParentLocationID: "", LocationType: "T0", CreatedAt: yesterday, UpdatedAt: yesterday, AccessPath: "O", IsUnauthorized: true},
			},
			locationType: map[string]*domain.LocationType{
				"T0": {LocationTypeID: "TO", Level: 0, Name: "Org"},
			},
			lowestType: "T0",
			locationTree: &domain.TreeLocation{
				LocationID:        "O",
				Name:              "Org",
				PartnerInternalID: "p_O",
				ParentLocationID:  "",
				LocationType:      "T0",
				IsArchived:        false,
				AccessPath:        "O",
				IsUnauthorized:    true,
				IsLowestLevel:     true,
				CreatedAt:         yesterday,
				UpdatedAt:         yesterday,
				Children:          nil,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			result, err := service.buildLocationTree(test.locations, test.locationType, test.locationType[test.lowestType])
			if test.expectedErr != nil {
				assert.Equal(t, test.expectedErr, err)
			} else {
				assert.Equal(t, nil, err)
				json, _ := json.Marshal(test.locationTree)
				assert.Equal(t, string(json), result)
			}
		})
	}
}
