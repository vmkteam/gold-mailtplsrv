package renderer

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// TestEchoTestTemplateFixture checks that the preview renders a template with its fixture in both modes.
// The embed case runs from an empty directory, as a container does: it ships only the binary.
func TestEchoTestTemplateFixture(t *testing.T) {
	tests := []struct {
		fixture string
		want    string
	}{
		{fixture: "", want: "SALEONSALE"},
		{fixture: "test", want: "BLACKFRIDAY"},
	}

	for _, isEmbed := range []bool{false, true} {
		r := newTestRenderer(t, isEmbed)
		if isEmbed {
			t.Chdir(t.TempDir())
		}

		for _, tc := range tests {
			req := httptest.NewRequest(http.MethodGet, "/template/test/welcome?fixture="+tc.fixture, nil)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)
			c.SetParamNames("path")
			c.SetParamValues("welcome")

			if err := r.EchoTestTemplate(c); err != nil {
				t.Fatalf("isEmbed=%v fixture=%q: EchoTestTemplate failed: %v", isEmbed, tc.fixture, err)
			}
			if !strings.Contains(rec.Body.String(), tc.want) {
				t.Errorf("isEmbed=%v fixture=%q: preview does not contain %q", isEmbed, tc.fixture, tc.want)
			}
		}
	}
}
