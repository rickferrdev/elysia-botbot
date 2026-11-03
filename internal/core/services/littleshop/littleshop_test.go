package littleshop

import (
	"context"
	"testing"

	mock_slipper "github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbslipper/mocks"
	mock_tomatoes "github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtomatoes/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type test struct {
	suite.Suite
	service  Interface
	tomatoes *mock_tomatoes.MockInterface
	slipper  *mock_slipper.MockInterface
	ctrl     *gomock.Controller
}

func (test *test) SetupTest() {
	test.ctrl = gomock.NewController(test.T())
	test.tomatoes = mock_tomatoes.NewMockInterface(test.ctrl)
	test.slipper = mock_slipper.NewMockInterface(test.ctrl)

	service := New(FxArgs{
		TomatoesDB: test.tomatoes,
		SlipperDB:  test.slipper,
	})

	test.service = service
}

func (test *test) TestPurchaseTomato() {
	test.slipper.EXPECT().
		GetInventory(gomock.Any(), "one", "user").
		Return(25, nil)
	test.tomatoes.EXPECT().
		Increment(gomock.Any(), "one", "user", 5).
		Return(5, nil)
	test.slipper.EXPECT().
		DecrementWith(gomock.Any(), "one", "user", 25).
		Return(0, nil)

	balance, err := test.service.PurchaseTomato(
		context.Background(), "one", "user", 5,
	)

	test.NoError(err)
	test.Equal(5, balance)
}

func (test *test) TestInsufficientBalance() {
	test.slipper.EXPECT().
		GetInventory(gomock.Any(), "one", "user").
		Return(20, nil)

	balance, err := test.service.PurchaseTomato(
		context.Background(), "one", "user", 5,
	)

	test.Error(err)
	test.ErrorIs(err, InsufficientBalanceError)
	test.Zero(balance)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
