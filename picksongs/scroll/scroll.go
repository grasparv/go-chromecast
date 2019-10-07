package scroll

const smoothScrollMargin = 20

type sizeFun func() (int, int)

type SliceScroller struct {
	view    int
	pos     int
	dimfun  sizeFun
	entries int
}

func New() *SliceScroller {
	return &SliceScroller{
		view:   0,
		pos:    0,
		dimfun: nil,
	}
}

func (s *SliceScroller) SetEntriesCount(n int) {
	s.entries = n
}

func (s *SliceScroller) SetSizeFunction(f sizeFun) {
	s.dimfun = f
}

func (s *SliceScroller) Index() int {
	return s.pos
}

func (s *SliceScroller) View() int {
	return s.view
}

func (s *SliceScroller) scrolldown(amount int) {
	_, y := s.dimfun()
	//fmt.Fprintf(os.Stderr, "scrolldown  start view=%d pos=%d\n", s.view, s.pos)
	s.view += amount
	for s.view+y >= s.entries+1 && s.view > 0 {
		s.view--
	}
	for s.pos < s.view {
		s.pos++
	}
	//fmt.Fprintf(os.Stderr, "scrolldown  done  view=%d pos=%d\n", s.view, s.pos)
}

func (s *SliceScroller) scrollup(amount int) {
	_, y := s.dimfun()
	s.view -= amount
	for s.view < 0 && s.view < s.entries {
		s.view++
	}
	for s.pos > s.view+y-1 {
		s.pos--
	}
}

func (s *SliceScroller) ScrollDown() {
	_, y := s.dimfun()
	s.scrolldown(y / 3)
}

func (s *SliceScroller) ScrollUp() {
	_, y := s.dimfun()
	s.scrollup(y / 3)
}

func (s *SliceScroller) MoveDown() {
	_, y := s.dimfun()
	if s.pos >= s.view+y-smoothScrollMargin {
		s.scrolldown(1)
	}
	if s.pos < s.view+y-1 && s.pos < s.entries-1 {
		s.pos++
	}
}

func (s *SliceScroller) MoveUp() {
	if s.pos <= s.view+smoothScrollMargin {
		s.scrollup(1)
	}
	if s.pos > s.view && s.pos > 0 {
		s.pos--
	}
}
