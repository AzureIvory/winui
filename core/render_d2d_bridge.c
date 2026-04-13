#include <windows.h>
#include <stdint.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <wchar.h>

#include <d2d1.h>
#include <dwrite.h>
#include <wincodec.h>

#ifndef D2D1_DRAW_TEXT_OPTIONS_ENABLE_COLOR_FONT
#define D2D1_DRAW_TEXT_OPTIONS_ENABLE_COLOR_FONT ((D2D1_DRAW_TEXT_OPTIONS)0x00000004)
#endif

// 该桥接层把 Go 侧的绘制请求映射到 Direct2D、DirectWrite 和 WIC。

typedef struct WinUIDPoint {
	int32_t x;
	int32_t y;
} WinUIDPoint;

typedef struct WinUID2DRenderer {
	ID2D1Factory* factory;
	IDWriteFactory* dwrite;
	IWICImagingFactory* wic;
	ID2D1DCRenderTarget* target;
	ID2D1SolidColorBrush* brush;
	int com_initialized;
} WinUID2DRenderer;

void winui_d2d_renderer_destroy(WinUID2DRenderer* renderer);

static const IID WINUI_IID_ID2D1Factory = {0x06152247, 0x6F50, 0x465A, {0x92, 0x45, 0x11, 0x8B, 0xFD, 0x3B, 0x60, 0x07}};
static const IID WINUI_IID_IDWriteFactory = {0xB859EE5A, 0xD838, 0x4B5B, {0xA2, 0xE8, 0x1A, 0xDC, 0x7D, 0x93, 0xDB, 0x48}};
static const IID WINUI_IID_IWICImagingFactory = {0xEC5EC8A9, 0xC395, 0x4314, {0x9C, 0x77, 0x54, 0xD7, 0xA9, 0x35, 0xFF, 0x70}};
static const CLSID WINUI_CLSID_WICImagingFactory = {0xCACAF262, 0x9370, 0x4615, {0xA1, 0x3B, 0x9F, 0x55, 0x39, 0xDA, 0x4C, 0x0A}};

static void winui_set_error(char* err, size_t err_len, const char* op, HRESULT hr) {
	if (err == NULL || err_len == 0) {
		return;
	}
	if (op == NULL) {
		op = "Direct2D";
	}
	snprintf(err, err_len, "%s failed: 0x%08lX", op, (unsigned long)hr);
	err[err_len - 1] = '\0';
}

static void winui_clear_error(char* err, size_t err_len) {
	if (err == NULL || err_len == 0) {
		return;
	}
	err[0] = '\0';
}

static D2D1_COLOR_F winui_color(uint32_t color) {
	D2D1_COLOR_F out;
	out.r = (FLOAT)(color & 0xFF) / 255.0f;
	out.g = (FLOAT)((color >> 8) & 0xFF) / 255.0f;
	out.b = (FLOAT)((color >> 16) & 0xFF) / 255.0f;
	out.a = 1.0f;
	return out;
}

static D2D1_RECT_F winui_rect(int32_t left, int32_t top, int32_t right, int32_t bottom) {
	D2D1_RECT_F rect;
	rect.left = (FLOAT)left;
	rect.top = (FLOAT)top;
	rect.right = (FLOAT)right;
	rect.bottom = (FLOAT)bottom;
	return rect;
}

static D2D1_ROUNDED_RECT winui_round_rect(int32_t left, int32_t top, int32_t right, int32_t bottom, FLOAT radius) {
	D2D1_ROUNDED_RECT rounded;
	rounded.rect = winui_rect(left, top, right, bottom);
	rounded.radiusX = radius;
	rounded.radiusY = radius;
	return rounded;
}

static HRESULT winui_ensure_target(WinUID2DRenderer* renderer) {
	if (renderer->target != NULL) {
		return S_OK;
	}

	D2D1_RENDER_TARGET_PROPERTIES props;
	props.type = D2D1_RENDER_TARGET_TYPE_DEFAULT;
	props.pixelFormat.format = DXGI_FORMAT_B8G8R8A8_UNORM;
	props.pixelFormat.alphaMode = D2D1_ALPHA_MODE_IGNORE;
	props.dpiX = 96.0f;
	props.dpiY = 96.0f;
	props.usage = D2D1_RENDER_TARGET_USAGE_NONE;
	props.minLevel = D2D1_FEATURE_LEVEL_DEFAULT;

	return ID2D1Factory_CreateDCRenderTarget(renderer->factory, &props, &renderer->target);
}

