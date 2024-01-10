package application

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"unsafe"

	"github.com/stringflow/gambatte-speedrun/assets"
	"github.com/veandco/go-sdl2/sdl"
)

const Charmap = " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

var (
	Font = NewSpritesheet(Palettize(ParsePNG(assets.FontPng), []color.Color{color.Gray{0x00}, color.Gray{0xa0}, color.Transparent}), 7, 9)
)

func ParsePNG(pngData []byte) image.Image {
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		panic(err)
	}

	return img
}

func Palettize(img image.Image, palette []color.Color) *image.Paletted {
	paletted := image.NewPaletted(img.Bounds(), palette)

	for y := 0; y < img.Bounds().Max.Y; y++ {
		for x := 0; x < img.Bounds().Max.X; x++ {
			color := img.At(x, y)
			red, _, _, _ := color.RGBA()
			paletteIndex := byte(red) / byte(256/len(palette))
			paletted.Pix[x+y*paletted.Stride] = paletteIndex
		}
	}

	return paletted
}

type Backbuffer struct {
	SdlTexture *sdl.Texture
	Image      *image.NRGBA
}

func NewBackbuffer(renderer *Renderer, width int, height int) *Backbuffer {
	sdlTexture, err := renderer.SdlRenderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(width), int32(height))
	if err != nil {
		panic(err)
	}

	return &Backbuffer{
		SdlTexture: sdlTexture,
		Image:      image.NewNRGBA(image.Rect(0, 0, width, height)),
	}
}

func (backbuffer *Backbuffer) Close() {
	backbuffer.SdlTexture.Destroy()
}

func (Backbuffer *Backbuffer) UpdateVRAM() {
	Backbuffer.SdlTexture.Update(nil, unsafe.Pointer(&Backbuffer.Image.Pix[0]), Backbuffer.Image.Stride)
}

func (backbuffer *Backbuffer) Clear(color color.Color) {
	backbuffer.FillRect(0, 0, backbuffer.Image.Bounds().Max.X, backbuffer.Image.Bounds().Max.Y, color)
}

func (backbuffer *Backbuffer) FillRect(x int, y int, width int, height int, color color.Color) {
	draw.Over.Draw(backbuffer.Image, image.Rect(x, y, x+width, y+height), &image.Uniform{color}, image.Point{X: 0, Y: 0})
}

func (backbuffer *Backbuffer) CopyPixelsRGBANative(pixels []byte, destX int, destY int, width int, height int) {
	// TODO(stringflow): I'm not sure if the standard library can provide a more efficient way to do this.
	srcRow := 0
	destRow := destX*4 + destY*backbuffer.Image.Stride

	for y := 0; y < height; y++ {
		src := srcRow
		dest := destRow

		for x := 0; x < width; x++ {
			backbuffer.Image.Pix[dest+0] = pixels[src+2]
			backbuffer.Image.Pix[dest+1] = pixels[src+1]
			backbuffer.Image.Pix[dest+2] = pixels[src+0]
			backbuffer.Image.Pix[dest+3] = pixels[src+3]

			src += 4
			dest += 4
		}

		srcRow += width * 4
		destRow += backbuffer.Image.Stride
	}
}

func (backbuffer *Backbuffer) CopyEntireSprite(spritesheet *Spritesheet, spriteIndex int, destX int, destY int) {
	backbuffer.CopySpriteSection(spritesheet, spriteIndex, 0, 0, destX, destY, spritesheet.SpriteWidth, spritesheet.SpriteHeight)
}

func (backbuffer *Backbuffer) CopySpriteSection(spritesheet *Spritesheet, spriteIndex int, spriteOffsetX int, spriteOffsetY int, destX int, destY int, width int, height int) {
	destRect := image.Rect(destX, destY, destX+width, destY+height)

	sourceX := (spriteIndex % spritesheet.SpritesPerRow) * spritesheet.SpriteWidth
	sourceY := (spriteIndex / spritesheet.SpritesPerRow) * spritesheet.SpriteHeight
	sourcePoint := image.Point{sourceX + spriteOffsetX, sourceY + spriteOffsetY}

	draw.Over.Draw(backbuffer.Image, destRect, spritesheet.Image, sourcePoint)
}

func (backbuffer *Backbuffer) DrawTextCentered(text string, areaX int, areaY int, areaWidth int, areaHeight int) {
	textWidth := len(text) * Font.SpriteWidth
	textHeight := Font.SpriteHeight

	x := areaX + (areaWidth-textWidth)/2
	y := areaY + (areaHeight-textHeight)/2

	backbuffer.DrawText(text, x, y)
}

func (backbuffer *Backbuffer) DrawText(text string, x int, y int) {
	for i, character := range text {
		characterIndex := strings.IndexRune(Charmap, character)
		if characterIndex == -1 {
			continue
		}

		backbuffer.CopyEntireSprite(Font, characterIndex, x+i*Font.SpriteWidth, y)
	}
}

func (backbuffer *Backbuffer) ScaleMode() sdl.ScaleMode {
	scaleMode, err := backbuffer.SdlTexture.GetScaleMode()
	if err != nil {
		return sdl.ScaleModeNearest
	}

	return scaleMode
}

func (backbuffer *Backbuffer) SetScaleMode(scaleMode sdl.ScaleMode) {
	backbuffer.SdlTexture.SetScaleMode(scaleMode)
}

type Spritesheet struct {
	Image         image.Image
	SpriteWidth   int
	SpriteHeight  int
	SpritesPerRow int
}

func NewSpritesheet(img image.Image, spriteWidth int, spriteHeight int) *Spritesheet {
	return &Spritesheet{
		Image:         img,
		SpriteWidth:   spriteWidth,
		SpriteHeight:  spriteHeight,
		SpritesPerRow: img.Bounds().Max.X / spriteWidth,
	}
}

type NativeRGBA struct {
	Pixels []byte
	Rect   image.Rectangle
	Pitch  int
}
