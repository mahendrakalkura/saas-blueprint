package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/config"
	"github.com/mahendrakalkura/saas-blueprint/internal/mfa"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
	"github.com/mahendrakalkura/saas-blueprint/internal/testutil"
	"github.com/mahendrakalkura/saas-blueprint/internal/worker"
)

func TestAuthHandler_Register(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	// Setup repositories
	userRepo := repository.NewUserRepository(testDB.DB.DB)
	sessionRepo := repository.NewSessionRepository(testDB.DB.DB)

	// Mock worker client (we're not testing background jobs here)
	workerClient := &worker.Client{}

	// Setup config
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:               "test-secret",
			AccessTokenDuration:  auth.AccessTokenDuration,
			RefreshTokenDuration: auth.RefreshTokenDuration,
		},
	}

	// Create handler
	mfaService := mfa.NewService("Test App")
	handler := NewAuthHandler(userRepo, sessionRepo, workerClient, cfg, mfaService)

	tests := []struct {
		name           string
		payload        RegisterRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp *AuthResponse)
	}{
		{
			name: "successful registration",
			payload: RegisterRequest{
				Email:    "newuser@example.com",
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, resp *AuthResponse) {
				if resp.AccessToken == "" {
					t.Error("Expected access token")
				}
				if resp.RefreshToken == "" {
					t.Error("Expected refresh token")
				}
				if resp.User == nil {
					t.Fatal("Expected user object")
				}
				if resp.User.Email != "newuser@example.com" {
					t.Errorf("Expected email newuser@example.com, got %s", resp.User.Email)
				}
			},
		},
		{
			name: "duplicate email",
			payload: RegisterRequest{
				Email:    "newuser@example.com", // Same as previous test
				Password: "AnotherPass123!",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "missing email",
			payload: RegisterRequest{
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			payload: RegisterRequest{
				Email: "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.Register(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check response for successful registration
			if tt.expectedStatus == http.StatusCreated && tt.checkResponse != nil {
				var resp AuthResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.checkResponse(t, &resp)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	// Setup repositories
	userRepo := repository.NewUserRepository(testDB.DB.DB)
	sessionRepo := repository.NewSessionRepository(testDB.DB.DB)

	// Create a test user
	testUser := testutil.CreateTestUser(t, userRepo, "testuser@example.com")

	// Mock worker client
	workerClient := &worker.Client{}

	// Setup config
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:               "test-secret",
			AccessTokenDuration:  auth.AccessTokenDuration,
			RefreshTokenDuration: auth.RefreshTokenDuration,
		},
	}

	// Create handler
	mfaService := mfa.NewService("Test App")
	handler := NewAuthHandler(userRepo, sessionRepo, workerClient, cfg, mfaService)

	tests := []struct {
		name           string
		payload        LoginRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp *AuthResponse)
	}{
		{
			name: "successful login",
			payload: LoginRequest{
				Email:    "testuser@example.com",
				Password: "TestPassword123!",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *AuthResponse) {
				if resp.AccessToken == "" {
					t.Error("Expected access token")
				}
				if resp.RefreshToken == "" {
					t.Error("Expected refresh token")
				}
				if resp.User == nil {
					t.Fatal("Expected user object")
				}
				if resp.User.ID != testUser.ID {
					t.Errorf("Expected user ID %s, got %s", testUser.ID, resp.User.ID)
				}
			},
		},
		{
			name: "wrong password",
			payload: LoginRequest{
				Email:    "testuser@example.com",
				Password: "WrongPassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "non-existent user",
			payload: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "SomePassword123!",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing email",
			payload: LoginRequest{
				Password: "TestPassword123!",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.Login(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Check response for successful login
			if tt.expectedStatus == http.StatusOK && tt.checkResponse != nil {
				var resp AuthResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				tt.checkResponse(t, &resp)
			}
		})
	}
}

func TestAuthHandler_LoginWithMFA(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	// Setup repositories
	userRepo := repository.NewUserRepository(testDB.DB.DB)
	sessionRepo := repository.NewSessionRepository(testDB.DB.DB)

	// Create a test user with MFA enabled
	testUser := testutil.CreateTestUserWithMFA(t, userRepo, "mfauser@example.com", "test-mfa-secret")

	// Mock worker client
	workerClient := &worker.Client{}

	// Setup config
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:               "test-secret",
			AccessTokenDuration:  auth.AccessTokenDuration,
			RefreshTokenDuration: auth.RefreshTokenDuration,
		},
	}

	// Create handler
	mfaService := mfa.NewService("Test App")
	handler := NewAuthHandler(userRepo, sessionRepo, workerClient, cfg, mfaService)

	t.Run("login with MFA requires MFA token", func(t *testing.T) {
		payload := LoginRequest{
			Email:    "mfauser@example.com",
			Password: "TestPassword123!",
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler.Login(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		// Should return MFA required response
		if !resp.MFARequired {
			t.Error("Expected MFA required")
		}

		if resp.MFAToken == "" {
			t.Error("Expected MFA token")
		}

		// Should NOT return access/refresh tokens yet
		if resp.AccessToken != "" {
			t.Error("Should not return access token before MFA verification")
		}

		if resp.RefreshToken != "" {
			t.Error("Should not return refresh token before MFA verification")
		}
	})
}

func TestAuthHandler_Me(t *testing.T) {
	// Setup test database
	testDB := testutil.SetupTestDB(t)
	defer testDB.TeardownTestDB(t)

	// Setup repositories
	userRepo := repository.NewUserRepository(testDB.DB.DB)
	sessionRepo := repository.NewSessionRepository(testDB.DB.DB)

	// Create a test user
	testUser := testutil.CreateTestUser(t, userRepo, "testuser@example.com")

	// Mock worker client
	workerClient := &worker.Client{}

	// Setup config
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:               "test-secret",
			AccessTokenDuration:  auth.AccessTokenDuration,
			RefreshTokenDuration: auth.RefreshTokenDuration,
		},
	}

	// Create handler
	mfaService := mfa.NewService("Test App")
	handler := NewAuthHandler(userRepo, sessionRepo, workerClient, cfg, mfaService)

	// Generate a valid JWT token
	token := testutil.GenerateTestJWT(t, testUser.ID, testUser.Email, cfg.JWT.Secret)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "authenticated request",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired/invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Add auth context (normally done by middleware)
			if tt.expectedStatus == http.StatusOK {
				req = req.WithContext(auth.WithUserContext(req.Context(), auth.UserContext{
					UserID: testUser.ID,
					Email:  testUser.Email,
				}))
			}

			w := httptest.NewRecorder()
			handler.Me(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
