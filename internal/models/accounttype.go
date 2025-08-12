package models

import "encoding/json"

type AccountType struct {
	ID        int    `db:"id"`
	TypeName  string `db:"type_name"`
	IsCredit  bool   `db:"is_credit"`
	IsDeleted bool   `db:"is_deleted"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

// MarshalJSON customizes JSON output to match AccountTypeDTO format
func (a *AccountType) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID       int    `json:"id"`
		TypeName string `json:"type_name"`
		IsCredit bool   `json:"is_credit"`
	}{
		ID:       a.ID,
		TypeName: a.TypeName,
		IsCredit: a.IsCredit,
	})
}

func (a *AccountType) String() string {
	isCreditStr := "false"
	if a.IsCredit {
		isCreditStr = "true"
	}
	return "[AccountType: " + a.TypeName + ", IsCredit: " + isCreditStr + "]"
}
