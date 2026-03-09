//go:build windows
// +build windows

package walk

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

func YellowLogger(widget string, funcName string, target string, args ...interface{}) {
	fmt.Println(fill2Len(widget, 16), fill2Len(funcName, 16), fill2Len(target, 16), args)
}

func fill2Len(s string, i int) string {

	if len(s) >= i {
		return s
	}
	for len(s) < i {
		s = s + strings.Repeat(" ", i-len(s))
	}
	return s

}

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(YellowStaticWindowClass)
		YellowStaticWndProcPtr = syscall.NewCallback(YellowLabelMWndProc)
	})
}

const YellowStaticWindowClass = `\o/ Yellow_Walk_Static_Class \o/`

var YellowStaticWndProcPtr uintptr

type LabelV2M struct {
	WidgetBase
	HwndStatic win.HWND

	MyTextColor   Color
	TextAlignment Alignment2D

	InnerText            string
	textChangedPublisher EventPublisher

	origStaticWndProcPtr uintptr
}

func (s *LabelV2M) SetText(param1 string) error {
	YellowLogger("YellowLabelM", "SetText", "param1", param1)

	s.InnerText = param1

	if err := s.WidgetBase.SetText2(param1); err != nil {
		return err
	}

	if err := SetWindowText(s.HwndStatic, param1); err != nil {
		return err
	}

	s.RequestLayout()

	return nil
}

func (s *LabelV2M) Text() string {
	return s.InnerText
}

func NewLabelV2WithStyle(parent Container, style uint32) (*LabelV2M, error) {

	fmt.Println("NewLabelWithStyle ------------")

	l := new(LabelV2M)

	if err := l.Init(l, parent, style); err != nil {
		return nil, err
	}

	l.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return l.Text()
		},
		func(v interface{}) error {
			return l.SetText(assertStringOr(v, ""))
		},
		l.textChangedPublisher.Event()),
	)

	return l, nil
}

func (s *LabelV2M) Init(widget Widget, parent Container, style uint32) error {
	if err := InitWidget(
		widget,
		parent,
		YellowStaticWindowClass,
		win.WS_VISIBLE|(style&win.WS_BORDER),
		win.WS_EX_CONTROLPARENT); err != nil {
		return err
	}

	// 子窗口
	if s.HwndStatic = win.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr("static"),
		nil,
		win.WS_CHILD|win.WS_CLIPSIBLINGS|win.WS_VISIBLE|win.SS_LEFT|win.SS_NOTIFY|(style&^win.WS_BORDER),
		win.CW_USEDEFAULT,
		win.CW_USEDEFAULT,
		100,
		100, // 让它自己计算大小
		s.HWnd,
		0,
		0,
		nil,
	); s.HwndStatic == 0 {
		return NewError("creating static failed")
	}

	if err := s.Group.toolTip.AddTool(s); err != nil {
		return err
	}

	s.origStaticWndProcPtr = win.SetWindowLongPtr(s.HwndStatic, win.GWLP_WNDPROC, YellowStaticWndProcPtr)
	if s.origStaticWndProcPtr == 0 {
		return lastError("SetWindowLongPtr")
	}

	s.applyFont(s.Font())

	s.SetBackground(nullBrushSingleton)

	s.SetAlignment(AlignHNearVCenter)

	return nil
}

func (s *LabelV2M) Dispose() {
	if s.HwndStatic != 0 {
		win.DestroyWindow(s.HwndStatic)
		s.HwndStatic = 0
	}

	s.WidgetBase.Dispose()
}

func (s *LabelV2M) handleForToolTip() win.HWND {
	return s.HwndStatic
}

func (s *LabelV2M) applyEnabled(enabled bool) {
	s.WidgetBase.applyEnabled(enabled)

	SetWindowEnabled(s.HwndStatic, enabled)
}

func (s *LabelV2M) applyFont(font *Font) {
	s.WidgetBase.applyFont(font)

	SetWindowFont(s.HwndStatic, font)
}

