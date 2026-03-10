// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
// +build windows

package declarative

import (
	"github.com/hwhaocool/walk"
	"github.com/lxn/win"
)

type LabelV2 struct {
	// Window

	Accessibility      Accessibility
	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size
	Name               string
	OnBoundsChanged    walk.EventHandler
	OnKeyDown          walk.KeyEventHandler
	OnKeyPress         walk.KeyEventHandler
	OnKeyUp            walk.KeyEventHandler
	OnMouseDown        walk.MouseEventHandler
	OnMouseMove        walk.MouseEventHandler
	OnMouseUp          walk.MouseEventHandler
	OnSizeChanged      walk.EventHandler
	Persistent         bool
	RightToLeftReading bool
	ToolTipText        Property
	Visible            Property

	// Widget

	Alignment          Alignment2D //控件水平和垂直对齐方式
	AlwaysConsumeSpace bool
	Column             int
	ColumnSpan         int
	GraphicsEffects    []walk.WidgetGraphicsEffect //图形效果
	Row                int
	RowSpan            int
	StretchFactor      int

	// Label

	AssignTo      **walk.LabelV2M
	EllipsisMode  EllipsisMode //文本超出控件宽度时省略显示
	NoPrefix      bool
	Text          Property
	TextAlignment Alignment1D //文本水平对齐方式
	TextColor     walk.Color  //文本颜色
}

func (l LabelV2) Create(builder *Builder) error {
	var style uint32
	if l.NoPrefix {
		style |= win.SS_NOPREFIX
	}

	w, err := walk.NewLabelV2WithStyle(builder.Parent(), style)
	if err != nil {
		return err
	}

	if l.AssignTo != nil {
		*l.AssignTo = w
	}

	return builder.InitWidget(l, w, func() error {
		// if err := w.SetEllipsisMode(walk.EllipsisMode(l.EllipsisMode)); err != nil {
		// 	return err
		// }

		// if err := w.SetTextAlignment(walk.Alignment1D(l.TextAlignment)); err != nil {
		// 	return err
		// }

		w.SetTextColor(l.TextColor)

		w.InitTextFormat()

		return nil
	})
}
