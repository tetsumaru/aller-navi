package linebot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// pdfToJPEGs は PDF バイト列を Ghostscript で JPEG 画像のスライスに変換します。
// 実行環境に ghostscript (gs コマンド) がインストールされている必要があります。
func pdfToJPEGs(pdfBytes []byte) ([][]byte, error) {
	// 一時ファイルに PDF を書き込む
	tmpPDF, err := os.CreateTemp("", "linebot-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("一時ファイル作成: %w", err)
	}
	defer os.Remove(tmpPDF.Name())

	if _, err := tmpPDF.Write(pdfBytes); err != nil {
		tmpPDF.Close()
		return nil, fmt.Errorf("PDF 書き込み: %w", err)
	}
	tmpPDF.Close()

	// 出力用の一時ディレクトリを作成する
	tmpDir, err := os.MkdirTemp("", "linebot-pages-*")
	if err != nil {
		return nil, fmt.Errorf("一時ディレクトリ作成: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Ghostscript で PDF を JPEG に変換する（150dpi、品質90）
	outPattern := filepath.Join(tmpDir, "page-%03d.jpg")
	cmd := exec.Command("gs",
		"-dNOPAUSE", "-dBATCH", "-dSAFER",
		"-sDEVICE=jpeg",
		"-r150",
		"-dJPEGQ=90",
		"-sOutputFile="+outPattern,
		tmpPDF.Name(),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ghostscript 変換失敗: %w\n%s", err, out)
	}

	// 出力ファイルを名前順に読み込む
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("出力ファイル一覧取得: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	images := make([][]byte, 0, len(entries))
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("画像読み込み %s: %w", entry.Name(), err)
		}
		images = append(images, data)
	}
	return images, nil
}
