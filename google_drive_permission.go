package fundrive

import (
    "google.golang.org/api/drive/v3"
)

type Permission string

const (
    PublicPermission  Permission = "public"
    PrivatePermission Permission = "private"
)

// getPermission returns the permission based on the permission type
// default permission is public
func getPermission(permission Permission) *drive.Permission {
    if permission == PrivatePermission {
        return &drive.Permission{
            Type: "anyone",
            Role: "owner",
        }
    }
    return &drive.Permission{
        AllowFileDiscovery: true,
        Type:               "anyone",
        Role:               "reader",
    }
}
