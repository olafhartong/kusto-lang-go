package kustolang

import (
	"fmt"
	"sync"
	"testing"
)

func TestKustoCodeServiceGetAnalyzersReturnsIndependentDiagnostics(t *testing.T) {
	service := newAnalyzerInfoTestService()

	first := service.GetAnalyzers()
	second := service.GetAnalyzers()
	assertAnalyzerInfoTestSnapshot(t, first)
	assertAnalyzerInfoTestSnapshot(t, second)

	first[0].Name = "changed analyzer"
	first[0].Diagnostics[0] = NewDiagnostic("KS999", "changed diagnostic")

	assertAnalyzerInfoTestSnapshot(t, second)
	assertAnalyzerInfoTestSnapshot(t, service.GetAnalyzers())
}

func TestKustoCodeServiceGetAnalyzersConcurrentMutationIsIsolated(t *testing.T) {
	service := newAnalyzerInfoTestService()
	assertAnalyzerInfoTestSnapshot(t, service.GetAnalyzers()) // Populate the shared cache before concurrency.

	const (
		workers    = 32
		iterations = 100
	)
	start := make(chan struct{})
	errors := make(chan error, workers)
	var wait sync.WaitGroup
	wait.Add(workers)
	for worker := 0; worker < workers; worker++ {
		worker := worker
		go func() {
			defer wait.Done()
			<-start
			for iteration := 0; iteration < iterations; iteration++ {
				infos := service.GetAnalyzers()
				if err := analyzerInfoTestSnapshotError(infos); err != nil {
					errors <- fmt.Errorf("worker %d iteration %d: %w", worker, iteration, err)
					return
				}
				infos[0].Diagnostics[0] = NewDiagnostic(fmt.Sprintf("KS%03d", worker), "caller mutation")
			}
		}()
	}
	close(start)
	wait.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}

	assertAnalyzerInfoTestSnapshot(t, service.GetAnalyzers())
}

func newAnalyzerInfoTestService() *KustoCodeService {
	analyzer := NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "test analyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{NewDiagnostic("KS998", "metadata")}
		},
	})
	return NewKustoCodeService("print 1").WithAnalyzers([]KustoAnalyzer{analyzer})
}

func assertAnalyzerInfoTestSnapshot(t *testing.T, infos []AnalyzerInfo) {
	t.Helper()
	if err := analyzerInfoTestSnapshotError(infos); err != nil {
		t.Fatal(err)
	}
}

func analyzerInfoTestSnapshotError(infos []AnalyzerInfo) error {
	if len(infos) != 1 {
		return fmt.Errorf("analyzer infos length = %d, want 1", len(infos))
	}
	if infos[0].Name != "test analyzer" {
		return fmt.Errorf("analyzer name = %q, want %q", infos[0].Name, "test analyzer")
	}
	if len(infos[0].Diagnostics) != 1 || infos[0].Diagnostics[0].Code != "KS998" {
		return fmt.Errorf("analyzer diagnostics = %#v, want KS998 metadata", infos[0].Diagnostics)
	}
	return nil
}
