package internal

import (
	"encoding/json"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// Config  config struct
type Config struct {
	SourceDSN     string                       `toml:"source"`
	DestDSN       string                       `toml:"dest"`
	AlterIgnore   map[string]*AlterIgnoreTable `toml:"alter_ignore"`
	Tables        []string                     `toml:"tables"`
	TablesIGNORE  []string                     `toml:"tables_ignore"`
	OverwriteData overwriteData                `toml:"overwrite_data"`
	Email         *EmailStruct                 `toml:"email"`
	ConfigPath    string
	Sync          bool
	Drop          bool
}

func (cfg *Config) String() string {
	ds, _ := json.MarshalIndent(cfg, "  ", "  ")
	return string(ds)
}

// AlterIgnoreTable table's ignore info
type AlterIgnoreTable struct {
	Column     []string `toml:"column"`
	Index      []string `toml:"index"`
	ForeignKey []string `toml:"foreign"` //外键
}

// overwriteData overwrite tables data
type overwriteData struct {
	Tables []string `toml:"tables"`
}

// IsIgnoreField isIgnore
func (cfg *Config) IsIgnoreField(table string, name string) bool {
	for tname, dit := range cfg.AlterIgnore {
		if simpleMatch(tname, table, "IsIgnoreField_table") {
			for _, col := range dit.Column {
				if simpleMatch(col, name, "IsIgnoreField_colum") {
					return true
				}
			}
		}
	}
	return false
}

// CheckMatchTables check table is match
func (cfg *Config) CheckMatchTables(name string) bool {
	if len(cfg.Tables) == 0 {
		return true
	}
	for _, tableName := range cfg.Tables {
		if simpleMatch(tableName, name, "CheckMatchTables") {
			return true
		}
	}
	return false
}

// CheckMatchIgnoreTables check table_Ignore is match
func (cfg *Config) CheckMatchIgnoreTables(name string) bool {
	if len(cfg.TablesIGNORE) == 0 {
		return false
	}
	for _, tableName := range cfg.TablesIGNORE {
		if simpleMatch(tableName, name, "CheckMatchTables") {
			return true
		}
	}
	return false
}

// Check check config
func (cfg *Config) Check() {
	if cfg.SourceDSN == "" {
		log.Fatal("source dns is empty")
	}
	if cfg.DestDSN == "" {
		log.Fatal("dest dns is empty")
	}
	log.Println("config:\n", cfg)
}

// IsIgnoreIndex is index ignore
func (cfg *Config) IsIgnoreIndex(table string, name string) bool {
	for tname, dit := range cfg.AlterIgnore {
		if simpleMatch(tname, table, "IsIgnoreIndex_table") {
			for _, index := range dit.Index {
				if simpleMatch(index, name) {
					return true
				}
			}
		}
	}
	return false
}

// IsIgnoreForeignKey 检查外键是否忽略掉
func (cfg *Config) IsIgnoreForeignKey(table string, name string) bool {
	for tname, dit := range cfg.AlterIgnore {
		if simpleMatch(tname, table, "IsIgnoreForeignKey_table") {
			for _, foreignName := range dit.ForeignKey {
				if simpleMatch(foreignName, name) {
					return true
				}
			}
		}
	}
	return false
}

// SendMailFail send fail mail
func (cfg *Config) SendMailFail(errStr string) {
	if cfg.Email == nil {
		log.Println("email conf is empty,skip send mail")
		return
	}
	_host, _ := os.Hostname()
	title := "[mysql-schema-sync][" + _host + "]failed"
	body := "error:<font color=red>" + errStr + "</font><br/>"
	body += "host:" + _host + "<br/>"
	body += "config-file:" + cfg.ConfigPath + "<br/>"
	body += "dest_dsn:" + cfg.DestDSN + "<br/>"
	pwd, _ := os.Getwd()
	body += "pwd:" + pwd + "<br/>"
	cfg.Email.SendMail(title, body)
}

// LoadConfig load config file
func LoadConfig(confPath string) *Config {
	var cfg *Config
	_, err := toml.DecodeFile(confPath, &cfg)
	if err != nil {
		log.Fatalln("load json conf:", confPath, "failed:", err)
	}
	cfg.ConfigPath = confPath
	//	if *mailTo != "" {
	//		cfg.Email.To = *mailTo
	//	}
	return cfg
}
