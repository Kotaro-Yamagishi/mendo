package order_test

import (
	"context"
	"errors"
	"testing"
	"time"

	queryorder "mendo/internal/application/query/order"
	"mendo/internal/domain/kitchen"
	"mendo/internal/domain/service"
	"mendo/internal/testutil"
)

func TestEstimateWaitTimeUsecase_Execute_Success(t *testing.T) {
	t.Parallel()
	// CookingCapacity = MaxConcurrentTasks(10) - activeTasks(0) = 10
	// pendingOrders=20, capacity=10 → estimatedMinutes = (20/10)*5 = 10
	k := kitchen.NewKitchen("kitchen-1")
	orderReader := &testutil.StubOrderReader{PendingCount: 20}
	kitchenReader := &testutil.StubKitchenReader{Kitchen: k}

	calc := service.NewWaitTimeCalculator(orderReader, kitchenReader)
	uc := queryorder.NewEstimateWaitTimeUsecase(calc, "kitchen-1")

	d, err := uc.Execute(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	expected := 10 * time.Minute
	if d != expected {
		t.Errorf("expected %v, got: %v", expected, d)
	}
}

func TestEstimateWaitTimeUsecase_Execute_OrderReaderError(t *testing.T) {
	t.Parallel()
	countErr := errors.New("order db error")
	orderReader := &testutil.StubOrderReader{CountErr: countErr}
	kitchenReader := &testutil.StubKitchenReader{}

	calc := service.NewWaitTimeCalculator(orderReader, kitchenReader)
	uc := queryorder.NewEstimateWaitTimeUsecase(calc, "kitchen-1")

	_, err := uc.Execute(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, countErr) {
		t.Errorf("expected wrapped count error, got: %v", err)
	}
}

func TestEstimateWaitTimeUsecase_Execute_KitchenReaderError(t *testing.T) {
	t.Parallel()
	findErr := errors.New("kitchen not found")
	orderReader := &testutil.StubOrderReader{PendingCount: 5}
	kitchenReader := &testutil.StubKitchenReader{FindErr: findErr}

	calc := service.NewWaitTimeCalculator(orderReader, kitchenReader)
	uc := queryorder.NewEstimateWaitTimeUsecase(calc, "kitchen-1")

	_, err := uc.Execute(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, findErr) {
		t.Errorf("expected wrapped find error, got: %v", err)
	}
}
