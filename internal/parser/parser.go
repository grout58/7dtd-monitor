package parser

import (
	"7dtd-monitor/internal/model"
	"regexp"
	"strconv"
	"strings"
)

// Helper to remove log lines from Telnet output
func sanitizeOutput(output string) string {
	lines := strings.Split(output, "\n")
	var cleanLines []string
	for _, line := range lines {
		// Filter out lines that look like logs
		// e.g. "2025-12-10T10:35:08 1991.171 INF Executing command..."
		if strings.Contains(line, "INF Executing command") {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	return strings.Join(cleanLines, "\n")
}

// Robust Key-Value parser instead of strict Regex
func ParsePlayers(output string) ([]model.Player, error) {
	output = sanitizeOutput(output)
	var players []model.Player
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Check for standard ID start
		// e.g. "1. id=..."
		if !strings.Contains(line, "id=") {
			continue
		}

		p := model.Player{}
		stats := make(map[string]string)

		// 7DTD format is essentially comma separated values, usually "key=value"
		// But the name field often doesn't have a key, e.g. "1. id=171, PlayerName, pos=..."
		parts := strings.Split(line, ",")

		for i, part := range parts {
			part = strings.TrimSpace(part)
			kv := strings.SplitN(part, "=", 2)

			if len(kv) == 2 {
				keyRAW := strings.ToLower(strings.TrimSpace(kv[0]))
				// Remove potential "1. " prefix "2. " etc.
				// We trim digits, dots, and spaces from the LEFT only.
				key := strings.TrimLeft(keyRAW, "0123456789. ")

				val := strings.TrimSpace(kv[1])
				stats[key] = val

				// Direct mapping for known IDs if needed, but map approach is safer
			} else if i == 1 {
				// The second field is usually the name if it has no "="
				// e.g. "1. id=..., Name Is Here, pos=..."
				// But sometimes it might be stuck to the previous parts if we split efficiently.
				// However, splitting by comma is usually safe for 7DTD unless name has comma.
				p.Name = part
			}
		}

		// Helper to safely get int
		getInt := func(key string) int {
			if v, ok := stats[key]; ok {
				i, _ := strconv.Atoi(v)
				return i
			}
			return 0
		}

		p.ID = stats["id"]
		// If name wasn't set by position (2nd field), try to see if it was extracted?
		// Actually, standard LPI output: "1. id=171, Name, pos=..."
		// stats["id"] will be "171" (or "171" from "1. id=171"?)
		// The first chunk is often "1. id=171".
		// Let's clean the ID lookup.
		if p.ID == "" {
			// Try to handle "1. id=171" case for first split
			// This logic happens in the loop above.
		}

		p.Level = getInt("level")
		p.Health = getInt("health")
		p.Deaths = getInt("deaths")
		p.Zombies = getInt("zombies")
		p.Score = getInt("score")
		p.Ping = getInt("ping")
		p.SteamID = stats["steamid"]
		p.IP = stats["ip"]

		// Fallback for Name if not found in 2nd position (unlikely but safe)
		if p.Name == "" {
			if n, ok := stats["name"]; ok {
				p.Name = n
			}
		}

		if p.ID != "" {
			// Filter out simple observers (Telnet connections)
			// Observers (like this tool) usually show up as "id=1" with no name.
			// Real players should have a name.
			if p.Name == "" {
				continue
			}
			players = append(players, p)
		}
	}
	return players, nil
}

func ParseMem(output string) (heapUsed string, heapMax string, fps string) {
	// Don't strict sanitize here because ParseMem relies on specific lines,
	// but sanitizing log lines is good.
	// However, the output "Time: 32.55m FPS: 14.07 Heap: ..." is one line.
	// The sanitizer won't hurt it.
	output = sanitizeOutput(output)

	// Example: Heap: 2500.5 MB, Max: 3500.0 MB, ... FPS: 58.4
	// Revert single-line replacement strategy because sanitize handles newlines
	// But let's keep it safe.
	output = strings.ReplaceAll(output, "\n", " ")

	// Robust extract (regex is safer for these mixed strings)
	reHeap := regexp.MustCompile(`Heap:\s*([\d\.]+\s*MB)`)
	reMax := regexp.MustCompile(`Max:\s*([\d\.]+\s*MB)`)
	reFps := regexp.MustCompile(`FPS:\s*([\d\.]+)`)

	if m := reHeap.FindStringSubmatch(output); len(m) > 1 {
		heapUsed = m[1]
	}
	if m := reMax.FindStringSubmatch(output); len(m) > 1 {
		heapMax = m[1]
	}
	if m := reFps.FindStringSubmatch(output); len(m) > 1 {
		fps = m[1]
	}

	return
}

func ParseEntities(output string) (zombies, animals, other int) {
	output = sanitizeOutput(output)
	// Simple line scan
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "type=zombie") {
			zombies++
		} else if strings.Contains(lower, "type=animal") || strings.Contains(lower, "type=entityanimal") {
			animals++
		} else if strings.Contains(lower, "type=player") {
			// ignore, we get players from lpi
		} else if strings.Contains(lower, "id=") {
			other++
		}
	}
	return
}

func ParseTime(output string) string {
	output = sanitizeOutput(output)
	// Example: Day 7, 21:45
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 {
		return lines[0] // usually just one line response for gettime after sanitize
	}
	return ""
}
