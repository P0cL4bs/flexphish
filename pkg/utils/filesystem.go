package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return os.ExpandEnv(path)
}

func GetAppRoot() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Dir(realPath)
}

func GetAppBinaryPath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		log.Fatal(err)
	}
	return realPath
}

func GetBasePath(relPath string) string {
	appRoot := GetAppRoot()

	var baseDir string
	if strings.HasPrefix(appRoot, "/opt/") {
		baseDir = filepath.Join(appRoot, "..")
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		baseDir = cwd
	}

	return filepath.Join(baseDir, relPath)
}

func SaveYAMLFile(path string, model any) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	defer encoder.Close()

	return encoder.Encode(model)
}

func GenerateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