func (s *LabelV2M) textAlignment1D() Alignment1D {
	switch s.TextAlignment {
	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		return AlignCenter

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		return AlignFar

	default:
		return AlignNear
	}
}

func (s *LabelV2M) setTextAlignment1D(alignment Alignment1D) error {
	var align Alignment2D

	switch alignment {
	case AlignCenter:
		align = AlignHCenterVCenter

	case AlignFar:
		align = AlignHFarVCenter

	default:
		align = AlignHNearVCenter
	}

	return s.setTextAlignment(align)
}

func (s *LabelV2M) setTextAlignment(alignment Alignment2D) error {
	if alignment == s.TextAlignment {
		return nil
	}

	var styleBit uint32

	switch alignment {
	case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
		styleBit |= win.SS_LEFT

	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		styleBit |= win.SS_CENTER

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		styleBit |= win.SS_RIGHT
	}

	if err := SetAndClearWindowLongBits(s.HwndStatic, win.GWL_STYLE, styleBit, win.SS_LEFT|win.SS_CENTER|win.SS_RIGHT); err != nil {
		return err
	}

	s.TextAlignment = alignment

	s.Invalidate()

	return nil
}

func (s *LabelV2M) SetText2(text string) (changed bool, err error) {
	fmt.Println("YellowLabelM", "SetText2", "text", s.text(), "<->", text)

	s.InnerText = text
	if text == s.text() {
		return false, nil
	}

	if err := s.WidgetBase.SetText2(text); err != nil {
		return false, err
	}

	if err := SetWindowText(s.HwndStatic, text); err != nil {
		return false, err
	}

	s.RequestLayout()

	return true, nil
}

func (s *LabelV2M) TextColor() Color {
	return s.MyTextColor
}

func (s *LabelV2M) SetTextColor(c Color) {
	s.MyTextColor = c

	s.Invalidate()
}

func (s *LabelV2M) shrinkable() (ret bool) {

	defer func() {
		// fmt.Println("YellowLabelM", "shrinkable", "return", "ret", ret)
	}()

	if em, ok := s.window.(interface{ EllipsisMode() EllipsisMode }); ok {
		ret = em.EllipsisMode() != EllipsisNone
	}

	return
}

func (s *LabelV2M) UpdateStaticBounds() {
	var format DrawTextFormat

	switch s.TextAlignment {
	case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
		format |= TextLeft

	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		format |= TextCenter

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		format |= TextRight
	}

	switch s.TextAlignment {
	case AlignHNearVNear, AlignHCenterVNear, AlignHFarVNear:
		format |= TextTop

	case AlignHNearVCenter, AlignHCenterVCenter, AlignHFarVCenter:
		format |= TextVCenter

	case AlignHNearVFar, AlignHCenterVFar, AlignHFarVFar:
		format |= TextBottom
	}

	cb := s.ClientBoundsPixels()
	YellowLogger("YellowLabelM", "UpdateStaticBounds", "ClientBoundsPixels", "cb", cb)

	if shrinkable := s.shrinkable(); shrinkable || format&TextVCenter != 0 || format&TextBottom != 0 {
		var size Size
		if _, ok := s.window.(HeightForWidther); ok {
			size = s.CalculateTextSizeForWidth(cb.Width)
		} else {
			size = s.CalculateTextSize()
		}

		if shrinkable {
			var text string
			if size.Width > cb.Width {
				text = s.InnerText
			}
			s.SetToolTipText(text)
		}

		if format&TextVCenter != 0 || format&TextBottom != 0 {
			if format&TextVCenter != 0 {
				cb.Y += (cb.Height - size.Height) / 2
			} else {
				cb.Y += cb.Height - size.Height
			}

			cb.Height = size.Height
		}
	}

	YellowLogger("YellowLabelM", "UpdateStaticBounds", "ClientBoundsPixels", "cb2", cb)

	win.MoveWindow(s.HwndStatic, int32(cb.X), int32(cb.Y), int32(cb.Width), int32(cb.Height), true)
	// win.MoveWindow(s.HwndStatic, int32(0), int32(0), int32(500), int32(500), true)

	// s.OnPaint()

	s.Invalidate()
}

