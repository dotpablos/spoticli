package spotifysession

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func loadEnv() error {
	if err := godotenv.Load(); err == nil && os.Getenv("SPOTIFY_ID") != "" {
		return nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if _, statErr := os.Stat(envPath); statErr == nil {
			if loadErr := godotenv.Load(envPath); loadErr != nil {
				return fmt.Errorf("load %s: %w", envPath, loadErr)
			}

			if os.Getenv("SPOTIFY_ID") != "" {
				return nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	if os.Getenv("SPOTIFY_ID") == "" {
		return fmt.Errorf("SPOTIFY_ID is not set; put it in .env or export it in the shell")
	}

	return nil
}