static HRESULT winui_ensure_brush(WinUID2DRenderer* renderer, uint32_t color) {
	HRESULT hr;

	hr = winui_ensure_target(renderer);
	if (FAILED(hr)) {
		return hr;
	}

	if (renderer->brush == NULL) {
		D2D1_COLOR_F initial = winui_color(color);
		hr = ID2D1DCRenderTarget_CreateSolidColorBrush(renderer->target, &initial, NULL, &renderer->brush);
		if (FAILED(hr)) {
			return hr;
		}
	}

	D2D1_COLOR_F value = winui_color(color);
	ID2D1SolidColorBrush_SetColor(renderer->brush, &value);
	return S_OK;
}

static HRESULT winui_create_text_format(
	WinUID2DRenderer* renderer,
	const uint16_t* font_family,
	FLOAT font_size,
	int32_t font_weight,
	uint32_t format_flags,
	IDWriteTextFormat** out_format
) {
	HRESULT hr;
	IDWriteTextFormat* format = NULL;
	const wchar_t* family = (const wchar_t*)font_family;
	const wchar_t* locale = L"en-us";
	DWRITE_FONT_WEIGHT weight = font_weight > 0 ? (DWRITE_FONT_WEIGHT)font_weight : DWRITE_FONT_WEIGHT_NORMAL;

	if (family == NULL || family[0] == L'\0') {
		family = L"Microsoft YaHei UI";
	}
	if (font_size <= 0.0f) {
		font_size = 16.0f;
	}

	hr = IDWriteFactory_CreateTextFormat(
		renderer->dwrite,
		family,
		NULL,
		weight,
		DWRITE_FONT_STYLE_NORMAL,
		DWRITE_FONT_STRETCH_NORMAL,
		font_size,
		locale,
		&format
	);
	if (FAILED(hr)) {
		return hr;
	}

	if (format_flags & 0x00000001) {
		IDWriteTextFormat_SetTextAlignment(format, DWRITE_TEXT_ALIGNMENT_CENTER);
	} else {
		IDWriteTextFormat_SetTextAlignment(format, DWRITE_TEXT_ALIGNMENT_LEADING);
	}

	if (format_flags & 0x00000004) {
		IDWriteTextFormat_SetParagraphAlignment(format, DWRITE_PARAGRAPH_ALIGNMENT_CENTER);
	} else {
		IDWriteTextFormat_SetParagraphAlignment(format, DWRITE_PARAGRAPH_ALIGNMENT_NEAR);
	}

	if (format_flags & 0x00000020) {
		IDWriteTextFormat_SetWordWrapping(format, DWRITE_WORD_WRAPPING_NO_WRAP);
	} else {
		IDWriteTextFormat_SetWordWrapping(format, DWRITE_WORD_WRAPPING_WRAP);
	}

	if (format_flags & 0x00008000) {
		DWRITE_TRIMMING trimming;
		IDWriteInlineObject* ellipsis = NULL;

		trimming.granularity = DWRITE_TRIMMING_GRANULARITY_CHARACTER;
		trimming.delimiter = 0;
		trimming.delimiterCount = 0;

		hr = IDWriteFactory_CreateEllipsisTrimmingSign(renderer->dwrite, format, &ellipsis);
		if (SUCCEEDED(hr)) {
			IDWriteTextFormat_SetTrimming(format, &trimming, ellipsis);
			IDWriteInlineObject_Release(ellipsis);
		}
	}

	*out_format = format;
	return S_OK;
}

