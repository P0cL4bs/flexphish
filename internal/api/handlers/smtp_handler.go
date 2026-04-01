package handlers

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"flexphish/internal/domain/smtp"
	"fmt"
	"net"
	"net/http"
	netsmtp "net/smtp"
	"strings"
	"time"
)

type SMTPHandler struct {
	repo smtp.Repository
}

func NewSMTPHandler(repo smtp.Repository) *SMTPHandler {
	return &SMTPHandler{repo: repo}
}

func (h *SMTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	var input struct {
		Name      string `json:"name"`
		IsGlobal  bool   `json:"is_global"`
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromName  string `json:"from_name"`
		FromEmail string `json:"from_email"`
		IsActive  bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Host = strings.TrimSpace(input.Host)
	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)
	input.FromName = strings.TrimSpace(input.FromName)
	input.FromEmail = strings.TrimSpace(input.FromEmail)

	if input.Name == "" || input.Host == "" || input.Port <= 0 || input.Username == "" || input.Password == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "missing required fields",
		})
		return
	}

	connectionExists, err := h.repo.ExistsByConnection(input.Host, input.Port, input.Username, userID, input.IsGlobal, nil)
	if err != nil {
		http.Error(w, "error validating smtp profile", http.StatusInternalServerError)
		return
	}
	if connectionExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": smtp.ErrConnectionAlreadyExists.Error(),
		})
		return
	}

	profile := smtp.SMTPProfile{
		Name:      input.Name,
		IsGlobal:  input.IsGlobal,
		Host:      input.Host,
		Port:      input.Port,
		Username:  input.Username,
		Password:  input.Password,
		FromName:  input.FromName,
		FromEmail: input.FromEmail,
		IsActive:  input.IsActive,
	}

	if !profile.IsGlobal {
		profile.UserId = &userID
	}

	if err := h.repo.Create(&profile); err != nil {
		http.Error(w, "error creating smtp profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

func (h *SMTPHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string `json:"name"`
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromName  string `json:"from_name"`
		FromEmail string `json:"from_email"`
		TestEmail string `json:"test_email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Host = strings.TrimSpace(input.Host)
	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)
	input.FromName = strings.TrimSpace(input.FromName)
	input.FromEmail = strings.TrimSpace(input.FromEmail)
	input.TestEmail = strings.TrimSpace(input.TestEmail)

	if input.Host == "" || input.Port <= 0 || input.Username == "" || input.Password == "" || input.TestEmail == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "host, port, username, password and test_email are required",
		})
		return
	}

	fromEmail := input.FromEmail
	if fromEmail == "" {
		fromEmail = input.Username
	}

	fromHeader := fromEmail
	if input.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", input.FromName, fromEmail)
	}

	subject := "SMTP Test Email - Flexphish"
	body := fmt.Sprintf(
		"This is a test email sent at %s.\n\nIf you received this message, your SMTP settings are working.",
		time.Now().Format(time.RFC3339),
	)

	msg := []byte(
		"From: " + fromHeader + "\r\n" +
			"To: " + input.TestEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	if err := sendSMTPTestMessage(input.Host, input.Port, input.Username, input.Password, fromEmail, input.TestEmail, msg); err != nil {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "smtp test failed: " + err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, map[string]string{
		"message": "test email sent successfully",
	})
}

func (h *SMTPHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int64)

	profiles, err := h.repo.GetAll(userID)
	if err != nil {
		http.Error(w, "error fetching smtp profiles", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profiles)
}

func (h *SMTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	profile := r.Context().Value("smtpProfile").(*smtp.SMTPProfile)
	json.NewEncoder(w).Encode(profile)
}

func (h *SMTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	existing := r.Context().Value("smtpProfile").(*smtp.SMTPProfile)
	userID := r.Context().Value("userID").(int64)

	var input struct {
		Name      string `json:"name"`
		IsGlobal  bool   `json:"is_global"`
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromName  string `json:"from_name"`
		FromEmail string `json:"from_email"`
		IsActive  bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Host = strings.TrimSpace(input.Host)
	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)
	input.FromName = strings.TrimSpace(input.FromName)
	input.FromEmail = strings.TrimSpace(input.FromEmail)

	if input.Name == "" || input.Host == "" || input.Port <= 0 || input.Username == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "missing required fields",
		})
		return
	}

	connectionExists, err := h.repo.ExistsByConnection(input.Host, input.Port, input.Username, userID, input.IsGlobal, &existing.Id)
	if err != nil {
		http.Error(w, "error validating smtp profile", http.StatusInternalServerError)
		return
	}
	if connectionExists {
		JSONResponse(w, http.StatusConflict, map[string]string{
			"error": smtp.ErrConnectionAlreadyExists.Error(),
		})
		return
	}

	existing.Name = input.Name
	existing.IsGlobal = input.IsGlobal
	existing.Host = input.Host
	existing.Port = input.Port
	existing.Username = input.Username
	if input.Password != "" {
		existing.Password = input.Password
	}
	existing.FromName = input.FromName
	existing.FromEmail = input.FromEmail
	existing.IsActive = input.IsActive

	if existing.IsGlobal {
		existing.UserId = nil
	} else {
		existing.UserId = &userID
	}

	if err := h.repo.Update(existing); err != nil {
		http.Error(w, "error updating smtp profile", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existing)
}

func (h *SMTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	profile := r.Context().Value("smtpProfile").(*smtp.SMTPProfile)

	if err := h.repo.Delete(profile.Id); err != nil {
		http.Error(w, "error deleting smtp profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func sendSMTPTestMessage(host string, port int, username string, password string, from string, to string, msg []byte) error {
	return sendSMTPMessage(host, port, username, password, from, []string{to}, msg)
}

func sendSMTPMessage(host string, port int, username string, password string, from string, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	const (
		connectTimeout = 10 * time.Second
		sessionTimeout = 45 * time.Second
	)

	if port == 465 {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		dialer := &net.Dialer{Timeout: connectTimeout}

		conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()
		if err := conn.SetDeadline(time.Now().Add(sessionTimeout)); err != nil {
			return err
		}

		client, err := netsmtp.NewClient(conn, host)
		if err != nil {
			return err
		}
		defer client.Quit()

		auth, err := pickSMTPAuth(client, host, username, password)
		if err != nil {
			return err
		}
		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}

		return writeSMTPMessage(client, from, to, msg)
	}

	conn, err := net.DialTimeout("tcp", addr, connectTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(sessionTimeout)); err != nil {
		return err
	}

	client, err := netsmtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	// Upgrade to TLS before authentication when the server supports STARTTLS.
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	} else if username != "" && !isLocalSMTPHost(host) {
		return errors.New("server does not support STARTTLS for authenticated SMTP submission")
	}

	auth, err := pickSMTPAuth(client, host, username, password)
	if err != nil {
		return err
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	return writeSMTPMessage(client, from, to, msg)
}

func writeSMTPMessage(client *netsmtp.Client, from string, to []string, msg []byte) error {
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func pickSMTPAuth(client *netsmtp.Client, host string, username string, password string) (netsmtp.Auth, error) {
	ok, authExt := client.Extension("AUTH")
	if !ok {
		return nil, nil
	}

	authExt = strings.ToUpper(authExt)
	switch {
	case strings.Contains(authExt, "LOGIN"):
		return &loginAuth{username: username, password: password}, nil
	case strings.Contains(authExt, "PLAIN"):
		return netsmtp.PlainAuth("", username, password, host), nil
	case strings.Contains(authExt, "CRAM-MD5"):
		return netsmtp.CRAMMD5Auth(username, password), nil
	default:
		return nil, fmt.Errorf("unsupported auth mechanisms advertised by server: %s", authExt)
	}
}

type loginAuth struct {
	username string
	password string
	step     int
}

func (a *loginAuth) Start(server *netsmtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		return "", nil, errors.New("refusing LOGIN auth over non-TLS connection")
	}

	a.step = 0
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	switch a.step {
	case 0:
		a.step = 1
		return []byte(a.username), nil
	case 1:
		a.step = 2
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected LOGIN challenge: %q", string(fromServer))
	}
}

func isLocalSMTPHost(host string) bool {
	normalized := strings.Trim(host, "[]")
	if normalized == "localhost" {
		return true
	}

	ip := net.ParseIP(normalized)
	return ip != nil && ip.IsLoopback()
}
