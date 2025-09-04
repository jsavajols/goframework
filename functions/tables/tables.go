package tables

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jsavajols/goframework/functions/database"
	"github.com/jsavajols/goframework/functions/fstrings"

	con "github.com/jsavajols/goframework/const"

	sqlFunctions "github.com/jsavajols/goframework/functions/sql"

	"github.com/gofiber/fiber/v2/log"
	logs "github.com/jsavajols/goframework/functions/logs"
)

// Validator interface pour la validation
type Validator interface {
	ValidateRecord() error
	BeforeInsert() error
	AfterInsert() error
	BeforeUpdate(values []interface{}) error
	AfterUpdate() bool
	BeforeDelete() error
	AfterDelete() bool
}

// Méthode par défaut de validation
type DefaultValidator struct{}

// Structure de base Table
type Table struct {
	Dialect          string
	Database         string
	TableName        string
	Source           string
	Struct           interface{}
	Db               *sql.DB
	Validator        Validator
	DefaultValidator DefaultValidator
	ReadOnly         bool
}

type ReturnFunction struct {
	StatusCode    int32  `json:"statusCode"`
	Message       string `json:"message"`
	ErrorMessage  string `json:"errorMessage"`
	GetRecords    int    `json:"getRecords"`
	Rows          interface{}
	InsertRecords int64 `json:"insertRecords"`
	UpdateRecords int64 `json:"updateRecords"`
	DeleteRecords int64 `json:"deleteRecords"`
	LastInsertId  int64 `json:"lastInsertId"`
}

func (dv DefaultValidator) ValidateRecord() error {
	logs.Logs("Validation de l'enregistrement (défaut)")
	return nil
}

// BeforeInsert méthode pour Table
func (dv DefaultValidator) BeforeInsert() error {
	logs.Logs("Préparation avant insertion (défaut)")
	// Appel de ValidateRecord et gestion de l'erreur
	err := dv.ValidateRecord()
	if err != nil {
		logs.Logs("Échec de la validation:", err)
		return err
	}

	return nil
}

// ExecSql exécute une requête SQL et retourne le résultat ou une erreur
func ExecSql(dbName, dialect, sql string) (sql.Result, error) {
	db, _, _ := database.ConnectDatabase(dbName, dialect)
	if db == nil {
		return nil, fmt.Errorf("Erreur de connexion à la base de données")
	}
	defer db.Close()
	logs.Logs("Exécution de la requête SQL:", sql)
	result, err := db.Exec(sql)
	if err != nil {
		log.Error("Erreur lors de l'exécution de la requête SQL:", err)
		return nil, err
	}
	return result, nil
}

