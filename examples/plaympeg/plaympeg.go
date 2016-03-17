package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/gen2brain/go-smpeg2/smpeg"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	flag.Usage = func() {
		us := "\nUsage: %s [options] <file.mpg>\n\n"
		fmt.Fprintf(os.Stderr, us, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Println()
	}

	fullscreen := flag.Bool("fullscreen", false, "Play MPEG in fullscreen mode")
	noAudio := flag.Bool("noaudio", false, "Don't play audio stream")
	noVideo := flag.Bool("novideo", false, "Don't play video stream")

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	useAudio := !*noAudio
	useVideo := !*noVideo
	fullScreen := *fullscreen

	var err error
	var window *sdl.Window
	var renderer *sdl.Renderer
	var texture *sdl.Texture

	context := &smpeg.Context{}

	context.Lock, err = sdl.CreateMutex()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	err = sdl.Init(sdl.INIT_VIDEO)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if useAudio {
		err = sdl.Init(sdl.INIT_AUDIO)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			useAudio = false
		}
	}

	mpg, err := smpeg.New(flag.Args()[0], useAudio)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	mpg.EnableAudio(useAudio)
	mpg.EnableVideo(useVideo)

	info := mpg.Info()

	if info.HasAudio && info.HasVideo {
		fmt.Println("MPEG system stream (audio/video)")
	} else if info.HasAudio {
		fmt.Println("MPEG audio stream")
	} else if info.HasVideo {
		fmt.Println("MPEG video stream")
	}

	if info.HasVideo {
		fmt.Printf("Video %dx%d resolution\n", info.Width, info.Height)
	}

	if info.HasAudio {
		fmt.Printf("Audio %s\n", info.AudioString)
	}

	fmt.Printf("Size: %d\n", info.TotalSize)
	fmt.Printf("Total time: %f\n", info.TotalTime)

	windowFlags := sdl.WINDOW_RESIZABLE
	if fullScreen {
		windowFlags |= sdl.WINDOW_FULLSCREEN
	}

	if info.HasVideo && useVideo {
		window, err = sdl.CreateWindow(filepath.Base(flag.Args()[0]), sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, info.Width, info.Height, uint32(windowFlags))
		if err != nil {
			return
		}

		renderer, err = sdl.CreateRenderer(window, -1, 0)
		if err != nil {
			return
		}

		textureWidth := info.Width + 15 & ^15
		textureHeight := info.Height + 15 & ^15
		texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_YV12, sdl.TEXTUREACCESS_STREAMING, textureWidth, textureHeight)
		if err != nil {
			return
		}

		mpg.SetDisplay(unsafe.Pointer(context), context.Lock)
	} else {
		sdl.QuitSubSystem(sdl.INIT_VIDEO)
		useVideo = false
	}

	mpg.Play()

	frameCount := 0
	running := true

	for running {

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.WindowEvent:
				if t.Event == sdl.WINDOWEVENT_SIZE_CHANGED {
					renderer.SetViewport(nil)
				}
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyDownEvent:
				if t.Keysym.Scancode == sdl.SCANCODE_ESCAPE || t.Keysym.Scancode == sdl.SCANCODE_Q {
					running = false
				} else if t.Keysym.Scancode == sdl.SCANCODE_RETURN {
					if t.Keysym.Mod&sdl.KMOD_ALT != 0 {
						fullScreen = !fullScreen
						flags := 0
						if fullScreen {
							flags = sdl.WINDOW_FULLSCREEN
						}
						window.SetFullscreen(uint32(flags))
					}
				} else if t.Keysym.Scancode == sdl.SCANCODE_SPACE {
					mpg.Pause()
				} else if t.Keysym.Scancode == sdl.SCANCODE_RIGHT {
					mpg.Skip(5)
				}
			}
		}

		if useVideo && context.FrameCount > frameCount {
			sdl.LockMutex(context.Lock)
			texture.Update(nil, unsafe.Pointer(context.Frame.Image), int(context.Frame.ImageWidth))
			sdl.UnlockMutex(context.Lock)

			src := &sdl.Rect{0, 0, int32(info.Width), int32(info.Height)}
			renderer.Copy(texture, src, nil)

			renderer.Present()
		} else {
			sdl.Delay(0)
		}

	}

	mpg.Delete()

	sdl.DestroyMutex(context.Lock)
	sdl.Quit()
}