WinUID2DRenderer* winui_d2d_renderer_create(char* err, size_t err_len) {
	WinUID2DRenderer* renderer;
	HRESULT hr;

	winui_clear_error(err, err_len);

	renderer = (WinUID2DRenderer*)calloc(1, sizeof(WinUID2DRenderer));
	if (renderer == NULL) {
		winui_set_error(err, err_len, "calloc", E_OUTOFMEMORY);
		return NULL;
	}

	hr = CoInitializeEx(NULL, COINIT_APARTMENTTHREADED);
	if (SUCCEEDED(hr)) {
		renderer->com_initialized = 1;
	} else if (hr != RPC_E_CHANGED_MODE) {
		winui_set_error(err, err_len, "CoInitializeEx", hr);
		free(renderer);
		return NULL;
	}

	hr = D2D1CreateFactory(D2D1_FACTORY_TYPE_SINGLE_THREADED, &WINUI_IID_ID2D1Factory, NULL, (void**)&renderer->factory);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "D2D1CreateFactory", hr);
		winui_d2d_renderer_destroy(renderer);
		return NULL;
	}

	hr = DWriteCreateFactory(DWRITE_FACTORY_TYPE_SHARED, &WINUI_IID_IDWriteFactory, (IUnknown**)&renderer->dwrite);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "DWriteCreateFactory", hr);
		winui_d2d_renderer_destroy(renderer);
		return NULL;
	}

	hr = CoCreateInstance(&WINUI_CLSID_WICImagingFactory, NULL, CLSCTX_INPROC_SERVER, &WINUI_IID_IWICImagingFactory, (void**)&renderer->wic);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CoCreateInstance(WIC)", hr);
		winui_d2d_renderer_destroy(renderer);
		return NULL;
	}

	return renderer;
}

void winui_d2d_renderer_destroy(WinUID2DRenderer* renderer) {
	if (renderer == NULL) {
		return;
	}
	if (renderer->brush != NULL) {
		ID2D1SolidColorBrush_Release(renderer->brush);
		renderer->brush = NULL;
	}
	if (renderer->target != NULL) {
		ID2D1DCRenderTarget_Release(renderer->target);
		renderer->target = NULL;
	}
	if (renderer->wic != NULL) {
		IWICImagingFactory_Release(renderer->wic);
		renderer->wic = NULL;
	}
	if (renderer->dwrite != NULL) {
		IDWriteFactory_Release(renderer->dwrite);
		renderer->dwrite = NULL;
	}
	if (renderer->factory != NULL) {
		ID2D1Factory_Release(renderer->factory);
		renderer->factory = NULL;
	}
	if (renderer->com_initialized) {
		CoUninitialize();
		renderer->com_initialized = 0;
	}
	free(renderer);
}

int winui_d2d_renderer_begin(WinUID2DRenderer* renderer, uintptr_t hdc, int32_t left, int32_t top, int32_t right, int32_t bottom, char* err, size_t err_len) {
	HRESULT hr;
	RECT bounds;
	D2D1_MATRIX_3X2_F identity;

	if (renderer == NULL || hdc == 0) {
		winui_set_error(err, err_len, "Direct2D begin", E_HANDLE);
		return 0;
	}

	winui_clear_error(err, err_len);

	hr = winui_ensure_target(renderer);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateDCRenderTarget", hr);
		return 0;
	}

	bounds.left = left;
	bounds.top = top;
	bounds.right = right;
	bounds.bottom = bottom;

	hr = ID2D1DCRenderTarget_BindDC(renderer->target, (HDC)hdc, &bounds);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "BindDC", hr);
		return 0;
	}

	identity._11 = 1.0f;
	identity._12 = 0.0f;
	identity._21 = 0.0f;
	identity._22 = 1.0f;
	identity._31 = 0.0f;
	identity._32 = 0.0f;

	ID2D1DCRenderTarget_BeginDraw(renderer->target);
	ID2D1DCRenderTarget_SetTransform(renderer->target, &identity);
	ID2D1DCRenderTarget_SetAntialiasMode(renderer->target, D2D1_ANTIALIAS_MODE_PER_PRIMITIVE);
	ID2D1DCRenderTarget_SetTextAntialiasMode(renderer->target, D2D1_TEXT_ANTIALIAS_MODE_CLEARTYPE);
	return 1;
}

