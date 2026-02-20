package client

import (
	"context"
	"encoding/json"
	"fmt"
	cfg "gophkeeper/internal/config/agent"
	dto "gophkeeper/internal/dto/model"
	"gophkeeper/internal/service/agent"
	"gophkeeper/internal/service/file"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type agentSendler interface {
	Register(ctx context.Context, user dto.User) error
	Login(ctx context.Context, user dto.User) error
	CreateSecret(ctx context.Context, secret dto.Secret, data []byte) error
	UpdateSecret(ctx context.Context, secret dto.Secret, data []byte) error
	GetSecret(ctx context.Context, secretName string, secretType string) (agent.Secret, error)
	ListSecrets(ctx context.Context, secretType string) ([]agent.Secret, error)
	DeleteSecret(ctx context.Context, secretName string, secretType string) error
}

type AppState string

const (
	StateMenu            AppState = "menu"
	StateRegister        AppState = "register"
	StateAuth            AppState = "auth"
	StateSecretType      AppState = "secType"
	StateSecretForm      AppState = "secForm"
	StateSecretNameInput AppState = "secretNameInput"
	StateSecretList      AppState = "secretList"
	StateResult          AppState = "result"
	StateError           AppState = "error"

	OpCreate SecretOperation = "createSecret"
	OpUpdate SecretOperation = "updateSecret"
	OpGet    SecretOperation = "getSecret"
	OpList   SecretOperation = "listSecret"
	OpDelete SecretOperation = "deleteSecret"
)

type SecretOperation string
type model struct {
	state     AppState
	cursor    int
	menuItems []string
	agent     agentSendler

	user     dto.User
	fieldIdx int

	secType     string
	secFields   map[int]string
	secFieldIdx int
	secOp       SecretOperation

	secretName string

	secretData []byte
	secrets    []agent.Secret

	message string
}

func initialModel(agent agentSendler) model {
	return model{
		state:     StateMenu,
		menuItems: []string{"Register", "Auth", "Create Secret", "Update Secret", "Get Secret", "List Secrets", "Delete Secret", "Quit"},
		secFields: make(map[int]string),
		agent:     agent,
	}
}

func (m model) Init() tea.Cmd { return nil }

// Update handlers
func (m *model) handleMenu(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.menuItems)-1 {
			m.cursor++
		}
	case "enter":
		m.handleMenuSelection()
	}
	return m, nil
}

func (m *model) handleMenuSelection() {
	switch m.cursor {
	case 0: // Register
		m.setState(StateRegister, 0, "", "")
	case 1: // Auth
		m.setState(StateAuth, 0, "", "")
	case 2: // Create Secret
		m.secOp = OpCreate
		m.state, m.cursor = StateSecretType, 0
		m.secFields = make(map[int]string)
	case 3: // Update Secret
		m.secOp = OpUpdate
		m.state, m.cursor = StateSecretType, 0
		m.secFields = make(map[int]string)
	case 4: // Get Secret
		m.secOp = OpGet
		m.state, m.cursor = StateSecretType, 0
		m.secFields = make(map[int]string)
	case 5: // List Secrets
		m.secOp = OpList
		m.state, m.cursor = StateSecretType, 0
		m.secFields = make(map[int]string)
		m.secrets = nil
	case 6: // Delete Secret
		m.secOp = OpDelete
		m.state, m.cursor = StateSecretType, 0
		m.secFields = make(map[int]string)
	case 7: // Quit
		m.state = StateMenu
	}
}

func (m *model) setState(state AppState, fieldIdx int, login, password string) {
	m.state = state
	m.fieldIdx = fieldIdx
	m.user.Login = login
	m.user.Password = password
}

func (m *model) handleAuthRegister(key string, isRegister bool) {
	switch key {
	case "esc":
		m.state = StateMenu
	case "tab":
		if m.fieldIdx < 1 {
			m.fieldIdx++
		}
	case "shift+tab":
		if m.fieldIdx > 0 {
			m.fieldIdx--
		}
	case "backspace":
		m.handleBackspace()
	case "enter":
		if m.fieldIdx == 1 {
			m.handleAuthRegisterSubmit(isRegister)
		}
	default:
		if len(key) == 1 {
			m.handleAuthRegisterInput(key)
		}
	}
}

func (m *model) handleBackspace() {
	if m.fieldIdx == 0 && len(m.user.Login) > 0 {
		m.user.Login = m.user.Login[:len(m.user.Login)-1]
	}
	if m.fieldIdx == 1 && len(m.user.Password) > 0 {
		m.user.Password = m.user.Password[:len(m.user.Password)-1]
	}
}

