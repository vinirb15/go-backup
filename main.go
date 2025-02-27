package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Aviso: Arquivo .env não encontrado ou não pôde ser carregado")
	}

	if err := os.MkdirAll("dump", os.ModePerm); err != nil {
		fmt.Println("Erro ao criar diretório de dump:", err)
		return
	}

	dbType := os.Getenv("DB_TYPE")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")

	if dbType == "" || host == "" || port == "" || dbName == "" || user == "" || pass == "" {
		fmt.Println("Erro: todas as variáveis de ambiente devem estar definidas.")
		os.Exit(1)
	}

	c := cron.New()
	c.AddFunc("0 0 * * *", func() {
		backupDatabase(dbType, host, port, dbName, user, pass)
	})
	c.Start()

	fmt.Println("Agendador iniciado. Backup será executado diariamente às 00:00.")
	select {}
}

func backupDatabase(dbType, host, port, dbName, user, pass string) {
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("dump/backup_%s_%s.sql", dbName, timestamp)
	var cmd *exec.Cmd

	switch dbType {
	case "mysql":
		cmd = exec.Command("mysqldump", "-h", host, "-P", port, "-u", user, fmt.Sprintf("--password=%s", pass), dbName)
	case "postgres":
		os.Setenv("PGPASSWORD", pass)
		cmd = exec.Command("pg_dump", "-h", host, "-p", port, "-U", user, "-d", dbName, "-F", "c")
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