int winui_d2d_renderer_end(WinUID2DRenderer* renderer, char* err, size_t err_len) {
	HRESULT hr;

	if (renderer == NULL || renderer->target == NULL) {
		winui_set_error(err, err_len, "Direct2D end", E_HANDLE);
		return 0;
	}

	winui_clear_error(err, err_len);
	hr = ID2D1DCRenderTarget_EndDraw(renderer->target, NULL, NULL);
	if (FAILED(hr)) {
		if (hr == D2DERR_RECREATE_TARGET) {
			if (renderer->brush != NULL) {
				ID2D1SolidColorBrush_Release(renderer->brush);
				renderer->brush = NULL;
			}
			ID2D1DCRenderTarget_Release(renderer->target);
			renderer->target = NULL;
		}
		winui_set_error(err, err_len, "EndDraw", hr);
		return 0;
	}
	return 1;
}

int winui_d2d_fill_rect(WinUID2DRenderer* renderer, int32_t left, int32_t top, int32_t right, int32_t bottom, uint32_t color, char* err, size_t err_len) {
	HRESULT hr;
	D2D1_RECT_F rect;

	winui_clear_error(err, err_len);
	hr = winui_ensure_brush(renderer, color);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateSolidColorBrush", hr);
		return 0;
	}

	rect = winui_rect(left, top, right, bottom);
	ID2D1DCRenderTarget_FillRectangle(renderer->target, &rect, (ID2D1Brush*)renderer->brush);
	return 1;
}

int winui_d2d_fill_round_rect(WinUID2DRenderer* renderer, int32_t left, int32_t top, int32_t right, int32_t bottom, float radius, uint32_t color, char* err, size_t err_len) {
	HRESULT hr;
	D2D1_ROUNDED_RECT rect;

	winui_clear_error(err, err_len);
	hr = winui_ensure_brush(renderer, color);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateSolidColorBrush", hr);
		return 0;
	}

	rect = winui_round_rect(left, top, right, bottom, radius);
	ID2D1DCRenderTarget_FillRoundedRectangle(renderer->target, &rect, (ID2D1Brush*)renderer->brush);
	return 1;
}

int winui_d2d_stroke_round_rect(WinUID2DRenderer* renderer, int32_t left, int32_t top, int32_t right, int32_t bottom, float radius, uint32_t color, float width, char* err, size_t err_len) {
	HRESULT hr;
	D2D1_ROUNDED_RECT rect;

	winui_clear_error(err, err_len);
	hr = winui_ensure_brush(renderer, color);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateSolidColorBrush", hr);
		return 0;
	}

	rect = winui_round_rect(left, top, right, bottom, radius);
	ID2D1DCRenderTarget_DrawRoundedRectangle(renderer->target, &rect, (ID2D1Brush*)renderer->brush, width, NULL);
	return 1;
}

