package ui

import (
	"7dtd-monitor/internal/model" // Added for model.Player
	"7dtd-monitor/internal/parser"
	"7dtd-monitor/internal/telnet"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	TviewApp     *tview.Application
	Client       *telnet.Client
	StatsText    *tview.TextView
	PlayersTable *tview.Table
	LogView      *tview.TextView
	Input        *tview.InputField
}

func NewApp(client *telnet.Client) *App {
	app := &App{
		TviewApp: tview.NewApplication(),
		Client:   client,
	}
	app.setupUI()
	return app
}

func (a *App) setupUI() {
	// 1. Stats View
	a.StatsText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText("Connecting...")
	a.StatsText.SetBorder(true).SetTitle(" Server Stats ")

	// 2. Players Table
	a.PlayersTable = tview.NewTable().
		SetBorders(true).
		SetSelectable(true, false) // Enable row selection
	a.PlayersTable.SetBorder(true).SetTitle(" Online Players ")

	// 3. Log/Debug View (bottom)
	a.LogView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetMaxLines(1000) // keep history
	a.LogView.SetBorder(true).SetTitle(" Server Log ")

	// 4. Input Field
	a.Input = tview.NewInputField().
		SetLabel("> ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	a.Input.SetBorder(true).SetTitle(" Admin Console ")

	a.Input.SetDoneFunc(func(key tcell.Key) { // Changed tview.Key to tcell.Key
		if key == tcell.KeyEnter {
			cmd := a.Input.GetText()
			if cmd == "" {
				return
			}
			a.Input.SetText("") // Clear

			// Exec Async
			go func() {
				a.TviewApp.QueueUpdateDraw(func() {
					a.LogView.Write([]byte(fmt.Sprintf("[yellow]> %s[white]\n", cmd)))
				})

				resp, err := a.Client.SendCommand(cmd)
				if err != nil {
					a.TviewApp.QueueUpdateDraw(func() {
						a.LogView.Write([]byte(fmt.Sprintf("[red]Error: %v[white]\n", err)))
						a.LogView.ScrollToEnd()
					})
				} else {
					a.TviewApp.QueueUpdateDraw(func() {
						_, logs := parser.SplitLogs(resp)
						for _, l := range logs {
							a.LogView.Write([]byte(fmt.Sprintf("[gray]%s[white]\n", l)))
						}
						clean, _ := parser.SplitLogs(resp)
						if clean != "" {
							a.LogView.Write([]byte(fmt.Sprintf("%s\n", clean)))
						}
						a.LogView.ScrollToEnd()
					})
				}
			}()
		}
	})

	// Layout: Flex
	// Top: Stats (Fixed Height?), Middle: Log/Players, Bottom: Input
	// Let's go with:
	// Top: Stats (25%) | Players (75%)
	// Middle: Log (Rest)
	// Bottom: Input (3 lines)

	topFlex := tview.NewFlex().
		AddItem(a.StatsText, 0, 1, false).
		AddItem(a.PlayersTable, 0, 2, true) // Focus table by default?

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topFlex, 15, 1, false).
		AddItem(a.LogView, 0, 1, false).
		AddItem(a.Input, 3, 1, true)

	a.TviewApp.SetRoot(mainFlex, true).SetFocus(a.Input)
}

func (a *App) Run() error {
	// Start refresh loop
	go a.refreshLoop()

	if err := a.Client.Connect(); err != nil {
		return err
	}

	return a.TviewApp.Run()
}

func (a *App) refreshLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		a.updateData()
	}
}

func (a *App) processResponse(cmd string) (string, error) {
	raw, err := a.Client.SendCommand(cmd)
	if err != nil {
		return "", err
	}
	clean, logs := parser.SplitLogs(raw)

	// Append logs to view safely
	a.TviewApp.QueueUpdateDraw(func() {
		for _, l := range logs {
			if isImportantLog(l) {
				// Colorize?
				color := "[gray]"
				if strings.Contains(l, "ERR") {
					color = "[red]"
				} else if strings.Contains(l, "Chat") {
					color = "[green]"
				} else if strings.Contains(l, "WRN") {
					color = "[yellow]"
				}

				a.LogView.Write([]byte(fmt.Sprintf("%s%s[white]\n", color, l)))
			}
		}
		a.LogView.ScrollToEnd()
	})

	return clean, nil
}

