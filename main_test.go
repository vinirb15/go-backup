package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupExpiredBackups(t *testing.T) {
	dumpDir := t.TempDir()
	now := time.Date(2026, time.November, 23, 12, 0, 0, 0, time.UTC)

	files := []struct {
		name string
		age  time.Duration
	}{
		{"backup_old_20261120_110000.sql", 72*time.Hour + time.Second},
		{"backup_new_20261120_120000.sql", 72*time.Hour - time.Second},
		{"backup_exact_20261120_120000.sql", 72 * time.Hour},
		{"other.sql", 96 * time.Hour},
	}
	for _, file := range files {
		path := filepath.Join(dumpDir, file.name)
		if err := os.WriteFile(path, []byte("dump"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, now.Add(-file.age), now.Add(-file.age)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dumpDir, "backup_directory_20261120_120000.sql"), 0o700); err != nil {
		t.Fatal(err)
	}

	cleanupExpiredBackups(dumpDir, now, 72*time.Hour)

	assertFileExists(t, filepath.Join(dumpDir, "backup_new_20261120_120000.sql"))
	assertFileExists(t, filepath.Join(dumpDir, "backup_exact_20261120_120000.sql"))
	assertFileExists(t, filepath.Join(dumpDir, "other.sql"))
	assertFileExists(t, filepath.Join(dumpDir, "backup_directory_20261120_120000.sql"))
	assertFileDoesNotExist(t, filepath.Join(dumpDir, "backup_old_20261120_110000.sql"))
}

func TestBackupDatabaseFailureDoesNotCleanUpExpiredBackups(t *testing.T) {
	workDir := t.TempDir()
	binDir := filepath.Join(workDir, "bin")
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Errorf("restaurar diretório de trabalho: %v", err)
		}
	})
	if err := os.Mkdir(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	command := filepath.Join(binDir, "mysqldump")
	if err := os.WriteFile(command, []byte("#!/bin/sh\nprintf partial\nexit 1\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := os.Mkdir("dump", 0o700); err != nil {
		t.Fatal(err)
	}
	expired := filepath.Join("dump", "backup_old_20261120_110000.sql")
	if err := os.WriteFile(expired, []byte("dump"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(expired, time.Now().Add(-73*time.Hour), time.Now().Add(-73*time.Hour)); err != nil {
		t.Fatal(err)
	}

	if err := backupDatabase(dbConfig{alias: "db", dbType: "mysql"}); err == nil {
		t.Fatal("backupDatabase() não retornou erro quando mysqldump falhou")
	}
	assertFileExists(t, expired)
	entries, err := os.ReadDir("dump")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(expired) {
		t.Fatalf("esperava apenas o dump expirado após falha, obteve: %+v", entries)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("esperava que %s existisse: %v", path, err)
	}
}

func assertFileDoesNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("esperava que %s não existisse: %v", path, err)
	}
}
