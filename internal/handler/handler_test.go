package handler

import (
	"bytes"
	"cruder/internal/controller"
	"cruder/internal/model"
	"cruder/internal/repository"
	"cruder/internal/service"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// setupTestDB initializes the test database connection using test configuration
// Configuration can be provided via environment variables (see LoadTestConfig)
// or via TEST_POSTGRES_DSN for a complete connection string
func setupTestDB(t *testing.T) *sql.DB {
	config := LoadTestConfig()
	dsn := config.BuildDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	return db
}

// setupRouter creates a router with test dependencies
func setupRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	repositories := repository.NewRepository(db)
	services := service.NewService(repositories)
	controllers := controller.NewController(services)
	r := gin.New()
	New(r, controllers.Users)
	return r
}

// insertTestUser inserts a user into the test database and returns the UUID
func insertTestUser(t *testing.T, db *sql.DB, user model.User) string {
	repos := repository.NewRepository(db)
	createdUser, err := repos.Users.Create(&user)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	return createdUser.UUID
}

// userExists checks if a user exists in the database by UUID
func userExists(t *testing.T, db *sql.DB, uuid string) bool {
	repos := repository.NewRepository(db)
	user, err := repos.Users.GetByUUID(uuid)
	if err != nil {
		t.Fatalf("failed to check if user exists: %v", err)
	}
	return user != nil
}

// getUserByUUID retrieves a user by UUID from the database
func getUserByUUID(t *testing.T, db *sql.DB, uuid string) *model.User {
	repos := repository.NewRepository(db)
	user, err := repos.Users.GetByUUID(uuid)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	return user
}

// cleanupTestDB removes all test data from the database
func cleanupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to cleanup test database: %v", err)
	}
}

func TestMain(m *testing.M) {
	// Run tests
	// Each test will set up its own database connection
	code := m.Run()
	os.Exit(code)
}

