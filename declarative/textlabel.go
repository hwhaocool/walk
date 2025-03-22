// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
// +build windows

package declarative

import (
	"github.com/hwhaocool/walk"
	"github.com/lxn/win"
)

// 水平和垂直对齐方式
type Alignment2D uint

const (
	// 默认的水平和垂直对齐方式
	AlignHVDefault = Alignment2D(walk.AlignHVDefault)

	// 水平靠左，垂直靠上对齐
	AlignHNearVNear = Alignment2D(walk.AlignHNearVNear)

	// 水平居中，垂直靠上对齐
	AlignHCenterVNear = Alignment2D(walk.AlignHCenterVNear)

	// 水平靠右，垂直靠上对齐
	AlignHFarVNear = Alignment2D(walk.AlignHFarVNear)

	// 水平靠左，垂直居中对齐
	AlignHNearVCenter = Alignment2D(walk.AlignHNearVCenter)

	// 水平居中，垂直居中对齐
	AlignHCenterVCenter = Alignment2D(walk.AlignHCenterVCenter)

	// 水平靠右，垂直居中对齐
	AlignHFarVCenter = Alignment2D(walk.AlignHFarVCenter)

	// 水平靠左，垂直靠下对齐
	AlignHNearVFar = Alignment2D(walk.AlignHNearVFar)

	// 水平居中，垂直靠下对齐
	AlignHCenterVFar = Alignment2D(walk.AlignHCenterVFar)

	// 水平靠右，垂直靠下对齐
	AlignHFarVFar = Alignment2D(walk.AlignHFarVFar)
)

type TextLabel struct {
	// Window

	Accessibility      Accessibility
	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size // Set MinSize.Width to a value > 0 to enable dynamic line wrapping.
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

	Alignment          Alignment2D
	AlwaysConsumeSpace bool
	Column             int
	ColumnSpan         int
	GraphicsEffects    []walk.WidgetGraphicsEffect
	Row                int
	RowSpan            int
	StretchFactor      int

	// static

	TextColor walk.Color

	// Text

	AssignTo      **walk.TextLabel
	NoPrefix      bool
	TextAlignment Alignment2D
	Text          Property
}

func (tl TextLabel) Create(builder *Builder) error {
	var style uint32
	if tl.NoPrefix {
		style |= win.SS_NOPREFIX
	}

	w, err := walk.NewTextLabelWithStyle(builder.Parent(), style)
	if err != nil {
		return err
	}

	if tl.AssignTo != nil {
		*tl.AssignTo = w
	}

	return builder.InitWidget(tl, w, func() error {
		w.SetTextColor(tl.TextColor)

		if err := w.SetTextAlignment(walk.Alignment2D(tl.TextAlignment)); err != nil {
			return err
		}

		return nil
	})
}
