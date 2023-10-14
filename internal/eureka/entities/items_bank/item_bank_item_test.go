package services

import (
	"testing"
)

func TestValidateItemID(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name    string
		item    ItemsBankItem
		isValid bool
	}{
		{
			name: "valid item id",
			item: ItemsBankItem{
				ItemID: "119aca7c-4e33-488b-8961-24e606fc8bac",
			},
			isValid: true,
		},
		{
			name: "invalid item id - too long",
			item: ItemsBankItem{
				ItemID: "119aca7c-4e33-488b-8961-24e606fc8bac119aca7c-4e33-488b-8961-24e606fc8bac119aca7c-4e33-488b-8961-24e606fc8bac119aca7c-4e33-488b-8961-24e606fc8bac119aca7c-4e33-488b-8961-24e606fc8bac119aca7c-4e33-488b-8961-24e606fc8bac119aca7c-4e33-488b-8961-24e606fc8bac119aca7c-4e33-488b-8961-24e606fc8bac"},
			isValid: false,
		},
		{
			name: "invalid item id - invalid character - accent",
			item: ItemsBankItem{
				ItemID: "119aca7c-4e33-488b-8961-24e606fc8bac`",
			},
			isValid: false,
		},
		{
			name: "invalid item id - invalid character - double quote",
			item: ItemsBankItem{
				ItemID: `119aca7c-4e33-488b-8961-24e606fc8bac"`,
			},
			isValid: false,
		},
		{
			name: "invalid item id - invalid character - single quote",
			item: ItemsBankItem{
				ItemID: `119aca7c-4e33-488b-8961-24e606fc8bac'`,
			},
			isValid: false,
		},
		{
			name: "invalid item id - invalid character - control character - line feed",
			item: ItemsBankItem{
				ItemID: "119aca7c-4e33-488b-8961-24e606fc8bac\n 119aca7c-4e33-488b-8961-24e606fc8bac0",
			},
			isValid: false,
		},
		{
			name: "invalid item id - invalid character - extended ascii",
			item: ItemsBankItem{
				ItemID: "119aca7c-4e33-488b-8961-24e606fc8bac€",
			},
			isValid: false,
		},
		{
			name: "valid item id - special characters",
			item: ItemsBankItem{
				ItemID: "!@#$%^&*()_+",
			},
			isValid: true,
		},
	}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			isValid := tc.item.IsItemIDValid()
			if isValid != tc.isValid {
				tt.Errorf("expected %v, got %v", tc.isValid, isValid)
			}
		})
	}

}
