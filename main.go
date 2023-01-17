package main

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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

var craneTurn bool = false // Если ручку крана не поворачивали не показываем предыдущее положение

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Позиция ККМ")
	w.Resize(fyne.NewSize(400, 400))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	// labelNull := widget.NewLabel("")

	topLabel := container.NewVBox(
		labelConnect,
		labelSpace,
		labelHeadCurPos,
		labelCurPos,
		labelSpace,
		labelHeadOldPos,
		labelOldPos,
	)

	content := container.NewWithoutLayout(topLabel) // создаем специальный контейнер в котором будут размещаться элементы (положение)

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
				stringHeadCurPos = "Текущее положение"
				stringCurPos = CurPos.NamePos()
				if craneTurn {
					stringHeadOldPos = "Предыдущее  положение"
					stringOldPos = OldPos.NamePos()
				}

			} else {

				stringConnected = "Ожидание соединения с ККМ..."

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

// Определяем наличие CAN адаптера
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

func processCAN() {
	<-canOk
	ch, _ := can.GetMsgChannelCopy()
	var posDefine bool
	var newPos KKMPosition
	for msg := range ch {

		if msg.ID == KKM_ID {
			bConnected = true // Проверяем наличие связи с ИПТМ по выдаваемым им сообщениям
			posDefine = false
			for !posDefine {
				if msg.Data[0] == gCANPosValues[newPos][0] && msg.Data[1] == gCANPosValues[newPos][1] {

					if CurPos != newPos {
						if OldPos != CurPos {
							OldPos = CurPos
							craneTurn = true
						}

						CurPos = newPos
					}
					newPos = 0
					posDefine = true

				} else {
					newPos++
				}
			}
		} else {
			resetInfo()
		}

		time.Sleep(20 * time.Millisecond)

	}
}
