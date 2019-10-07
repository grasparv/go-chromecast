package pick

import (
	"fmt"
	//"os"
	"strings"

	"github.com/grasparv/go-chromecast/picksongs/scroll"
	"github.com/jroimartin/gocui"
)

type songNameList []songName

type songName struct {
	label string
	name  string
}

type songPickerApp struct {
	g *gocui.Gui

	queue songNameList
	songs songNameList

	scrollQueue *scroll.SliceScroller
	scrollSongs *scroll.SliceScroller
}

const paneSongs = "songs"
const paneQueue = "queue"
const titleQueue = " Your initial queue "
const titleSongs = " Songs selection "

func createLabel(s string) string {
	s = strings.ReplaceAll(s, ".mp4", "")
	s = strings.ReplaceAll(s, ".mp3", "")
	s = strings.ReplaceAll(s, ".ogg", "")
	s = strings.ReplaceAll(s, "_", " ")
	return s
}

func newSongNameList(input []string) songNameList {
	list := make([]songName, 0, len(input))
	for _, i := range input {
		sn := songName{
			label: createLabel(i),
			name:  i,
		}
		list = append(list, sn)
	}
	return list
}

func (list songNameList) Strings() []string {
	strings := make([]string, 0, len(list))
	for _, sn := range list {
		strings = append(strings, sn.name)
	}
	return strings
}

func Picksongs(filenames []string) (queue []string, remaining []string) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	app := songPickerApp{
		g:           g,
		queue:       songNameList{},
		songs:       newSongNameList(filenames),
		scrollSongs: scroll.New(),
		scrollQueue: scroll.New(),
	}

	g.SetManagerFunc(app.views)
	g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, quit)
	g.SetKeybinding("", 'q', gocui.ModNone, quit)
	g.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, app.switchFocus)
	g.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, app.switchFocus)

	g.SetKeybinding(paneSongs, gocui.KeyPgdn, gocui.ModNone, app.scrollDownSongs)
	g.SetKeybinding(paneSongs, gocui.KeyPgup, gocui.ModNone, app.scrollUpSongs)
	g.SetKeybinding(paneSongs, gocui.KeyArrowDown, gocui.ModNone, app.moveDownSongs)
	g.SetKeybinding(paneSongs, gocui.KeyArrowUp, gocui.ModNone, app.moveUpSongs)
	g.SetKeybinding(paneSongs, gocui.KeyEnter, gocui.ModNone, app.pickSong)

	g.SetKeybinding(paneQueue, gocui.KeyPgdn, gocui.ModNone, app.scrollDownQueue)
	g.SetKeybinding(paneQueue, gocui.KeyPgup, gocui.ModNone, app.scrollUpQueue)
	g.SetKeybinding(paneQueue, gocui.KeyArrowDown, gocui.ModNone, app.moveDownQueue)
	g.SetKeybinding(paneQueue, gocui.KeyArrowUp, gocui.ModNone, app.moveUpQueue)
	g.SetKeybinding(paneQueue, gocui.KeyEnter, gocui.ModNone, app.pickQueue)

	g.SetKeybinding(paneQueue, 'w', gocui.ModNone, app.prioritizeUp)
	g.SetKeybinding(paneQueue, 'W', gocui.ModNone, app.prioritizeUp)
	g.SetKeybinding(paneQueue, 's', gocui.ModNone, app.prioritizeDown)
	g.SetKeybinding(paneQueue, 'S', gocui.ModNone, app.prioritizeDown)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		empty := []string{}
		return empty, filenames
	}

	//fmt.Fprintf(os.Stderr, "a = %v\n", app.queue)
	//fmt.Fprintf(os.Stderr, "b = %v\n", app.songs)
	return app.queue.Strings(), app.songs.Strings()
}

func (m *songPickerApp) views(g *gocui.Gui) error {
	var err error
	err = m.viewSongs()
	if err != nil {
		return err
	}
	err = m.viewQueue()
	if err != nil {
		return err
	}
	return nil
}

func (m *songPickerApp) viewSongs() error {
	x, y := m.g.Size()

	v, err := m.g.SetView(paneSongs, 0, 0, x/2-1, y-2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	} else {
		v.Title = titleSongs
		m.scrollSongs.SetSizeFunction(v.Size)
		v.Frame = true
	}

	m.scrollSongs.SetEntriesCount(len(m.songs))
	v.Clear()

	start := m.scrollSongs.View()
	end := m.scrollSongs.View() + y
	if end >= len(m.songs) {
		end = len(m.songs)
	}

	//fmt.Fprintf(os.Stderr, "view =%d, index=%d, start, end = %d, %d\n",
	//	m.scrollSongs.View(),
	//	m.scrollSongs.Index(),
	//	start, end)

	for i, sn := range m.songs[start:end] {
		if m.scrollSongs.Index()-m.scrollSongs.View() == i {
			fmt.Fprintf(v, "* %s\n", sn.label)
		} else {
			fmt.Fprintf(v, "  %s\n", sn.label)
		}
	}

	if m.g.CurrentView() == nil {
		m.g.SetCurrentView(paneSongs)
	}

	return nil
}

