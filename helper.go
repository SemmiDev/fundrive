package fundrive

const (
	fileURLPrefix = "https://drive.google.com/uc?id="
)

func GetFileURL(resourceID string) string {
	return fileURLPrefix + resourceID
}

func PanicIfNeeded(err error) {
	if err != nil {
		panic(err)
	}
}
