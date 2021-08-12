package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/0xAX/notificator"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/getlantern/systray"
)

//go:embed icons/network-idle.png
var networkIdleIcon []byte

//go:embed icons/network-error.png
var networkErrorIcon []byte

//go:embed icons/system-log-out.png
var systemLogOutIcon []byte

//go:embed sounds/online.wav
var onlineSound []byte

//go:embed sounds/offline.wav
var offlineSound []byte

var notify *notificator.Notificator

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(networkIdleIcon)
	// systray.SetTitle("NetStatus")

	notify = notificator.New(notificator.Options{
		AppName: "NetStatus",
	})

	ctx, cancel := context.WithCancel(context.Background())

	mDNS := systray.AddMenuItem("DNS test: checking", "Checking")
	mHTTP := systray.AddMenuItem("HTTP test: checking", "Checking")

	systray.AddSeparator()

	mSounds := systray.AddMenuItemCheckbox("Sounds", "Enable sounds", true)
	mNotifications := systray.AddMenuItemCheckbox("Notifications", "Enable notifications", true)

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit application")
	mQuit.SetIcon(systemLogOutIcon)

	sounds := true
	notifications := true

	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			select {
			case <-mQuit.ClickedCh:
				log.Println("Clicked")
				cancel()
				return
			case <-mSounds.ClickedCh:
				sounds = !sounds
			case <-mNotifications.ClickedCh:
				notifications = !notifications
			}
		}
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		lastState := false
		for {
			log.Println(mSounds.Checked())
			log.Println("Checking")
			dnsErr := CheckDNS()
			UpdateItem(mDNS, "DNS test", dnsErr)
			httpErr := CheckHTTP()
			UpdateItem(mHTTP, "HTTP test", httpErr)
			newState := dnsErr == nil && httpErr == nil
			if newState != lastState {
				if newState {
					systray.SetIcon(networkIdleIcon)
					if notifications {
						notify.Push("NetStatus", "Internet connection is up.", "", notificator.UR_NORMAL)
					}
					if sounds {
						PlaySound(onlineSound)
						<-time.After(time.Second / 4)
					}
				} else {
					systray.SetIcon(networkErrorIcon)
					if notifications {
						notify.Push("NetStatus", "Internet connection is DOWN.", "", notificator.UR_CRITICAL)
					}
					if sounds {
						PlaySound(offlineSound)
						<-time.After(time.Second / 4)
					}
				}
			}
			lastState = newState
			select {
			case <-time.After(time.Second * 5):
			case <-ctx.Done():
				log.Println("Stopping")
				return
			}
		}
	}()
	<-ctx.Done()
	wg.Wait()
	systray.Quit()
}

func UpdateItem(item *systray.MenuItem, title string, err error) {
	if err != nil {
		item.SetIcon(networkErrorIcon)
		item.SetTitle(title + ": failed")
		item.SetTooltip(err.Error())
	} else {
		item.SetIcon(networkIdleIcon)
		item.SetTitle(title + ": ok")
		item.SetTooltip("OK")
	}
}

func CheckDNS() error {
	_, err := LookupHost("www.google.com", time.Second*2)
	return err
}

func CheckHTTP() error {
	client := http.Client{Timeout: time.Second * 2}
	_, err := client.Get("http://clients3.google.com/generate_204")
	return err
}

func onExit() {
}

func LookupHost(hostname string, timeout time.Duration) ([]string, error) {
	c1 := make(chan []string)
	c2 := make(chan error)

	var ipaddr []string
	var err error

	go func() {
		var ipaddr []string
		ipaddr, err := net.LookupHost(hostname)
		if err != nil {
			c2 <- err
		}

		c1 <- ipaddr
	}()

	select {
	case ipaddr = <-c1:
	case err = <-c2:
	case <-time.After(timeout):
		return ipaddr, errors.New("timeout")
	}

	if err != nil {
		return ipaddr, errors.New("timeout")
	}

	return ipaddr, nil
}

func PlaySound(data []byte) {
	streamer, format, err := wav.Decode(bytes.NewBuffer(data))
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}
	speaker.Play(streamer)
}
