package screenshot

import (
	"fmt"
	"image"
	"log"
	"math"
	"sync"
	"sync/atomic"

	"github.com/ghp3000/screenshot/d3d"
)

type DXGIScreenshot struct {
	rect      image.Rectangle
	device    *d3d.ID3D11Device
	deviceCtx *d3d.ID3D11DeviceContext
	ddup      *d3d.OutputDuplicator
	display   int // 截取哪个屏幕
	cursor    int32
}

func NewDXGIScreenshot() *DXGIScreenshot {
	return &DXGIScreenshot{}
}

func (s *DXGIScreenshot) Init(display int) error {
	if display < 0 {
		return fmt.Errorf("display %d is invalid", display)
	}
	s.Release()
	s.display = display
	s.rect = GetDisplayBounds(display)
	var err error
	s.device, s.deviceCtx, err = d3d.NewD3D11Device()
	if err != nil {
		return err
	}
	s.ddup, err = d3d.NewIDXGIOutputDuplication(s.device, s.deviceCtx, uint(s.display))
	if err != nil {
		return err
	}

	return nil
}

var once sync.Once

func (s *DXGIScreenshot) Capture() (*image.RGBA, error) {
	if s.ddup == nil {
		return nil, fmt.Errorf("no init,please run Init first")
	}
	rect := GetDisplayBounds(s.display)
	// 如果发现屏幕范围发生了变化就重新初始化
	if rect != s.rect {
		s.Release()
		if err := s.Init(s.display); err != nil {
			return nil, err
		}
	}

	// Do not use s.rect directly, because left and top may be negative
	width := s.rect.Dx()
	height := s.rect.Dy()
	// Also get Absolute of width and height
	width = int(math.Abs(float64(width)))
	height = int(math.Abs(float64(height)))
	var imgBuf *image.RGBA
	once.Do(
		func() {
			log.Printf("DXGI screenshot: display %d, rect %v", s.display, s.rect)
		},
	)
	err := s.ddup.GetImage(imgBuf, 0, atomic.LoadInt32(&s.cursor) == 1)
	if err != nil {
		return nil, err
	}
	return imgBuf, err
}

func (s *DXGIScreenshot) CaptureBGRA() (*image.RGBA, error) {
	// todo
	return s.Capture()
}

func (s *DXGIScreenshot) Release() {
	if s.device != nil {
		s.device.Release()
	}
	if s.deviceCtx != nil {
		s.deviceCtx.Release()
	}
	if s.ddup != nil {
		s.ddup.Release()
	}
}

func (s *DXGIScreenshot) GetDisplayId() int {
	return s.display
}
func (s *DXGIScreenshot) GetCaptureName() string {
	return "DXGI"
}
func (s *DXGIScreenshot) DrawCursor(cursor int32) {
	atomic.StoreInt32(&s.cursor, cursor)
}
