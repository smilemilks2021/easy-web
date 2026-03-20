package cookie

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Entry struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain"`
	Path     string    `json:"path"`
	Expires  time.Time `json:"expires"`
	Secure   bool      `json:"secure"`
	HTTPOnly bool      `json:"http_only"`
}

type Cache struct{ dir string }

func NewCache(dir string) *Cache { return &Cache{dir: dir} }

func (c *Cache) file(domain string) string {
	safe := strings.ReplaceAll(strings.ReplaceAll(domain, ":", "_"), "/", "_")
	return filepath.Join(c.dir, filepath.Base(safe)+".json")
}

func (c *Cache) Save(domain string, entries []*Entry) error {
	if err := os.MkdirAll(c.dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.file(domain), data, 0600)
}

func (c *Cache) Load(domain string) ([]*Entry, error) {
	data, err := os.ReadFile(c.file(domain))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []*Entry
	return entries, json.Unmarshal(data, &entries)
}

func (c *Cache) Delete(domain string) error { return os.Remove(c.file(domain)) }

func (c *Cache) List() ([]string, error) {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var domains []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			domains = append(domains, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return domains, nil
}

func (c *Cache) Clear() error { return os.RemoveAll(c.dir) }

// CookieHeader builds a "Cookie: name=value; ..." header string from entries.
func CookieHeader(entries []*Entry) string {
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		parts = append(parts, e.Name+"="+e.Value)
	}
	return strings.Join(parts, "; ")
}
