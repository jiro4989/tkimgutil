package command

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"runtime"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/codegangsta/cli"
	jimage "github.com/jiro4989/lib-go/image"
	jlog "github.com/jiro4989/tkimgutil/internal/log"
)

type TOML struct {
	Image Image `toml:"image"`
}

type Image struct {
	SaveFilenameFormat string     `toml:"save_filename_format"`
	Pattern            [][]string `toml:"pattern"`
}

func CmdGenerate(c *cli.Context) {
	// 出力先ディレクトリの作成
	outDir := c.String("d")
	cfgPath := c.String("config")
	generate(outDir, cfgPath)
}

func generate(outDir, cfgPath string) {
	err := os.MkdirAll(outDir, os.ModePerm)
	jlog.FatalError(err)

	var cfg TOML
	_, err = toml.DecodeFile(cfgPath, &cfg)
	jlog.FatalError(err)

	// 画像の幅が必要なので先行して1枚だけload
	src, err := jimage.ReadFile(cfg.Image.Pattern[0][0])
	jlog.FatalError(err)

	var (
		width  = src.Bounds().Size().X                       // 生成する画像の横幅
		height = src.Bounds().Size().Y                       // 生成する画像の縦幅
		onFmt  = outDir + "/" + cfg.Image.SaveFilenameFormat // 生成する画像ファイルのパス
	)

	ch := make(chan indexedPattern, len(cfg.Image.Pattern))
	var wg sync.WaitGroup

	// CPU数だけワーカースレッドの起動
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, ch chan indexedPattern) {
			defer wg.Done()
			for {
				ip, ok := <-ch
				if !ok {
					return
				}
				i, p := ip.index, ip.pattern

				// 画像ファイルの組み合わせ分ひたすら重ねて書き込む
				bi := newBufImage(width, height)
				bi.drawImageFromFile(p)

				// 画像ファイル出力
				on := fmt.Sprintf(onFmt, (i + 1))
				err = jimage.WriteFile(on, bi.image)
				jlog.FatalError(err)

				// パイプで処理を続けるために標準出力
				fmt.Println(on)
			}
		}(&wg, ch)
	}

	// 組み合わせをファイル名インデックスと一緒にチャネルに送信
	for i, p := range cfg.Image.Pattern {
		v := indexedPattern{
			index:   i,
			pattern: p,
		}

		ch <- v
	}
	close(ch)
	wg.Wait()
}

// indexedPattern 画像組み合わせに番号をもたせただけの型。チャネル用
type indexedPattern struct {
	index   int
	pattern []string
}

// bufImage は自身のデータを加工する画像型です。
type bufImage struct {
	image *image.RGBA
}

// newBufImage はbufImageの初期値をセット済みのものを返します。
func newBufImage(w, h int) bufImage {
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	return bufImage{image: rgba}
}

// drawImageFromFile // は引数で指定したファイルパスの画像を重ねて描画します。
func (bi *bufImage) drawImageFromFile(p []string) {
	for _, f := range p {
		img, err := jimage.ReadFile(f)
		jlog.FatalError(err)
		bi.drawImage(img)
	}
}

// drawImage は指定のイメージを書き込みます。
func (bi *bufImage) drawImage(img image.Image) {
	size := img.Bounds().Size()
	w, h := size.X, size.Y

	// 画像を貼り付け
	rect := image.Rectangle{
		image.ZP,
		image.Pt(w, h),
	}
	draw.Draw(bi.image, rect, img, image.ZP, draw.Over)
}
