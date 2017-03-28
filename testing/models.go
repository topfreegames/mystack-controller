package testing

import (
	"github.com/topfreegames/mystack/mystack-controller/models"
	runner "gopkg.in/mgutz/dat.v2/sqlx-runner"
)

//GetTestDB returns a connection to the test database
func GetTestDB() (runner.Connection, error) {
	return models.GetDB(
		"localhost", "mystack_controller_test", 8585, "disable",
		"mystack_controller_test", "",
		10, 10, 100,
	)
}
