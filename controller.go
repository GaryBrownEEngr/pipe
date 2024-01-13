package pipe

type Controller struct {
	blockSize    int
	channelDepth int
	timeStepSec  float64
	started      bool
	startChan    chan bool
	done         bool
	doneChan     chan bool
}

func NewController(blockSize, channelDepth int, timeStepSec float64) *Controller {
	ret := &Controller{
		blockSize:    blockSize,
		channelDepth: channelDepth,
		timeStepSec:  timeStepSec,
		startChan:    make(chan bool),
		doneChan:     make(chan bool),
	}

	return ret
}

func (s *Controller) BlockSize() int {
	return s.blockSize
}

func (s *Controller) TimeStepSec() float64 {
	return s.timeStepSec
}

func (s *Controller) Start() {
	if !s.started {
		close(s.startChan)
		s.started = true
	}
}

func (s *Controller) WaitForStart() {
	// wait until the start chan is closed
	<-s.startChan
}

func (s *Controller) Stop() {
	if !s.done {
		close(s.doneChan)
		s.done = true
	}
}
