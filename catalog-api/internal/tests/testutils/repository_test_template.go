package testutils

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// RepositoryTestTemplate provides common test setup for repositories
type RepositoryTestTemplate struct {
	t       *testing.T
	DB      *sql.DB // Exported for direct access
	mock    sqlmock.Sqlmock
	cleanup func()
}

// NewRepositoryTestTemplate creates a new repository test template
func NewRepositoryTestTemplate(t *testing.T) *RepositoryTestTemplate {
	db, mock := MockDB(t)

	return &RepositoryTestTemplate{
		t:    t,
		DB:   db,
		mock: mock,
		cleanup: func() {
			db.Close()
		},
	}
}

// Cleanup cleans up test resources
func (rtt *RepositoryTestTemplate) Cleanup() {
	if rtt.cleanup != nil {
		rtt.cleanup()
	}
}

// ExpectQuery sets up an expected SQL query
func (rtt *RepositoryTestTemplate) ExpectQuery(query string) *sqlmock.ExpectedQuery {
	return rtt.mock.ExpectQuery(query)
}

// ExpectExec sets up an expected SQL exec
func (rtt *RepositoryTestTemplate) ExpectExec(query string) *sqlmock.ExpectedExec {
	return rtt.mock.ExpectExec(query)
}

// ExpectBegin sets up an expected transaction begin
func (rtt *RepositoryTestTemplate) ExpectBegin() {
	rtt.mock.ExpectBegin()
}

// ExpectCommit sets up an expected transaction commit
func (rtt *RepositoryTestTemplate) ExpectCommit() {
	rtt.mock.ExpectCommit()
}

// ExpectRollback sets up an expected transaction rollback
func (rtt *RepositoryTestTemplate) ExpectRollback() {
	rtt.mock.ExpectRollback()
}

// VerifyAll verifies all expectations were met
func (rtt *RepositoryTestTemplate) VerifyAll() {
	if err := rtt.mock.ExpectationsWereMet(); err != nil {
		rtt.t.Errorf("Unfulfilled expectations: %v", err)
	}
}

// RepositoryTestSuite defines interface for repository test suites
type RepositoryTestSuite interface {
	SetupTest()
	TeardownTest()
	TestCreate()
	TestFindByID()
	TestFindAll()
	TestUpdate()
	TestDelete()
}

// RunRepositoryTestSuite runs a complete repository test suite
func RunRepositoryTestSuite(t *testing.T, suite RepositoryTestSuite) {
	t.Run("Setup", func(t *testing.T) {
		suite.SetupTest()
	})

	t.Run("Create", func(t *testing.T) {
		suite.TestCreate()
	})

	t.Run("FindByID", func(t *testing.T) {
		suite.TestFindByID()
	})

	t.Run("FindAll", func(t *testing.T) {
		suite.TestFindAll()
	})

	t.Run("Update", func(t *testing.T) {
		suite.TestUpdate()
	})

	t.Run("Delete", func(t *testing.T) {
		suite.TestDelete()
	})

	t.Run("Teardown", func(t *testing.T) {
		suite.TeardownTest()
	})
}
