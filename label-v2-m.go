//go:build windows
// +build windows

package walk

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/bupjae/direct"
	"github.com/bupjae/direct/d2d"
	"github.com/bupjae/direct/dwrite"
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

	// Direct2D 工厂对象
	D2dFactory *d2d.IFactory
	// 渲染目标对象
	RenderTarget *d2d.IHwndRenderTarget
	// 渐变停止集合
	GradientStopCollection *d2d.IGradientStopCollection
	// 黑色画笔
	BrushBlack *d2d.IBrush
	// 渐变画笔
	BrushGradient *d2d.IBrush

	// DirectWrite 工厂对象
	DwriteFactory *dwrite.IFactory
	// 文本格式对象
	TextFormat *dwrite.ITextFormat
}

func (s *LabelV2M) SetText(param1 string) error {
	YellowLogger("LabelV2M", "SetText", "param1", param1)
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

	// 创建一个新的 Direct2D 工厂
	d2dFactory, err := d2d.CreateFactory(
		d2d.D2D1_FACTORY_TYPE_SINGLE_THREADED,
		&d2d.FactoryOptions{DebugLevel: d2d.D2D1_DEBUG_LEVEL_NONE})
	if err != nil {
		panic(err)
	}
	l.D2dFactory = d2dFactory

	// 创建一个新的 DirectWrite 工厂
	l.DwriteFactory, err = dwrite.CreateFactory(dwrite.DWRITE_FACTORY_TYPE_SHARED)
	if err != nil {
		panic(err)
	}

	// 创建一个新的文本格式
	l.TextFormat = l.newTextFormat()

	return l, nil
}

func (l *LabelV2M) newTextFormat() *dwrite.ITextFormat {
	var textTF *dwrite.ITextFormat

	// 创建新的文本格式
	t, err := l.DwriteFactory.CreateTextFormat("Segoe UI Emoji", nil, dwrite.DWRITE_FONT_WEIGHT_REGULAR, dwrite.DWRITE_FONT_STYLE_NORMAL, dwrite.DWRITE_FONT_STRETCH_NORMAL, 10, "en-US")
	if err != nil {
		panic(err)
	}

	// 设置文本对齐方式为左对齐
	if err = t.SetTextAlignment(dwrite.DWRITE_TEXT_ALIGNMENT_LEADING); err != nil {
		panic(err)
	}

	// 设置段落对齐方式为顶部对齐
	if err = t.SetParagraphAlignment(dwrite.DWRITE_PARAGRAPH_ALIGNMENT_NEAR); err != nil {
		panic(err)
	}

	textTF = t

	return textTF
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
	YellowLogger("LabelV2M", "SetText2", "text", s.text(), "<->", text)

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
		// fmt.Println("LabelV2M", "shrinkable", "return", "ret", ret)
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
	YellowLogger("LabelV2M", "UpdateStaticBounds", "ClientBoundsPixels", "cb", cb)

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

	YellowLogger("LabelV2M", "UpdateStaticBounds", "ClientBoundsPixels", "cb2", cb)

	win.MoveWindow(s.HwndStatic, int32(cb.X), int32(cb.Y), int32(cb.Width), int32(cb.Height), true)
	// win.MoveWindow(s.HwndStatic, int32(0), int32(0), int32(500), int32(500), true)

	// s.OnPaint()

	s.Invalidate()
}

func (l *LabelV2M) AsStatic() *LabelV2M {
	return l
}

func (s *LabelV2M) OnPaint() {

	// bp := parent.BoundsPixels()
	// YellowLogger("LabelV2M", "NewLabelWithStyle", "BoundsPixels", bp)

	bRect := s.BoundsPixels()
	YellowLogger("LabelV2M", "OnPaint", "BoundsPixels", bRect, s.ClientBoundsPixels())
	// s.Static.AsWidgetBase().RectangleTo96DPI()

	if bRect.Width == 0 || bRect.Height == 0 {
		return
	}

	if s.RenderTarget == nil {
		YellowLogger("LabelV2M", "OnPaint", "HwndStatic", s.HwndStatic)
		t, err := s.D2dFactory.CreateHwndRenderTarget(
			&d2d.RenderTargetProperties{},
			&d2d.HwndRenderTargetProperties{
				Hwnd: uintptr(s.HwndStatic),
				PixelSize: d2d.SizeU{
					Width:  uint32(bRect.Width),
					Height: uint32(bRect.Height),
				},
			},
		)
		if err != nil {
			panic(err)
		}
		s.RenderTarget = t
	}

	s.RenderTarget.BeginDraw()
	defer s.RenderTarget.EndDraw()

	b, err := s.RenderTarget.CreateSolidColorBrush(
		&d2d.ColorF{R: 0, G: 255, B: 0, A: 1},
		&d2d.BrushProperties{
			Opacity:   1,
			Transform: d2d.Matrix3x2F{{1, 0}, {0, 1}, {0, 0}}})
	if err != nil {
		panic(err)
	}
	brushBlack := &b.IBrush

	s.RenderTarget.DrawText(
		"✅你好拖拽❤️",
		s.TextFormat,
		&d2d.RectF{
			Left:   0,
			Top:    0,
			Right:  50,
			Bottom: 40,
		},
		brushBlack,
		// 使用裁剪选项绘制文本
		d2d.D2D1_DRAW_TEXT_OPTIONS_ENABLE_COLOR_FONT,
		// 使用自然测量模式
		direct.DWRITE_MEASURING_MODE_NATURAL)

}

func (s *LabelV2M) WndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {
	// YellowLogger("LabelV2M", "WndProc", "msg", msg, "wp", wp, "lp", lp)
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

		YellowLogger("LabelV2M", "WndProc", "WM_SIZE", "hwnd", hwnd, "wp", wp, "lp", lp, "bRect", bRect)

		// s.OnPaint()

	case win.WM_WINDOWPOSCHANGED:
		// 位置移动

		YellowLogger("LabelV2M", "WndProc", "WM_WINDOWPOSCHANGED", "hwnd", hwnd, "wp", wp, "lp", lp)
		wp := (*win.WINDOWPOS)(unsafe.Pointer(lp))

		if wp.Flags&win.SWP_NOSIZE != 0 {
			YellowLogger("LabelV2M", "WndProc", "WM_WINDOWPOSCHANGED", "break")
			break
		}

		s.UpdateStaticBounds()

	case win.WM_PAINT:
		YellowLogger("LabelV2M", "WndProc", "WM_PAINT", "hwnd", hwnd, "wp", wp, "lp", lp)
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
	YellowLogger("LabelV2M", "YellowLabelMWndProc", "params", "hwnd", hwnd, "pHwnd", pHwnd, "msg", msg, "wp", wp, "lp", lp)

	// hwnd 就是子窗口

	// as, ok := WindowFromHandle(win.GetParent(hwnd)).(interface{ AsStatic() *LabelV2M })
	as, ok := WindowFromHandle(pHwnd).(interface{ AsStatic() *LabelV2M })
	if !ok {
		fmt.Println("YellowLabelMWndProc: not a LabelV2M")
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
