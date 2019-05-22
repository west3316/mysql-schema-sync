package internal

import (
	"bytes"
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"text/template"
)

const tmpl = `INSERT INTO {{ .Name }} VALUES {{ .Values }}`

type table struct {
	Name   string
	Values string
}

func createTableValues(db *sql.DB, name string) (string, error) {
	// Get Data
	rows, err := db.Query("SELECT * FROM " + name)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	// Get columns
	columns, err := rows.Columns()
	if err != nil {
		return "", err
	}
	if len(columns) == 0 {
		return "", errors.New("No columns in table " + name + ".")
	}

	// Read data
	dataText := make([]string, 0)
	for rows.Next() {
		data := make([]*sql.NullString, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range data {
			ptrs[i] = &data[i]
		}

		// Read data
		if err := rows.Scan(ptrs...); err != nil {
			return "", err
		}

		dataStrings := make([]string, len(columns))

		for key, value := range data {
			if value != nil && value.Valid {
				dataStrings[key] = "'" + value.String + "'"
			} else {
				dataStrings[key] = "null"
			}
		}

		dataText = append(dataText, "("+strings.Join(dataStrings, ",")+")")
	}

	return strings.Join(dataText, ","), rows.Err()
}

var (
	dumpBuffer = make([]byte, 0, 4094)
)

// OverwriteData 重写表中数据
func OverwriteData(cfg *Config) {
	sc := NewSchemaSync(cfg)

	t, err := template.New("mysqldump").Parse(tmpl)
	if err != nil {
		log.Fatalf("mysqldump new template: %v", err)
	}

	buffer := bytes.NewBuffer(dumpBuffer)
	for _, tbl := range cfg.OverwriteData.Tables {
		sqlText, err := createTableValues(sc.SourceDb.Db, tbl)
		if err != nil {
			log.Printf("-- create table value error: %v \n", err)
			continue
		}

		log.Println("-- overwrite table: " + tbl)
		err = t.Execute(buffer, &table{Name: tbl, Values: sqlText})
		if err != nil {
			log.Fatalf("overwrite table: %s, err: %v", tbl, err)
		}

		data, _ := ioutil.ReadAll(buffer)
		result, err := sc.DestDb.Db.Exec(string(data))
		if err != nil {
			log.Fatalf("exec table: %s, err: %v", tbl, err)
		}

		inserted, _ := result.LastInsertId()
		log.Printf("table %s inserted %d\n", tbl, inserted)
	}

}
