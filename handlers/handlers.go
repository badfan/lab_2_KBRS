package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"lab_2_KBRS/cbc"
	"lab_2_KBRS/models"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	socketio "github.com/googollee/go-socket.io"
	"go.uber.org/zap"
)

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

type Handler struct {
	sessions map[string]session
	logger   *zap.SugaredLogger
}

func NewHandler(logger *zap.SugaredLogger) *Handler {
	return &Handler{sessions: make(map[string]session), logger: logger}
}

func (h *Handler) Login(s socketio.Conn, msg string) string {
	h.logger.Infof("user with ID - %v is signing in", s.ID())

	var input models.LogInInput

	err := json.Unmarshal([]byte(msg), &input)
	if err != nil {
		h.logger.Errorf("error occurred while unmarshalling input : %v", err)
		return "could not unmarshal input"
	}

	expectedPassword, ok := users[input.Username]
	if !ok || expectedPassword != input.Password {
		h.logger.Infof("invalid username or password for user with ID - %v", s.ID())
		return "invalid username or password"
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(300 * time.Second)

	h.sessions[sessionToken] = session{
		username: input.Username,
		expiry:   expiresAt,
	}

	n := new(big.Int)
	n, ok = n.SetString(input.N, 10)
	if !ok {
		h.logger.Errorf("SetString error : %v", err)
		return "не могу создать БИГ ИНТ"
	}
	e, err := strconv.Atoi(input.E)
	if err != nil {
		h.logger.Errorf("Atoi error : %v", err)
		return "не могу создать ИНТ"
	}

	encryptedKey, err := rsa.EncryptPKCS1v15(rand.Reader, &rsa.PublicKey{N: n, E: e}, []byte(sessionToken))
	if err != nil {
		h.logger.Errorf("не могу зашифровать : %v", err)
		return "не могу зашифровать"
	}

	return string(encryptedKey)
}

func (h *Handler) CreateFile(s socketio.Conn, msg string) string {
	h.logger.Infof("user with ID - %v is creating a file", s.ID())

	url := s.URL()
	token := url.Query().Get("token")
	if !tokenVerifier(h.sessions, token) {
		h.logger.Errorf("unauthorized user with ID - %v", s.ID())
		return "unauthorized"
	}

	var input models.CreateUpdateFileInput

	err := json.Unmarshal([]byte(msg), &input)
	if err != nil {
		h.logger.Errorf("error occurred while unmarshalling input : %v", err)
		return "could not unmarshal input"
	}

	plainText, err := cbc.CBCDecrypter([]byte(token), []byte(input.Text))
	if err != nil {
		h.logger.Errorf("error occurred while decrypting data : %v", err)
		return "could not decrypt data"
	}

	err = os.WriteFile("./files/"+h.sessions[token].username+"/"+input.FileName, plainText, 0666)
	if err != nil {
		h.logger.Errorf("error occurred while creating data : %v", err)
		return "could not create file"
	}

	return "file created"
}

func (h *Handler) UpdateFile(s socketio.Conn, msg string) string {
	h.logger.Infof("user with ID - %v is editing a file", s.ID())

	url := s.URL()
	token := url.Query().Get("token")
	if !tokenVerifier(h.sessions, token) {
		h.logger.Errorf("unauthorized user with ID - %v", s.ID())
		return "unauthorized"
	}

	var input models.CreateUpdateFileInput

	err := json.Unmarshal([]byte(msg), &input)
	if err != nil {
		h.logger.Errorf("error occurred while unmarshalling input : %v", err)
		return "could not unmarshal input"
	}

	plainText, err := cbc.CBCDecrypter([]byte(token), []byte(input.Text))
	if err != nil {
		h.logger.Errorf("error occurred while decrypting data : %v", err)
		return "could not decrypt data"
	}

	err = os.WriteFile("./files/"+h.sessions[token].username+"/"+input.FileName, plainText, 0666)
	if err != nil {
		h.logger.Errorf("error occurred while writing data : %v", err)
		return "could not edit data"
	}

	return "file edited"
}

func (h *Handler) DeleteFile(s socketio.Conn, msg string) string {
	h.logger.Infof("user with ID - %v is deleting a file", s.ID())

	url := s.URL()
	token := url.Query().Get("token")
	if !tokenVerifier(h.sessions, token) {
		h.logger.Errorf("unauthorized user with ID - %v", s.ID())
		return "unauthorized"
	}

	var input models.GetDeleteFileInput

	err := json.Unmarshal([]byte(msg), &input)
	if err != nil {
		h.logger.Errorf("error occurred while unmarshalling input : %v", err)
		return "could not unmarshal input"
	}

	err = os.Remove("./files/" + h.sessions[token].username + "/" + input.FileName)
	if err != nil {
		h.logger.Errorf("error occurred while deleting file : %v", err)
		return "could not delete file"
	}

	return "file deleted"
}

func (h *Handler) GetFile(s socketio.Conn, msg string) string {
	h.logger.Infof("user with ID - %v is getting a file", s.ID())

	url := s.URL()
	token := url.Query().Get("token")
	if !tokenVerifier(h.sessions, token) {
		h.logger.Errorf("unauthorized user with ID - %v", s.ID())
		return "unauthorized"
	}

	var input models.GetDeleteFileInput

	err := json.Unmarshal([]byte(msg), &input)
	if err != nil {
		h.logger.Errorf("error occurred while unmarshalling input : %v", err)
		return "could not unmarshal input"
	}

	text, err := os.ReadFile("./files/" + h.sessions[token].username + "/" + input.FileName)
	if err != nil {
		h.logger.Errorf("error occurred while getting file : %v", err)
		return "could not get file"
	}

	return string(text)
}

type session struct {
	username string
	expiry   time.Time
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func tokenVerifier(sessions map[string]session, token string) bool {
	userSession, exists := sessions[token]
	if !exists {
		return false
	}

	if userSession.isExpired() {
		delete(sessions, token)
		return false
	}

	return true
}
