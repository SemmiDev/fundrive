package fundrive

import (
	"google.golang.org/api/drive/v3"
)

var (
	publicPermission = &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}

	privatePermission = &drive.Permission{
		Type: "default",
		Role: "owner",
	}
)

func getPermission(permission string) *drive.Permission {
	if permission == "private" {
		return privatePermission
	}
	return publicPermission
}
