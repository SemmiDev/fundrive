package fundrive

const (
	FileURLPrefix = "https://drive.google.com/uc?id="
)

func GetFileURL(resourceID string) string {
	return FileURLPrefix + resourceID
}
