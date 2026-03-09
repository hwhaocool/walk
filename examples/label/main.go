package main

import (
	"fmt"

	. "github.com/hwhaocool/walk/declarative"
)

func main() {

	n, err := MainWindow{
		Title: "Label Demo",
		Size:  Size{Width: 800, Height: 800},
		Layout: VBox{
			Alignment: AlignHNearVNear,
		},

		Children: []Widget{
			Composite{
				Layout: VBox{
					Alignment: AlignHNearVNear,
				},
				Children: []Widget{
					Label{
						Text: "请选择文件，可以输入，也可以拖拽❤️",
					},
					LabelV2{
						EllipsisMode: EllipsisEnd,
						Text:         "2 ❤️ 2 22222222",
						Font:         Font{PointSize: 18, Family: "DejaVu Sans Mono"},
					},
				},
			},
		},
	}.Run()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("main end", n)
}
