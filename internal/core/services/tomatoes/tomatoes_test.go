package tomatoes

import (
	"context"
	"testing"

	mock_tomatoes "github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtomatoes/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type test struct {
	suite.Suite
	service  Interface
	tomatoes *mock_tomatoes.MockInterface
	ctrl     *gomock.Controller
}

func (test *test) SetupTest() {
	test.ctrl = gomock.NewController(test.T())
	test.tomatoes = mock_tomatoes.NewMockInterface(test.ctrl)

	service := New(FxArgs{
		TomatoesDB: test.tomatoes,
	})

	test.service = service
}

func (test *test) TestThrowTomato() {
	test.tomatoes.EXPECT().
		GetInventory(gomock.Any(), "one", "user").
		Return(1, nil)
	test.tomatoes.EXPECT().
		Decrement(gomock.Any(), "one", "user", 1)

	inventory, err := test.service.ThrowTomato(
		context.Background(), "one", "user",
	)

	test.NoError(err)
	test.Equal(0, inventory)
}

func (test *test) TestInsufficientBalance() {
	test.tomatoes.EXPECT().
		GetInventory(gomock.Any(), "one", "user").
		Return(0, nil)

	inventory, err := test.service.ThrowTomato(
		context.Background(), "one", "user",
	)

	test.Error(err)
	test.ErrorIs(err, InsufficientBalanceError)
	test.Zero(inventory)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
