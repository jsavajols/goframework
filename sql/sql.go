package sql

func DateDiff(unit, date1, date2, dialect string) string {
	if dialect == "mysql" {
		return "DATEDIFF('" + date1 + "','" + date2 + "')"
	} else if dialect == "postgres" {
		return "DATE_PART('" + unit + "','" + date1 + "') - DATE_PART('" + unit + "','" + date2 + "')"
	} else if dialect == "sqlite3" {
		return "JULIANDAY('" + date1 + "') - JULIANDAY('" + date2 + "')"
	} else {
		return ""
	}
}