func (m *model) handleAuthRegisterInput(key string) {
	if m.fieldIdx == 0 {
		m.user.Login += key
	}
	if m.fieldIdx == 1 {
		m.user.Password += key
	}
}

func (m *model) handleAuthRegisterSubmit(isRegister bool) {
	var err error
	ctx := context.Background()

	if isRegister {
		err = m.agent.Register(ctx, m.user)
	} else {
		err = m.agent.Login(ctx, m.user)
	}

	if err != nil {
		m.message = err.Error()
		m.state = StateError
	} else {
		if isRegister {
			m.message = "User successfully registered"
		} else {
			m.message = "Authorization successful"
		}
		m.state = StateResult
	}
}

func (m *model) handleSecretType(key string) {
	secretTypes := m.getSecretTypes()

	switch key {
	case "esc":
		m.state = StateMenu
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(secretTypes)-1 {
			m.cursor++
		}
	case "enter":
		m.handleSecretTypeSelection(secretTypes)
	}
}

func (m *model) getSecretTypes() []string {
	types := []string{"login_pass", "text", "bank_card", "binary_data"}
	if m.secOp == OpList {
		types = append(types, "all")
	}
	return types
}

func (m *model) handleSecretTypeSelection(types []string) {
	m.secType = types[m.cursor]

	switch m.secOp {
	case OpGet, OpDelete:
		m.state = StateSecretNameInput
		m.secretName = ""
	case OpList:
		m.fetchSecretList()
	default:
		m.state = StateSecretForm
		m.secFieldIdx = 0
		m.secFields = make(map[int]string)
	}
}

func (m *model) fetchSecretList() {
	secrets, err := m.agent.ListSecrets(context.Background(), m.secType)
	if err != nil {
		m.message = err.Error()
		m.state = StateError
	} else {
		m.secrets = secrets
		m.state = StateSecretList
		m.cursor = 0
	}
}

func (m *model) handleSecretList(key string) {
	switch key {
	case "esc":
		m.state = StateSecretType
		m.secrets = nil
		m.cursor = 0
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.secrets)-1 {
			m.cursor++
		}
	case "enter":
		m.showSecretDetails()
	}
}

func (m *model) showSecretDetails() {
	if len(m.secrets) == 0 || m.cursor >= len(m.secrets) {
		return
	}

	secret := m.secrets[m.cursor]
	data, err := json.MarshalIndent(secret, "", "  ")
	if err == nil {
		m.message = fmt.Sprintf("Secret '%s' (%s)\n\n%s",
			secret.Name,
			secret.SecretType,
			string(data))
	} else {
		m.message = fmt.Sprintf("Secret '%s' (%s)", secret.Name, secret.SecretType)
	}
	m.state = StateResult
}

func (m *model) handleSecretNameInput(key string) {
	switch key {
	case "esc":
		m.state = StateSecretType
		m.secretName = ""
	case "backspace":
		if len(m.secretName) > 0 {
			m.secretName = m.secretName[:len(m.secretName)-1]
		}
	case "enter":
		if m.secretName != "" {
			m.handleSecretNameSubmit()
		}
	default:
		if len(key) == 1 {
			m.secretName += key
		}
	}
}

func (m *model) handleSecretNameSubmit() {
	var err error
	ctx := context.Background()

	switch m.secOp {
	case OpGet:
		var secret agent.Secret
		secret, err = m.agent.GetSecret(ctx, m.secretName, m.secType)
		if err == nil {
			m.secretData, err = json.MarshalIndent(secret, "", "  ")
		}
	case OpDelete:
		err = m.agent.DeleteSecret(ctx, m.secretName, m.secType)
	}

	if err != nil {
		m.message = err.Error()
		m.state = StateError
	} else {
		m.formatSecretResult()
		m.state = StateResult
	}
}

func (m *model) formatSecretResult() {
	if m.secOp == OpGet {
		m.message = fmt.Sprintf("Secret '%s' successfully retrieved\n\n%s",
			m.secretName, m.secretData)
		m.secretData = nil
	} else {
		m.message = fmt.Sprintf("Secret '%s' successfully deleted", m.secretName)
	}
}

func (m *model) handleSecretForm(key string) {
	maxField := m.getMaxField() - 1

	switch key {
	case "esc":
		m.state = StateSecretType
		m.secFields = make(map[int]string)
	case "tab":
		if m.secFieldIdx < maxField {
			m.secFieldIdx++
		}
	case "shift+tab":
		if m.secFieldIdx > 0 {
			m.secFieldIdx--
		}
	case "backspace":
		m.handleSecretFormBackspace()
	case "enter":
		if m.secFieldIdx == maxField {
			m.handleSecretFormSubmit()
		}
	default:
		if len(key) == 1 {
			m.secFields[m.secFieldIdx] += key
		}
	}
}

