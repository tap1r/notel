package tools

import (
	"os"
	"strings"
	"testing"
)

func TestCatalogMarkdownIsStableAndCoversEveryTool(t *testing.T) {
	registry := Registry()
	markdown := CatalogMarkdown(registry)
	for _, tool := range registry {
		if !strings.Contains(markdown, tool.Name) {
			t.Fatalf("catalog markdown missing %s", tool.Name)
		}
		if !strings.Contains(markdown, tool.DocumentationURL) {
			t.Fatalf("catalog markdown missing docs URL for %s", tool.Name)
		}
	}
	if strings.Contains(markdown, "@nuxt/telemetry") && strings.Contains(markdown, "Next.js") {
		// The historical README put a Nuxt command on the Next.js row.
		// The generator must not invent that mix-up. A Nuxt command is fine
		// on the Nuxt row; this assertion is about the Next.js cell.
		next := markdown[strings.Index(markdown, "Next.js"):]
		if end := strings.Index(next, "\n"); end >= 0 {
			next = next[:end]
		}
		if strings.Contains(next, "@nuxt/telemetry") {
			t.Fatalf("Next.js catalog row contains a Nuxt command:\n%s", next)
		}
	}
	if strings.Contains(markdown, "Unset means already off") {
		t.Fatal("catalog still claims an unset variable is already the disabled state")
	}
}

func TestCatalogMarkdownRendersFirstSetInOrder(t *testing.T) {
	markdown := CatalogMarkdown([]Tool{{
		Name:             "GitHub CLI",
		DocumentationURL: "https://example.com",
		DataCollection:   "usage",
		OptOut:           FirstSet,
		EnvChecks: []EnvCheck{
			{Name: "GH_TELEMETRY", ValidValues: []string{"false", ""}, IgnoreCase: true},
			{Name: "DO_NOT_TRACK", ValidValues: []string{"1"}},
		},
	}})
	if strings.Contains(markdown, "`GH_TELEMETRY` / `DO_NOT_TRACK`") {
		t.Fatalf("FirstSet rendered as an OR:\n%s", markdown)
	}
	if !strings.Contains(markdown, "`GH_TELEMETRY`<br>`DO_NOT_TRACK`") {
		t.Fatalf("FirstSet env column missing ordered line break:\n%s", markdown)
	}
	if !strings.Contains(markdown, "`(empty)`") {
		t.Fatalf("blank opt-out missing (empty):\n%s", markdown)
	}
	if !strings.Contains(markdown, "The first variable that is set decides.") {
		t.Fatalf("missing precedence note:\n%s", markdown)
	}
}

func TestREADMECatalogMatchesRegistry(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	const startMarker = "<!-- notel:catalog:start -->"
	const endMarker = "<!-- notel:catalog:end -->"
	text := string(readme)
	start := strings.Index(text, startMarker)
	end := strings.Index(text, endMarker)
	if start < 0 || end < 0 || end < start {
		t.Fatal("README is missing the notel catalog markers")
	}
	got := strings.TrimSpace(text[start+len(startMarker) : end])
	want := strings.TrimSpace(CatalogMarkdown(Registry()))
	if got != want {
		t.Fatalf("README catalog drifted from the registry.\nRegenerate it with `go run . list --format=markdown` and replace the block between the catalog markers.\n\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}
