package command

import (
	"io/ioutil"
	"path/filepath"

	"github.com/Jabba-Team/jabba/cfg"
	"strings"
)

func LsAlias() (map[string]string, error) {
	dir := cfg.Dir() // Assuming cfg.Dir() provides the directory containing alias files
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	aliases := make(map[string]string)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".alias") {
			filePath := filepath.Join(dir, file.Name())
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			name := strings.TrimSuffix(file.Name(), ".alias")
			aliases[name] = string(content)
		}
	}

	return aliases, nil
}