func (m *model) getMaxField() int {
	fieldMap := map[string]int{
		"all":         4,
		"login_pass":  4,
		"text":        3,
		"bank_card":   5,
		"binary_data": 3,
	}
	return fieldMap[m.secType]
}

func (m *model) handleSecretFormBackspace() {
	if len(m.secFields[m.secFieldIdx]) > 0 {
		s := m.secFields[m.secFieldIdx]
		m.secFields[m.secFieldIdx] = s[:len(s)-1]
	}
}

func (m *model) handleSecretFormSubmit() {
	secret := dto.Secret{
		Name:        m.secFields[0],
		Description: m.secFields[1],
		SecretType:  m.secType,
	}

	secretData, err := m.prepareSecretData()
	if err != nil {
		m.message = err.Error()
		m.state = StateError
		return
	}

	secret.Data = secretData
	err = m.submitSecret(secret, secretData)

	if err != nil {
		m.message = err.Error()
		m.state = StateError
	} else {
		m.message = fmt.Sprintf("Secret '%s' successfully %s",
			m.secFields[0],
			map[SecretOperation]string{
				OpCreate: "created",
				OpUpdate: "updated",
			}[m.secOp])
		m.state = StateResult
	}
}

func (m *model) prepareSecretData() ([]byte, error) {
	var data interface{}

	switch m.secType {
	case "login_pass":
		data = dto.LoginPass{
			Login:    m.secFields[2],
			Password: m.secFields[3],
		}
	case "text":
		data = dto.SecretText{
			Text: m.secFields[2],
		}
	case "bank_card":
		expiry := m.secFields[3]
		data = dto.BankCard{
			CardNumber:  m.secFields[2],
			ExpiryMonth: expiry[:2],
			ExpiryYear:  expiry[2:],
			CVV:         m.secFields[4],
		}
	case "binary_data":
		readFile, err := file.ReadFile(m.secFields[2])
		if err != nil {
			return nil, err
		}
		data = dto.BinaryData{
			Data: readFile,
		}
	}

	return json.Marshal(data)
}

func (m *model) submitSecret(secret dto.Secret, secretData []byte) error {
	ctx := context.Background()

	switch m.secOp {
	case OpCreate:
		return m.agent.CreateSecret(ctx, secret, secretData)
	case OpUpdate:
		return m.agent.UpdateSecret(ctx, secret, secretData)
	}
	return nil
}

func (m *model) handleResultError(key string) {
	if key == "enter" || key == "esc" {
		m.state = StateMenu
		m.cursor = 0
		m.message = ""
	}
}

// Main Update function
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch m.state {
		case StateMenu:
			return m.handleMenu(key)

		case StateRegister:
			m.handleAuthRegister(key, true)
			return m, nil

		case StateAuth:
			m.handleAuthRegister(key, false)
			return m, nil

		case StateSecretType:
			m.handleSecretType(key)
			return m, nil

		case StateSecretList:
			m.handleSecretList(key)
			return m, nil

		case StateSecretNameInput:
			m.handleSecretNameInput(key)
			return m, nil

		case StateSecretForm:
			m.handleSecretForm(key)
			return m, nil

		case StateResult, StateError:
			m.handleResultError(key)
			return m, nil
		}
	}
	return m, nil
}

// View functions
func (m model) View() string {
	switch m.state {
	case StateMenu:
		return m.viewMenu()
	case StateRegister, StateAuth:
		return m.viewAuthRegister()
	case StateSecretType:
		return m.viewSecretType()
	case StateSecretList:
		return m.viewSecretList()
	case StateSecretNameInput:
		return m.viewSecretNameInput()
	case StateSecretForm:
		return m.viewSecretForm()
	case StateResult:
		return m.viewResult()
	case StateError:
		return m.viewError()
	default:
		return ""
	}
}

func (m model) viewMenu() string {
	var s strings.Builder
	s.WriteString("=== Secret Manager ===\n\n")

	for i, item := range m.menuItems {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}
		s.WriteString(cursor + item + "\n")
	}
	s.WriteString("\n↑/↓: navigate • enter: select • q: quit")
	return s.String()
}

func (m model) viewAuthRegister() string {
	var s strings.Builder

	if m.state == StateRegister {
		s.WriteString("=== Registration ===\n\n")
	} else {
		s.WriteString("=== Authorization ===\n\n")
	}

	s.WriteString(m.renderField("Login", m.user.Login, 0))
	s.WriteString(m.renderField("Password", strings.Repeat("*", len(m.user.Password)), 1))

	s.WriteString("\ntab: next • enter: submit • esc: back")
	return s.String()
}