int winui_d2d_fill_polygon(WinUID2DRenderer* renderer, const WinUIDPoint* points, int32_t count, uint32_t color, char* err, size_t err_len) {
	HRESULT hr;
	ID2D1PathGeometry* geometry = NULL;
	ID2D1GeometrySink* sink = NULL;
	ID2D1SimplifiedGeometrySink* simple_sink = NULL;
	D2D1_POINT_2F* converted = NULL;
	int32_t i;

	if (renderer == NULL || points == NULL || count < 3) {
		return 1;
	}

	winui_clear_error(err, err_len);
	hr = winui_ensure_brush(renderer, color);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateSolidColorBrush", hr);
		return 0;
	}

	hr = ID2D1Factory_CreatePathGeometry(renderer->factory, &geometry);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreatePathGeometry", hr);
		return 0;
	}

	hr = ID2D1PathGeometry_Open(geometry, &sink);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "PathGeometry::Open", hr);
		ID2D1PathGeometry_Release(geometry);
		return 0;
	}

	converted = (D2D1_POINT_2F*)calloc((size_t)count, sizeof(D2D1_POINT_2F));
	if (converted == NULL) {
		winui_set_error(err, err_len, "calloc", E_OUTOFMEMORY);
		ID2D1GeometrySink_Release(sink);
		ID2D1PathGeometry_Release(geometry);
		return 0;
	}

	for (i = 0; i < count; i++) {
		converted[i].x = (FLOAT)points[i].x;
		converted[i].y = (FLOAT)points[i].y;
	}

	simple_sink = (ID2D1SimplifiedGeometrySink*)sink;
	ID2D1SimplifiedGeometrySink_SetFillMode(simple_sink, D2D1_FILL_MODE_WINDING);
	ID2D1SimplifiedGeometrySink_BeginFigure(simple_sink, converted[0], D2D1_FIGURE_BEGIN_FILLED);
	ID2D1SimplifiedGeometrySink_AddLines(simple_sink, &converted[1], (UINT32)(count - 1));
	ID2D1SimplifiedGeometrySink_EndFigure(simple_sink, D2D1_FIGURE_END_CLOSED);
	hr = ID2D1SimplifiedGeometrySink_Close(simple_sink);
	ID2D1GeometrySink_Release(sink);
	free(converted);

	if (FAILED(hr)) {
		winui_set_error(err, err_len, "GeometrySink::Close", hr);
		ID2D1PathGeometry_Release(geometry);
		return 0;
	}

	ID2D1DCRenderTarget_FillGeometry(renderer->target, (ID2D1Geometry*)geometry, (ID2D1Brush*)renderer->brush, NULL);
	ID2D1PathGeometry_Release(geometry);
	return 1;
}

int winui_d2d_draw_text(WinUID2DRenderer* renderer, const uint16_t* text, const uint16_t* font_family, float font_size, int32_t font_weight, uint32_t color, uint32_t format, uint32_t options, int32_t left, int32_t top, int32_t right, int32_t bottom, char* err, size_t err_len) {
	HRESULT hr;
	IDWriteTextFormat* text_format = NULL;
	IDWriteTextLayout* layout = NULL;
	D2D1_POINT_2F origin;
	size_t text_len;

	if (renderer == NULL || text == NULL || text[0] == 0) {
		return 1;
	}

	winui_clear_error(err, err_len);
	hr = winui_ensure_brush(renderer, color);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateSolidColorBrush", hr);
		return 0;
	}

	hr = winui_create_text_format(renderer, font_family, font_size, font_weight, format, &text_format);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateTextFormat", hr);
		return 0;
	}

	text_len = wcslen((const wchar_t*)text);
	hr = IDWriteFactory_CreateTextLayout(
		renderer->dwrite,
		(const wchar_t*)text,
		(UINT32)text_len,
		text_format,
		(FLOAT)(right - left),
		(FLOAT)(bottom - top),
		&layout
	);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateTextLayout", hr);
		IDWriteTextFormat_Release(text_format);
		return 0;
	}

	origin.x = (FLOAT)left;
	origin.y = (FLOAT)top;
	ID2D1DCRenderTarget_DrawTextLayout(renderer->target, origin, layout, (ID2D1Brush*)renderer->brush, (D2D1_DRAW_TEXT_OPTIONS)options);

	IDWriteTextLayout_Release(layout);
	IDWriteTextFormat_Release(text_format);
	return 1;
}

int winui_d2d_measure_text(WinUID2DRenderer* renderer, const uint16_t* text, const uint16_t* font_family, float font_size, int32_t font_weight, int32_t* width, int32_t* height, char* err, size_t err_len) {
	HRESULT hr;
	IDWriteTextFormat* text_format = NULL;
	IDWriteTextLayout* layout = NULL;
	DWRITE_TEXT_METRICS metrics;
	size_t text_len;

	if (width != NULL) {
		*width = 0;
	}
	if (height != NULL) {
		*height = 0;
	}
	if (renderer == NULL || text == NULL || text[0] == 0) {
		return 1;
	}

	winui_clear_error(err, err_len);
	hr = winui_create_text_format(renderer, font_family, font_size, font_weight, 0x00000020u, &text_format);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateTextFormat", hr);
		return 0;
	}

	text_len = wcslen((const wchar_t*)text);
	hr = IDWriteFactory_CreateTextLayout(
		renderer->dwrite,
		(const wchar_t*)text,
		(UINT32)text_len,
		text_format,
		4096.0f,
		4096.0f,
		&layout
	);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateTextLayout", hr);
		IDWriteTextFormat_Release(text_format);
		return 0;
	}

	hr = IDWriteTextLayout_GetMetrics(layout, &metrics);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "GetMetrics", hr);
		IDWriteTextLayout_Release(layout);
		IDWriteTextFormat_Release(text_format);
		return 0;
	}

	if (width != NULL) {
		*width = (int32_t)(metrics.widthIncludingTrailingWhitespace + 0.999f);
	}
	if (height != NULL) {
		*height = (int32_t)(metrics.height + 0.999f);
	}

	IDWriteTextLayout_Release(layout);
	IDWriteTextFormat_Release(text_format);
	return 1;
}

