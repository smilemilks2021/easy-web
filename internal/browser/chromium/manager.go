package chromium

import (
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct {
	Dir     string
	Current string
}

func NewManager(dir, current string) *Manager { return &Manager{Dir: dir, Current: current} }

func (m *Manager) List() ([]string, error) {
	entries, err := os.ReadDir(m.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var revs []string
	for _, e := range entries {
		if e.IsDir() {
			revs = append(revs, e.Name())
		}
	}
	return revs, nil
}

func (m *Manager) Info() {
	revs, _ := m.List()
	fmt.Printf("Chromium cache: %s\nCurrent:        %s\n", m.Dir, m.Current)
	for _, r := range revs {
		mark := "  "
		if r == m.Current {
			mark = "* "
		}
		fmt.Printf("  %s%s\n", mark, r)
	}
}

func (m *Manager) Clean() error {
	revs, err := m.List()
	if err != nil {
		return err
	}
	for _, r := range revs {
		if r == m.Current {
			continue
		}
		path := filepath.Join(m.Dir, r)
		fmt.Printf("Removing %s\n", path)
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) ExecutablePath() string {
	return findExecutable(filepath.Join(m.Dir, m.Current))
}
