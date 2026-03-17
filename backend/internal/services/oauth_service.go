package services

import (
"crypto/rand"
"encoding/hex"
"encoding/json"
"errors"
"fmt"
"io"
"net/http"
"net/url"
"strings"
"time"

"github.com/google/uuid"
"github.com/jmoiron/sqlx"

"github.com/vikukumar/Pushpaka/internal/config"
"github.com/vikukumar/Pushpaka/internal/models"
"github.com/vikukumar/Pushpaka/internal/repositories"
)

var ErrOAuthStateMismatch = errors.New("OAuth state mismatch or expired")
var ErrOAuthExchangeFailed = errors.New("OAuth token exchange failed")

type OAuthService struct {
userRepo *repositories.UserRepository
cfg      *config.Config
authSvc  *AuthService
db       *sqlx.DB
}

func NewOAuthService(userRepo *repositories.UserRepository, cfg *config.Config, authSvc *AuthService, db *sqlx.DB) *OAuthService {
return &OAuthService{userRepo: userRepo, cfg: cfg, authSvc: authSvc, db: db}
}

func (s *OAuthService) GenerateState(provider, redirect string) (string, error) {
b := make([]byte, 16)
if _, err := rand.Read(b); err != nil {
return "", err
}
state := hex.EncodeToString(b)
expiresAt := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
q := s.db.Rebind(`INSERT INTO oauth_states (state, provider, redirect, expires_at) VALUES (?, ?, ?, ?)`)
if _, err := s.db.Exec(q, state, provider, redirect, expiresAt); err != nil {
return "", fmt.Errorf("storing oauth state: %w", err)
}
return state, nil
}

func (s *OAuthService) ValidateState(state string) error {
var expiresAt string
err := s.db.Get(&expiresAt, s.db.Rebind(`SELECT expires_at FROM oauth_states WHERE state = ?`), state)
if err != nil {
return ErrOAuthStateMismatch
}
s.db.Exec(s.db.Rebind(`DELETE FROM oauth_states WHERE state = ?`), state) //nolint:errcheck
t, err := time.Parse(time.RFC3339, expiresAt)
if err != nil || time.Now().UTC().After(t) {
return ErrOAuthStateMismatch
}
return nil
}

