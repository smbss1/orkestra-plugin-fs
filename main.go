package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/smbss1/orkestra-plugin-fs/shared"

	"github.com/hashicorp/go-plugin"
)

type FSExecutor struct{}

func (e *FSExecutor) GetCapabilities() ([]string, error) {
	return []string{"fs/read"}, nil
}

func (e *FSExecutor) Execute(node shared.Node, ctx shared.ExecutionContext) (interface{}, error) {
	var withMap map[string]interface{}
	if err := json.Unmarshal(node.With, &withMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal 'with' params for node '%s': %w", node.ID, err)
	}

	switch node.Uses {
	case "fs/read":
		return e.executeFsRead(withMap)
	default:
		return nil, fmt.Errorf("type de noeud inconnu dans le plugin fs: '%s'", node.Uses)
	}
}

func (e *FSExecutor) executeFsRead(with map[string]interface{}) (interface{}, error) {
	path, ok := with["path"].(string)
	if !ok {
		return nil, fmt.Errorf("le paramètre 'path' est requis et doit être une chaîne pour fs/read")
	}

	cleanPath := filepath.Clean(path)
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") {
		return nil, fmt.Errorf("chemin de fichier invalide ou non autorisé: '%s'", path)
	}

	log.Printf("LOG | Reading file: %s", cleanPath)
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("erreur de lecture du fichier '%s': %w", path, err)
	}

	return map[string]interface{}{
		"content": string(content),
		"path":    cleanPath,
		"size":    len(content),
	}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"executor": &shared.NodeExecutorPlugin{Impl: &FSExecutor{}},
		},
	})
}