func (a *App) updateData() {
	// 1. Get Time
	timeStr, _ := a.processResponse("gettime")
	gameTime := parser.ParseTime(timeStr)

	// 2. Get Mem & FPS
	memStr, _ := a.processResponse("mem")
	heap, max, fps := parser.ParseMem(memStr)

	// 3. Get Players
	playersStr, _ := a.processResponse("lp")
	players, _ := parser.ParsePlayers(playersStr)

	// 4. Get Entities
	entStr, _ := a.processResponse("le")
	zombies, animals, _ := parser.ParseEntities(entStr)

	// Calculate Avg Ping
	var totalPing int
	for _, p := range players {
		totalPing += p.Ping
	}
	avgPing := 0
	if len(players) > 0 {
		avgPing = totalPing / len(players)
	}

	a.TviewApp.QueueUpdateDraw(func() {
		// Update Stats
		statsText := fmt.Sprintf("\n [green]Host:[white] %s\n [green]Port:[white] %s\n\n [yellow]Game Time:[white] %s\n [yellow]Server FPS:[white] %s\n\n [blue]Heap:[white] %s / %s\n [blue]Players:[white] %d\n [blue]Avg Ping:[white] %d ms\n\n [red]Zombies:[white] %d\n [green]Animals:[white] %d",
			a.Client.Host, a.Client.Port, gameTime, fps, heap, max, len(players), avgPing, zombies, animals)
		a.StatsText.SetText(statsText)

		// Update Table
		a.PlayersTable.Clear()
		headers := []string{"ID", "Name", "Score", "Lvl", "Z-Kills", "P-Kills", "Deaths", "Ping", "IP"}
		for i, h := range headers {
			a.PlayersTable.SetCell(0, i,
				tview.NewTableCell(h).
					SetTextColor(tview.Styles.SecondaryTextColor).
					SetAlign(tview.AlignCenter).
					SetSelectable(false))
		}

		for i, p := range players {
			row := i + 1
			// Helper to make cells
			c := func(text string) *tview.TableCell {
				return tview.NewTableCell(text).SetTextColor(tview.Styles.PrimaryTextColor)
			}
			center := func(text string) *tview.TableCell {
				return tview.NewTableCell(text).SetTextColor(tview.Styles.PrimaryTextColor).SetAlign(tview.AlignCenter)
			}

			// Store Player Struct or ID in the reference for actions
			// We store ID in the first cell
			idCell := c(p.ID)
			idCell.SetReference(p) // Store full player object

			a.PlayersTable.SetCell(row, 0, idCell)
			a.PlayersTable.SetCell(row, 1, c(p.Name))
			a.PlayersTable.SetCell(row, 2, center(fmt.Sprintf("%d", p.Score)))
			a.PlayersTable.SetCell(row, 3, center(fmt.Sprintf("%d", p.Level)))
			a.PlayersTable.SetCell(row, 4, center(fmt.Sprintf("%d", p.Zombies)))     // Z-Kills
			a.PlayersTable.SetCell(row, 5, center(fmt.Sprintf("%d", p.PlayerKills))) // P-Kills
			a.PlayersTable.SetCell(row, 6, center(fmt.Sprintf("%d", p.Deaths)))      // Deaths
			a.PlayersTable.SetCell(row, 7, center(fmt.Sprintf("%d", p.Ping)))
			a.PlayersTable.SetCell(row, 8, c(p.IP))
		}

		// Ensure selection behavior
		a.PlayersTable.SetSelectable(true, false)

		a.PlayersTable.SetSelectedFunc(func(row, column int) {
			// Enter on row does nothing or maybe shows info?
		})

		a.PlayersTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			r, _ := a.PlayersTable.GetSelection()
			if r <= 0 { // Ignore header row
				return event
			}
			// Get player from cell 0 reference
			cell := a.PlayersTable.GetCell(r, 0)
			if cell == nil {
				return event
			}
			ref := cell.GetReference()
			if ref == nil {
				return event
			}
			p := ref.(model.Player) // Correctly cast to model.Player

			switch event.Rune() {
			case 'k':
				a.Input.SetText(fmt.Sprintf("kick %s \"Kicked by Console\"", p.ID))
				a.TviewApp.SetFocus(a.Input)
			case 'b':
				a.Input.SetText(fmt.Sprintf("ban %s 10 year \"Banned by Console\"", p.ID))
				a.TviewApp.SetFocus(a.Input)
			case 't':
				a.Input.SetText(fmt.Sprintf("teleport %s ", p.ID))
				a.TviewApp.SetFocus(a.Input)
			}
			return event
		})
	})
}

func isImportantLog(log string) bool {
	// Filter for Chat, Errors, Warnings, Player Activity
	// Chat in 7DTD: "Chat (from 'SteamId'): <Name>: Message"
	// or similar patterns.
	// We want to keep it simple but effective.

	// Keywords to Keep
	keywords := []string{
		"Chat",
		"PlayerConnected",
		"PlayerDisconnected",
		"ERR",
		"WRN",
		"Kicked",
		"Banned",
	}

	for _, kw := range keywords {
		if strings.Contains(log, kw) {
			return true
		}
	}

	return false
}
