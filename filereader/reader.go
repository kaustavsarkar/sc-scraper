package filereader

import (
	"io/ioutil"
	"os"
)

func GetHtml(filePath string) (string, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Convert the file content to a string
	return string(fileContent), nil
}

func TraverseOutputDir(root string) ([]string, error) {
	dirEntries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	dirNames := make([]string, 0)

	for _, entry := range dirEntries {
		if entry.IsDir() {
			dirNames = append(dirNames, entry.Name())
		}
	}
	return dirNames, nil
}
