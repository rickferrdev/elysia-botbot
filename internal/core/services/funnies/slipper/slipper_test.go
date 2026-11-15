package slipper

import (
	"context"
	"testing"

	mock_slipper "github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbslipper/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type test struct {
	suite.Suite
	service Interface
	slipper *mock_slipper.MockInterface
	ctrl    *gomock.Controller
}

func (test *test) SetupTest() {
	test.ctrl = gomock.NewController(test.T())
	test.slipper = mock_slipper.NewMockInterface(test.ctrl)

	service := New(FxArgs{
		SlipperDB: test.slipper,
	})

	test.service = service
}

func (test *test) TestThrowSlipper() {
	test.slipper.EXPECT().
		IncrementWith(gomock.Any(), "one", "user", 1).
		Return(1, nil)

	inventory, err := test.service.ThrowSlipper(
		context.Background(), "one", "user",
	)

	test.NoError(err)
	test.Equal(1, inventory)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(test))
}
