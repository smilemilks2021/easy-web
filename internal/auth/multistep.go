package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/smilemilks2021/easy-web/internal/browser"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/request"
)

func RunMultiStep(name string) error {
	authCfg, ok := config.C.MultiStepAuth[name]
	if !ok {
		return fmt.Errorf("multi_step_auth %q not found in ~/.easy-web.yaml", name)
	}
	fmt.Printf("Running multi-step auth: %s (%s)\n", name, authCfg.Description)
	vars := map[string]string{}

	for _, step := range authCfg.Steps {
		fmt.Printf("  [%s] type=%s\n", step.ID, step.Type)
		var err error
		switch step.Type {
		case "browser_capture":
			err = execBrowserCapture(step, vars)
		case "http_request":
			err = execHTTPRequest(step, vars)
		default:
			err = fmt.Errorf("unknown step type: %s", step.Type)
		}
		if err != nil {
			return fmt.Errorf("step %s: %w", step.ID, err)
		}
	}

	fmt.Println("Extracted variables:")
	for k, v := range vars {
		disp := v
		if len(disp) > 40 {
			disp = disp[:40] + "..."
		}
		fmt.Printf("  %s = %s\n", k, disp)
	}
	return nil
}

func execBrowserCapture(step config.MultiStepStep, vars map[string]string) error {
	d := browser.NewDriver(browser.Options{Headless: false, ReuseProfile: true, ProfileDir: config.ProfileDir()})
	entries, err := d.LoginAndGetCookies(step.URL, 5*time.Minute)
	if err != nil {
		return err
	}

	for _, ex := range step.Extract {
		switch ex.Source {
		case "cookie":
			for _, c := range entries {
				if c.Name == ex.Key {
					vars[ex.Variable] = c.Value
					break
				}
			}
		case "localStorage":
			val, err := d.ExtractLocalStorageToken(step.URL, []string{ex.Key})
			if err == nil {
				vars[ex.Variable] = val
			}
		}
	}
	return nil
}

func execHTTPRequest(step config.MultiStepStep, vars map[string]string) error {
	reqURL := interpolate(step.URL, vars)
	headers := map[string]string{}
	for k, v := range step.Headers {
		headers[k] = interpolate(v, vars)
	}

	method := step.Method
	if method == "" {
		method = "GET"
	}

	c, err := request.NewClient(nil, headers)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}
	resp, err := c.Do(method, reqURL, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	for _, ex := range step.Extract {
		switch ex.Source {
		case "header":
			vars[ex.Variable] = resp.Header.Get(ex.Key)
		case "json_response":
			vars[ex.Variable] = extractJSONPath(body, ex.Key)
		}
	}
	return nil
}

func extractJSONPath(body []byte, path string) string {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}
	parts := strings.Split(path, ".")
	current := data
	for _, p := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = m[p]
	}
	if s, ok := current.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", current)
}

func interpolate(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	return s
}
