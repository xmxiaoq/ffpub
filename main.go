package main

import (
	"github.com/spf13/afero"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"os"
	"github.com/ungerik/go-dry"
	_"time"
	"strings"
	"encoding/json"
	"github.com/urfave/cli"
	"os/exec"
	"time"
	"fmt"
	"regexp"
	"sort"
)

type Prop struct {
	EgretVersion string `json:"egret_version"`
	Modules []struct {
		Name string `json:"name"`
	} `json:"modules"`
	Native struct {
		AndroidAsPath string   `json:"android_as_path"`
		PathIgnore    []string `json:"path_ignore"`
	} `json:"native"`
	Publish struct {
		Native int64  `json:"native"`
		Path   string `json:"path"`
		Web    int64  `json:"web"`
	} `json:"publish"`
}

type VersionInfo struct {
	CodeURL   string `json:"code_url"`
	UpdateURL string `json:"update_url"`
	Major     int64 `json:"major"`
	Minor     int64 `json:"minor"`
	Patch     int64 `json:"patch"`
	TS        string `json:"ts"`
}

type GroupDef struct {
	Keys string `json:"keys"`
	Name string `json:"name"`
}
type ResourceDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

type ResConfig struct {
	Groups    []GroupDef `json:"groups"`
	Resources []ResourceDef`json:"resources"`
}

var FS = afero.Afero{Fs: afero.NewOsFs()}
var sugar *zap.SugaredLogger

var (
	pubType string
	ip      string
	debug   bool
)

