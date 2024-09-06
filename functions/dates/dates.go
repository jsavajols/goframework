package dates

import "time"

func GetParisTime() time.Time {
	// Charger le fuseau horaire Europe/Paris
	location, _ := time.LoadLocation("Europe/Paris")

	// Obtenir l'heure actuelle dans la timezone du serveur
	now := time.Now()

	// Convertir l'heure actuelle dans le fuseau horaire de Paris
	parisTime := now.In(location)

	return parisTime
}

func LastDayOfMonth(year int, month int) int {
	// Convertir l'entier du mois en time.Month
	tMonth := time.Month(month)

	// Passer au premier jour du mois suivant
	firstDayNextMonth := time.Date(year, tMonth+1, 1, 0, 0, 0, 0, time.UTC)

	// Reculer d'un jour pour obtenir le dernier jour du mois actuel
	lastDay := firstDayNextMonth.AddDate(0, 0, -1)

	// Retourner seulement le jour sous forme d'entier
	return lastDay.Day()
}

func FirstAndLastDayOfWeek(year int, month int, week int) (int, int) {
	// Définir le premier jour du mois donné
	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	// Obtenir le jour de la semaine pour le premier jour du mois
	weekdayOfFirst := int(firstDayOfMonth.Weekday())
	if weekdayOfFirst == 0 { // En Go, Sunday est représenté par 0, donc on le remplace par 7
		weekdayOfFirst = 7
	}

	// Calculer le premier jour de la semaine donnée
	startOfWeek := firstDayOfMonth.AddDate(0, 0, (week-1)*7-(weekdayOfFirst-1))

	// Calculer le dernier jour de la semaine donnée
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	// Vérifier si la semaine dépasse le mois donné
	if endOfWeek.Month() != time.Month(month) {
		endOfWeek = time.Date(year, time.Month(month), LastDayOfMonth(year, month), 0, 0, 0, 0, time.UTC)
	}

	return startOfWeek.Day(), endOfWeek.Day()
}

func WeekNumberFromDate(dateStr string) (int, error) {
	// Convertir la chaîne de caractères en type time.Time
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, err
	}

	// Obtenir le numéro de la semaine et l'année correspondante
	_, week := date.ISOWeek()

	return week, nil
}

func ConvertDateToFrench(dateStr, withHour string) (string, error) {
	// Analyser la chaîne de caractères au format "yyyy-mm-dd hh:mm:ss"
	if len(dateStr) == 10 {
		dateStr += " 00:00:00"
	}
	format := "2006-01-02 15:04:05"
	date, err := time.Parse(format, dateStr)
	if err != nil {
		return "", err
	}

	// Reformater la date au format "dd/mm/yyyy"
	formattedDate := date.Format("02/01/2006")
	if withHour == "h" {
		formattedDate = date.Format("02/01/2006 15:04:05")
	}

	return formattedDate, nil
}
