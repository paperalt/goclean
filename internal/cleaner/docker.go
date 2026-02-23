package cleaner

import (
	"fmt"
	"os/exec"
	"strings"
)

type DockerCleaner struct{}

func (c *DockerCleaner) Name() string {
	return "Docker System"
}

func (c *DockerCleaner) RequiresRoot() bool {
	return true
}

func (c *DockerCleaner) Scan() (int64, error) {
	// Check if docker exists
	if _, err := exec.LookPath("docker"); err != nil {
		return 0, nil
	}

	// We can't easily get the *exact* reclaimable size of 'docker system prune'
	// without actually running it or parsing a lot of 'docker system df -v' output.
	// 'docker system df' gives a summary.
	
	cmd := exec.Command("docker", "system", "df", "--format", "{{.Reclaimable}}")
	output, err := cmd.Output()
	if err != nil {
		return 0, nil // Docker might not be running or permission denied
	}

	// Output format example: "1.2GB (20%)"
	// We need to parse this. This is tricky to parse exactly to bytes without a heavy library.
	// For now, let's just return a non-zero dummy value if we detect reclaimable text, 
	// or try simple parsing.
	// Actually, just returning 0 with a special "Detected" status in UI might be better if we can't parse?
	// But let's try to parse bytes if possible.
	// A simpler approach for MVP: checking if there is ANYTHING reclaimable.
	
	lines := strings.Split(string(output), "\n")
	var totalBytes int64
	
	// Very accumulation string parsing is error prone.
	// Let's rely on 'docker system df --format "{{.Size}}"' of images/containers? No.
	// Let's simplify: if docker exists, we assume we might clean something.
	// Better: just try to read the "Reclaimable" string and log it? The interface expects int64.
	// Let's implement a rough parser for "GB", "MB", "KB".
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// line like "1.23GB (23%)"
		parts := strings.Fields(line)
		if len(parts) > 0 {
			valStr := parts[0]
			bytes := parseSize(valStr)
			totalBytes += bytes
		}
	}

	return totalBytes, nil
}

func parseSize(s string) int64 {
	// s = "1.23GB"
	s = strings.ToUpper(s)
	var mult int64 = 1
	if strings.Contains(s, "KB") {
		mult = 1024
		s = strings.Replace(s, "KB", "", 1)
	} else if strings.Contains(s, "MB") {
		mult = 1024 * 1024
		s = strings.Replace(s, "MB", "", 1)
	} else if strings.Contains(s, "GB") {
		mult = 1024 * 1024 * 1024
		s = strings.Replace(s, "GB", "", 1)
	} else if strings.Contains(s, "B") {
		s = strings.Replace(s, "B", "", 1)
	}
	
	var val float64
	fmt.Sscanf(s, "%f", &val)
	return int64(val * float64(mult))
}

func (c *DockerCleaner) Clean() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found")
	}

	// docker system prune --volumes -f
	// CAREFUL: --volumes removes anonymous volumes. Maybe too aggressive?
	// User asked for "more space". Standard 'system prune' removes unused data.
	cmd := exec.Command("docker", "system", "prune", "-f")
	// cmd.Stdout = os.Stdout // Let's not spam unless needed, or maybe we do?
	return cmd.Run()
}
