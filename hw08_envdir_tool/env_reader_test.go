package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadDir(t *testing.T) {
	// Place your code here
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "API_KEY"), []byte("secret_token_123"), 0o644)
	if err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "DB_HOST"), []byte("localhost   \t"), 0o644)
	if err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "MULTILINE_VAR"), []byte("line1\nline2\nline3"), 0o644)
	if err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "OLD_VAR"), []byte(""), 0o644)
	if err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "NULL_BYTE_VAR"), []byte{0x00, 0x00}, 0o644)
	if err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}

	err = os.Mkdir(filepath.Join(tmpDir, "some_sub_dir"), 0o755)
	if err != nil {
		t.Fatalf("не удалось создать тестовую папку: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "INVALID=NAME"), []byte("some_val"), 0o644)
	if err != nil {
		t.Fatalf("не удалось создать тестовый файл: %v", err)
	}

	expectedEnv := Environment{
		"API_KEY":       EnvValue{Value: "secret_token_123", NeedRemove: false},
		"DB_HOST":       EnvValue{Value: "localhost", NeedRemove: false},
		"MULTILINE_VAR": EnvValue{Value: "line1", NeedRemove: false},
		"OLD_VAR":       EnvValue{Value: "", NeedRemove: true},
		"NULL_BYTE_VAR": EnvValue{Value: "\n\n", NeedRemove: false},
	}

	t.Run("Успешное чтение валидной директории", func(t *testing.T) {
		gotEnv, err := ReadDir(tmpDir)
		if err != nil {
			t.Fatalf("ReadDir() вернул ошибку: %v", err)
		}

		if !reflect.DeepEqual(gotEnv, expectedEnv) {
			t.Errorf("ReadDir() результат не совпадает с ожидаемым.\nПолучено: %+v\nОжидалось: %+v", gotEnv, expectedEnv)
		}
	})
}