int winui_d2d_draw_icon(WinUID2DRenderer* renderer, uintptr_t hicon, int32_t left, int32_t top, int32_t right, int32_t bottom, char* err, size_t err_len) {
	HRESULT hr;
	IWICBitmap* wic_bitmap = NULL;
	ID2D1Bitmap* bitmap = NULL;
	D2D1_RECT_F dest;

	if (renderer == NULL || hicon == 0) {
		return 1;
	}

	winui_clear_error(err, err_len);
	hr = IWICImagingFactory_CreateBitmapFromHICON(renderer->wic, (HICON)hicon, &wic_bitmap);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateBitmapFromHICON", hr);
		return 0;
	}

	hr = ID2D1DCRenderTarget_CreateBitmapFromWicBitmap(renderer->target, (IWICBitmapSource*)wic_bitmap, NULL, &bitmap);
	IWICBitmap_Release(wic_bitmap);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateBitmapFromWicBitmap(icon)", hr);
		return 0;
	}

	dest = winui_rect(left, top, right, bottom);
	ID2D1DCRenderTarget_DrawBitmap(renderer->target, bitmap, &dest, 1.0f, D2D1_BITMAP_INTERPOLATION_MODE_LINEAR, NULL);
	ID2D1Bitmap_Release(bitmap);
	return 1;
}

int winui_d2d_draw_bitmap(WinUID2DRenderer* renderer, uintptr_t hbitmap, int32_t left, int32_t top, int32_t right, int32_t bottom, uint8_t alpha, char* err, size_t err_len) {
	HRESULT hr;
	IWICBitmap* wic_bitmap = NULL;
	ID2D1Bitmap* bitmap = NULL;
	D2D1_RECT_F dest;

	if (renderer == NULL || hbitmap == 0) {
		return 1;
	}

	winui_clear_error(err, err_len);
	hr = IWICImagingFactory_CreateBitmapFromHBITMAP(renderer->wic, (HBITMAP)hbitmap, NULL, WICBitmapUsePremultipliedAlpha, &wic_bitmap);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateBitmapFromHBITMAP", hr);
		return 0;
	}

	hr = ID2D1DCRenderTarget_CreateBitmapFromWicBitmap(renderer->target, (IWICBitmapSource*)wic_bitmap, NULL, &bitmap);
	IWICBitmap_Release(wic_bitmap);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "CreateBitmapFromWicBitmap(bitmap)", hr);
		return 0;
	}

	dest = winui_rect(left, top, right, bottom);
	ID2D1DCRenderTarget_DrawBitmap(renderer->target, bitmap, &dest, (FLOAT)alpha / 255.0f, D2D1_BITMAP_INTERPOLATION_MODE_LINEAR, NULL);
	ID2D1Bitmap_Release(bitmap);
	return 1;
}

int winui_d2d_flush(WinUID2DRenderer* renderer, char* err, size_t err_len) {
	HRESULT hr;

	if (renderer == NULL || renderer->target == NULL) {
		return 1;
	}

	winui_clear_error(err, err_len);
	hr = ID2D1DCRenderTarget_Flush(renderer->target, NULL, NULL);
	if (FAILED(hr)) {
		winui_set_error(err, err_len, "Flush", hr);
		return 0;
	}
	return 1;
}
