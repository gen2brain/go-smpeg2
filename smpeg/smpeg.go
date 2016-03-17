// Go bindings for smpeg2 http://icculus.org/smpeg/
package smpeg

//#cgo linux LDFLAGS: -lsmpeg2 -lSDL2
//#cgo linux CFLAGS: -I/usr/include/smpeg2 -I/usr/include/SDL2
//#include <stdlib.h>
//#include <smpeg.h>
//extern void displayCallback(void *data, SMPEG_Frame *frame);
import "C"

import (
	"errors"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// MPEG status codes
const (
	SMPEG_ERROR = iota
	SMPEG_STOPPED
	SMPEG_PLAYING
)

// SMPEG object
type SMPEG struct {
	mpeg *C.SMPEG
}

// Informations about the SMPEG object
type Info struct {
	HasAudio          bool
	HasVideo          bool
	Width             int
	Height            int
	CurrentFrame      int
	CurrentFps        float64
	AudioString       string
	AudioCurrentFrame int
	CurrentOffset     uint32
	TotalSize         uint32
	CurrentTime       float64
	TotalTime         float64
}

// YV12 format video frame
type Frame struct {
	W           uint32
	H           uint32
	ImageWidth  uint32
	ImageHeight uint32
	Image       *uint8
}

// Update context
type Context struct {
	Frame      *Frame
	FrameCount int
	Lock       *sdl.Mutex
}

// Creates a new SMPEG object from an MPEG file.
// The sdl_audio parameter indicates if SMPEG should initialize the SDL audio subsystem.
// If not, you will have to use the PlayAudio() function to extract the decoded data.
func New(file string, sdl_audio bool) (s *SMPEG, err error) {
	s = &SMPEG{}

	f := C.CString(file)
	defer C.free(unsafe.Pointer(f))

	a := 0
	if sdl_audio {
		a = 1
	}

	s.mpeg = C.SMPEG_new(f, nil, C.int(a))

	e := C.SMPEG_error(s.mpeg)
	if e != nil {
		err = errors.New(C.GoString(e))
	}

	return
}

// Creates a new SMPEG object from a file descriptor
func NewDescr(file int, sdl_audio bool) (s *SMPEG, err error) {
	s = &SMPEG{}

	a := 0
	if sdl_audio {
		a = 1
	}

	s.mpeg = C.SMPEG_new_descr(C.int(file), nil, C.int(a))

	e := C.SMPEG_error(s.mpeg)
	if e != nil {
		err = errors.New(C.GoString(e))
	}

	return
}

// Creates a new SMPEG object from a raw chunk of data
func NewData(data []byte, sdl_audio bool) (s *SMPEG, err error) {
	s = &SMPEG{}

	a := 0
	if sdl_audio {
		a = 1
	}

	s.mpeg = C.SMPEG_new_data(unsafe.Pointer(&data[0]), C.int(len(data)), nil, C.int(a))

	e := C.SMPEG_error(s.mpeg)
	if e != nil {
		err = errors.New(C.GoString(e))
	}

	return
}

// Creates a new SMPEG object from a generic SDL_RWops structure
func NewRWops(src *sdl.RWops, freesrc bool, sdl_audio bool) (s *SMPEG, err error) {
	s = &SMPEG{}

	f := 0
	if freesrc {
		f = 1
	}

	a := 0
	if sdl_audio {
		a = 1
	}

	s.mpeg = C.SMPEG_new_rwops((*C.SDL_RWops)(unsafe.Pointer(src)), nil, C.int(f), C.int(a))

	e := C.SMPEG_error(s.mpeg)
	if e != nil {
		err = errors.New(C.GoString(e))
	}

	return
}

// Current information about an SMPEG object
func (s *SMPEG) Info() (info *Info) {
	info = &Info{}

	i := C.struct__SMPEG_Info{}
	C.SMPEG_getinfo(s.mpeg, (*C.struct__SMPEG_Info)(unsafe.Pointer(&i)))

	info.HasAudio = false
	if int(i.has_audio) == 1 {
		info.HasAudio = true
	}

	info.HasVideo = false
	if int(i.has_video) == 1 {
		info.HasVideo = true
	}

	info.Width = int(i.width)
	info.Height = int(i.height)
	info.CurrentFrame = int(i.current_frame)
	info.CurrentFps = float64(i.current_fps)
	info.AudioCurrentFrame = int(i.audio_current_frame)
	info.AudioString = C.GoString(&i.audio_string[0])
	info.AudioCurrentFrame = int(i.audio_current_frame)
	info.CurrentOffset = uint32(i.current_offset)
	info.TotalSize = uint32(i.total_size)
	info.CurrentTime = float64(i.current_time)
	info.TotalTime = float64(i.total_time)

	return
}

// Enables or disables audio playback in MPEG stream
func (s *SMPEG) EnableAudio(enable bool) {
	e := 0
	if enable {
		e = 1
	}

	C.SMPEG_enableaudio(s.mpeg, C.int(e))
}

// Enables or disables video playback in MPEG stream
func (s *SMPEG) EnableVideo(enable bool) {
	e := 0
	if enable {
		e = 1
	}

	C.SMPEG_enablevideo(s.mpeg, C.int(e))
}

// Deletes an SMPEG object and releases the memory
func (s *SMPEG) Delete() {
	C.SMPEG_delete(s.mpeg)
}

// Current status of an SMPEG object
func (s *SMPEG) Status() int {
	return int(C.SMPEG_status(s.mpeg))
}

// Sets the audio volume of an MPEG stream, in the range 0-100
func (s *SMPEG) SetVolume(volume int) {
	if volume < 0 {
		volume = 0
	} else if volume > 100 {
		volume = 100
	}

	C.SMPEG_setvolume(s.mpeg, C.int(volume))
}

// Sets the frame display callback for MPEG video
// 'lock' is a mutex used to synchronize access to the frame data,
// and is held during the update callback
func (s *SMPEG) SetDisplay(data unsafe.Pointer, lock *sdl.Mutex) {
	C.SMPEG_setdisplay(s.mpeg, (C.SMPEG_DisplayCallback)(unsafe.Pointer(C.displayCallback)), data, (*C.SDL_mutex)(unsafe.Pointer(lock)))
}

//export displayCallback
func displayCallback(data unsafe.Pointer, frame *C.SMPEG_Frame) {
	context := (*Context)(data)

	f := &Frame{}
	f.W = uint32(frame.w)
	f.H = uint32(frame.h)
	f.ImageWidth = uint32(frame.image_width)
	f.ImageHeight = uint32(frame.image_height)
	f.Image = (*uint8)(frame.image)

	context.Frame = f
	context.FrameCount += 1
}

// Sets or clears looping play on an SMPEG object
func (s *SMPEG) Loop(repeat int) {
	C.SMPEG_loop(s.mpeg, C.int(repeat))
}

// Plays an SMPEG object
func (s *SMPEG) Play() {
	C.SMPEG_play(s.mpeg)
}

// Pauses/Resumes playback of an SMPEG object
func (s *SMPEG) Pause() {
	C.SMPEG_pause(s.mpeg)
}

// Stops playback of an SMPEG object
func (s *SMPEG) Stop() {
	C.SMPEG_stop(s.mpeg)
}

// Rewinds the play position of an SMPEG object to the beginning of the MPEG
func (s *SMPEG) Rewind() {
	C.SMPEG_rewind(s.mpeg)
}

// Seeks in the MPEG stream
func (s *SMPEG) Seek(bytes int) {
	C.SMPEG_seek(s.mpeg, C.int(bytes))
}

// Skips seconds in the MPEG stream
func (s *SMPEG) Skip(seconds float32) {
	C.SMPEG_skip(s.mpeg, C.float(seconds))
}

// Renders a particular frame in the MPEG video
func (s *SMPEG) RenderFrame(framenum int) {
	C.SMPEG_renderFrame(s.mpeg, C.int(framenum))
}

// Renders the last frame of an MPEG video
func (s *SMPEG) RenderFinal() {
	C.SMPEG_renderFinal(s.mpeg)
}

// Returns error if there was a fatal error in the MPEG stream for the SMPEG object
func (s *SMPEG) Error() error {
	e := C.SMPEG_error(s.mpeg)
	if e == nil {
		return nil
	}
	return errors.New(C.GoString(e))
}

// Callback function for audio playback.
// Takes a buffer and the amount of data to fill, and returns
// the amount of data in bytes that was actually written
func (s *SMPEG) PlayAudio(stream []byte, len int) int {
	return int(C.SMPEG_playAudio(s.mpeg, (*C.Uint8)(&stream[0]), C.int(len)))
}

// Wrapper for PlayAudio() that can be passed to SDL and SDL_mixer
func (s *SMPEG) PlayAudioSDL(stream []byte, len int) {
	C.SMPEG_playAudioSDL(s.mpeg, (*C.Uint8)(&stream[0]), C.int(len))
}

// Gets the best SDL audio spec for the audio stream
func (s *SMPEG) WantedSpec(wanted *sdl.AudioSpec) {
	C.SMPEG_wantedSpec(s.mpeg, (*C.SDL_AudioSpec)(unsafe.Pointer(wanted)))
}

// Informs SMPEG of the actual SDL audio spec used for sound playback
func (s *SMPEG) ActualSpec(spec *sdl.AudioSpec) {
	C.SMPEG_actualSpec(s.mpeg, (*C.SDL_AudioSpec)(unsafe.Pointer(spec)))
}
