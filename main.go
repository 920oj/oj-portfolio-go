package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/russross/blackfriday/v2"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
)

// InputData 入力データ
type InputData struct {
	Name            string `json:"name"`
	HiraName        string `json:"hira_name"`
	AlpName         string `json:"alp_name"`
	School          string `json:"school"`
	Department      string `json:"department"`
	Grade           string `json:"grade"`
	ProfileURL      string `json:"profile_url"`
	LogoURL         string `json:"logo_url"`
	MetaDescription string `json:"meta_description"`
	MetaTitle       string `json:"meta_title"`
	MetaOgpImg      string `json:"meta_ogp_img"`
	MetaURL         string `json:"meta_url"`
	MetaTwitter     string `json:"meta_twitter"`
	GoogleAnalytics string `json:"google_analytics"`
	HTML            string
	CSS             string
}

func main() {
	// distディレクトリを削除
	if err := os.RemoveAll("./dist"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	// Markdownファイル読み込み
	md, err := ioutil.ReadFile("./contents/index.md")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// 埋め込むCSSファイル読み込み
	style, err := ioutil.ReadFile("./templates/style.css")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// 設定ファイル読み込み
	conf, err := ioutil.ReadFile("./conf/conf.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// MarkdownをHTML化
	exts := blackfriday.Autolink // リンクを自動でaタグにする
	htmlData := blackfriday.Run(md, blackfriday.WithExtensions(exts))

	// 設定ファイルとHTML化したMarkdownを一つの構造体に移す
	data := new(InputData)
	err = json.Unmarshal([]byte(conf), data)
	data.HTML = string(htmlData)
	data.CSS = string(style)

	// 書き出し準備
	templateBuff := new(bytes.Buffer)
	templateFw := io.Writer(templateBuff)

	// テンプレートファイルを読み込んで変換する
	t := template.Must(template.ParseFiles("./templates/index.html.tql"))
	if err := t.ExecuteTemplate(templateFw, "index.html.tql", data); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// HTMLをminifyする
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	r := io.Reader(templateBuff)
	outputBuff := new(bytes.Buffer)
	outputFw := io.Writer(outputBuff)
	if err = m.Minify("text/html", outputFw, r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(outputBuff)

	// distフォルダが無ければ作成
	if _, err := os.Stat("./dist"); os.IsNotExist(err) {
		os.Mkdir("./dist", 0777)
	}

	// ファイル書き込み
	err = writeBytes("./dist/index.html", outputBuff.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

// writeBytes 指定した名前と入力内容でファイルを出力する
func writeBytes(filename string, lines []byte) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(lines)
	if err != nil {
		return err
	}
	return nil
}
