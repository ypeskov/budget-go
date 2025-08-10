package models

type AccountType struct {
	ID        int    `db:"id"`
	TypeName  string `db:"type_name"`
	IsCredit  bool   `db:"is_credit"`
	IsDeleted bool   `db:"is_deleted"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (a *AccountType) String() string {
	isCreditStr := "false"
	if a.IsCredit {
		isCreditStr = "true"
	}
	return "[AccountType: " + a.TypeName + ", IsCredit: " + isCreditStr + "]"
}
