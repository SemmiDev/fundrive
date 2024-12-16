package fundrive

const (
	fileURLPrefix = "https://drive.google.com/uc?id="
)

func GetFileURL(resourceID string) string {
	return fileURLPrefix + resourceID
}