func (l *LabelV2M) AsStatic() *LabelV2M {
	return l
}

func (s *LabelV2M) OnPaint() {

	bRect := s.BoundsPixels()
	YellowLogger("YellowLabelM", "OnPaint", "BoundsPixels", bRect, s.ClientBoundsPixels())

	if bRect.Width == 0 || bRect.Height == 0 {
		return
	}

	// 获取设备上下文
	hdc := win.GetDC(s.HwndStatic)
	defer win.ReleaseDC(s.HwndStatic, hdc)

	if hdc == 0 {
		return
	}

	// 设置背景为透明
	win.SetBkMode(hdc, win.TRANSPARENT)

	// 设置文本颜色（使用 s.MyTextColor）
	textColor := win.RGB(
		s.MyTextColor.R(),
		s.MyTextColor.G(),
		s.MyTextColor.B())
	win.SetTextColor(hdc, textColor)

	// 获取当前字体或设置默认字体
	hFont := s.Font().handleForDPI(s.DPI())
	var deleteEmojiFont bool

	if hFont == 0 {
		// 使用 Segoe UI Emoji 字体以支持 emoji
		var lf win.LOGFONT
		lf.LfHeight = -win.MulDiv(int32(s.Font().PointSize()), int32(s.DPI()), 72)
		if s.Font().Bold() {
			lf.LfWeight = win.FW_BOLD
		} else {
			lf.LfWeight = win.FW_NORMAL
		}
		if s.Font().Italic() {
			lf.LfItalic = 1
		}
		if s.Font().Underline() {
			lf.LfUnderline = 1
		}
		if s.Font().StrikeOut() {
			lf.LfStrikeOut = 1
		}
		lf.LfCharSet = win.DEFAULT_CHARSET
		lf.LfOutPrecision = win.OUT_TT_PRECIS
		lf.LfClipPrecision = win.CLIP_DEFAULT_PRECIS
		lf.LfQuality = win.CLEARTYPE_QUALITY
		lf.LfPitchAndFamily = win.VARIABLE_PITCH | win.FF_DONTCARE

		fontName := "Segoe UI Emoji"
		src := syscall.StringToUTF16(fontName)
		dest := lf.LfFaceName[:]
		copy(dest, src)

		hFont = win.CreateFontIndirect(&lf)
		deleteEmojiFont = (hFont != 0)
	}
	oldFont := win.SelectObject(hdc, win.HGDIOBJ(hFont))
	defer win.SelectObject(hdc, oldFont)
	defer func() {
		if deleteEmojiFont && hFont != 0 {
			win.DeleteObject(win.HGDIOBJ(hFont))
		}
	}()

	// 设置绘制格式
	var drawFormat uint32 = win.DT_SINGLELINE | win.DT_VCENTER

	// 设置水平对齐
	switch s.TextAlignment {
	case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
		drawFormat |= win.DT_LEFT
	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		drawFormat |= win.DT_CENTER
	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		drawFormat |= win.DT_RIGHT
	}

	// 设置垂直对齐
	switch s.TextAlignment {
	case AlignHNearVNear, AlignHCenterVNear, AlignHFarVNear:
		drawFormat |= win.DT_TOP
	case AlignHNearVCenter, AlignHCenterVCenter, AlignHFarVCenter:
		drawFormat |= win.DT_VCENTER
	case AlignHNearVFar, AlignHCenterVFar, AlignHFarVFar:
		drawFormat |= win.DT_BOTTOM
	}

	// 如果启了省略号模式，添加 DT_END_ELLIPSIS
	if em, ok := s.window.(interface{ EllipsisMode() EllipsisMode }); ok {
		if em.EllipsisMode() != EllipsisNone {
			drawFormat |= win.DT_END_ELLIPSIS
		}
	}

	// 使用 GDI 绘制文本，支持 emoji
	rect := win.RECT{
		Left:   0,
		Top:    0,
		Right:  int32(bRect.Width),
		Bottom: int32(bRect.Height),
	}
	// 转换为 UTF16 指针
	textUTF16, err := syscall.UTF16PtrFromString(s.InnerText)
	if err != nil {
		return
	}
	win.DrawTextEx(hdc, textUTF16, -1, &rect, drawFormat, nil)

}

