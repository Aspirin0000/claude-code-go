package porting

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type PortingStatus string

const (
	StatusPending    PortingStatus = "pending"
	StatusInProgress PortingStatus = "in-progress"
	StatusSkeleton   PortingStatus = "skeleton"
	StatusComplete   PortingStatus = "complete"
)

type PortingModule struct {
	Name         string        `json:"name"`
	SourceFile   string        `json:"sourceFile"`
	TargetFile   string        `json:"targetFile"`
	Status       PortingStatus `json:"status"`
	Priority     int           `json:"priority"`
	LinesTS      int           `json:"linesTS"`
	LinesGo      int           `json:"linesGo"`
	Dependencies []string      `json:"dependencies"`
	Blockers     []string      `json:"blockers"`
}

type PortingBacklog struct {
	Title   string          `json:"title"`
	Modules []PortingModule `json:"modules"`
}

func NewBacklog(title string) *PortingBacklog {
	return &PortingBacklog{
		Title:   title,
		Modules: []PortingModule{},
	}
}

func (pb *PortingBacklog) AddModule(module PortingModule) {
	pb.Modules = append(pb.Modules, module)
}

func (pb *PortingBacklog) UpdateModule(name string, updater func(*PortingModule)) bool {
	for i := range pb.Modules {
		if pb.Modules[i].Name == name {
			updater(&pb.Modules[i])
			return true
		}
	}
	return false
}

func (pb *PortingBacklog) GetByStatus(status PortingStatus) []PortingModule {
	var result []PortingModule
	for _, m := range pb.Modules {
		if m.Status == status {
			result = append(result, m)
		}
	}
	return result
}

func (pb *PortingBacklog) GetBlockedItems() []PortingModule {
	var result []PortingModule
	for _, m := range pb.Modules {
		if len(m.Blockers) > 0 {
			result = append(result, m)
		}
	}
	return result
}

func (pb *PortingBacklog) GetModule(name string) (PortingModule, bool) {
	for _, m := range pb.Modules {
		if m.Name == name {
			return m, true
		}
	}
	return PortingModule{}, false
}

func (pb *PortingBacklog) RemoveModule(name string) bool {
	for i, m := range pb.Modules {
		if m.Name == name {
			pb.Modules = append(pb.Modules[:i], pb.Modules[i+1:]...)
			return true
		}
	}
	return false
}

type BacklogStats struct {
	Total        int            `json:"total"`
	ByStatus     map[string]int `json:"byStatus"`
	TotalLinesTS int            `json:"totalLinesTS"`
	TotalLinesGo int            `json:"totalLinesGo"`
	PercentDone  float64        `json:"percentDone"`
}

func (pb *PortingBacklog) GetStats() BacklogStats {
	stats := BacklogStats{
		ByStatus: make(map[string]int),
	}

	for _, m := range pb.Modules {
		stats.Total++
		stats.ByStatus[string(m.Status)]++
		stats.TotalLinesTS += m.LinesTS
		stats.TotalLinesGo += m.LinesGo

		if m.Status == StatusComplete {
			stats.PercentDone += 1.0
		}
	}

	if stats.Total > 0 {
		stats.PercentDone = (stats.PercentDone / float64(stats.Total)) * 100
	}

	return stats
}

func (pb *PortingBacklog) GetNextBatch(limit int) []PortingModule {
	var candidates []PortingModule

	for _, m := range pb.Modules {
		if m.Status == StatusPending && len(m.Blockers) == 0 {
			candidates = append(candidates, m)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Priority > candidates[j].Priority
	})

	if len(candidates) > limit {
		return candidates[:limit]
	}
	return candidates
}

func (pb *PortingBacklog) UpdateModuleStatus(name string, status PortingStatus) bool {
	return pb.UpdateModule(name, func(m *PortingModule) {
		m.Status = status
	})
}

func LoadBacklog(filepath string) (*PortingBacklog, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backlog file: %w", err)
	}

	var backlog PortingBacklog
	if err := json.Unmarshal(data, &backlog); err != nil {
		return nil, fmt.Errorf("failed to parse backlog JSON: %w", err)
	}

	return &backlog, nil
}

func SaveBacklog(backlog *PortingBacklog, filepath string) error {
	data, err := json.MarshalIndent(backlog, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backlog: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backlog file: %w", err)
	}

	return nil
}
