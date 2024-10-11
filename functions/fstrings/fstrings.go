package fstrings

import (
	"strconv"
	"strings"
	"time"

	logs "github.com/jsavajols/goframework/functions/logs"
)

// Vérifie si l'élément passé en paramètre est nil
// Retourne une chaine vide si c'est le cas
func NilString(s interface{}) string {
	if s == nil {
		return ""
	}
	return s.(string)
}

func UpperString(s interface{}) string {
	if s == nil {
		return ""
	}
	return strings.ToUpper(s.(string))
}

func LowerString(s interface{}) string {
	if s == nil {
		return ""
	}
	return strings.ToLower(s.(string))
}

func NilInt(s interface{}) string {
	result, err := strconv.Atoi(s.(string))
	if err != nil {
		return ""
	} else {
		return strconv.Itoa(result)
	}
}

func ToString(s interface{}) string {
	switch v := s.(type) {
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return v.Format("2006-01-02")
	case string:
		return v
	default:
		logs.Logs(v)
		return ""
	}
}

func ToInt(s interface{}) int {
	switch v := s.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		result, err := strconv.Atoi(v)
		if err != nil {
			return 0
		} else {
			return result
		}
	default:
		return 0
	}
}

func ToFloat(s interface{}) float64 {
	switch v := s.(type) {
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		result, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		} else {
			return result
		}
	default:
		return 0
	}
}

// Purge la chaine de caractères des caractères
// des doubles espaces, des tabulations et des retours à la ligne
func PurgeElementText(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.ReplaceAll(s, "  ", "")
	return s
}

// Vérifie si un element existe déjà dans un tableau de strings
func ElementExistsInArray(s string, array []string) bool {
	for _, v := range array {
		if v == s {
			return true
		}
	}
	return false
}

// Transforme un tableau de chaines en string avec un séparateur , par défaut
func ArrayToString(array []string, separator string) string {
	if separator == "" {
		separator = ","
	}
	return strings.Join(array, separator)
}
