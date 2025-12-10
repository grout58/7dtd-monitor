package ui

import (
	"7dtd-monitor/internal/parser"
	"7dtd-monitor/internal/telnet"
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type App struct {
	TviewApp     *tview.Application
	Client       *telnet.Client
	StatsText    *tview.TextView
	PlayersTable *tview.Table
	LogView      *tview.TextView
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
		SetBorders(true)
	a.PlayersTable.SetBorder(true).SetTitle(" Online Players ")

	// 3. Log/Debug View (bottom)
	a.LogView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	a.LogView.SetBorder(true).SetTitle(" Log ")

	// Layout: Flex
	// Top: Stats (1/3), Players (2/3)
	// Bottom: Log
	topFlex := tview.NewFlex().
		AddItem(a.StatsText, 0, 1, false).
		AddItem(a.PlayersTable, 0, 2, false)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topFlex, 0, 3, false).
		AddItem(a.LogView, 10, 1, false)

	a.TviewApp.SetRoot(mainFlex, true)
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

func (a *App) updateData() {
	// 1. Get Time
	timeStr, _ := a.Client.SendCommand("gettime")
	gameTime := parser.ParseTime(timeStr)

	// 2. Get Mem
	memStr, _ := a.Client.SendCommand("mem")
	heap, max := parser.ParseMem(memStr)

	// 3. Get Players
	playersStr, _ := a.Client.SendCommand("lpi")
	players, _ := parser.ParsePlayers(playersStr)

	a.TviewApp.QueueUpdateDraw(func() {
		// Update Stats
		statsText := fmt.Sprintf("\n [green]Host:[white] %s\n [green]Port:[white] %s\n\n [yellow]Game Time:[white] %s\n\n [blue]Heap:[white] %s / %s\n [blue]Players:[white] %d",
			a.Client.Host, a.Client.Port, gameTime, heap, max, len(players))
		a.StatsText.SetText(statsText)

		// Update Table
		a.PlayersTable.Clear()
		headers := []string{"ID", "Name", "Score", "Lvl", "Ping", "IP"}
		for i, h := range headers {
			a.PlayersTable.SetCell(0, i,
				tview.NewTableCell(h).
					SetTextColor(tview.Styles.SecondaryTextColor).
					SetAlign(tview.AlignCenter).
					SetSelectable(false))
		}

		for i, p := range players {
			row := i + 1
			a.PlayersTable.SetCell(row, 0, tview.NewTableCell(p.ID))
			a.PlayersTable.SetCell(row, 1, tview.NewTableCell(p.Name))
			a.PlayersTable.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%d", p.Score)))
			a.PlayersTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%d", p.Level)))
			a.PlayersTable.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%d", p.Ping)))
			a.PlayersTable.SetCell(row, 5, tview.NewTableCell(p.IP))
		}

		a.LogView.SetText("[gray]Refreshed data at " + time.Now().Format("15:04:05"))
	})
}
