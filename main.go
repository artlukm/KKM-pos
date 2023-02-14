package main

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/amdf/ixxatvci3"
	"github.com/amdf/ixxatvci3/candev"
)

var canOk = make(chan int)
var can *candev.Device
var b candev.Builder

var bOkCAN bool
var bConnected bool

var labelConnect = widget.NewLabel("")
var labelSpace = widget.NewLabel("")
var labelHeadCurPos = widget.NewLabel("")
var labelCurPos = widget.NewLabel("")
var labelHeadOldPos = widget.NewLabel("")
var labelOldPos = widget.NewLabel("")

var craneTurn bool = false // If the crane handle was not turned, we do not show the previous position

var redLine1 = canvas.NewLine(color.NRGBA{255, 0, 0, 255})

var redLine2 = canvas.NewLine(color.NRGBA{255, 0, 0, 255})

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Позиция ККМ")
	w.Resize(fyne.NewSize(400, 400))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	labelCurPos.TextStyle = fyne.TextStyle{Bold: true}

	redLine1.StrokeWidth = 5
	redLine1.Hide()
	redLine2.StrokeWidth = 5
	redLine2.Hide()

	content := container.NewVBox(
		labelConnect,
		labelSpace,
		redLine1,
		labelHeadCurPos,
		labelCurPos,
		redLine2,
		labelSpace,
		labelHeadOldPos,
		labelOldPos,
	)

	w.SetContent(content)

	go connectCAN()
	go processCAN()
	go processScreen()

	defer func() {
		bOkCAN = false
		resetInfo()
		can.Stop()
	}()

	w.ShowAndRun()
}

func resetInfo() {
	bConnected = false
	craneTurn = false
}

func processScreen() {
	sec := time.NewTicker(200 * time.Millisecond)
	for range sec.C {

		stringConnected := ""
		stringText := ""
		stringHeadCurPos := ""
		stringCurPos := ""
		stringHeadOldPos := ""
		stringOldPos := ""

		if bOkCAN {
			if bConnected {
				stringConnected = "Соединено с ККМ"
				stringHeadCurPos = "Текущее положение:"
				stringCurPos = CurPos.NamePos()
				redLine1.Show()
				redLine2.Show()
				if craneTurn {
					stringHeadOldPos = "Предыдущее  положение:"
					stringOldPos = OldPos.NamePos()
				}
			} else {
				stringConnected = "Ожидание соединения с ККМ..."
				redLine1.Hide()
				redLine2.Hide()

			}
		} else {
			stringConnected = "Не обнаружен адаптер USB-to-CAN"
			stringText = "Подключите адаптер, перезапустите программу"

		}

		labelConnect.SetText(stringConnected)
		labelSpace.SetText(stringText)
		labelHeadCurPos.SetText(stringHeadCurPos)
		labelCurPos.SetText(stringCurPos)
		labelHeadOldPos.SetText(stringHeadOldPos)
		labelOldPos.SetText(stringOldPos)
	}
}

// We determine the presence of a CAN adapter
func connectCAN() {
	var err error
	can, err = b.Speed(ixxatvci3.Bitrate25kbps).Get()
	for {
		if err == nil {
			can.Run()
			canOk <- 1
			bOkCAN = true
			time.Sleep(500 * time.Millisecond)
		} else {
			bOkCAN = false
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// We check the availability of communication by the messages issued to them
func threadActivity() {
	for {
		_, err1 := can.GetMsgByID(KKM_ID1, 2*time.Second)
		_, err2 := can.GetMsgByID(KKM_ID2, 2*time.Second)
		if err1 == nil || err2 == nil {
			bConnected = true
		} else {
			resetInfo()
		}
	}
}

func processCAN() {

	<-canOk
	bOkCAN = true
	var newPos KKMPosition

	ch, _ := can.GetMsgChannelCopy()
	go threadActivity()

	for msg := range ch {

		switch msg.ID {
		case KKM_ID1:
			fallthrough
		case KKM_ID2:

			for {
				if msg.Data[0] == gCANPosValues[newPos][0] && msg.Data[1] == gCANPosValues[newPos][1] {
					if CurPos != newPos {
						if OldPos != CurPos {
							OldPos = CurPos
							craneTurn = true
						}
						CurPos = newPos
					}
					break
				} else {
					newPos++
				}
			}
		}

		newPos = 0
		time.Sleep(20 * time.Millisecond)
	}
}