func (m model) renderField(label, value string, fieldIdx int) string {
	cursor := "  "
	if m.fieldIdx == fieldIdx {
		cursor = "> "
	}

	fieldCursor := ""
	if m.fieldIdx == fieldIdx {
		fieldCursor = "_"
	}

	return fmt.Sprintf("%s%s: %s%s\n\n", cursor, label, value, fieldCursor)
}

func (m model) viewSecretType() string {
	var s strings.Builder
	s.WriteString("=== Select Secret Type ===\n\n")

	types := m.getSecretTypes()
	for i, t := range types {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}
		s.WriteString(cursor + t + "\n")
	}
	s.WriteString("\n↑/↓: navigate • enter: select • esc: back")
	return s.String()
}

func (m model) viewSecretList() string {
	var s strings.Builder
	s.WriteString("=== Secret List ===\n\n")
	s.WriteString(fmt.Sprintf("Type: %s\n\n", m.secType))

	if len(m.secrets) == 0 {
		s.WriteString("  No secrets found\n")
	} else {
		for i, secret := range m.secrets {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}
			s.WriteString(fmt.Sprintf("%s%s [%s]\n", cursor, secret.Name, secret.SecretType))
			if secret.Description != "" {
				s.WriteString(fmt.Sprintf("    📝 %s\n", secret.Description))
			}
			s.WriteString("\n")
		}
	}
	s.WriteString("\n↑/↓: navigate • enter: view • esc: back")
	return s.String()
}

func (m model) viewSecretNameInput() string {
	var s strings.Builder
	s.WriteString("=== Enter Secret Name ===\n\n")

	if m.secOp == OpGet {
		s.WriteString("Get Secret\n")
	} else {
		s.WriteString("Delete Secret\n")
	}

	s.WriteString(fmt.Sprintf("Type: %s\n\n", m.secType))
	s.WriteString(fmt.Sprintf("Name: %s_\n\n", m.secretName))
	s.WriteString("enter: submit • esc: back")
	return s.String()
}

func (m model) viewSecretForm() string {
	var s strings.Builder

	opName := map[SecretOperation]string{
		OpCreate: "Create",
		OpUpdate: "Update",
	}[m.secOp]

	s.WriteString(fmt.Sprintf("=== %s %s ===\n\n", opName, m.secType))

	fields, placeholders := m.getFormFields()
	for i, field := range fields {
		cursor := "  "
		if m.secFieldIdx == i {
			cursor = "> "
		}

		value := m.secFields[i]
		if value == "" && m.secFieldIdx != i {
			value = "[" + placeholders[i] + "]"
		}

		fieldCursor := ""
		if m.secFieldIdx == i {
			fieldCursor = "_"
		}

		s.WriteString(fmt.Sprintf("%s%s: %s%s\n\n", cursor, field, value, fieldCursor))
	}

	s.WriteString("tab: next • enter: submit • esc: back")
	return s.String()
}

func (m model) getFormFields() ([]string, []string) {
	switch m.secType {
	case "login_pass":
		return []string{"Name", "Description", "Login", "Password"},
			[]string{"enter name", "enter description", "enter login", "enter password"}
	case "text":
		return []string{"Name", "Description", "Text"},
			[]string{"enter name", "enter description", "enter text"}
	case "bank_card":
		return []string{"Name", "Description", "Card Number", "Expiry (MMYY)", "CVV"},
			[]string{"enter name", "enter description", "0000 0000 0000 0000", "0125", "123"}
	case "binary_data":
		return []string{"Name", "Description", "Path to file"},
			[]string{"enter name", "enter description", "enter path to file"}
	default:
		return []string{}, []string{}
	}
}

func (m model) viewResult() string {
	return fmt.Sprintf("=== Success ===\n\n%s\n\nenter: continue", m.message)
}

func (m model) viewError() string {
	return fmt.Sprintf("=== Error ===\n\n%s\n\nenter: continue", m.message)
}

type Client struct {
	program *tea.Program
}

func New(config cfg.Config) (*Client, error) {
	a, err := agent.New(config)
	if err != nil {
		return nil, err
	}

	if config.User.Login != "" && config.User.Password != "" {
		if err = a.Login(context.Background(), config.User); err != nil {
			return nil, err
		}
	}

	return &Client{
		program: tea.NewProgram(initialModel(a)),
	}, nil
}

func (c *Client) Start() error {
	if _, err := c.program.Run(); err != nil {
		return err
	}
	return nil
}
