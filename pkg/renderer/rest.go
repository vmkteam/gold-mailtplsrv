package renderer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
)

func fixturePath(fixture string) string {
	return "tmpl/fixture/" + fixture + ".json"
}

// fixtureExists looks the fixture up where the templates are: in embed mode a container has no tmpl/ on disk.
func (r *Renderer) fixtureExists(fixture string) bool {
	_, err := fs.Stat(r.fsys(), fixturePath(fixture))
	return err == nil
}

// EchoTestTemplate renders a template for testing purposes.
// It accepts a template path as URL parameter and optional fixture data via query parameter.
// Returns HTML response with rendered template or error.
func (r *Renderer) EchoTestTemplate(c echo.Context) error {
	var data map[string]any

	templateName := c.Param("path")
	fixture := c.QueryParam("fixture")
	hasDefaultFixture := r.fixtureExists(templateName)

	// set default fixture or from param
	if fixture != "" {
		// check for existance
		if !r.fixtureExists(fixture) {
			return echo.NewHTTPError(http.StatusNotFound, "fixture not found")
		}
	} else if hasDefaultFixture {
		fixture = templateName
	}

	// parse fixture
	if fixture != "" {
		f, err := fs.ReadFile(r.fsys(), fixturePath(fixture))
		if err != nil {
			return fmt.Errorf("read fixture failed: %w", err)
		}

		if err = json.Unmarshal(f, &data); err != nil {
			return fmt.Errorf("parse fixture failed: %w", err)
		}
	}

	// render request
	result, err := r.Render(c.Request().Context(), templateName, data)
	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, result)
}

// EchoRenderTemplate renders a template with JSON body data.
// This is the backward-compatible REST endpoint (POST /template/:name).
func (r *Renderer) EchoRenderTemplate(c echo.Context) error {
	templateName := c.Param("name")

	var data map[string]any
	if err := json.NewDecoder(c.Request().Body).Decode(&data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json body")
	}

	result, err := r.Render(c.Request().Context(), templateName, data)
	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, result)
}

// EchoList returns an HTML page listing all available templates.
// The list is sorted alphabetically and includes embedded template indicator if applicable.
func (r *Renderer) EchoList(c echo.Context) error {
	paths, err := r.List()
	if err != nil {
		return err
	}
	sort.Strings(paths)

	// render template
	var (
		sb        strings.Builder
		embedTmpl string
	)

	if r.cfg.IsEmbed {
		embedTmpl = " <i>[embed]</i>"
	}

	fmt.Fprintf(&sb, "<html><body><h1>templates test%s</h1><ul>", embedTmpl)
	for _, path := range paths {
		fmt.Fprintf(&sb, `<li><a href="test/%s">%s</a></li>`, path, path)
	}
	sb.WriteString("</ul></body></html>")

	return c.HTML(http.StatusOK, sb.String())
}
