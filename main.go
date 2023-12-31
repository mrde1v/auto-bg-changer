package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/joho/godotenv"
	"github.com/kardianos/service"
	"github.com/lxn/walk"
)

const (
	SPI_SETDESKWALLPAPER = 0x0014
	SPIF_UPDATEINIFILE   = 0x01
	SPIF_SENDCHANGE      = 0x02
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

func setDesktopBackground(imagePath string) error {
	imagePathPtr, err := syscall.UTF16PtrFromString(imagePath)
	if err != nil {
		return err
	}

	ret, _, err := systemParametersInfo.Call(
		uintptr(SPI_SETDESKWALLPAPER),
		0,
		uintptr(unsafe.Pointer(imagePathPtr)),
		uintptr(SPIF_UPDATEINIFILE|SPIF_SENDCHANGE),
	)
	if ret == 0 {
		return fmt.Errorf("failed to set desktop background: %v", err)
	}

	return nil
}

func getRandomImage(imagesFolder string) (string, error) {
	imageFiles, err := filepath.Glob(filepath.Join(imagesFolder, "*.*"))
	if err != nil {
		return "", err
	}

	var availableImages []string
	availableImages = append(availableImages, imageFiles...)

	if len(availableImages) == 0 {
		availableImages = imageFiles
	}

	randomIndex := rand.Intn(len(availableImages))
	randomImage := availableImages[randomIndex]

	return randomImage, nil
}

type myService struct {
	UsedImages map[string]bool
}

func (s *myService) Start(svc service.Service) error {
	go s.run()
	return nil
}

func (s *myService) Stop(svc service.Service) error {
	s.Stop(svc)
	return nil
}

func (s *myService) run() {
	godotenv.Load()
	imagesPath := os.Getenv("IMAGES_PATH")

	walk.MsgBox(nil, "Automatic Background Changer", "Time to Change Background!", walk.MsgBoxOK)

	imagePath, err := getRandomImage(imagesPath)
	if err != nil {
		return
	}
	setDesktopBackground(imagePath)
	os.Exit(1)
}

func main() {
	err := moveExecutable()
	if err != nil {
		walk.MsgBox(nil, "Automatic Background Changer", err.Error(), walk.MsgBoxIconError)
		return
	}

	err = createStartup()
	if err != nil {
		walk.MsgBox(nil, "Automatic Background Changer", err.Error(), walk.MsgBoxIconError)
		return
	}

	svcConfig := &service.Config{
		Name:        "auto-bg-changer",
		DisplayName: "Automatic Background Changer",
		Description: "Changes background on startup",
	}

	prg := &myService{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		walk.MsgBox(nil, "Automatic Background Changer", err.Error(), walk.MsgBoxIconError)
	}

	// if len(os.Args) > 1 {
	// 	err := service.Control(s, os.Args[1])
	// 	if err != nil {
	// 		walk.MsgBox(nil, "Automatic Background Changer", err.Error(), walk.MsgBoxIconError)
	// 	}
	// 	return
	// }

	err = s.Run()
	if err != nil {
		walk.MsgBox(nil, "Automatic Background Changer", err.Error(), walk.MsgBoxIconError)
	}
	os.Exit(1)
}

func moveExecutable() error {
	targetFolder := filepath.Join(os.Getenv("APPDATA"), "auto-bg-changer")
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}

	if filepath.Dir(executablePath) != targetFolder {
		err = os.MkdirAll(targetFolder, os.ModePerm)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetFolder, "autobackgroundchanger.exe")
		err = os.Rename(executablePath, targetPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func createStartup() error {
	shortcutPath := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "AutoBackgroundChanger.lnk")

	if _, err := os.Stat(shortcutPath); os.IsNotExist(err) {
		executablePath, err := os.Executable()
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(shortcutPath, []byte(fmt.Sprintf(`
[InternetShortcut]
URL=file:///%s
`, filepath.ToSlash(executablePath))), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
