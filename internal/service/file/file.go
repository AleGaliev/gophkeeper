package file

import (
	"os"
	"path/filepath"
)

func ReadFile(filename string) ([]byte, error) {
	filename, err := createFullPath(filename)
	if err != nil {
		return nil, err
	}
	keyBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return keyBytes, nil
}

func WriteFile(filename string, data []byte) error {
	filename, err := createFullPath(filename)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return err
	}

	// Записываем/перезаписываем файл
	return os.WriteFile(filename, data, 0644)
}

func createFullPath(filename string) (string, error) {
	if filename[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		filename = filepath.Join(homeDir, filename[2:])
	}
	return filename, nil
}
