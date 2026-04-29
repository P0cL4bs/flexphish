package handlers

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flexphish/internal/domain/smtp"
	"fmt"
	"net"
	"net/http"
	netsmtp "net/smtp"
	"strings"
	"time"

	"gorm.io/gorm"
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
		Name         string `json:"name"`
		IsGlobal     bool   `json:"is_global"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		SecurityMode string `json:"security_mode"`
		Username     string `json:"username"`
		Password     string `json:"password"`
		FromName     string `json:"from_name"`
		FromEmail    string `json:"from_email"`
		IsActive     bool   `json:"is_active"`
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
	if !isValidSecurityModeInput(input.SecurityMode) {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid security_mode",
		})
		return
	}
	input.SecurityMode = normalizeSecurityMode(input.SecurityMode)

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
		Name:         input.Name,
		IsGlobal:     input.IsGlobal,
		Host:         input.Host,
		Port:         input.Port,
		SecurityMode: input.SecurityMode,
		Username:     input.Username,
		Password:     input.Password,
		FromName:     input.FromName,
		FromEmail:    input.FromEmail,
		IsActive:     input.IsActive,
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
		SMTPProfileID int64  `json:"smtp_profile_id"`
		Name          string `json:"name"`
		Host          string `json:"host"`
		Port          int    `json:"port"`
		SecurityMode  string `json:"security_mode"`
		UseAuth       *bool  `json:"use_authentication"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		FromName      string `json:"from_name"`
		FromEmail     string `json:"from_email"`
		TestEmail     string `json:"test_email"`
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
	if !isValidSecurityModeInput(input.SecurityMode) {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid security_mode",
		})
		return
	}
	input.SecurityMode = normalizeSecurityMode(input.SecurityMode)

	useAuth := true
	if input.UseAuth != nil {
		useAuth = *input.UseAuth
	}

	if input.Host == "" || input.Port <= 0 || input.TestEmail == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "host, port and test_email are required",
		})
		return
	}

	if useAuth && input.Username == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "username is required when authentication is enabled",
		})
		return
	}

	if useAuth && input.Password == "" && input.SMTPProfileID > 0 {
		userID := r.Context().Value("userID").(int64)
		profile, err := h.repo.GetByID(input.SMTPProfileID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				JSONResponse(w, http.StatusNotFound, map[string]string{
					"error": "smtp profile not found",
				})
				return
			}
			JSONResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "error fetching smtp profile",
			})
			return
		}

		if !profile.IsGlobal && (profile.UserId == nil || *profile.UserId != userID) {
			JSONResponse(w, http.StatusForbidden, map[string]string{
				"error": "forbidden",
			})
			return
		}
		input.Password = strings.TrimSpace(profile.Password)
	}

	if useAuth && input.Password == "" {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "password is required when testing a new SMTP profile",
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

	if !useAuth {
		input.Username = ""
		input.Password = ""
	}

	if err := sendSMTPTestMessage(input.Host, input.Port, input.SecurityMode, input.Username, input.Password, fromEmail, input.TestEmail, msg); err != nil {
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
		Name         string `json:"name"`
		IsGlobal     bool   `json:"is_global"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		SecurityMode string `json:"security_mode"`
		Username     string `json:"username"`
		Password     string `json:"password"`
		FromName     string `json:"from_name"`
		FromEmail    string `json:"from_email"`
		IsActive     bool   `json:"is_active"`
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
	if !isValidSecurityModeInput(input.SecurityMode) {
		JSONResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid security_mode",
		})
		return
	}
	input.SecurityMode = normalizeSecurityMode(input.SecurityMode)

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
	existing.SecurityMode = input.SecurityMode
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

func sendSMTPTestMessage(host string, port int, securityMode string, username string, password string, from string, to string, msg []byte) error {
	return sendSMTPMessage(host, port, securityMode, username, password, from, []string{to}, msg)
}

func sendSMTPMessage(host string, port int, securityMode string, username string, password string, from string, to []string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	const (
		connectTimeout = 10 * time.Second
		sessionTimeout = 45 * time.Second
	)
	securityMode = normalizeSecurityMode(securityMode)

	if securityMode == smtp.SecurityModeImplicitTLS {
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

		auth, err := pickSMTPAuth(client, host, username, password, securityMode != smtp.SecurityModeNone, securityMode == smtp.SecurityModeNone)
		if err != nil {
			return err
		}
		if auth != nil {
			if err := authWithFallback(client, auth, username, password, securityMode); err != nil {
				return fmt.Errorf("smtp auth failed: %w", err)
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

	if securityMode == smtp.SecurityModeStartTLS {
		// Upgrade to TLS before authentication when the server supports STARTTLS.
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: host}); err != nil {
				return err
			}
		} else if username != "" && !isLocalSMTPHost(host) {
			return errors.New("server does not support STARTTLS for authenticated SMTP submission")
		}
	}

	auth, err := pickSMTPAuth(client, host, username, password, securityMode != smtp.SecurityModeNone, securityMode == smtp.SecurityModeNone)
	if err != nil {
		return err
	}
	if auth != nil {
		if err := authWithFallback(client, auth, username, password, securityMode); err != nil {
			return fmt.Errorf("smtp auth failed: %w", err)
		}
	}

	return writeSMTPMessage(client, from, to, msg)
}

func normalizeSecurityMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", smtp.SecurityModeStartTLS:
		return smtp.SecurityModeStartTLS
	case smtp.SecurityModeImplicitTLS:
		return smtp.SecurityModeImplicitTLS
	case smtp.SecurityModeNone:
		return smtp.SecurityModeNone
	default:
		return smtp.SecurityModeStartTLS
	}
}

func isValidSecurityModeInput(mode string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", smtp.SecurityModeStartTLS, smtp.SecurityModeImplicitTLS, smtp.SecurityModeNone:
		return true
	default:
		return false
	}
}

func writeSMTPMessage(client *netsmtp.Client, from string, to []string, msg []byte) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("rcpt to %s failed: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("data command failed: %w", err)
	}
	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("message body write failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("message finalization failed: %w", err)
	}

	return nil
}

func pickSMTPAuth(client *netsmtp.Client, host string, username string, password string, allowLoginAuth bool, allowInsecureLogin bool) (netsmtp.Auth, error) {
	ok, authExt := client.Extension("AUTH")
	if !ok {
		return nil, nil
	}

	authExt = strings.ToUpper(authExt)
	switch {
	case allowInsecureLogin && strings.Contains(authExt, "PLAIN"):
		return &plainAuth{
			identity:      "",
			username:      username,
			password:      password,
			host:          host,
			allowInsecure: true,
		}, nil
	case allowLoginAuth && strings.Contains(authExt, "LOGIN"):
		return &loginAuth{username: username, password: password, allowInsecure: allowInsecureLogin}, nil
	case allowInsecureLogin && strings.Contains(authExt, "LOGIN"):
		return &loginAuth{username: username, password: password, allowInsecure: true}, nil
	case strings.Contains(authExt, "PLAIN"):
		return netsmtp.PlainAuth("", username, password, host), nil
	case strings.Contains(authExt, "CRAM-MD5"):
		return netsmtp.CRAMMD5Auth(username, password), nil
	default:
		return nil, fmt.Errorf("unsupported auth mechanisms advertised by server: %s", authExt)
	}
}

type loginAuth struct {
	username      string
	password      string
	step          int
	allowInsecure bool
}

type plainAuth struct {
	identity      string
	username      string
	password      string
	host          string
	allowInsecure bool
}

func (a *plainAuth) Start(server *netsmtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !a.allowInsecure {
		if !isLocalSMTPHost(server.Name) {
			return "", nil, errors.New("unencrypted connection")
		}
	}

	if server.Name != a.host && !isLocalSMTPHost(server.Name) {
		return "", nil, errors.New("wrong host name")
	}

	resp := []byte(a.identity)
	resp = append(resp, 0)
	resp = append(resp, []byte(a.username)...)
	resp = append(resp, 0)
	resp = append(resp, []byte(a.password)...)

	return "PLAIN", resp, nil
}

func (a *plainAuth) Next(_ []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("unexpected server challenge during PLAIN auth")
	}
	return nil, nil
}

func (a *loginAuth) Start(server *netsmtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !a.allowInsecure {
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

func authWithFallback(client *netsmtp.Client, auth netsmtp.Auth, username, password, securityMode string) error {
	if err := client.Auth(auth); err != nil {
		// Some SMTP servers reply to AUTH LOGIN with non-base64 challenge text,
		// which net/smtp rejects. In insecure mode we try a manual AUTH LOGIN flow.
		if securityMode == smtp.SecurityModeNone && strings.Contains(strings.ToLower(err.Error()), "illegal base64 data") {
			return manualAuthLogin(client, username, password)
		}
		return err
	}
	return nil
}

func manualAuthLogin(client *netsmtp.Client, username, password string) error {
	if username == "" || password == "" {
		return errors.New("username and password are required for AUTH LOGIN")
	}

	if _, authExt := client.Extension("AUTH"); !strings.Contains(strings.ToUpper(authExt), "LOGIN") {
		return fmt.Errorf("server does not advertise AUTH LOGIN")
	}

	encUser := base64.StdEncoding.EncodeToString([]byte(username))
	encPass := base64.StdEncoding.EncodeToString([]byte(password))

	id, err := client.Text.Cmd("AUTH LOGIN")
	if err != nil {
		return err
	}
	client.Text.StartResponse(id)
	defer client.Text.EndResponse(id)

	code, msg, err := client.Text.ReadResponse(334)
	if err != nil {
		return err
	}
	_ = code
	_ = msg

	id, err = client.Text.Cmd("%s", encUser)
	if err != nil {
		return err
	}
	client.Text.StartResponse(id)
	code, _, err = client.Text.ReadResponse(334)
	client.Text.EndResponse(id)
	if err != nil {
		return err
	}
	_ = code

	id, err = client.Text.Cmd("%s", encPass)
	if err != nil {
		return err
	}
	client.Text.StartResponse(id)
	defer client.Text.EndResponse(id)

	if _, _, err := client.Text.ReadResponse(235); err != nil {
		return err
	}

	return nil
}
