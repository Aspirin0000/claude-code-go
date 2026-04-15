package porting

import (
	"path/filepath"
	"testing"
)

func TestNewBacklog(t *testing.T) {
	pb := NewBacklog("Test Backlog")
	if pb.Title != "Test Backlog" {
		t.Errorf("unexpected title: %s", pb.Title)
	}
	if len(pb.Modules) != 0 {
		t.Error("expected empty modules")
	}
}

func TestAddAndGetModule(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "mod1", Status: StatusPending})

	m, ok := pb.GetModule("mod1")
	if !ok {
		t.Fatal("expected to find mod1")
	}
	if m.Status != StatusPending {
		t.Errorf("unexpected status: %s", m.Status)
	}
}

func TestUpdateModule(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "mod1", Status: StatusPending})

	ok := pb.UpdateModule("mod1", func(m *PortingModule) {
		m.Status = StatusComplete
	})
	if !ok {
		t.Error("expected update to succeed")
	}

	m, _ := pb.GetModule("mod1")
	if m.Status != StatusComplete {
		t.Errorf("expected status complete, got %s", m.Status)
	}
}

func TestGetByStatus(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "m1", Status: StatusPending})
	pb.AddModule(PortingModule{Name: "m2", Status: StatusComplete})
	pb.AddModule(PortingModule{Name: "m3", Status: StatusPending})

	pending := pb.GetByStatus(StatusPending)
	if len(pending) != 2 {
		t.Errorf("expected 2 pending, got %d", len(pending))
	}
}

func TestGetBlockedItems(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "m1", Status: StatusPending, Blockers: []string{"dep"}})
	pb.AddModule(PortingModule{Name: "m2", Status: StatusPending})

	blocked := pb.GetBlockedItems()
	if len(blocked) != 1 {
		t.Errorf("expected 1 blocked, got %d", len(blocked))
	}
}

func TestRemoveModule(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "m1"})

	if !pb.RemoveModule("m1") {
		t.Error("expected remove to succeed")
	}
	if pb.RemoveModule("m1") {
		t.Error("expected remove to fail for missing module")
	}
}

func TestGetStats(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "m1", Status: StatusComplete, LinesTS: 100, LinesGo: 80})
	pb.AddModule(PortingModule{Name: "m2", Status: StatusPending, LinesTS: 50, LinesGo: 40})

	stats := pb.GetStats()
	if stats.Total != 2 {
		t.Errorf("expected total 2, got %d", stats.Total)
	}
	if stats.TotalLinesTS != 150 {
		t.Errorf("expected 150 TS lines, got %d", stats.TotalLinesTS)
	}
	if stats.TotalLinesGo != 120 {
		t.Errorf("expected 120 Go lines, got %d", stats.TotalLinesGo)
	}
	if stats.PercentDone != 50.0 {
		t.Errorf("expected 50%% done, got %f", stats.PercentDone)
	}
}

func TestGetNextBatch(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "m1", Status: StatusPending, Priority: 1})
	pb.AddModule(PortingModule{Name: "m2", Status: StatusPending, Priority: 3})
	pb.AddModule(PortingModule{Name: "m3", Status: StatusInProgress, Priority: 5})
	pb.AddModule(PortingModule{Name: "m4", Status: StatusPending, Priority: 2, Blockers: []string{"dep"}})

	batch := pb.GetNextBatch(10)
	if len(batch) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(batch))
	}
	if batch[0].Name != "m2" {
		t.Errorf("expected m2 first (highest priority), got %s", batch[0].Name)
	}
}

func TestLoadSaveBacklog(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "m1", Status: StatusComplete})

	tmpFile := filepath.Join(t.TempDir(), "backlog.json")
	if err := SaveBacklog(pb, tmpFile); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	loaded, err := LoadBacklog(tmpFile)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if loaded.Title != "Test" {
		t.Errorf("unexpected title: %s", loaded.Title)
	}
	if len(loaded.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(loaded.Modules))
	}
}

func TestLoadBacklogMissingFile(t *testing.T) {
	_, err := LoadBacklog("/nonexistent/backlog.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestUpdateModuleStatus(t *testing.T) {
	pb := NewBacklog("Test")
	pb.AddModule(PortingModule{Name: "m1", Status: StatusPending})

	pb.UpdateModuleStatus("m1", StatusInProgress)
	m, _ := pb.GetModule("m1")
	if m.Status != StatusInProgress {
		t.Errorf("expected in-progress, got %s", m.Status)
	}
}
