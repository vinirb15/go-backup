package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

type dbConfig struct {
	alias  string
	dbType string
	host   string
	port   string
	dbName string
	user   string
	pass   string
}

var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// readDBConfig lê as variáveis <prefix>DB_* e retorna o nome das que estiverem faltando.
func readDBConfig(alias, prefix string) (dbConfig, []string) {
	var missing []string
	get := func(name string) string {
		v := os.Getenv(prefix + name)
		if v == "" {
			missing = append(missing, prefix+name)
		}
		return v
	}

	cfg := dbConfig{
		alias:  alias,
		dbType: get("DB_TYPE"),
		host:   get("DB_HOST"),
		port:   get("DB_PORT"),
		dbName: get("DB_NAME"),
		user:   get("DB_USER"),
		pass:   get("DB_PASS"),
	}

	return cfg, missing
}

func loadDBConfigs() []dbConfig {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Aviso: Arquivo .env não encontrado ou não pôde ser carregado")
	}

	var configs []dbConfig
	var errs []string

	databases := strings.TrimSpace(os.Getenv("DATABASES"))
	if databases == "" {
		cfg, missing := readDBConfig("", "")
		cfg.alias = cfg.dbName
		for _, name := range missing {
			errs = append(errs, "variável não definida: "+name)
		}
		configs = append(configs, cfg)
	} else {
		seen := map[string]bool{}
		for _, alias := range strings.Split(databases, ",") {
			alias = strings.TrimSpace(alias)
			if !aliasPattern.MatchString(alias) {
				errs = append(errs, fmt.Sprintf("alias inválido em DATABASES: %q (use apenas letras, números e _)", alias))
				continue
			}
			key := strings.ToUpper(alias)
			if seen[key] {
				errs = append(errs, "alias duplicado em DATABASES: "+alias)
				continue
			}
			seen[key] = true

			cfg, missing := readDBConfig(alias, key+"_")
			for _, name := range missing {
				errs = append(errs, "variável não definida: "+name)
			}
			configs = append(configs, cfg)
		}
	}

	for _, cfg := range configs {
		if cfg.dbType != "" && cfg.dbType != "mysql" && cfg.dbType != "postgres" {
			errs = append(errs, fmt.Sprintf("[%s] tipo de banco de dados não suportado: %s", cfg.alias, cfg.dbType))
		}
	}

	if len(errs) > 0 {
		fmt.Println("Erro: configuração de banco de dados inválida:")
		for _, e := range errs {
			fmt.Println("  -", e)
		}
		os.Exit(1)
	}

	return configs
}

func aliases(configs []dbConfig) string {
	names := make([]string, len(configs))
	for i, cfg := range configs {
		names[i] = cfg.alias
	}
	return strings.Join(names, ", ")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "restore" {
		runRestore(os.Args[2:])
		return
	}

	if err := os.MkdirAll("dump", os.ModePerm); err != nil {
		fmt.Println("Erro ao criar diretório de dump:", err)
		return
	}

	configs := loadDBConfigs()

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
		for _, cfg := range configs {
			backupDatabase(cfg)
		}
	}); err != nil {
		fmt.Println("Erro: expressão CRON_TIME inválida:", err)
		os.Exit(1)
	}
	c.Start()

	fmt.Printf("Agendador iniciado. Backup de [%s] agendado para: %s (%s)\n", aliases(configs), cronTime, loc)
	select {}
}

func runRestore(args []string) {
	if len(args) < 1 || len(args) > 2 {
		fmt.Println("Uso: db_backup restore [alias] <arquivo-de-dump>")
		os.Exit(1)
	}

	configs := loadDBConfigs()

	var cfg dbConfig
	var file string
	if len(args) == 1 {
		if len(configs) > 1 {
			fmt.Printf("Erro: há %d bancos configurados; informe o alias: db_backup restore <alias> <arquivo-de-dump> (disponíveis: %s)\n", len(configs), aliases(configs))
			os.Exit(1)
		}
		cfg, file = configs[0], args[0]
	} else {
		found := false
		for _, c := range configs {
			if strings.EqualFold(c.alias, args[0]) {
				cfg, found = c, true
				break
			}
		}
		if !found {
			fmt.Printf("Erro: alias desconhecido: %s (disponíveis: %s)\n", args[0], aliases(configs))
			os.Exit(1)
		}
		file = args[1]
	}

	if err := restoreDatabase(cfg, file); err != nil {
		fmt.Printf("[%s] Erro ao restaurar backup: %v\n", cfg.alias, err)
		os.Exit(1)
	}

	fmt.Printf("[%s] Restauração concluída com sucesso a partir de: %s\n", cfg.alias, file)
}

func backupDatabase(cfg dbConfig) {
	timestamp := time.Now().Format("20060102_150405")
	backupFile := fmt.Sprintf("dump/backup_%s_%s.sql", cfg.alias, timestamp)
	var cmd *exec.Cmd

	switch cfg.dbType {
	case "mysql":
		cmd = exec.Command("mysqldump", "-h", cfg.host, "-P", cfg.port, "-u", cfg.user, fmt.Sprintf("--password=%s", cfg.pass), cfg.dbName)
	case "postgres":
		cmd = exec.Command("pg_dump", "-h", cfg.host, "-p", cfg.port, "-U", cfg.user, "-d", cfg.dbName, "-F", "c")
		cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.pass)
	default:
		fmt.Printf("[%s] Erro: Tipo de banco de dados não suportado.\n", cfg.alias)
		return
	}

	outFile, err := os.Create(backupFile)
	if err != nil {
		fmt.Printf("[%s] Erro ao criar arquivo de backup: %v\n", cfg.alias, err)
		return
	}

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	outFile.Close()
	if err != nil {
		os.Remove(backupFile)
		fmt.Printf("[%s] Erro ao executar backup: %v\n", cfg.alias, err)
	} else {
		fmt.Printf("[%s] Backup realizado com sucesso: %s\n", cfg.alias, backupFile)
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
		cmd = exec.Command("pg_restore", "-h", cfg.host, "-p", cfg.port, "-U", cfg.user, "-d", cfg.dbName, "--clean", "--if-exists", "--no-owner", file)
		cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.pass)
	default:
		return fmt.Errorf("tipo de banco de dados não suportado: %s", cfg.dbType)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
