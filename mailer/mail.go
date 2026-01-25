package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	apimail "github.com/ainsleyclark/go-mail"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain                string
	Templates             string
	Host                  string
	Port                  int
	Username, Password    string
	Encryption            string
	FromAddress, FromName string
	Jobs                  chan Message
	Results               chan Result
	API, APIKey, APIUrl   string
}

type Message struct {
	To, Subject    string
	From, FromName string
	Template       string
	Attachments    []string
	Data           any
}

type Result struct {
	Success bool
	Error   error
}

// ListenForMail listens for the mail channel and sends mail
// when it receives a payload. It runs continually in the background,
// and sends error/success msg back on the results channel.
// Note that if api and api key are set, it will prefer using
// an api to send mail, otherwise it will use smtp.
func (m *Mail) ListenForMail() {
	log.Println(">> ListenForMail")
	fmt.Println(">> ListenForMail")
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		if err != nil {
			m.Results <- Result{Success: false, Error: err}
			continue
		}
		m.Results <- Result{Success: true, Error: nil}
	}
}

func (m *Mail) Send(msg Message) error {
	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.API != "smtp" {
		log.Println(">> SendUsingAPI:", m.API)
		fmt.Println(">> SendUsingAPI:", m.API)
		return m.ChooseAPI(msg)
	}
	log.Println(">> SendSMTPMessage")
	fmt.Println(">> SendSMTPMessage")
	return m.SendSMTPMessage(msg)
}

func (m *Mail) ChooseAPI(msg Message) error {
	switch strings.ToLower(m.API) {
	case "mailgun", "sparkpost", "sendgrid", "postmark":
		log.Println(">> SendUsingAPI:", m.API)
		fmt.Println(">> SendUsingAPI:", m.API)
		return m.SendUsingAPI(msg, m.API)
	default:
		return fmt.Errorf("unknown mail api: %s", m.API)
	}
}

func (m *Mail) SendUsingAPI(msg Message, transport string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	cfg := apimail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	driver, err := apimail.NewClient(transport, cfg)
	if err != nil {
		return err
	}

	formattedMsg, plainTextMsg, err := m.formatMsg(msg)
	if err != nil {
		return err
	}

	tx := &apimail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       formattedMsg,
		PlainText:  plainTextMsg,
	}

	// add attachement
	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	_, err = driver.Send(tx)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) addAPIAttachments(msg Message, tx *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachement []apimail.Attachment
		for _, attachment := range msg.Attachments {
			var attach apimail.Attachment
			content, err := os.ReadFile(attachment)
			if err != nil {
				return err
			}
			fileName := filepath.Base(attachment)
			attach.Bytes = content
			attach.Filename = fileName
			attachement = append(attachement, attach)
		}
		tx.Attachments = attachement
	}

	return nil
}

// SendSMTPMessage builds and sends an email using SMTP.
// This is called by ListenForMail, and also can be called
// directly when necessary.
func (m *Mail) SendSMTPMessage(msg Message) error {

	formattedMsg, plainTextMsg, err := m.formatMsg(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextHTML, formattedMsg).
		AddAlternative(mail.TextPlain, plainTextMsg)

	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			email.AddAttachment(attachment)
		}
	}

	return email.Send(smtpClient)
}

func (m *Mail) formatMsg(msg Message) (string, string, error) {
	formattedMsg, err := m.buildHTMLMessage(msg)
	if err != nil {
		return "", "", err
	}

	plainTextMsg, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return "", "", err
	}

	return formattedMsg, plainTextMsg, nil
}

func (m *Mail) getEncryption(encryption string) mail.Encryption {
	switch strings.ToLower(encryption) {
	case "ssl":
		return mail.EncryptionSSL
	case "tls":
		return mail.EncryptionSTARTTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {

	templateToRender := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMsg := tpl.String()
	formattedMsg, err = m.inlineCSS(formattedMsg)
	if err != nil {
		return "", err
	}

	return formattedMsg, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {

	templateToRender := fmt.Sprintf("%s/%s.plain.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMsg := tpl.String()

	return formattedMsg, nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	option := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &option)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