func (m *songPickerApp) viewQueue() error {
	x, y := m.g.Size()

	v, err := m.g.SetView(paneQueue, x/2, 0, x-1, y-2)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	} else {
		v.Title = titleQueue
		m.scrollQueue.SetSizeFunction(v.Size)
	}

	m.scrollQueue.SetEntriesCount(len(m.queue))
	v.Clear()

	start := m.scrollQueue.View()
	end := m.scrollQueue.View() + y
	if end >= len(m.queue) {
		end = len(m.queue)
	}

	for i, sn := range m.queue[start:end] {
		if m.scrollQueue.Index()-m.scrollQueue.View() == i {
			fmt.Fprintf(v, "* %s\n", sn.label)
		} else {
			fmt.Fprintf(v, "  %s\n", sn.label)
		}
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (m *songPickerApp) switchFocus(g *gocui.Gui, v *gocui.View) error {
	v = m.g.CurrentView()
	if v.Name() == paneQueue {
		//fmt.Fprintf(os.Stderr, "focus songs\n")
		m.g.SetCurrentView(paneSongs)
	} else {
		//fmt.Fprintf(os.Stderr, "focus queue\n")
		m.g.SetCurrentView(paneQueue)
	}
	return nil
}

func (m *songPickerApp) pickSong(g *gocui.Gui, v *gocui.View) error {
	i := m.scrollSongs.Index()
	if i >= len(m.songs) {
		return nil
	}
	sn := m.songs[i]
	m.songs = append(m.songs[:i], m.songs[i+1:]...)
	m.queue = append(m.queue, sn)
	m.scrollQueue.SetEntriesCount(len(m.queue))
	m.scrollSongs.SetEntriesCount(len(m.songs))
	for m.scrollSongs.Index() >= len(m.songs) && len(m.songs) > 0 {
		m.scrollSongs.MoveUp()
	}
	return nil
}

func (m *songPickerApp) pickQueue(g *gocui.Gui, v *gocui.View) error {
	i := m.scrollQueue.Index()
	if i >= len(m.queue) {
		return nil
	}
	sn := m.queue[i]
	m.queue = append(m.queue[:i], m.queue[i+1:]...)
	m.songs = append(m.songs, sn)
	m.scrollQueue.SetEntriesCount(len(m.queue))
	m.scrollSongs.SetEntriesCount(len(m.songs))
	for m.scrollQueue.Index() >= len(m.queue) && len(m.queue) > 0 {
		m.scrollQueue.MoveUp()
	}
	return nil
}

func (m *songPickerApp) prioritizeUp(g *gocui.Gui, v *gocui.View) error {
	if m.scrollQueue.Index() == 0 {
		return nil
	}

	this := m.scrollQueue.Index()
	prev := this - 1

	m.queue[this], m.queue[prev] = m.queue[prev], m.queue[this]
	return m.moveUpQueue(g, v)
}

func (m *songPickerApp) prioritizeDown(g *gocui.Gui, v *gocui.View) error {
	if m.scrollQueue.Index() >= len(m.queue)-1 {
		return nil
	}

	this := m.scrollQueue.Index()
	next := this + 1

	m.queue[this], m.queue[next] = m.queue[next], m.queue[this]
	return m.moveDownQueue(g, v)
}

func (m *songPickerApp) scrollDownSongs(g *gocui.Gui, v *gocui.View) error {
	m.scrollSongs.ScrollDown()
	return nil
}

func (m *songPickerApp) scrollUpSongs(g *gocui.Gui, v *gocui.View) error {
	m.scrollSongs.ScrollUp()
	return nil
}

func (m *songPickerApp) moveDownSongs(g *gocui.Gui, v *gocui.View) error {
	m.scrollSongs.MoveDown()
	return nil
}

func (m *songPickerApp) moveUpSongs(g *gocui.Gui, v *gocui.View) error {
	m.scrollSongs.MoveUp()
	return nil
}

func (m *songPickerApp) scrollDownQueue(g *gocui.Gui, v *gocui.View) error {
	m.scrollQueue.ScrollDown()
	return nil
}

func (m *songPickerApp) scrollUpQueue(g *gocui.Gui, v *gocui.View) error {
	m.scrollQueue.ScrollUp()
	return nil
}

func (m *songPickerApp) moveDownQueue(g *gocui.Gui, v *gocui.View) error {
	m.scrollQueue.MoveDown()
	return nil
}

func (m *songPickerApp) moveUpQueue(g *gocui.Gui, v *gocui.View) error {
	m.scrollQueue.MoveUp()
	return nil
}