// Insert méthode pour Table
func (t *Table) Insert(fields string, values []interface{}) ReturnFunction {
	errorMessage := ""
	var statusCode int32
	var message string
	var rowsAffected int64
	var lastInsertId int64
	logs.Logs("Début de l'insertion dans", t.TableName)
	// Si la table est en lecture seule, on renvoie une erreur
	if t.ReadOnly {
		return ReturnFunction{
			StatusCode:    500,
			Message:       "Insert error",
			ErrorMessage:  "Table is read only",
			InsertRecords: 0,
			LastInsertId:  0,
		}
	}

	// Appel de BeforeInsert et gestion de l'erreur
	err := t.Validator.BeforeInsert()
	if err != nil {
		logs.Logs("Échec lors de la préparation avant insertion:", err)
		return ReturnFunction{
			StatusCode:    500,
			Message:       "Insert error",
			ErrorMessage:  err.Error(),
			InsertRecords: 0,
			LastInsertId:  0,
		}

	}

	// Ici, insérez la logique d'insertion réelle si BeforeInsert réussit

	var sqlResult sql.Result
	db, dialect, _ := t.Open()
	t.Dialect = dialect
	nbPoints := ""
	if t.Dialect == "postgres" {
		nbPoints = "("
		for i := 0; i < len(values); i++ {
			if i == len(values)-1 {
				nbPoints = nbPoints + "$" + fstrings.ToString(i+1) + ") "
			} else {
				nbPoints = nbPoints + "$" + fstrings.ToString(i+1) + ", "
			}
		}
	} else {
		nbPoints = "("
		for i := 0; i < len(values); i++ {
			if i == len(values)-1 {
				nbPoints = nbPoints + "?) "
			} else {
				nbPoints = nbPoints + "?, "
			}
		}
	}
	logs.Logs("INSERT INTO "+t.TableName+" "+fields+"  VALUES "+nbPoints, values)
	sqlResult, err = db.Exec("INSERT INTO "+t.TableName+" "+fields+"  VALUES "+nbPoints, values...)
	if err != nil {
		errorMessage = err.Error()
		log.Error(errorMessage)
	}

	if errorMessage == "" {
		statusCode = 200
		message = "Insert success"
		rowsAffected, _ = sqlResult.RowsAffected()
		lastInsertId, _ = sqlResult.LastInsertId()
	} else {
		statusCode = 500
		message = "Insert error"
		rowsAffected = 0
		lastInsertId = 0
	}

	t.Validator.AfterInsert()
	t.Close(db)
	returnFunction := ReturnFunction{
		StatusCode:    statusCode,
		Message:       message,
		ErrorMessage:  errorMessage,
		InsertRecords: rowsAffected,
		LastInsertId:  lastInsertId,
	}

	logs.Logs("Insertion réussie dans", t.TableName)

	return returnFunction
}

func (t Table) Open() (*sql.DB, string, error) {
	logs.Logs("database " + t.Database + " " + t.TableName + " open.")
	return database.ConnectDatabase(t.Database, t.Dialect)
}

func (t Table) Close(db *sql.DB) {
	logs.Logs(t.TableName + " close.")
	defer db.Close()
}

func (t Table) Get(fields string, search string, sort string, start int, limit int) ReturnFunction {
	message := ""
	errorMessage := ""
	getRecords := 0
	limits := ""
	var statusCode int32

	db, dialect, _ := t.Open()
	t.Dialect = dialect

	if fields == "" {
		fields = "*"
	}
	if search == "" {
		search = "true"
	}
	if sort != "" {
		sort = " order by " + sort
	}
	// Gère start et limit
	if start != 0 || limit != 0 {
		if t.Dialect == "postgres" {
			limits = " offset " + fmt.Sprint(start) + " limit " + fmt.Sprint(limit)
		} else {
			limits = " limit " + fmt.Sprint(start) + ", " + fmt.Sprint(limit)
		}
	}
	// Limite au nombre de lignes maximum defini dans const/const.go
	if limit > con.ROWS_LIMIT {
		errorMessage = "Limit too high"
		limit = con.ROWS_LIMIT
	}

	sql := t.buildQuery(fields, search, sort, limits)
	logs.Logs(sql)
	stmt, err := db.Prepare(sql)
	var tableData []map[string]interface{}
	if err != nil {
		statusCode = 500
		message = "Get error"
		errorMessage = err.Error()
		// Retourne un tableau vide
		tableData = make([]map[string]interface{}, 0)
		getRecords = 0

	} else {
		statusCode = 200
		message = "Get success"
		rows, _ := stmt.Query()
		tableData = t.fetchData(rows)
		getRecords = len(tableData)
		defer rows.Close()
		defer stmt.Close()
	}

	t.Close(db)
	returnFunction := ReturnFunction{
		StatusCode:   statusCode,
		Message:      message,
		ErrorMessage: errorMessage,
		GetRecords:   getRecords,
		Rows:         tableData,
	}
	return returnFunction
}

