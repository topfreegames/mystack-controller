package testing

import (
	"github.com/jmoiron/sqlx"
	"github.com/topfreegames/mystack-controller/models"
)

//GetTestDB returns a connection to the test database
func GetTestDB() (*sqlx.DB, error) {
	return models.GetDB(
		"localhost", "mystack_controller_test", 8585, "disable",
		"mystack_controller_test", "",
		10, 10, 100,
	)
}
