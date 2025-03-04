package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/gdamore/tcell/v2"
)

type Player struct {
	ctrl        *beep.Ctrl
	format      beep.Format
	streamer    beep.StreamSeekCloser
	currentFile string
	isPlaying   bool
	volume      float64
	position    float64
	duration    float64
}

func NewPlayer() *Player {
	return &Player{
		volume: 1.0,
	}
}
func (p *Player) LoadFile(path string) error {
	if p.streamer != nil {
		p.streamer.Close()
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	var streamer beep.StreamSeekCloser
	var format beep.Format
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	default:
		return fmt.Errorf("Error occuredl: unsupported file format %v", filepath.Ext(path))
	}
	if err != nil {
		return err
	}
	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return err
	}
	p.ctrl = &beep.Ctrl{Streamer: streamer, Paused: true}
	p.streamer = streamer
	p.format = format
	p.currentFile = path
	p.isPlaying = false
	p.duration = float64(p.streamer.Len()) / float64(p.format.SampleRate)
	return nil
}

type PlayerUI struct {
	screen tcell.Screen
	player *Player
	quit   chan struct{}
}

func (p *Player) Play() {
	if p.ctrl == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = false
	p.isPlaying = true
	speaker.Unlock()
}
func (p *Player) Pause() {
	if p.ctrl == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = true
	p.isPlaying = false
	speaker.Unlock()
}
func (p *Player) Stop() {
	if p.streamer == nil {
		return
	}
	speaker.Lock()
	p.ctrl.Paused = true
	p.isPlaying = false
	p.streamer.Seek(0)
	speaker.Unlock()
}
func (p *Player) Seek(seconds float64) {
	if p.streamer == nil {
		return
	}
	position := p.format.SampleRate.N(time.Duration(seconds) * time.Second)
	if position < 0 {
		position = 0
	}
	if position > p.streamer.Len() {
		position = p.streamer.Len() - 1
	}
	speaker.Lock()
	p.streamer.Seek(position)
	speaker.Unlock()
}
func (p *Player) GetPosition() float64 {
	if p.streamer == nil {
		return 0
	}
	speaker.Lock()
	position := float64(p.streamer.Position()) / float64(p.format.SampleRate)
	speaker.Unlock()
	return position
}
func (ui *PlayerUI) Close() {
	if ui.screen != nil {
		ui.screen.Fini()
	}
}
func (p *Player) GetDuration() float64 {
	return p.duration
}
func (p *Player) SetVolume(volume float64) {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	p.volume = volume
}
func (ui *PlayerUI) handleKeyEvent(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		close(ui.quit)
		return
	case tcell.KeyRune:
		switch ev.Rune() {
		case ' ':
			if ui.player.isPlaying {
				ui.player.Pause()
			} else {
				ui.player.Play()
			}

		case 's':
			ui.player.Stop()
		case '+':
			ui.player.SetVolume(ui.player.volume + 0.1)
		case '-':
			ui.player.SetVolume(ui.player.volume - 0.1)
		}
	case tcell.KeyLeft:
		ui.player.Seek(ui.player.GetPosition() - 5)
	case tcell.KeyRight:
		ui.player.Seek(ui.player.GetPosition() + 5)
	}
}
func (ui *PlayerUI) handleEvents() {
	for {
		ev := ui.screen.PollEvent()
		select {
		case <-ui.quit:
			return
		default:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				ui.handleKeyEvent(ev)
			case *tcell.EventResize:
				ui.screen.Sync()
			}
		}
	}
}
func drawText(s tcell.Screen, x, y int, text string) {
	for i, r := range text {
		s.SetContent(x+i, y, r, nil, tcell.StyleDefault)
	}
}
func playbackStatus(isPlaying bool) string {
	if isPlaying {
		return "Playing"
	}
	return "Paused"
}
func (ui *PlayerUI) render() {
	ui.screen.Clear()
	drawText(ui.screen, 0, 0, fmt.Sprintf("Now playing: %s", ui.player.currentFile))
	drawText(ui.screen, 0, 2, fmt.Sprintf("Status: %s", playbackStatus(ui.player.isPlaying)))
	drawText(ui.screen, 0, 3, fmt.Sprintf("Volume: %.1f", ui.player.volume))
	position := ui.player.GetPosition()
	duration := ui.player.GetDuration()
	drawText(ui.screen, 0, 5, fmt.Sprintf("Position: %.1f / %.1f seconds", position, duration))
	progressWidth := 50
	progress := int(position / duration * float64(progressWidth))
	progressBar := "["
	for i := 0; i < progressWidth; i++ {
		if i < progress {
			progressBar += "="
		} else {
			progressBar += " "
		}
	}
	progressBar += "]"
	drawText(ui.screen, 0, 6, progressBar)
	drawText(ui.screen, 0, 8, "Controls:")
	drawText(ui.screen, 0, 9, "Space: Play/Pause")
	drawText(ui.screen, 0, 10, "Left/Right: Seek: -/+ 5 seconds")
	drawText(ui.screen, 0, 11, "+/-: Volume up/down")
	drawText(ui.screen, 0, 12, "s: Stop")

	ui.screen.Show()
}
func (ui *PlayerUI) Start() {
	go ui.handleEvents()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ui.quit:
			return
		case <-ticker.C:
			ui.render()
		}
	}
}

func NewPlayerUI(player *Player) (*PlayerUI, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("error creating screen : %v", err)
	}
	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("Error initializing screen: %v", err)
	}
	return &PlayerUI{
		screen: screen,
		player: player,
		quit:   make(chan struct{}),
	}, nil
}

func main() {
	audioFile := "sample.mp3"
	player := NewPlayer()
	if err := player.LoadFile(audioFile); err != nil {
		log.Fatalf("Error loading audio file: %v", err)
	}
	ui, err := NewPlayerUI(player)
	if err != nil {
		log.Fatalf("Error in uI %v", err)
	}
	defer ui.Close()
	speaker.Play(player.ctrl)
	player.Play()
	ui.Start()
}