func (t Table) buildQuery(fields string, search string, sort string, limits string) string {
	toReturn := ""
	if search != "-" {
		search = " where " + search
	} else {
		search = ""
	}
	if t.Source != "" {
		toReturn = t.Source + search + sort + limits
	} else {
		toReturn = "select " + fields + " from " + t.TableName + search + sort + limits
	}
	if t.Dialect != "mysql" {
		toReturn = strings.ReplaceAll(toReturn, "RAND", "RANDOM")
		toReturn = strings.ReplaceAll(toReturn, "ucase", "upper")
	}
	if t.Dialect == "postgres" {
		toReturn = strings.ReplaceAll(toReturn, `""`, `**++--`)
		toReturn = strings.ReplaceAll(toReturn, `"`, `'`)
		toReturn = strings.ReplaceAll(toReturn, `**++--`, `"`)
		// ajoute des " pour les noms de colonnes
		fieldsList := strings.Split(fields, ",")
		for i, field := range fieldsList {
			fieldsList[i] = "\"" + strings.TrimSpace(field) + "\""
		}
		fields = strings.Join(fieldsList, ",")
		// Si la source est définie, on recherche les noms de colonnes dans la source pour y ajouter des " si besoin
		if t.Source != "" {
			fieldsList := strings.Split(t.Source, "as")
			for i, field := range fieldsList {
				startField := strings.Index(field, " ")
				endField := strings.LastIndex(field, " ")
				if endField == -1 {
					endField = strings.LastIndex(field, ",")
				}
				fieldsList[i] = "\"" + field[startField:endField] + "\""
			}
			fields = strings.Join(fieldsList, ",")
		}
	}
	if sqlFunctions.CheckForSQLInjection(toReturn) {
		return ""
	}
	return toReturn
}

func (t Table) fetchData(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()

	tableData := make([]map[string]interface{}, 0)

	colCount := len(columns)
	values := make([]interface{}, colCount)
	scanArgs := make([]interface{}, colCount)
	for i := range values {
		scanArgs[i] = &values[i]
	}
	records := 0
	for rows.Next() {
		records++
		err := rows.Scan(scanArgs...)
		if err != nil {
			log.Error(err)
		}
		entry := make(map[string]interface{})
		for i, col := range columns {
			v := values[i]

			b, ok := v.([]byte)
			if ok {
				entry[col] = string(b)
			} else {
				entry[col] = v
			}
		}
		tableData = append(tableData, entry)
	}
	return tableData
}

func (dv DefaultValidator) AfterInsert() error {
	logs.Logs("after insert (défaut)")
	return nil
}

func (dv DefaultValidator) BeforeUpdate(values []interface{}) error {
	logs.Logs("before update (défaut)")
	// Appel de ValidateRecord et gestion de l'erreur
	err := dv.ValidateRecord()
	if err != nil {
		logs.Logs("Échec de la validation:", err)
		return err
	}
	return nil
}

func (dv DefaultValidator) AfterUpdate() bool {
	logs.Logs("after update (défaut)")
	return true
}

