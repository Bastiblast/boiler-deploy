package logging

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Reader struct {
	environment string
	logDir      string
}

func NewReader(environment string) *Reader {
	return &Reader{
		environment: environment,
		logDir:      filepath.Join("logs", environment),
	}
}

func (r *Reader) GetServerLogs(serverName string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(r.logDir, fmt.Sprintf("%s_*.log", serverName)))
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}

func (r *Reader) GetLatestLog(serverName string) (string, error) {
	files, err := r.GetServerLogs(serverName)
	if err != nil || len(files) == 0 {
		return "", fmt.Errorf("no logs found")
	}

	return files[len(files)-1], nil
}

func (r *Reader) ReadLog(logFile string, maxLines int) ([]string, error) {
	file, err := os.Open(logFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if maxLines > 0 && len(lines) > maxLines {
		return lines[len(lines)-maxLines:], nil
	}

	return lines, nil
}

func (r *Reader) TailLog(logFile string, lines int) ([]string, error) {
	return r.ReadLog(logFile, lines)
}

func (r *Reader) FormatLogLine(line string) string {
	line = strings.TrimSpace(line)
	
	if strings.Contains(line, `"failed"`) || strings.Contains(line, "FAILED") {
		return "❌ " + line
	}
	if strings.Contains(line, `"ok"`) || strings.Contains(line, "SUCCESS") {
		return "✓ " + line
	}
	if strings.Contains(line, `"changed"`) || strings.Contains(line, "CHANGED") {
		return "⚡ " + line
	}

	return line
}

func (r *Reader) GetAllLogs() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(r.logDir, "*.log"))
	if err != nil {
		return nil, err
	}

	sort.Strings(files)
	return files, nil
}
