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

func ParseMem(output string) (heapUsed string, heapMax string, fps string) {
	// Example: Heap: 2500.5 MB, Max: 3500.0 MB, ... FPS: 58.4
	if strings.Contains(output, "Heap:") {
		output = strings.ReplaceAll(output, "\n", " ")
		parts := strings.Split(output, ",")
		if len(parts) >= 2 {
			heapUsed = strings.TrimPrefix(strings.TrimSpace(parts[0]), "Heap: ")
			// Max is bit trickier if order varies, but usually 2nd.
			// Let's rely on string searching for robustness
		}
	}

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
	// Simple line scan
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "type=zombie") {
			zombies++
		} else if strings.Contains(lower, "type=animal") {
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
	// Example: Day 7, 21:45
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 {
		return lines[0] // usually just one line response for gettime
	}
	return ""
}