func (t Table) Update(fields []string, values []interface{}, types []interface{}, filter string) ReturnFunction {
	errorMessage := ""
	var statusCode int32
	var message string
	var rowsAffected int64
	logs.Logs("Début de la mise à jour dans", t.TableName)

	// Si la table est en lecture seule, on renvoie une erreur
	if t.ReadOnly {
		return ReturnFunction{
			StatusCode:    500,
			Message:       "Update error",
			ErrorMessage:  "Table is read only",
			InsertRecords: 0,
			LastInsertId:  0,
		}
	}

	// Appel de BeforeUpdate et gestion de l'erreur
	err := t.Validator.BeforeUpdate(values)
	if err != nil {
		logs.Logs("Échec lors de la préparation avant mise à jour:", err)
		return ReturnFunction{
			StatusCode:    500,
			Message:       "Update error",
			ErrorMessage:  err.Error(),
			InsertRecords: 0,
			LastInsertId:  0,
		}
	}

	// Ici, insérez la logique de mise à jour réelle si BeforeUpdate réussit
	db, dialect, _ := t.Open()
	t.Dialect = dialect
	toUpdate := ""
	quote := con.QUOTE
	if t.Dialect == "postgres" {
		quote = "'"
	}
	for i := 0; i < len(fields); i++ {
		if i == len(values)-1 {
			if types[i] == "string" || types[i] == "date" || types[i] == "datetime" {
				toUpdate = toUpdate + fields[i] + " = " + quote + fmt.Sprintf("%v", values[i]) + quote
			} else {
				toUpdate = toUpdate + fields[i] + " = " + fmt.Sprintf("%v", values[i])
			}
		} else {
			if types[i] == "string" || types[i] == "date" || types[i] == "datetime" {
				toUpdate = toUpdate + fields[i] + " = " + quote + fmt.Sprintf("%v", values[i]) + quote + ", "
			} else {
				toUpdate = toUpdate + fields[i] + " = " + fmt.Sprintf("%v", values[i]) + ", "
			}
		}
	}

	// Corrige les valeurs null pour enlever les quotes
	toUpdate = strings.ReplaceAll(toUpdate, quote+"null"+quote, "null")

	if filter != "" {
		filter = " where " + filter
	} else {
		filter = " where true"
	}

	logs.Logs("UPDATE " + t.TableName + " set " + toUpdate + filter)

	sqlResult, err := db.Exec("UPDATE " + t.TableName + " set " + toUpdate + filter)
	if err != nil {
		errorMessage = err.Error()
		rowsAffected = 0
		log.Error(errorMessage)
	} else {
		rowsAffected, _ = sqlResult.RowsAffected()
	}

	if errorMessage == "" {
		statusCode = 200
		message = "Update success"
	} else {
		statusCode = 500
		rowsAffected = 0
		message = "Update error"
	}

	t.Validator.AfterUpdate()
	t.Close(db)
	returnFunction := ReturnFunction{
		StatusCode:    statusCode,
		Message:       message,
		ErrorMessage:  errorMessage,
		InsertRecords: 0,
		UpdateRecords: rowsAffected,
		LastInsertId:  0,
	}

	logs.Logs("Mise à jour réussie dans", t.TableName)

	return returnFunction
}

func (dv DefaultValidator) BeforeDelete() error {
	logs.Logs("before delete (défaut)")
	return nil
}

func (t Table) Delete(search string) ReturnFunction {
	errorMessage := ""
	var statusCode int32
	// Si la table est en lecture seule, on renvoie une erreur
	if t.ReadOnly {
		return ReturnFunction{
			StatusCode:    500,
			Message:       "Delete error",
			ErrorMessage:  "Table is read only",
			InsertRecords: 0,
			LastInsertId:  0,
		}
	}

	if search == "" {
		search = "true"
	}
	db, dialect, _ := t.Open()
	t.Dialect = dialect
	err := t.Validator.BeforeDelete()
	if err != nil {
		logs.Logs("Échec lors de la préparation avant suppression:", err)
		return ReturnFunction{
			StatusCode:    500,
			Message:       "Delete error",
			ErrorMessage:  err.Error(),
			InsertRecords: 0,
			LastInsertId:  0,
		}
	}

	sql := "DELETE from " + t.TableName + " where " + search
	sqlResult, err := db.Exec(sql)
	var rowsAffected int64
	if err != nil {
		errorMessage = err.Error()
	} else {
		rowsAffected, _ = sqlResult.RowsAffected()
	}
	t.Validator.AfterDelete()
	t.Close(db)
	var message string
	if errorMessage == "" {
		statusCode = 200
		message = "Delete success"
	} else {
		statusCode = 500
		message = "Delete error"
	}
	// Si rowsAffected = 0, il y a eu une erreur
	if rowsAffected == 0 {
		statusCode = 500
		message = "0 rows affected"
	}
	returnFunction := ReturnFunction{
		StatusCode:    statusCode,
		Message:       message,
		ErrorMessage:  errorMessage,
		DeleteRecords: rowsAffected,
	}
	return returnFunction
}

func (dv DefaultValidator) AfterDelete() bool {
	logs.Logs("after delete (défaut)")
	return true
}
