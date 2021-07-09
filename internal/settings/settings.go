package settings

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/PatrickRudolph/filebin/internal/filedata/backends"
)

var (
	settings *Settings
)

type Settings struct {
	AuthRealm       string
	AuthUsername    string
	AuthPassword    string
	BaseUrl         string
	HighlightStyle  string
	IdLength        uint8
	ListenAddr      string
	MaxAge          time.Duration
	StorageDir      string
	UploadMaxSizeMb uint
	Backend         backends.Backend
}

func getString(key string, def string, required bool) (string, error) {
	if v, found := os.LookupEnv(key); found {
		if required && v == "" {
			return "", fmt.Errorf("settings: %s empty", key)
		}
		return v, nil
	}
	if required && def == "" {
		return "", fmt.Errorf("settings: %s missing", key)
	}
	return def, nil
}

func getUint(key string, def uint64, required bool, base int, bitSize int) (uint64, error) {
	v, err := getString(key, strconv.FormatUint(def, base), required)
	if err != nil {
		return 0, err
	}
	v2, err := strconv.ParseUint(v, base, bitSize)
	if err != nil {
		return 0, err
	}
	if required && v2 == 0 {
		return 0, fmt.Errorf("settings: %s empty", key)
	}
	return v2, nil
}

func getBool(key string, def bool) (bool, error) {
	v, err := getString(key, strconv.FormatBool(def), true)
	if err != nil {
		return false, err
	}
	v2, err := strconv.ParseBool(v)
	if err != nil {
		return false, err
	}
	return v2, nil
}

func Get() (*Settings, error) {
	if settings != nil {
		return settings, nil
	}

	var err error
	s := &Settings{}

	s.AuthRealm, err = getString("FILEBIN_AUTH_REALM", "filebin", true)
	if err != nil {
		return nil, err
	}

	s.AuthUsername, err = getString("FILEBIN_AUTH_USERNAME", "", true)
	if err != nil {
		return nil, err
	}

	s.AuthPassword, err = getString("FILEBIN_AUTH_PASSWORD", "", true)
	if err != nil {
		return nil, err
	}

	s.BaseUrl, err = getString("FILEBIN_BASE_URL", "", false)
	if err != nil {
		return nil, err
	}

	s.HighlightStyle, err = getString("FILEBIN_HIGHLIGHT_STYLE", "trac", true)
	if err != nil {
		return nil, err
	}

	idLength, err := getUint("FILEBIN_ID_LENGTH", 8, true, 10, 8)
	if err != nil {
		return nil, err
	}
	if idLength < 8 {
		return nil, errors.New("FILEBIN_ID_LENGTH must be >= 8")
	}
	s.IdLength = uint8(idLength)

	s.ListenAddr, err = getString("FILEBIN_LISTEN_ADDR", ":8000", true)
	if err != nil {
		return nil, err
	}

	s.StorageDir, err = getString("FILEBIN_STORAGE_DIR", "", false)
	if err != nil {
		return nil, err
	}

	maxAgeinHours, err := getUint("FILEBIN_MAX_AGE", 24*14, false, 10, 0)
	if err != nil {
		return nil, err
	}
	s.MaxAge, err = time.ParseDuration(fmt.Sprintf("%dh", maxAgeinHours))
	if err != nil {
		return nil, err
	}
	uploadMaxSizeMb, err := getUint("FILEBIN_UPLOAD_MAX_SIZE_MB", 10, true, 10, 0)
	if err != nil {
		return nil, err
	}
	if uploadMaxSizeMb == 0 {
		return nil, errors.New("FILEBIN_UPLOAD_MAX_SIZE_MB must be > 0")
	}
	s.UploadMaxSizeMb = uint(uploadMaxSizeMb)

	s.Backend, err = backends.Lookup(
		s.StorageDir,
	)
	if err != nil {
		return nil, err
	}

	settings = s

	return s, nil
}
