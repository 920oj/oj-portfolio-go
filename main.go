package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/template"

	Copy "github.com/otiai10/copy"
	"github.com/russross/blackfriday/v2"
)

// InputData 入力データ
type InputData struct {
	Name       string `json:"name"`
	HiraName   string `json:"hira_name"`
	AlpName    string `json:"alp_name"`
	School     string `json:"school"`
	Department string `json:"department"`
	Grade      string `json:"grade"`
	ProfileURL string `json:"profile_url"`
	LogoURL    string `json:"logo_url"`
	HTML       string
}

func main() {
	// Markdownファイル読み込み
	md, err := ioutil.ReadFile("./contents/index.md")
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
	html := blackfriday.Run(md, blackfriday.WithNoExtensions())

	// 設定ファイルとHTML化したMarkdownを一つの構造体に移す
	data := new(InputData)
	err = json.Unmarshal([]byte(conf), data)
	data.HTML = string(html)

	// 書き出し準備
	buff := new(bytes.Buffer)
	fw := io.Writer(buff)

	// テンプレートファイルを読み込んで変換する
	t := template.Must(template.ParseFiles("./templates/index.html.tql"))
	if err := t.ExecuteTemplate(fw, "index.html.tql", data); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// ファイル書き込み
	err = writeBytes("./dist/index.html", buff.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	err = Copy.Copy("./public", "./dist")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func writeBytes(filename string, lines []byte) error {
	file, err := os.OpenFile(filename, os.O_CREATE, 0666)
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
