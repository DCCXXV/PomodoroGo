package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// types
type C = layout.Context
type D = layout.Dimensions

// timer variables
var (
	progress    float32
	incrementer chan float32

	// is it running?
	isRunning bool
	isBreak   = false

	// autorun?
	auto bool

	// time
	studyTime = 25 * 60
	breakTime = 5 * 60

	laps = 0
)

func draw(w *app.Window) error {

	// timer logic
	go func() {
		for range incrementer {
			if isRunning && progress < 1 {
				if !isBreak {
					progress += 1.0 / 25.0 / float32(studyTime)
				} else {
					progress += 1.0 / 25.0 / float32(breakTime)
				}

				if progress >= 1 {
					if auto {
						if !isBreak {
							laps++
						}
						progress = 0
						isBreak = !isBreak
					} else {
						progress = 1
					}
				}

				w.Invalidate()
			}
		}
	}()

	th := material.NewTheme()

	var ops op.Ops

	var mainButton widget.Clickable
	var resetButton widget.Clickable
	var autorun widget.Bool
	var studyEd widget.Editor
	var breakEd widget.Editor

	for {
		evt := w.Event()

		switch e := evt.(type) {

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// logic
			if mainButton.Clicked(gtx) {
				isRunning = !isRunning

				if studyEd.Text() != "" {
					aux, _ := strconv.ParseInt(strings.TrimSpace(studyEd.Text()), 10, 64)
					studyTime = int(aux * 60)
				}

				if breakEd.Text() != "" {
					aux, _ := strconv.ParseInt(strings.TrimSpace(breakEd.Text()), 10, 64)
					breakTime = int(aux * 60)
				}

				if progress >= 1 {
					if !isBreak {
						laps++
					}
					progress = 0
					isBreak = !isBreak
					isRunning = true
				}
			}

			auto = autorun.Value

			if resetButton.Clicked(gtx) {
				laps = 0
				isRunning = false
				isBreak = false
				progress = 0
			}

			var remainingTime int

			if !isBreak { // studyTime
				remainingTime = studyTime - int(progress*float32(studyTime))
			} else { // breakTime
				remainingTime = breakTime - int(progress*float32(breakTime))
			}

			minutes := remainingTime / 60
			seconds := remainingTime % 60
			timeLabel := fmt.Sprintf("%02d:%02d", minutes, seconds)

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceBetween,
			}.Layout(gtx,
				// timer bar
				layout.Rigid(
					func(gtx C) D {
						bar := material.ProgressBar(th, 1-progress)
						bar.Height = 10
						bar.Radius = 0

						if isBreak {
							bar.Color = color.NRGBA{R: 100, G: 149, B: 237, A: 255}
						} else {
							bar.Color = color.NRGBA{R: 181, G: 63, B: 77, A: 255}
						}

						bar.TrackColor = color.NRGBA{A: 0}
						return bar.Layout(gtx)
					},
				),

				// laps + autorun checkbox
				layout.Rigid(
					func(gtx C) D {
						return layout.Flex{
							Axis:    layout.Horizontal,
							Spacing: layout.SpaceBetween,
						}.Layout(gtx,

							layout.Rigid(
								func(gtx C) D {
									checkbox := material.CheckBox(th, &autorun, "AutoRun")
									checkbox.TextSize = unit.Sp(16)
									checkbox.Font.Weight = font.Medium
									return checkbox.Layout(gtx)
								},
							),

							layout.Rigid(
								func(gtx C) D {
									label := material.H6(th, fmt.Sprintf(" Laps: %d ", laps))
									label.Font.Weight = font.Medium
									return label.Layout(gtx)
								},
							),
						)
					},
				),

				// timer label
				layout.Rigid(
					func(gtx C) D {
						return layout.Center.Layout(gtx,
							func(gtx C) D {
								h1 := material.H1(th, timeLabel)
								h1.Font.Weight = font.Medium
								return h1.Layout(gtx)
							},
						)
					},
				),

				// main button (Start/Stop)
				layout.Rigid(
					func(gtx C) D {
						margins := layout.Inset{
							Top:    unit.Dp(35),
							Bottom: unit.Dp(35),
							Left:   unit.Dp(55),
							Right:  unit.Dp(55),
						}

						return margins.Layout(gtx,
							func(gtx C) D {
								return layout.Flex{}.Layout(gtx,
									layout.Flexed(1,
										func(gtx C) D {
											margins := layout.Inset{
												Right: unit.Dp(10),
											}

											return margins.Layout(gtx,
												func(gtx C) D {
													var text string
													text = "Go!"

													if isRunning && progress < 1 {
														text = "Stop!"
													}
													if isRunning && progress >= 1 {
														text = "Repeat"
													}
													btn := material.Button(th, &mainButton, text)
													btn.TextSize = unit.Sp(16)
													return btn.Layout(gtx)
												},
											)
										},
									),

									layout.Rigid(
										func(gtx C) D {
											btn := material.Button(th, &resetButton, "Reset")
											btn.TextSize = unit.Sp(16)
											btn.Background = color.NRGBA{R: 66, G: 70, B: 96, A: 255}
											return btn.Layout(gtx)
										},
									),
								)
							},
						)
					},
				),

				// bottom bar
				layout.Rigid(
					func(gtx C) D {
						return layout.Flex{
							Axis:    layout.Horizontal,
							Spacing: layout.SpaceEvenly,
						}.Layout(gtx,
							layout.Rigid(
								func(gtx C) D {
									return material.Label(th, unit.Sp(16), "Study time: ").Layout(gtx)
								},
							),

							layout.Rigid(
								func(gtx C) D {
									minutes := studyTime / 60
									seconds := studyTime % 60
									text := fmt.Sprintf("%02d:%02d", minutes, seconds)

									ed := material.Editor(th, &studyEd, text)
									ed.Editor.SingleLine = true

									return ed.Layout(gtx)
								},
							),

							layout.Rigid(
								func(gtx C) D {
									return material.Label(th, unit.Sp(16), "Break time: ").Layout(gtx)
								},
							),

							layout.Rigid(
								func(gtx C) D {
									minutes := breakTime / 60
									seconds := breakTime % 60
									text := fmt.Sprintf("%02d:%02d", minutes, seconds)

									ed := material.Editor(th, &breakEd, text)
									ed.Editor.SingleLine = true

									return ed.Layout(gtx)
								},
							),
						)
					},
				),
			)

			e.Frame(gtx.Ops)

		case app.DestroyEvent:
			return e.Err
		}
	}
}

func main() {

	incrementer = make(chan float32)

	go func() {
		for {
			time.Sleep(time.Second / 25)
			incrementer <- 0.004
		}
	}()

	go func() {
		w := new(app.Window)
		w.Option(app.Title("PomodoroGo"))
		w.Option(app.Size(unit.Dp(600), unit.Dp(310)))
		//w.Option(app.Decorated(false))

		if err := draw(w); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	app.Main()
}
