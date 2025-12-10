package parser

import (
	"7dtd-monitor/internal/model"
	"regexp"
	"strconv"
	"strings"
)

// Regex for 'lpi' output
// Example: 1. id=171, Survivor (PL), pos=..., health=150, deaths=2, zombies=0, players=0, score=15, level=13, steamid=..., ip=..., ping=24
var playerRegex = regexp.MustCompile(`id=(\d+),\s+(.*?),\s+pos=.*health=(\d+),\s+deaths=(\d+),\s+zombies=(\d+),\s+players=.*score=(\d+),\s+level=(\d+),\s+steamid=(\d+),\s+ip=([\d\.]+),\s+ping=(\d+)`)

func ParsePlayers(output string) ([]model.Player, error) {
	var players []model.Player
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		matches := playerRegex.FindStringSubmatch(line)
		if len(matches) > 10 {
			// clean up name (removes potential trailing commas or spaces if regex is greedy)
			// name is at index 2
			
			level, _ := strconv.Atoi(matches[7])
			health, _ := strconv.Atoi(matches[3])
			deaths, _ := strconv.Atoi(matches[4])
			zombies, _ := strconv.Atoi(matches[5])
			score, _ := strconv.Atoi(matches[6])
			ping, _ := strconv.Atoi(matches[10])

			p := model.Player{
				ID:      matches[1],
				Name:    matches[2],
				Level:   level,
				Health:  health,
				Deaths:  deaths,
				Zombies: zombies,
				Score:   score,
				Ping:    ping,
				SteamID: matches[8],
				IP:      matches[9],
			}
			players = append(players, p)
		}
	}
	return players, nil
}

func ParseMem(output string) (heapUsed string, heapMax string) {
	// Example: Heap: 2500.5 MB, Max: 3500.0 MB...
	// Simple string contains check or crude split for now
	if strings.Contains(output, "Heap:") {
		parts := strings.Split(output, ",")
		if len(parts) >= 2 {
			heapUsed = strings.TrimPrefix(strings.TrimSpace(parts[0]), "Heap: ")
			heapMax = strings.TrimPrefix(strings.TrimSpace(parts[1]), "Max: ")
		}
	}
	return
}

func ParseTime(output string) string {
	// Example: Day 7, 21:45
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 {
		return lines[0] // usually just one line response for gettime
	}
	return ""
}