//var h5Path = `..\..\mj_h5\`
var h5Path = `D:\fanfan\mj_h5\`
var propPath = h5Path + "egretProperties.json"
var javaCodePath = h5Path + `..\mj_android\proj.android\app\src\main\java\org\egret\java\mj_android\mj_android.java`
var verFilePath = h5Path + `bin-release\native/version.json`

func init() {
	_, _ = FS.Exists(os.Args[0] + ".log")
	_ = cast.ToString(100)
	_ = dry.SyncMap{}
}

func copyRes() (err error) {
	dirs, err := dry.ListDirDirectories(h5Path + `bin-release\native\`)
	if err != nil {
		sugar.Error(err)
		return
	}

	sort.Strings(dirs)
	vs := dirs[len(dirs)-1]
	src := h5Path + `bin-release\web\` + vs + `\resource\Channel`
	dst := h5Path + `bin-release\native\` + vs + `\resource\Channel`
	err = os.RemoveAll(dst)
	if err != nil {
		sugar.Error(err)
		return
	}

	err = dry.FileCopyDir(src, dst)
	if err != nil {
		sugar.Error(err)
		return
	}

	return nil
}

func getVersionStr() (string, string) {
	ts := time.Now()
	verStr := fmt.Sprintf(`%d%02d%02d%02d%02d%02d`, ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second())
	verStr2 := fmt.Sprintf(`%d-%02d-%02d %02d:%02d:%02d`, ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second())
	return verStr, verStr2
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer logger.Sync() // flushes buffer, if any
	sugar = logger.Sugar()

	app := cli.NewApp()
	app.Name = "ffpub"
	app.Usage = "ff -t app"
	app.Version = "1.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "pubType, t",
			Usage:       "publish app `PUBLISH_TYPE`",
			Value:       "app",
			Destination: &pubType,
		},
		cli.StringFlag{
			Name:        "ip",
			Usage:       "资源服务器ip地址",
			Value:       "http://10.0.0.35",
			Destination: &ip,
		},
		cli.BoolTFlag{
			Name:        "debug, d",
			Usage:       "0:正式版 1:测试版",
			Destination: &debug,
		},
	}

	app.Action = func(c *cli.Context) error {
		sugar.With("pubType", pubType, "ip", ip, "debug", debug).Info("publishing...")

		var orig Prop
		b, _ := dry.FileGetBytes(propPath, 0)
		err := json.Unmarshal(b, &orig)
		if err != nil {
			sugar.Error(err)
		}

		if pubType == "res" {
			orig.Native.PathIgnore = []string{}
			return copyRes()
		} else if pubType == "app" {
			orig.Native.PathIgnore = []string{"Channel"}
		}
		dry.FileSetJSONIndent(propPath, orig, "    ")

		//var resCfg ResConfig
		//b, _ = dry.FileGetBytes(h5Path+`resource/default.res.json`, 0)
		//dry.FileSetBytes(h5Path+`resource\default-backup.res.json`, b)
		//json.Unmarshal(b, &resCfg)
		//
		//var channelCfg ResConfig
		//var commonCfg ResConfig
		//for _, value := range resCfg.Resources {
		//	if strings.HasPrefix(value.URL, "Channel") {
		//		channelCfg.Resources = append(channelCfg.Resources, value)
		//	} else {
		//		commonCfg.Resources = append(commonCfg.Resources, value)
		//	}
		//}
		//for _, value := range resCfg.Groups {
		//	if value.Name== "preload" || value.Name== "GameCenter" {
		//		commonCfg.Groups = append(commonCfg.Groups, value)
		//	} else {
		//		channelCfg.Groups = append(channelCfg.Groups, value)
		//	}
		//}

		//folders :=dry.ListDirDirectories(`D:\fanfan\mj_h5\resource\Channel`)
		//CHANNEL_LIST:=[]string{`GuangDong`, `GuiZhou`}
		//for _, value := range CHANNEL_LIST {
		//	idx :=strings.LastIndex(value, "_")
		//	if idx != -1 {
		//		cn :=value[idx:]
		//
		//	}
		//}
		//dry.FileSetJSONIndent(h5Path+`bin-release\native\channel.res.json`, channelCfg, "    ")
		//dry.FileSetJSONIndent(h5Path+`bin-release\native\default.res.json`, commonCfg, "    ")
		//
		//dry.FileSetJSONIndent(h5Path+`resource\channel.res.json`, channelCfg, "    ")
		//dry.FileSetJSONIndent(h5Path+`resource\default.res.json`, commonCfg, "    ")

		verStr, verTS := getVersionStr()
		sugar.Info(verStr)

		// 发布html5
		cmd := exec.Command("egret", "publish", "-compile", "--runtime", "html5", "--version", verStr)
		cmd.Dir = h5Path
		out, err := cmd.Output()
		if err != nil {
			sugar.Fatal(err)
		}
		sugar.Infof("publish html5 %s\n", out)

		// 发布native
		cmd = exec.Command("egret", "publish", "-compile", "--runtime", "native", "--version", verStr)
		cmd.Dir = h5Path
		out, err = cmd.Output()
		if err != nil {
			sugar.Fatal(err)
		}
		sugar.Infof("publish native %s\n", out)

		// 更新version.json
		var verInfo VersionInfo
		b, _ = dry.FileGetBytes(verFilePath, 0)
		json.Unmarshal(b, &verInfo)
		if verInfo.Major <= 0 {
			verInfo.Major = 1
		}
		verInfo.Patch += 1
		verInfo.CodeURL = fmt.Sprintf("%s/%s/game_code_%s.zip", ip, verStr, verStr)
		verInfo.UpdateURL = fmt.Sprintf("%s/%s", ip, verStr)
		verInfo.TS = verTS
		dry.FileSetJSON(verFilePath, verInfo)

		// 更新java代码
		code, _ := dry.FileGetString(javaCodePath)
		newCode := strings.Replace(code, "setLoaderUrl(0);", "setLoaderUrl(1);", -1)

		// 更新java代码
		code = newCode
		reg := regexp.MustCompile(`loaderUrl = ".+?"`)
		repStr := fmt.Sprintf(`loaderUrl = "%s/version.json"`, ip)
		newCode = reg.ReplaceAllString(code, repStr)

		// 更新java代码
		code = newCode
		reg = regexp.MustCompile(`updateUrl = ".+?"`)
		repStr = fmt.Sprintf(`updateUrl = "%s/%s/"`, ip, verStr)
		newCode = reg.ReplaceAllString(code, repStr)

		dry.FileSetString(javaCodePath, newCode)

		return nil
	}

	app.Run(os.Args)
}
