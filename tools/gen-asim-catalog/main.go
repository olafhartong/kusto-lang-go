// gen-asim-catalog refreshes the offline ASIM metadata from a local checkout of
// Microsoft's Azure/Azure-Sentinel repository. It never executes parser queries.
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
	"go.yaml.in/yaml/v3"
)

type parser struct {
	Name          string `yaml:"ParserName"`
	BuiltIn       string `yaml:"EquivalentBuiltInParser"`
	Normalization struct {
		Schema string `yaml:"Schema"`
	} `yaml:"Normalization"`
	Parameters []struct {
		Name    string    `yaml:"Name"`
		Type    string    `yaml:"Type"`
		Default yaml.Node `yaml:"Default"`
	} `yaml:"ParserParams"`
}

type definition struct {
	Schema    string                        `json:"schema"`
	Arguments []environment.ArgumentDetails `json:"arguments"`
}

type catalog struct {
	Source    string                              `json:"source"`
	Revision  string                              `json:"revision"`
	Schemas   map[string]environment.TableDetails `json:"schemas"`
	Functions map[string]definition               `json:"functions"`
}

func main() {
	root := flag.String("source-dir", "", "Local Azure/Azure-Sentinel checkout (required)")
	output := flag.String("output", "kustolang/environment/sentinel_catalog.json", "Catalog output path")
	flag.Parse()
	if err := run(*root, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root, output string) error {
	if root == "" {
		return fmt.Errorf("--source-dir is required")
	}
	revision, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("read upstream revision: %w", err)
	}
	c, err := generate(root)
	if err != nil {
		return err
	}
	c.Revision = strings.TrimSpace(string(revision))
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(output, append(data, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("Wrote %d parser names across %d schemas to %s (revision %s)\n", len(c.Functions), len(c.Schemas), output, c.Revision)
	return nil
}

func generate(root string) (catalog, error) {
	c := catalog{Source: "https://github.com/Azure/Azure-Sentinel", Schemas: map[string]environment.TableDetails{}, Functions: map[string]definition{}}
	file, err := os.Open(filepath.Join(root, "ASIM", "dev", "ASimTester", "ASimTester.csv"))
	if err != nil {
		return c, err
	}
	defer func() { _ = file.Close() }()
	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return c, err
	}
	indexes := map[string]int{}
	for i, name := range header {
		indexes[strings.TrimPrefix(name, "\ufeff")] = i
	}
	for _, name := range []string{"ColumnName", "ColumnType", "Schema"} {
		if _, ok := indexes[name]; !ok {
			return c, fmt.Errorf("ASimTester.csv missing %s column", name)
		}
	}
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return c, err
		}
		name, typ, schema := row[indexes["ColumnName"]], row[indexes["ColumnType"]], row[indexes["Schema"]]
		if typ == "double" {
			typ = "real"
		}
		if c.Schemas[schema] == nil {
			c.Schemas[schema] = environment.TableDetails{}
		}
		if previous := c.Schemas[schema][name]; previous != "" && previous != typ {
			return c, fmt.Errorf("schema %s column %s has conflicting types %s and %s", schema, name, previous, typ)
		}
		c.Schemas[schema][name] = typ
	}
	common := c.Schemas["Common"]
	delete(c.Schemas, "Common")
	for _, columns := range c.Schemas {
		for name, typ := range common {
			if _, exists := columns[name]; !exists {
				columns[name] = typ
			}
		}
	}
	paths, err := filepath.Glob(filepath.Join(root, "Parsers", "ASim*", "Parsers", "*.yaml"))
	if err != nil {
		return c, err
	}
	if len(paths) == 0 || len(common) == 0 {
		return c, fmt.Errorf("upstream checkout has no ASIM parser definitions or common schema")
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return c, err
		}
		var p parser
		if err := yaml.Unmarshal(data, &p); err != nil {
			return c, fmt.Errorf("%s: %w", path, err)
		}
		if p.Name == "" || p.BuiltIn == "" || c.Schemas[p.Normalization.Schema] == nil {
			return c, fmt.Errorf("%s: missing parser name, built-in name, or known schema %q", path, p.Normalization.Schema)
		}
		d := definition{Schema: p.Normalization.Schema, Arguments: []environment.ArgumentDetails{}}
		for _, param := range p.Parameters {
			if param.Name == "" || param.Type == "" {
				return c, fmt.Errorf("%s: incomplete parameter", path)
			}
			d.Arguments = append(d.Arguments, environment.ArgumentDetails{Name: param.Name, Type: param.Type, Optional: param.Default.Kind != 0})
		}
		for _, name := range []string{p.Name, p.BuiltIn} {
			if _, duplicate := c.Functions[name]; duplicate {
				return c, fmt.Errorf("%s: duplicate parser name %q", path, name)
			}
			c.Functions[name] = d
		}
	}
	return c, nil
}
