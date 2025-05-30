package gdi

import (
	"syscall"
	"unsafe"
)

const (
	DISPLAY_DEVICE_ACTIVE           = 0x00000001
	DISPLAY_DEVICE_MIRRORING_DRIVER = 0x00000008
	DISPLAY_DEVICE_PRIMARY_DEVICE   = 0x00000004
	ENUM_CURRENT_SETTINGS           = 0xFFFFFFFF
)

type DISPLAY_DEVICEW struct {
	Cb           uint32
	DeviceName   [32]uint16
	DeviceString [128]uint16
	StateFlags   uint32
	DeviceID     [128]uint16
	DeviceKey    [128]uint16
}

type DEVMODEW struct {
	DeviceName         [32]uint16
	SpecVersion        uint16
	DriverVersion      uint16
	Size               uint16
	DriverExtra        uint16
	Fields             uint32
	Position           POINTL
	DisplayOrientation uint32
	DisplayFixedOutput uint32
	Color              int16
	Duplex             int16
	YResolution        int16
	TTOption           int16
	Collate            int16
	FormName           [32]uint16
	LogPixels          uint16
	BitsPerPel         uint32
	PelsWidth          uint32
	PelsHeight         uint32
	DisplayFlags       uint32
	DisplayFrequency   uint32
	ICMMethod          uint32
	ICMIntent          uint32
	MediaType          uint32
	DitherType         uint32
	Reserved1          uint32
	Reserved2          uint32
	PanningWidth       uint32
	PanningHeight      uint32
}

type POINTL struct {
	X int32
	Y int32
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type DisplayDesc struct {
	DeviceName         [32]uint16
	DesktopCoordinates RECT
	AttachedToDesktop  int32
}

type Display struct {
	Desc DisplayDesc
	GDI  bool
}

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	procEnumDisplayDevicesW    = user32.NewProc("EnumDisplayDevicesW")
	procEnumDisplaySettingsExW = user32.NewProc("EnumDisplaySettingsExW")
)

func enumDisplayDevicesW(deviceName *uint16, deviceNum uint32, displayDevice *DISPLAY_DEVICEW, flags uint32) bool {
	ret, _, _ := procEnumDisplayDevicesW.Call(
		uintptr(unsafe.Pointer(deviceName)),
		uintptr(deviceNum),
		uintptr(unsafe.Pointer(displayDevice)),
		uintptr(flags),
	)
	return ret != 0
}

func enumDisplaySettingsExW(deviceName *uint16, modeNum uint32, devMode *DEVMODEW, flags uint32) bool {
	ret, _, _ := procEnumDisplaySettingsExW.Call(
		uintptr(unsafe.Pointer(deviceName)),
		uintptr(modeNum),
		uintptr(unsafe.Pointer(devMode)),
		uintptr(flags),
	)
	return ret != 0
}

func GetFromGDI() []Display {
	var all []Display
	var i uint32 = 0

	for {
		var d DISPLAY_DEVICEW
		d.Cb = uint32(unsafe.Sizeof(d))

		if !enumDisplayDevicesW(nil, i, &d, 0) {
			break
		}

		i++

		// Skip inactive displays and mirroring drivers
		if (d.StateFlags&DISPLAY_DEVICE_ACTIVE) == 0 ||
			(d.StateFlags&DISPLAY_DEVICE_MIRRORING_DRIVER) > 0 {
			continue
		}

		// Create display struct
		var disp Display
		disp.GDI = true
		disp.Desc.DeviceName = d.DeviceName

		// Get display settings
		var m DEVMODEW
		m.Size = uint16(unsafe.Sizeof(m))
		m.DriverExtra = 0

		if !enumDisplaySettingsExW(&d.DeviceName[0], ENUM_CURRENT_SETTINGS, &m, 0) {
			continue
		}

		// Set coordinates
		disp.Desc.DesktopCoordinates.Left = m.Position.X
		disp.Desc.DesktopCoordinates.Top = m.Position.Y
		disp.Desc.DesktopCoordinates.Right = disp.Desc.DesktopCoordinates.Left + int32(m.PelsWidth)
		disp.Desc.DesktopCoordinates.Bottom = disp.Desc.DesktopCoordinates.Top + int32(m.PelsHeight)
		disp.Desc.AttachedToDesktop = 1

		all = append(all, disp)
	}

	return all
}

// Helper function to convert UTF-16 device name to string
func deviceNameToString(deviceName [32]uint16) string {
	return syscall.UTF16ToString(deviceName[:])
}
