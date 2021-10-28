package detect

import (
	"context"
	"os"
	"path/filepath"

	godocutil "golang.org/x/tools/godoc/util"

	"golang.org/x/sync/errgroup"

	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/report"
)

// FromFiles opens the directory or file specified in source and checks each file against the rules
// from the configuration. If any secrets are found, they are added to the list of findings.
func FromFiles(source string, cfg config.Config, outputOptions Options) ([]report.Finding, error) {
	var findings []report.Finding
	g, _ := errgroup.WithContext(context.Background())
	paths := make(chan string)
	g.Go(func() error {
		defer close(paths)
		return filepath.Walk(source,
			func(path string, fInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if fInfo.Name() == ".git" {
					return filepath.SkipDir
				}
				if fInfo.Mode().IsRegular() {
					paths <- path
				}
				return nil
			})
	})
	for pa := range paths {
		p := pa
		g.Go(func() error {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}

			if !godocutil.IsText(b) {
				return nil
			}
			findings = append(findings, processBytes(cfg, b, filepath.Ext(p))...)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return findings, err
	}

	return findings, nil
}