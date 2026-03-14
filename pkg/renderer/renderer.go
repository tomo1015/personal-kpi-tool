// Package renderer は internal/renderer の公開ラッパーです。
package renderer

import (
	"text/template"

	"github.com/tomo1015/personal-kpi-tool/internal/renderer"
	"github.com/tomo1015/personal-kpi-tool/pkg/kpidef"
)

// Options は renderer.Render に渡すオプションです。
type Options struct {
	OutputPath   string
	TemplatePath string
	ChartJS      string
	ExtraFuncs   template.FuncMap
}

// Render は RenderData を HTML ファイルとして出力します。
func Render(rd *kpidef.RenderData, opts Options) error {
	return renderer.Render(rd, renderer.Options{
		OutputPath:   opts.OutputPath,
		TemplatePath: opts.TemplatePath,
		ChartJS:      opts.ChartJS,
		ExtraFuncs:   opts.ExtraFuncs,
	})
}