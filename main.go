package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

type dbConfig struct {
	dbType string
	host   string
	port   string
	dbName string
	user   string
	pass   string
}

func loadDBConfig() dbConfig {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Aviso: Arquivo .env não encontrado ou não pôde ser carregado")
	}

	cfg := dbConfig{
		dbType: os.Getenv("DB_TYPE"),
		host:   os.Getenv("DB_HOST"),
		port:   os.Getenv("DB_PORT"),
		dbName: os.Getenv("DB_NAME"),
		user:   os.Getenv("DB_USER"),
		pass:   os.Getenv("DB_PASS"),
	}

	if cfg.dbType == "" || cfg.host == "" || cfg.port == "" || cfg.dbName == "" || cfg.user == "" || cfg.pass == "" {
		fmt.Println("Erro: todas as variáveis de ambiente devem estar definidas.")
		os.Exit(1)
	}

	return cfg
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "restore" {
		if len(os.Args) < 3 {
			fmt.Println("Uso: db_backup restore <arquivo-de-dump>")
			os.Exit(1)
		}

		cfg := loadDBConfig()
		if err := restoreDatabase(cfg, os.Args[2]); err != nil {
			fmt.Println("Erro ao restaurar backup:", err)
			os.Exit(1)
		}

		fmt.Println("Restauração concluída com sucesso a partir de:", os.Args[2])
		return
	}

	if err := os.MkdirAll("dump", os.ModePerm); err != nil {
		fmt.Println("Erro ao criar diretório de dump:", err)
		return
	}

	cfg := loadDBConfig()

	cronTime := os.Getenv("CRON_TIME")
	if cronTime == "" {
		cronTime = "0 0 * * *"
	}

	loc := time.Local
	if tz := os.Getenv("TZ"); tz != "" {
		l, err := time.LoadLocation(tz)
		if err != nil {
			fmt.Println("Erro: fuso horário TZ inválido:", err)
			os.Exit(1)
		}
		loc = l
	}

	c := cron.New(cron.WithLocation(loc))
	if _, err := c.AddFunc(cronTime, func() {
		backupDatabase(cfg)
	}); err != nil {
		fmt.Println("Erro: expressão CRON_TIME inválida:", err)
		os.Exit(1)
	}
	c.Start()

	fmt.Printf("Agendador iniciado. Backup agendado para: %s (%s)\n", cronTime, loc)
	select {}
}

func backupDatabase(cfg dbConfig) {
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("dump/backup_%s_%s.sql", cfg.dbName, timestamp)
	var cmd *exec.Cmd

	switch cfg.dbType {
	case "mysql":
		cmd = exec.Command("mysqldump", "-h", cfg.host, "-P", cfg.port, "-u", cfg.user, fmt.Sprintf("--password=%s", cfg.pass), cfg.dbName)
	case "postgres":
		os.Setenv("PGPASSWORD", cfg.pass)
		cmd = exec.Command("pg_dump", "-h", cfg.host, "-p", cfg.port, "-U", cfg.user, "-d", cfg.dbName, "-F", "c")
	default:
		fmt.Println("Erro: Tipo de banco de dados não suportado.")
		return
	}

	outFile, err := os.Create(backupFile)
	if err != nil {
		fmt.Println("Erro ao criar arquivo de backup:", err)
		return
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Erro ao executar backup:", err)
	} else {
		fmt.Println("Backup realizado com sucesso:", backupFile)
	}
}

func restoreDatabase(cfg dbConfig, file string) error {
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("arquivo de dump não encontrado: %w", err)
	}

	var cmd *exec.Cmd

	switch cfg.dbType {
	case "mysql":
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd = exec.Command("mysql", "-h", cfg.host, "-P", cfg.port, "-u", cfg.user, fmt.Sprintf("--password=%s", cfg.pass), cfg.dbName)
		cmd.Stdin = f
	case "postgres":
		os.Setenv("PGPASSWORD", cfg.pass)
		cmd = exec.Command("pg_restore", "-h", cfg.host, "-p", cfg.port, "-U", cfg.user, "-d", cfg.dbName, "--clean", "--if-exists", "--no-owner", file)
	default:
		return fmt.Errorf("tipo de banco de dados não suportado: %s", cfg.dbType)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
