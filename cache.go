package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const EmojiURL = "https://github.com/muan/emojilib/raw/refs/tags/v4.0.0/dist/emoji-en-US.json"

func getDataDir() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(os.Getenv("HOME"), ".local", "share")
	}
	return filepath.Join(dataHome, "emoji-picker")
}

type CacheMetadata struct {
	Version string `json:"version"`
}

func GetEmojis() (map[string][]string, error) {
	dataDir := getDataDir()
	cachePath := filepath.Join(dataDir, "emojis.json")
	metaPath := filepath.Join(dataDir, "metadata.json")

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		if err := downloadEmojis(cachePath, metaPath); err != nil {
			return nil, fmt.Errorf("failed to download emojis: %w", err)
		}
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var emojis map[string][]string
	if err := json.Unmarshal(data, &emojis); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	return emojis, nil
}

func loadMetadata(path string) (CacheMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CacheMetadata{}, err
	}

	var metadata CacheMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return CacheMetadata{}, err
	}

	return metadata, nil
}

func downloadEmojis(cachePath, metaPath string) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Get(EmojiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	metadata := CacheMetadata{
		Version: "1.0",
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(metaPath, metadataJSON, 0644)
}
