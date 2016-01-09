package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

/**
 * 変換可能な拡張子(かも)
 */
var models = []string{
	".collada",
	".3ds",
	".obj",
	".lwo",
	".fbx",
	".blend",
	".x",
	".stl",
	".ply",
	".ms3d",
	".b3d",
	".md3",
	".mdl",
	".dxf",
	".ifc",
	".dae",
	".mtl",
}

const TEMPLATE_DIR = "template/"

var PRODUCTION = "0"
var URL = "localhost"

const PORT = "8080"

/**************************************************************************************************/
/*!
 *  テンプレート呼び出し
 */
/**************************************************************************************************/
func display(w http.ResponseWriter, tmpl string, data interface{}) {
	html := tmpl + ".html"
	var templates = template.Must(template.ParseFiles(TEMPLATE_DIR + html))
	templates.ExecuteTemplate(w, tmpl+".html", data)
}

/**************************************************************************************************/
/*!
 *  データアップロード
 */
/**************************************************************************************************/
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//GET displays the upload form.
	case "GET":
		var data = map[string]interface{}{}
		data["models"] = strings.Join(models, " ")
		display(w, "upload", data)

	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		isModelFind := false
		//get the multipart reader for the request.
		reader, err := r.MultipartReader()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// アップロードするフォルダを設定
		folder := strconv.FormatInt(time.Now().UnixNano(), 10)

		// ディレクトリ取得
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		// imagesがなければ作成
		upd := path.Join(dir, "uploads")
		uploadPath := path.Join(upd, folder)
		if err := os.Mkdir(uploadPath, 0755); err != nil && !os.IsExist(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//copy each part to destination.
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			// ファイル名
			fileName := part.FileName()

			// 空で来たらスルー
			if fileName == "" {
				continue
			}

			// ファイル生成
			dst, err := os.Create(uploadPath + "/" + fileName)

			if err != nil {
				dst.Close()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(dst, part); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			dst.Close()

			// モデルデータチェック & JSON化
			if isModelData(fileName) {
				// 重複は許さない
				if isModelFind {
					http.Error(w, "複数のモデルデータは設定できません", http.StatusInternalServerError)
					return
				}

				jsonData, errStr := jsonize(uploadPath, fileName)
				if len(errStr) > 0 {
					http.Error(w, errStr, http.StatusInternalServerError)
					return
				}

				err := ioutil.WriteFile(uploadPath+"/data.json", jsonData, 0644)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				isModelFind = true
			}

		}

		// モデルデータがない場合もエラー
		if !isModelFind {
			http.Error(w, "モデルデータを含めてください", http.StatusInternalServerError)
			return
		}

		// 生成したURLを返す
		fmt.Fprintf(w, "http://"+URL+"/view?d="+folder)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

/**************************************************************************************************/
/*!
 *  サンプルモデルデータ表示
 */
/**************************************************************************************************/
func viewSampleHandler(w http.ResponseWriter, r *http.Request) {
	display(w, "view_sample", nil)
}

type Data struct {
	Id string
}

/**************************************************************************************************/
/*!
 *  モデルデータ表示
 */
/**************************************************************************************************/
func viewHandler(w http.ResponseWriter, r *http.Request) {

	// クエリパラメータをmapにする
	m, _ := url.ParseQuery(r.URL.RawQuery)

	// クエリパラメータを取得
	data := new(Data)
	data.Id = m["d"][0]

	display(w, "view", data)
}

/**************************************************************************************************/
/*!
 *  初期化
 */
/**************************************************************************************************/
func init() {
	// ランダムシード
	rand.Seed(time.Now().UnixNano())

	// 本番か
	PRODUCTION = os.Getenv("PRODUCTION")
	if PRODUCTION == "1" {
		URL = os.Getenv("HOSTNAME")
	} else {
		URL = "localhost:" + PORT
	}

	// uploadsフォルダがなければ作成する
	checkFolder()
}

/**************************************************************************************************/
/*!
 *  エントリポイント
 */
/**************************************************************************************************/
func main() {
	// static file handler.
	static := http.FileServer(http.Dir("static"))
	uploads := http.FileServer(http.Dir("uploads"))
	http.Handle("/static/", http.StripPrefix("/static/", static))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", uploads))

	// handler
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/view", viewHandler)
	http.HandleFunc("/view_sample", viewSampleHandler)

	//Listen on port 8080
	http.ListenAndServe(":"+PORT, nil)
}

/**************************************************************************************************/
/*!
 *  拡張子からモデルデータか判定
 */
/**************************************************************************************************/
func isModelData(fileName string) bool {
	ext := path.Ext(fileName)

	for _, model := range models {
		if model == ext || strings.ToUpper(model) == ext {
			return true
		}
	}
	return false
}

/**************************************************************************************************/
/*!
 *  アップフォルダ確認
 */
/**************************************************************************************************/
func checkFolder() {
	// ディレクトリ取得
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
		return
	}

	// なければ作成
	uploads := path.Join(dir, "uploads")
	if err := os.Mkdir(uploads, 0755); err != nil && !os.IsExist(err) {
		panic(err)
		return
	}
}

/**************************************************************************************************/
/*!
 *  モデルデータをJSON化
 */
/**************************************************************************************************/
func jsonize(folder, fileName string) ([]byte, string) {

	// 実行コマンド実体
	path, err := exec.LookPath("assimp2json")
	if err != nil {
		panic(err)
	}

	cmds := []string{folder + "/" + fileName}
	cmd := exec.Command(path, cmds...)

	var stdErr, stdOut bytes.Buffer
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut

	// exec
	err = cmd.Run()
	if err != nil {
		return []byte{}, stdOut.String()
	}

	return stdOut.Bytes(), ""
}
