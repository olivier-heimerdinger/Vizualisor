package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/olivier-heimerdinger/Vizualisor/logs"
)

// OpenLogWindow ouvre une nouvelle fenêtre pour visualiser les logs.
// Si serviceName est vide, utilise logFilePath pour un fichier custom.
func OpenLogWindow(app *App, serverName, serviceName, logFilePath string) {
	client, ok := app.sshPool.Get(serverName)
	if !ok {
		return
	}

	// Titre de la fenêtre
	var title string
	if serviceName != "" {
		title = fmt.Sprintf("Logs - %s @ %s", serviceName, serverName)
	} else {
		title = fmt.Sprintf("Logs - %s @ %s", logFilePath, serverName)
	}

	logWindow := app.fyneApp.NewWindow(title)
	logWindow.Resize(fyne.NewSize(900, 600))

	// Zone de texte pour les logs
	logText := widget.NewLabel("")
	logText.Wrapping = fyne.TextWrapBreak
	logText.TextStyle = fyne.TextStyle{Monospace: true}

	logScroll := container.NewVScroll(logText)

	// Filtre
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("🔍 Filtrer les logs...")

	// Boutons
	streamBtn := widget.NewButton("▶ Streaming", nil)
	refreshBtn := widget.NewButton("↻ Actualiser", nil)
	clearBtn := widget.NewButton("🗑 Effacer", nil)

	var fullContent string
	var streamer *logs.Streamer
	var streaming bool

	// Fonction de chargement initial
	loadLogs := func() {
		reader := logs.NewReader(client, serverName)
		var content string
		var err error

		if serviceName != "" {
			content, err = reader.ReadServiceLog(serviceName, 200)
		} else if logFilePath != "" {
			content, err = reader.ReadFileLog(logFilePath, 200)
		}

		if err != nil {
			logText.SetText(fmt.Sprintf("Erreur : %v", err))
			return
		}

		fullContent = content
		if filterEntry.Text != "" {
			logText.SetText(logs.FilterLines(content, filterEntry.Text))
		} else {
			logText.SetText(content)
		}

		// Défiler vers le bas
		logScroll.ScrollToBottom()
	}

	// Chargement initial
	go loadLogs()

	// Actualiser
	refreshBtn.OnTapped = func() {
		go loadLogs()
	}

	// Effacer
	clearBtn.OnTapped = func() {
		fullContent = ""
		logText.SetText("")
	}

	// Filtre temps réel
	filterEntry.OnChanged = func(filter string) {
		if fullContent != "" {
			logText.SetText(logs.FilterLines(fullContent, filter))
		}
	}

	// Streaming (tail -f)
	streamBtn.OnTapped = func() {
		if streaming {
			// Arrêter le streaming
			if streamer != nil {
				streamer.Stop()
			}
			streaming = false
			streamBtn.SetText("▶ Streaming")
			return
		}

		// Démarrer le streaming
		streamer = logs.NewStreamer(client, serverName)
		var lines <-chan string
		var err error

		if serviceName != "" {
			lines, err = streamer.StreamServiceLog(serviceName)
		} else if logFilePath != "" {
			lines, err = streamer.StreamFileLog(logFilePath)
		}

		if err != nil {
			logText.SetText(fmt.Sprintf("Erreur streaming : %v", err))
			return
		}

		streaming = true
		streamBtn.SetText("■ Arrêter")

		// Goroutine de lecture du stream
		go func() {
			var buffer strings.Builder
			ticker := time.NewTicker(100 * time.Millisecond) // Batch UI updates
			defer ticker.Stop()

			for {
				select {
				case line, ok := <-lines:
					if !ok {
						streaming = false
						streamBtn.SetText("▶ Streaming")
						return
					}
					buffer.WriteString(line + "\n")
					fullContent += line + "\n"

				case <-ticker.C:
					if buffer.Len() > 0 {
						newText := buffer.String()
						buffer.Reset()

						current := logText.Text
						if filterEntry.Text != "" {
							filtered := logs.FilterLines(newText, filterEntry.Text)
							if filtered != "" {
								logText.SetText(current + filtered)
								logScroll.ScrollToBottom()
							}
						} else {
							logText.SetText(current + newText)
							logScroll.ScrollToBottom()
						}
					}
				}
			}
		}()
	}

	// Arrêter le streaming quand la fenêtre est fermée
	logWindow.SetOnClosed(func() {
		if streamer != nil {
			streamer.Stop()
		}
	})

	// Layout
	toolbar := container.NewHBox(refreshBtn, streamBtn, clearBtn)
	topBar := container.NewBorder(nil, nil, nil, toolbar, filterEntry)

	content := container.NewBorder(
		topBar,    // Haut
		nil,       // Bas
		nil,       // Gauche
		nil,       // Droite
		logScroll, // Centre
	)

	logWindow.SetContent(content)
	logWindow.Show()
}