func (s *LabelV2M) WndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	// YellowLogger("YellowLabelM", "WndProc", "msg", msg, "wp", wp, "lp", lp)
	switch msg {
	case win.WM_CTLCOLORSTATIC:
		if hBrush := s.HandleWMCTLCOLOR(wp, uintptr(s.HWnd)); hBrush != 0 {
			return hBrush
		}
	case win.WM_SETTEXT:
		fmt.Println("yellow WndProc SetText", s.text(), "<->", s.InnerText)

		fmt.Println("yellow WndProc SetText", "lp文本内容:", GetStringByPtr(lp))

		// s.OnPaint()
	case win.WM_SIZE:
		bRect := s.BoundsPixels()

		YellowLogger("YellowLabelM", "WndProc", "WM_SIZE", "hwnd", hwnd, "wp", wp, "lp", lp, "bRect", bRect)

		// s.OnPaint()

	case win.WM_WINDOWPOSCHANGED:
		// 位置移动

		YellowLogger("YellowLabelM", "WndProc", "WM_WINDOWPOSCHANGED", "hwnd", hwnd, "wp", wp, "lp", lp)
		wp := (*win.WINDOWPOS)(unsafe.Pointer(lp))

		if wp.Flags&win.SWP_NOSIZE != 0 {
			YellowLogger("YellowLabelM", "WndProc", "WM_WINDOWPOSCHANGED", "break")
			break
		}

		s.UpdateStaticBounds()

	case win.WM_PAINT:
		YellowLogger("YellowLabelM", "WndProc", "WM_PAINT", "hwnd", hwnd, "wp", wp, "lp", lp)
		var ps win.PAINTSTRUCT
		win.BeginPaint(hwnd, &ps)
		defer win.EndPaint(hwnd, &ps)

		// TODO: 调用 DirectWrite 绘制
		s.OnPaint()
		return 0
	}

	return s.WidgetBase.WndProc(hwnd, msg, wp, lp)
}

func YellowLabelMWndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	pHwnd := win.GetParent(hwnd)
	YellowLogger("YellowLabelM", "YellowLabelMWndProc", "params", "hwnd", hwnd, "pHwnd", pHwnd, "msg", msg, "wp", wp, "lp", lp)

	// hwnd 就是子窗口

	// as, ok := WindowFromHandle(win.GetParent(hwnd)).(interface{ AsStatic() *YellowLabelM })
	as, ok := WindowFromHandle(pHwnd).(interface{ AsStatic() *LabelV2M })
	if !ok {
		fmt.Println("YellowLabelMWndProc: not a YellowLabelM")
		return 0
	}

	s := as.AsStatic()

	switch msg {
	case win.WM_NCHITTEST:
		return win.HTCLIENT

	case win.WM_CREATE:
		// 创建独立资源
		// createIndependentResource()
		return 0
	case win.WM_SIZE:
		// 当窗口大小改变时，触发重绘
		win.InvalidateRect(s.HwndStatic, nil, false)
		return 0

	case win.WM_PAINT:
		s.OnPaint()
		// return 0

	}

	return win.CallWindowProc(s.origStaticWndProcPtr, hwnd, msg, wp, lp)
}

func (s *LabelV2M) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	var layoutFlags LayoutFlags
	if s.textAlignment1D() != AlignNear {
		layoutFlags = GrowableHorz
	} else if s.shrinkable() {
		layoutFlags = ShrinkableHorz
	}

	idealSize := s.CalculateTextSize()
	if s.hasStyleBits(win.WS_BORDER) {
		border := s.IntFrom96DPI(1) * 2
		idealSize.Width += border
		idealSize.Height += border * 2
	}

	return &staticLayoutItem{
		layoutFlags: layoutFlags,
		idealSize:   idealSize,
	}
}
