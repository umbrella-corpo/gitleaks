package git

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gitleaks/go-gitdiff/gitdiff"
	"github.com/rs/zerolog/log"
)

func GitLog(source string, logOpts string) (<-chan *gitdiff.File, error) {
	sourceClean := filepath.Clean(source)
	var cmd *exec.Cmd
	if logOpts != "" {
		args := []string{"-C", sourceClean, "log", "-p", "-U0"}
		args = append(args, strings.Split(logOpts, " ")...)
		cmd = exec.Command("git", args...)
	} else {
		cmd = exec.Command("git", "-C", sourceClean, "log", "-p", "-U0", "--full-history", "--simplify-merges", "--show-pulls", "--all")
	}

	log.Debug().Msgf("executing: %s", cmd.String())

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go listenForStdErr(stderr)

	return gitdiff.Parse(stdout)
}

func GitDiff(source string) (<-chan *gitdiff.File, error) {
	sourceClean := filepath.Clean(source)
	cmd := exec.Command("git", "-C", sourceClean, "diff", "-U0", ".")

	log.Debug().Msgf("executing: %s", cmd.String())

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go listenForStdErr(stderr)

	return gitdiff.Parse(stdout)
}

// listenForStdErr listens for stderr output from git and prints it to stdout
// then exits with exit code 1
func listenForStdErr(stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	errEncountered := false
	for scanner.Scan() {
		log.Error().Msg(scanner.Text())
		errEncountered = true
	}
	if errEncountered {
		os.Exit(1)
	}
}