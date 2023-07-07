package filereader

import (
	"io/ioutil"
)

func GetHtml(filePath string) (string, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Convert the file content to a string
	return string(fileContent), nil
}
