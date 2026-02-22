package irs

import (
	"context"
	"testing"

	"isac-cran-system/internal/model"
)

func TestSimulator_Connect(t *testing.T) {
	simulator := NewSimulator(64, "2.4GHz")

	ctx := context.Background()
	err := simulator.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !simulator.IsConnected() {
		t.Error("Expected simulator to be connected")
	}
}

func TestSimulator_Disconnect(t *testing.T) {
	simulator := NewSimulator(64, "2.4GHz")
	ctx := context.Background()

	_ = simulator.Connect(ctx)
	simulator.Disconnect()

	if simulator.IsConnected() {
		t.Error("Expected simulator to be disconnected")
	}
}

func TestSimulator_SetPhaseShifts(t *testing.T) {
	simulator := NewSimulator(64, "2.4GHz")
	ctx := context.Background()
	_ = simulator.Connect(ctx)

	phaseShifts := make([]float64, 64)
	for i := range phaseShifts {
		phaseShifts[i] = float64(i) * 0.1
	}

	err := simulator.SetPhaseShifts(ctx, phaseShifts)
	if err != nil {
		t.Fatalf("SetPhaseShifts failed: %v", err)
	}

	status, err := simulator.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if len(status.PhaseShifts) != 64 {
		t.Errorf("Expected 64 phase shifts, got %d", len(status.PhaseShifts))
	}
}

func TestSimulator_SetPhaseShifts_WrongCount(t *testing.T) {
	simulator := NewSimulator(64, "2.4GHz")
	ctx := context.Background()
	_ = simulator.Connect(ctx)

	phaseShifts := make([]float64, 32)

	err := simulator.SetPhaseShifts(ctx, phaseShifts)
	if err == nil {
		t.Error("Expected error for wrong phase shift count")
	}
}

func TestSimulator_GetStatus(t *testing.T) {
	simulator := NewSimulator(64, "2.4GHz")
	ctx := context.Background()
	_ = simulator.Connect(ctx)

	status, err := simulator.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.ElementCount != 64 {
		t.Errorf("Expected element count 64, got %d", status.ElementCount)
	}

	if status.FrequencyBand != "2.4GHz" {
		t.Errorf("Expected frequency band 2.4GHz, got %s", status.FrequencyBand)
	}

	if !status.PowerStatus {
		t.Error("Expected power status to be true")
	}
}

func TestSimulator_GetStatus_NotConnected(t *testing.T) {
	simulator := NewSimulator(64, "2.4GHz")
	ctx := context.Background()

	_, err := simulator.GetStatus(ctx)
	if err == nil {
		t.Error("Expected error when not connected")
	}
}

func TestController_Configure(t *testing.T) {
	simulator := NewSimulator(64, "2.4GHz")
	controller := NewController(simulator)

	ctx := context.Background()
	_ = simulator.Connect(ctx)

	req := &model.IRSConfigRequest{
		Name:          "test-config",
		ElementCount:  64,
		PhaseShifts:   make([]float64, 64),
		FrequencyBand: "2.4GHz",
	}

	err := controller.Configure(ctx, req)
	if err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	config := controller.GetCurrentConfig()
	if config == nil {
		t.Fatal("Expected config to be set")
	}

	if config.Name != "test-config" {
		t.Errorf("Expected name 'test-config', got '%s'", config.Name)
	}
}

func TestController_GetStatus(t *testing.T) {
	simulator := NewSimulator(64, "2.4GHz")
	controller := NewController(simulator)

	ctx := context.Background()
	_ = simulator.Connect(ctx)

	status, err := controller.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.ElementCount != 64 {
		t.Errorf("Expected element count 64, got %d", status.ElementCount)
	}
}

func TestDriverFactory_Create_Simulator(t *testing.T) {
	factory := NewDriverFactory()

	driver, err := factory.Create(DriverTypeSimulator,
		WithElementCount(32),
		WithFrequencyBand("5GHz"),
	)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	ctx := context.Background()
	err = driver.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !driver.IsConnected() {
		t.Error("Expected driver to be connected")
	}
}
