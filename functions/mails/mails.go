package mails

import (
	"log"
	"net/smtp"
	"os"
	"strings"

	logs "github.com/jsavajols/goframework/functions/logs"
)

func GetMailTemplate(filename string) string {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	fileContentString := string(fileContent)
	return fileContentString
}

/*
// SendMailSMTP envoi un mail via SMTP
func SendMailSMTP(email, template string, vars map[string]string) string {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	to := []string{email}

	template = strings.ToUpper(template)
	templateToApply := ""
	subject := ""
	switch template {
	case "NEW-CONTACT":
		templateToApply = GetMailTemplate("shared/public/templates/mails/new-contact.html")
		templateToApply = strings.Replace(templateToApply, "{{firstname}}", vars["firstname"], -1)
		templateToApply = strings.Replace(templateToApply, "{{lastname}}", vars["lastname"], -1)
		subject = "Votre intérêt pour le podcast Traces"
	}
	msg := "From: podcast Traces <" + from + ">\n" +
		"To: " + email + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		templateToApply

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"),
		smtp.PlainAuth("", from, pass, os.Getenv("SMTP_HOST")),
		from, to, []byte(msg))

	if err != nil {
		logs.Logs("smtp error: ", err)
	}
	return "ok"
}
*/

func SendMail(vars map[string]string) string {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	to := []string{vars["email"]}
	templateToApply := ""

	templateToApply = GetMailTemplate(vars["template"])
	for key, value := range vars {
		templateToApply = strings.Replace(templateToApply, "{{"+key+"}}", value, -1)
	}

	msg := "From: " + vars["sender"] + " <" + from + ">\n" +
		"To: " + vars["email"] + "\n" +
		"Subject: " + vars["subject"] + "\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		templateToApply

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"),
		smtp.PlainAuth("", from, pass, os.Getenv("SMTP_HOST")),
		from, to, []byte(msg))

	if err != nil {
		logs.Logs("smtp error: ", err)
	}
	return "ok"
}
