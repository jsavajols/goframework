package mails

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

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

func SendMail(vars map[string]string) error {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	to := []string{vars["email"]}
	templateToApply := ""

	templateToApply = GetMailTemplate(vars["template"])
	for key, value := range vars {
		templateToApply = strings.Replace(templateToApply, "{{"+key+"}}", value, -1)
	}

	// Formatage de l'en-tête "Date"
	date := time.Now().Format(time.RFC1123Z)
	dateHeader := fmt.Sprintf("Date: %s\r\n", date)

	// Création d'un Message-ID unique
	messageID := fmt.Sprintf("<%d.%d@"+os.Getenv("DOMAIN")+"", time.Now().UnixNano(), 12345)
	messageIDHeader := fmt.Sprintf("Message-ID: %s\r\n", messageID)

	msg := "From: " + vars["sender"] + " <" + from + ">\n" +
		"To: " + vars["email"] + "\n" +
		"Subject: " + vars["subject"] + "\n" +
		dateHeader +
		messageIDHeader +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		templateToApply

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"),
		smtp.PlainAuth("", from, pass, os.Getenv("SMTP_HOST")),
		from, to, []byte(msg))
	if err != nil {
		logs.Logs("smtp error: ", err)
	}
	return err
}

// ChkMail verifie la cohérence d'une adresse mail
func ChkMail(pMail string, pChkValid string) (float32, string) {
	var errorCode float32 = 0
	message := ""
	if errorCode == 0 && !strings.Contains(pMail, ".") {
		errorCode = -1
	}
	if errorCode == 0 && !strings.Contains(pMail, "@") {
		errorCode = -2
	}
	if errorCode == 0 && strings.TrimSpace(pMail) == "" {
		errorCode = -98
	}
	if errorCode == 0 && strings.LastIndex(pMail, "@") > strings.LastIndex(pMail, ".") {
		errorCode = -3
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if errorCode == 0 && !emailRegex.MatchString(pMail) {
		errorCode = -4
	}
	if len(pMail) < 3 && len(pMail) > 254 {
		errorCode = -5
	}
	if errorCode == 0 {
		parts := strings.Split(pMail, "@")
		mx, err := net.LookupMX(parts[1])
		if err != nil || len(mx) == 0 {
			errorCode = -6
		}
	}

	if errorCode == 0 && pChkValid == "Y" {
		err := ValidateHostAndUser("ssl0.ovh.net", "contact@happyapi.fr", pMail)
		if smtpErr, ok := err.(SmtpError); ok && err != nil {
			fmt.Printf("Code: %s, Msg: %s", smtpErr.Code(), smtpErr)
		}
		if err != nil {
			if strings.Contains(strings.ToLower(err.(SmtpError).Error()), "blocked") {
				errorCode = -8
			} else {
				errorCode = -7
			}
		}
	}

	return errorCode, message
}

type SmtpError struct {
	Err error
}

func (e SmtpError) Error() string {
	return e.Err.Error()
}

func (e SmtpError) Code() string {
	return e.Err.Error()[0:3]
}

func NewSmtpError(err error) SmtpError {
	return SmtpError{
		Err: err,
	}
}

const forceDisconnectAfter = time.Second * 5

var (
	ErrBadFormat        = errors.New("invalid format")
	ErrUnresolvableHost = errors.New("unresolvable host")
)

func ValidateHostAndUser(serverHostName, serverMailAddress, email string) error {
	_, host := split(email)
	mx, err := net.LookupMX(host)
	if err != nil {
		return ErrUnresolvableHost
	}
	client, err := DialTimeout(fmt.Sprintf("%s:%d", mx[0].Host, 25), forceDisconnectAfter)
	if err != nil {
		return NewSmtpError(err)
	}
	defer client.Close()

	err = client.Hello(serverHostName)
	if err != nil {
		return NewSmtpError(err)
	}
	err = client.Mail(serverMailAddress)
	if err != nil {
		return NewSmtpError(err)
	}
	err = client.Rcpt(email)
	if err != nil {
		return NewSmtpError(err)
	}
	return nil
}

func DialTimeout(addr string, timeout time.Duration) (*smtp.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	t := time.AfterFunc(timeout, func() { conn.Close() })
	defer t.Stop()

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

func split(email string) (account, host string) {
	i := strings.LastIndexByte(email, '@')
	// If no @ present, not a valid email.
	if i < 0 {
		return
	}
	account = email[:i]
	host = email[i+1:]
	return
}
