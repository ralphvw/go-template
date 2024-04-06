package helpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/ralphvw/go-template/models"
)

const (
	SERVER_ERROR = "Server error"
)

func LogAction(message string) {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log page")
	}

	defer file.Close()

	customFormat := "Mon 02 Jan, 2006 @ 15:04"
	logger := log.New(file, time.Now().Format(customFormat)+" ", 0)
	logger.Printf(message)
	fmt.Println(message)
}

func GetDataHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, pageSize int, page int, query string, countQuery string, message string, args []interface{}, keys []string, destinations ...interface{}) {
	EnableCors(w)
	offset := (page - 1) * pageSize

	queryString := fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pageSize, offset)

	rows, err := db.Query(queryString, args...)

	if err != nil {
		LogAction(err.Error())
		http.Error(w, SERVER_ERROR, http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var results []map[string]interface{}

	for rows.Next() {
		if err := rows.Scan(destinations...); err != nil {
			LogAction(err.Error())
			http.Error(w, SERVER_ERROR, http.StatusInternalServerError)
			return
		}
		result := make(map[string]interface{})

		for i := 0; i < len(destinations); i += 1 {
			key := keys[i]
			value := reflect.Indirect(reflect.ValueOf(destinations[i])).Interface()
			result[key] = value
		}

		results = append(results, result)

	}

	totalRows := 0
	err = db.QueryRow(countQuery, args...).Scan(&totalRows)
	if err != nil {
		LogAction(err.Error())
		http.Error(w, SERVER_ERROR, http.StatusInternalServerError)
		return
	}

	totalPages := (totalRows + pageSize - 1) / pageSize

	response := models.Response{
		Message:    message,
		Data:       results,
		TotalPages: totalPages,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		LogAction(err.Error())
		http.Error(w, SERVER_ERROR, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

func SendResponse(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	response := models.SingleResponse{
		Message: message,
		Data:    data,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		LogAction(err.Error())
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func CheckFields(obj interface{}, fields []string) (bool, []string) {
	objValue := reflect.ValueOf(obj)
	var missingFields []string

	for _, field := range fields {
		fieldValue := objValue.FieldByName(field)

		if !fieldValue.IsValid() {
			missingFields = append(missingFields, field)
		} else {
			zeroValue := reflect.Zero(fieldValue.Type())
			if reflect.DeepEqual(fieldValue.Interface(), zeroValue.Interface()) {
				missingFields = append(missingFields, field)
			}
		}
	}

	fieldsExist := len(missingFields) == 0

	return fieldsExist, missingFields
}

func EnableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func HandleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	w.WriteHeader(http.StatusOK)
}

func GetPointerValue(pointer interface{}) any {
	return reflect.Indirect(reflect.ValueOf(pointer)).Interface()
}