func (s *OAuthService) GithubAuthURL(state string) string {
params := url.Values{
"client_id":    {s.cfg.GithubClientID},
"redirect_uri": {s.cfg.BaseURL + "/api/v1/auth/github/callback"},
"scope":        {"user:email"},
"state":        {state},
}
return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (s *OAuthService) GitlabAuthURL(state string) string {
base := s.cfg.GitlabBaseURL
if base == "" {
base = "https://gitlab.com"
}
params := url.Values{
"client_id":     {s.cfg.GitlabClientID},
"redirect_uri":  {s.cfg.BaseURL + "/api/v1/auth/gitlab/callback"},
"response_type": {"code"},
"scope":         {"read_user"},
"state":         {state},
}
return base + "/oauth/authorize?" + params.Encode()
}

type githubUser struct {
ID    int    `json:"id"`
Login string `json:"login"`
Email string `json:"email"`
Name  string `json:"name"`
}

type githubEmail struct {
Email    string `json:"email"`
Primary  bool   `json:"primary"`
Verified bool   `json:"verified"`
}

func (s *OAuthService) ExchangeGithub(code string) (*models.AuthResponse, error) {
token, err := githubTokenExchange(s.cfg.GithubClientID, s.cfg.GithubClientSecret, code,
s.cfg.BaseURL+"/api/v1/auth/github/callback")
if err != nil {
return nil, fmt.Errorf("%w: %v", ErrOAuthExchangeFailed, err)
}
ghUser, err := githubFetchUser(token)
if err != nil {
return nil, fmt.Errorf("fetching github profile: %w", err)
}
email := ghUser.Email
if email == "" {
if emails, err2 := githubFetchEmails(token); err2 == nil {
for _, e := range emails {
if e.Primary && e.Verified {
email = e.Email
break
}
}
}
}
if email == "" {
email = fmt.Sprintf("github_%d@noreply.pushpaka", ghUser.ID)
}
name := ghUser.Name
if name == "" {
name = ghUser.Login
}
return s.findOrCreateUser(email, name, "github", fmt.Sprintf("%d", ghUser.ID))
}

type gitlabUser struct {
ID       int    `json:"id"`
Username string `json:"username"`
Email    string `json:"email"`
Name     string `json:"name"`
}

func (s *OAuthService) ExchangeGitlab(code string) (*models.AuthResponse, error) {
base := s.cfg.GitlabBaseURL
if base == "" {
base = "https://gitlab.com"
}
token, err := gitlabTokenExchange(base, s.cfg.GitlabClientID, s.cfg.GitlabClientSecret, code,
s.cfg.BaseURL+"/api/v1/auth/gitlab/callback")
if err != nil {
return nil, fmt.Errorf("%w: %v", ErrOAuthExchangeFailed, err)
}
gUser, err := gitlabFetchUser(base, token)
if err != nil {
return nil, fmt.Errorf("fetching gitlab profile: %w", err)
}
email := gUser.Email
if email == "" {
email = fmt.Sprintf("gitlab_%d@noreply.pushpaka", gUser.ID)
}
name := gUser.Name
if name == "" {
name = gUser.Username
}
return s.findOrCreateUser(email, name, "gitlab", fmt.Sprintf("%d", gUser.ID))
}

func (s *OAuthService) findOrCreateUser(email, name, provider, providerID string) (*models.AuthResponse, error) {
user, err := s.userRepo.FindByEmail(email)
if err != nil {
now := models.NowUTC()
user = &models.User{
ID:           uuid.New().String(),
Email:        email,
Name:         name,
PasswordHash: "$oauth$" + provider + "$" + providerID,
APIKey:       uuid.New().String(),
Role:         "user",
CreatedAt:    now,
UpdatedAt:    now,
}
if createErr := s.userRepo.Create(user); createErr != nil {
return nil, fmt.Errorf("creating oauth user: %w", createErr)
}
}
jwtToken, err := s.authSvc.generateToken(user.ID)
if err != nil {
return nil, err
}
return &models.AuthResponse{Token: jwtToken, User: *user}, nil
}

func githubTokenExchange(clientID, clientSecret, code, redirectURI string) (string, error) {
data := url.Values{
"client_id":     {clientID},
"client_secret": {clientSecret},
"code":          {code},
"redirect_uri":  {redirectURI},
}
req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token",
strings.NewReader(data.Encode()))
req.Header.Set("Accept", "application/json")
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Do(req)
if err != nil {
return "", err
}
defer resp.Body.Close()
var result struct {
AccessToken string `json:"access_token"`
Err         string `json:"error"`
}
if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
return "", err
}
if result.Err != "" {
return "", errors.New(result.Err)
}
return result.AccessToken, nil
}

func githubFetchUser(token string) (*githubUser, error) {
	u, err := githubGET[githubUser]("https://api.github.com/user", token)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func githubFetchEmails(token string) ([]githubEmail, error) {
return githubGET[[]githubEmail]("https://api.github.com/user/emails", token)
}

func githubGET[T any](endpoint, token string) (T, error) {
var zero T
req, _ := http.NewRequest("GET", endpoint, nil)
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Accept", "application/vnd.github+json")
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Do(req)
if err != nil {
return zero, err
}
defer resp.Body.Close()
body, _ := io.ReadAll(resp.Body)
var result T
if err := json.Unmarshal(body, &result); err != nil {
return zero, err
}
return result, nil
}

func gitlabTokenExchange(baseURL, clientID, clientSecret, code, redirectURI string) (string, error) {
data := url.Values{
"client_id":     {clientID},
"client_secret": {clientSecret},
"code":          {code},
"grant_type":    {"authorization_code"},
"redirect_uri":  {redirectURI},
}
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.PostForm(baseURL+"/oauth/token", data)
if err != nil {
return "", err
}
defer resp.Body.Close()
var result struct {
AccessToken string `json:"access_token"`
Err         string `json:"error"`
}
if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
return "", err
}
if result.Err != "" {
return "", errors.New(result.Err)
}
return result.AccessToken, nil
}

func gitlabFetchUser(baseURL, token string) (*gitlabUser, error) {
req, _ := http.NewRequest("GET", baseURL+"/api/v4/user", nil)
req.Header.Set("Authorization", "Bearer "+token)
client := &http.Client{Timeout: 10 * time.Second}
resp, err := client.Do(req)
if err != nil {
return nil, err
}
defer resp.Body.Close()
var u gitlabUser
if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
return nil, err
}
return &u, nil
}