func TestGetAllUsers_Success(t *testing.T) {
	// Given: Multiple users exist in the database
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	user1 := model.User{
		Username: "user1_test",
		Email:    "user1_test@example.com",
		FullName: "User One",
	}
	user2 := model.User{
		Username: "user2_test",
		Email:    "user2_test@example.com",
		FullName: "User Two",
	}
	_ = insertTestUser(t, db, user1)
	_ = insertTestUser(t, db, user2)

	router := setupRouter(db)

	// When: Sending a GET request to /api/v1/users
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 200 OK and return all users
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var users []model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestGetAllUsers_Empty(t *testing.T) {
	// Given: No users exist in the database
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	router := setupRouter(db)

	// When: Sending a GET request to /api/v1/users
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 200 OK and return empty array
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var users []model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestGetUserByUsername_Success(t *testing.T) {
	// Given: A user exists in the database with a specific username
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	testUser := model.User{
		Username: "getuser_test",
		Email:    "getuser_test@example.com",
		FullName: "Get User",
	}
	_ = insertTestUser(t, db, testUser)

	router := setupRouter(db)

	// When: Sending a GET request to /api/v1/users/username/{username}
	req, _ := http.NewRequest("GET", "/api/v1/users/username/getuser_test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 200 OK and return the user
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var user model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if user.Username != testUser.Username {
		t.Errorf("expected username %s, got %s", testUser.Username, user.Username)
	}
	if user.Email != testUser.Email {
		t.Errorf("expected email %s, got %s", testUser.Email, user.Email)
	}
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	// Given: No user exists with the given username
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	router := setupRouter(db)

	// When: Sending a GET request to /api/v1/users/username/{non-existent-username}
	req, _ := http.NewRequest("GET", "/api/v1/users/username/nonexistent_test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestGetUserByID_Success(t *testing.T) {
	// Given: A user exists in the database with a specific ID
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	testUser := model.User{
		Username: "getbyid_test",
		Email:    "getbyid_test@example.com",
		FullName: "Get By ID User",
	}
	uuid := insertTestUser(t, db, testUser)
	createdUser := getUserByUUID(t, db, uuid)

	router := setupRouter(db)

	// When: Sending a GET request to /api/v1/users/id/{id}
	req, _ := http.NewRequest("GET", "/api/v1/users/id/"+strconv.FormatInt(int64(createdUser.ID), 10), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 200 OK and return the user
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var user model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("expected ID %d, got %d", createdUser.ID, user.ID)
	}
	if user.Username != testUser.Username {
		t.Errorf("expected username %s, got %s", testUser.Username, user.Username)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	// Given: No user exists with the given ID
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	router := setupRouter(db)

	// When: Sending a GET request to /api/v1/users/id/{non-existent-id}
	req, _ := http.NewRequest("GET", "/api/v1/users/id/99999", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestGetUserByID_InvalidID(t *testing.T) {
	// Given: An invalid ID format is provided
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	router := setupRouter(db)

	// When: Sending a GET request to /api/v1/users/id/{invalid-id}
	req, _ := http.NewRequest("GET", "/api/v1/users/id/invalid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	// Given: A user exists in the database with UUID
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	testUser := model.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		FullName: "Test User",
	}
	uuid := insertTestUser(t, db, testUser)

	router := setupRouter(db)

	// When: Sending a DELETE request to /api/v1/users/{uuid}
	req, _ := http.NewRequest("DELETE", "/api/v1/users/"+uuid, nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 204 No Content and user should be removed from the database
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rr.Code)
	}
	if userExists(t, db, uuid) {
		t.Errorf("user was not deleted from the database")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	// Given: No user exists with the given UUID
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	router := setupRouter(db)

	// When: Sending a DELETE request to /api/v1/users/{non-existent-uuid}
	req, _ := http.NewRequest("DELETE", "/api/v1/users/123e4567-e89b-12d3-a456-426614174000", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	// Given: A user exists in the database with UUID
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	testUser := model.User{
		Username: "updateuser_test",
		Email:    "updateuser_test@example.com",
		FullName: "Update User",
	}
	uuid := insertTestUser(t, db, testUser)

	router := setupRouter(db)

	// When: Sending a PATCH request to /api/v1/users/{uuid} with updated data
	updatedUser := model.User{
		Username: "updateduser_test",
		Email:    "updateduser_test@example.com",
		FullName: "Updated User",
	}
	body, _ := json.Marshal(updatedUser)
	req, _ := http.NewRequest("PATCH", "/api/v1/users/"+uuid, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 200 OK and user should be updated in the database
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Verify user was updated in database
	updatedUserFromDB := getUserByUUID(t, db, uuid)
	if updatedUserFromDB == nil {
		t.Fatal("user was deleted from database")
	}
	if updatedUserFromDB.Username != updatedUser.Username {
		t.Errorf("expected username %s, got %s", updatedUser.Username, updatedUserFromDB.Username)
	}
	if updatedUserFromDB.Email != updatedUser.Email {
		t.Errorf("expected email %s, got %s", updatedUser.Email, updatedUserFromDB.Email)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	// Given: No user exists with the given UUID
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	router := setupRouter(db)

	// When: Sending a PATCH request to /api/v1/users/{non-existent-uuid}
	updatedUser := model.User{
		Username: "newuser",
		Email:    "newuser@example.com",
		FullName: "New User",
	}
	body, _ := json.Marshal(updatedUser)
	req, _ := http.NewRequest("PATCH", "/api/v1/users/123e4567-e89b-12d3-a456-426614174000", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestCreateUser_Success(t *testing.T) {
	// Given: No user exists with the given username/email
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	router := setupRouter(db)

	// When: Sending a POST request to /api/v1/users with user data
	newUser := model.User{
		Username: "newuser_test",
		Email:    "newuser_test@example.com",
		FullName: "New User",
	}
	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 201 Created and user should be created in the database
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}

	var createdUser model.User
	if err := json.Unmarshal(rr.Body.Bytes(), &createdUser); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify user was created in database
	userFromDB := getUserByUUID(t, db, createdUser.UUID)
	if userFromDB == nil {
		t.Fatal("user was not created in database")
	}
	if userFromDB.Username != newUser.Username {
		t.Errorf("expected username %s, got %s", newUser.Username, userFromDB.Username)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	// Given: A user already exists with the same username
	db := setupTestDB(t)
	defer db.Close()
	cleanupTestDB(t, db)

	existingUser := model.User{
		Username: "duplicateuser_test",
		Email:    "duplicateuser_test@example.com",
		FullName: "Duplicate User",
	}
	_ = insertTestUser(t, db, existingUser)

	router := setupRouter(db)

	// When: Sending a POST request to /api/v1/users with duplicate username
	newUser := model.User{
		Username: "duplicateuser_test", // Same username
		Email:    "different_test@example.com",
		FullName: "Different User",
	}
	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then: The response status should be 409 Conflict
	if rr.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rr.Code)
	}
}
