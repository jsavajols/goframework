package tables

import (
	"database/sql"
	"fmt"

	"github.com/jsavajols/goframework/functions/database"

	con "github.com/jsavajols/goframework/const"

	logs "github.com/jsavajols/goframework/functions/logs"

	"github.com/gofiber/fiber/v2/log"
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

	db, _ := t.Open()
	nbPoints := "("
	for i := 0; i < len(values); i++ {
		if i == len(values)-1 {
			nbPoints = nbPoints + "?) "
		} else {
			nbPoints = nbPoints + "?, "
		}
	}
	logs.Logs("INSERT INTO "+t.TableName+" "+fields+"  VALUES "+nbPoints, values)
	sqlResult, err := db.Exec("INSERT INTO "+t.TableName+" "+fields+"  VALUES "+nbPoints, values...)
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

func (t Table) Open() (*sql.DB, error) {
	logs.Logs("database " + t.Database + " " + t.TableName + " open.")
	return database.ConnectDatabase(t.Database)
}

func (t Table) Close(db *sql.DB) {
	logs.Logs(t.TableName + " close.")
	defer db.Close()
	defer db.Free()
}

func (t Table) Get(fields string, search string, sort string, start int, limit int) ReturnFunction {
	message := ""
	errorMessage := ""
	getRecords := 0
	limits := ""
	var statusCode int32
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
		limits = " limit " + fmt.Sprint(start) + ", " + fmt.Sprint(limit)
	}
	// Limite au nombre de lignes maximum defini dans const/const.go
	if limit > con.ROWS_LIMIT {
		errorMessage = "Limit too high"
		limit = con.ROWS_LIMIT
	}

	db, _ := t.Open()
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
	db, _ := t.Open()
	toUpdate := ""
	for i := 0; i < len(fields); i++ {
		if i == len(values)-1 {
			if types[i] == "string" {
				toUpdate = toUpdate + fields[i] + " = " + con.QUOTE + fmt.Sprintf("%v", values[i]) + con.QUOTE
			} else {
				toUpdate = toUpdate + fields[i] + " = " + fmt.Sprintf("%v", values[i])
			}
		} else {
			if types[i] == "string" {
				toUpdate = toUpdate + fields[i] + " = " + con.QUOTE + fmt.Sprintf("%v", values[i]) + con.QUOTE + ", "
			} else {
				toUpdate = toUpdate + fields[i] + " = " + fmt.Sprintf("%v", values[i]) + ", "
			}
		}
	}
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
	db, _ := t.Open()
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
