package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/88250/gulu"
	"github.com/dop251/goja"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	dynamicSkinPath = "assets/res/ui/face/skin/dynamicSkin" // 动态皮肤路径
	skinPath        = "assets/res/ui/face/skin"             // 静态头像路径
	bfacePath       = "assets/res/ui/bface/skin"            // 预览皮肤
	horSkinPath     = "assets/res/ui/bface/skin/horSkin"    // 预览全屏皮肤
	effectPath      = "assets/res/audio/effect"             // 音效路径
	skillPath       = "assets/res/audio/skill"

	// 动皮输出路径  skin/数字key/文件夹路径/文件
	outputSkin = "skins/skin"
	// audio输出路径 audio/数字key/文件夹路径/文件
	outputAudioEffect = "skins/audio/effect"
	outputAudioSkill  = "skins/audio/skill"

	finalDynamicSkin = "final/skin/dynamic"
	finalSkin        = "final/skin/sanguoskin"
	finalPreviewSkin = "final/skin/大图预览"
	finalAudioEffect = "final/audio/effect"
	finalAudioSkill  = "final/audio/sanguoaudio"
)

var (
	excelHeader = []string{"皮肤id", "所属武将", "武将英文", "皮肤名称", "前缀", "是否是真动皮"}
	config      Config
	apkPath     = "" // apk路径
)

type Args struct {
	X     []float64
	Y     []float64
	Scale float64
}

type Config struct {
	ExcelName string
	Beijing   Args
	Daiji     Args
	Gongji    Args
}

type SkinInfo struct {
	Id        string   // 皮肤id
	Name      string   // 皮肤名称
	WuJiangCn string   // 武将中文名称
	WuJiangEn string   // 武将英文名称
	Prefix    []string // 武将前缀
	IsReal    bool     // 是否是真动皮
}

func Init() {
	flag.StringVar(&apkPath, "apkPath", "./sgs", "请输入解压后的三国杀apk路径")
}

// 初始化
func initial() error {
	Init()
	flag.Parse()
	viper.SetConfigFile("./config.yml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到错误；如果需要可以忽略
			fmt.Println(err)
		} else {
			fmt.Println(err)
			// 配置文件被找到，但产生了另外的错误
		}
		return errors.New("配置文件错误")
	}
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Println(err)
		return errors.New("配置文件错误")
	}
	fmt.Println("apkPath: ", apkPath)
	if !gulu.File.IsExist(apkPath) {
		return errors.New("三国杀路径错误")
	}
	return nil
}

func main() {
	err := initial()

	if err != nil {
		fmt.Println(err)
		return
	}

	skinIds := createInitialData()
	// // 创建新的excel, 查看是否有新的数据需要进行填写
	createExcel(skinIds)

	genDynamicSkinParams()

	fmt.Println("end!!!!")
}

// 提取手杀的所有动皮, 创建文件夹, 存放.
func createInitialData() []string {
	// 路径拼接，并得到简化的路径
	dsp := filepath.Join(apkPath, dynamicSkinPath)
	lutouP := filepath.Join(apkPath, skinPath)
	bfaceP := filepath.Join(apkPath, bfacePath)
	horSkinP := filepath.Join(apkPath, horSkinPath)
	audioEffectP := filepath.Join(apkPath, effectPath)
	audioskillP := filepath.Join(apkPath, skillPath)

	// 获取音效的集合, 以皮肤id为键值, 值为对应的音效列表
	audios, _ := ioutil.ReadDir(audioskillP)
	skinAudioMap := make(map[string][]string, 0)
	for _, f := range audios {
		fileName := f.Name()
		key := fileName[6:]
		numKeys := strings.SplitN(key, "_", 2)
		if len(numKeys) == 1 {
			continue
		}
		skinId := numKeys[0]
		skinAudioMap[skinId] = append(skinAudioMap[skinId], fileName)
	}

	files, _ := ioutil.ReadDir(dsp)

	skinIds := make([]string, 0)
	skinIdMap := make(map[string]struct{})

	for _, f := range files {
		fileName := f.Name()
		if !strings.HasSuffix(fileName, ".zip") {
			continue
		}
		key := fileName[0 : len(fileName)-4]
		numKeys := strings.SplitN(key, "_", 2)
		if len(numKeys) == 1 {
			continue
		}
		numKey := numKeys[0]
		// id去重
		if _, ok := skinIdMap[numKey]; !ok {
			skinIds = append(skinIds, numKey)
		}

		// 解压zip文件
		zipFile := path.Join(dsp, fileName)
		dstPath := path.Join(outputSkin, numKey)

		// 如果解压存在, 则不进行解压, 节省时间, 添加配置, 是否覆盖解压
		if !gulu.File.IsExist(path.Join(dstPath, key)) {
			if err := gulu.Zip.Unzip(zipFile, dstPath); err != nil {
				fmt.Println(fmt.Sprintf("%s解压失败", fileName))
				continue
			}
		}

		// 获取三种动皮静皮, 便于区分

		// 露头皮肤
		lutou := path.Join(lutouP, numKey+".png")

		if gulu.File.IsExist(lutou) {
			// 复制到动皮处
			if !gulu.File.IsExist(path.Join(dstPath, numKey+".png")) {
				err := gulu.File.CopyFile(lutou, path.Join(dstPath, numKey+".png"))
				if err != nil {
					fmt.Println(fmt.Sprintf("%s复制露头皮肤失败", numKey))
				}
			}
		}

		// 大屏预览静皮
		for i := 0; i < 3; i++ {
			filename := fmt.Sprintf("%s_%d.jpg", numKey, i)
			bface := path.Join(bfaceP, filename)
			if gulu.File.IsExist(bface) {
				if !gulu.File.IsExist(path.Join(dstPath, key+".jpg")) {
					err := gulu.File.CopyFile(bface, path.Join(dstPath, key+".jpg"))
					if err != nil {
						fmt.Println(fmt.Sprintf("%s复制露头皮肤失败", numKey))
						break
					}
					break
				}
			}
		}

		// 全屏预览静皮
		for i := 0; i < 3; i++ {
			filename := fmt.Sprintf("%s_%d_horizon.jpg", numKey, i)
			horSkin := path.Join(horSkinP, filename)
			if gulu.File.IsExist(horSkin) {
				if !gulu.File.IsExist(path.Join(dstPath, path.Join(dstPath, key+"_horizon.jpg"))) {
					err := gulu.File.CopyFile(horSkin, path.Join(dstPath, key+"_horizon.jpg"))
					if err != nil {
						fmt.Println(fmt.Sprintf("%s复制皮肤失败", filename))
						break
					}
					break
				}

			}
		}

		// 获取动皮特效
		audioDstPath := path.Join(outputAudioEffect, numKey, key)
		audioPath := path.Join(audioEffectP, fmt.Sprintf("effect_dynamicSkin_%s.mp3", numKey))
		if gulu.File.IsExist(audioPath) {
			if !gulu.File.IsExist(path.Join(dstPath, path.Join(audioDstPath, "XingXiang.mp3"))) {
				err := gulu.File.CopyFile(audioPath, path.Join(audioDstPath, "XingXiang.mp3"))
				if err != nil {
					fmt.Println(fmt.Sprintf("%s复制音效文件失败", numKey))
				}
			}
		}
		skillAudioDstPath := path.Join(outputAudioSkill, numKey, key)
		// 获取皮肤技能音效
		if sAudios, ok := skinAudioMap[numKey]; ok {
			for _, audio := range sAudios {
				skillAudioPath := path.Join(audioskillP, audio)
				if !gulu.File.IsExist(path.Join(dstPath, path.Join(skillAudioDstPath, audio))) {
					err := gulu.File.CopyFile(skillAudioPath, path.Join(skillAudioDstPath, audio))
					if err != nil {
						fmt.Println(fmt.Sprintf("%s复制技能音效文件失败", numKey))
					}
				}
			}
		}
	}
	return skinIds
}

// 融合原来的dynamicSkin参数, 生成新的文件夹路径与千幻聆音需要的技能格式
func genDynamicSkinParams() {
	dynamicSkin, extend := parseSkinJs()
	// 读取excel获取参数并生成新的文件夹路径等.
	excelName := config.ExcelName
	toReadFileName := excelName + ".xlsx"

	skinInfoMap := make(map[string]*SkinInfo)
	if gulu.File.IsExist(toReadFileName) {
		output, _ := xlsx.FileToSlice(toReadFileName)
		// 读取第一个sheet的数据
	GetSkinInfo:
		for rowIndex, row := range output[0] {
			if rowIndex == 0 {
				continue
			}
			if len(row) == 0 {
				continue
			}
			info := new(SkinInfo)
			if len(row) < 3 {
				continue
			}
			for i, v := range row {
				if i <= 3 && v == "" {
					goto GetSkinInfo
				}
				switch i {
				case 0:
					info.Id = v
				case 1:
					info.WuJiangCn = v
				case 2:
					info.WuJiangEn = v
				case 3:
					info.Name = v
				case 4:
					info.Prefix = strings.Split(v, ",")
				case 5:
					info.IsReal = v == "是"
				}
			}

			skinInfoMap[info.Id] = info
		}
	}

	// 遍历动皮生成路径, 组成新的组织形式
	files, _ := ioutil.ReadDir(outputSkin)

	// 生成需要产出的map
	toOutputMap := make(map[string]map[string]interface{})
	toOutputExtendMap := make(map[string]string)

	for idx, f := range files {
		numKey := f.Name()
		key := ""
		if info, ok := skinInfoMap[numKey]; ok {
			curDir := path.Join(outputSkin, numKey)
			subFiles, _ := ioutil.ReadDir(curDir)
			dsPath := ""
			for _, innerF := range subFiles {
				p := path.Join(curDir, innerF.Name())
				if gulu.File.IsDir(p) {
					// 复制对应皮肤文件
					dsPath = p
					key = innerF.Name()
					break
				}
			}
			// 复制相关信息到新的路径中
			if dsPath == "" {
				continue
			}
			fmt.Println(fmt.Sprintf("<%d> [%s]: 动皮: %s", idx, info.WuJiangCn, info.Name))
			if !gulu.File.IsExist(path.Join(finalDynamicSkin, info.WuJiangCn, info.Name)) {
				_ = gulu.File.CopyDir(dsPath, path.Join(finalDynamicSkin, info.WuJiangCn, info.Name))

			}

			copyFile := func(src, dst string) {
				if gulu.File.IsExist(src) && !gulu.File.IsExist(dst) {
					_ = gulu.File.CopyFile(src, dst)
				}
			}

			// 复制静皮
			copyFile(path.Join(curDir, key+".jpg"), path.Join(finalSkin, info.WuJiangEn, info.Name+".jpg"))
			copyFile(path.Join(curDir, numKey+".png"), path.Join(finalPreviewSkin, info.WuJiangCn, info.Name+".png"))
			copyFile(path.Join(curDir, key+"_horizon.jpg"), path.Join(finalPreviewSkin, info.WuJiangCn, info.Name+"_horizon.jpg"))

			// 音效
			if !gulu.File.IsExist(path.Join(finalAudioEffect, info.WuJiangCn, info.Name)) {
				_ = gulu.File.CopyDir(path.Join(outputAudioEffect, numKey, key), path.Join(finalAudioEffect, info.WuJiangCn, info.Name))
			}
			// 复制音频技能音频等信息
			subFiles, _ = ioutil.ReadDir(path.Join(outputAudioSkill, numKey, key))
			for _, innerF := range subFiles {
				p := path.Join(outputAudioSkill, numKey, key, innerF.Name())

				skillName := strings.TrimPrefix(innerF.Name(), "SKILL_"+numKey+"_")
				if len(skillName) > 0 && !(skillName[0] >= '0' && skillName[0] <= '9') {
					skillName = strings.TrimSuffix(skillName, ".mp3")
					if strings.HasSuffix(skillName, "_1") || strings.HasSuffix(skillName, "_2") || strings.HasSuffix(skillName, "_3") {
						skillName = skillName[:len(skillName)-2] + skillName[len(skillName)-1:]
					}
					copyFile(p, path.Join(finalAudioSkill, info.WuJiangEn, info.Name, skillName+".mp3"))
				}
			}

			// js文件需要的map
			toAddInfo := map[string]interface{}{
				"x":     config.Daiji.X,
				"y":     config.Daiji.Y,
				"scale": config.Daiji.Scale,
				"name":  fmt.Sprintf("%s/%s/XingXiang", info.WuJiangCn, info.Name),
				"beijing": map[string]interface{}{
					"x":     config.Beijing.X,
					"y":     config.Beijing.Y,
					"scale": config.Beijing.Scale,
					"name":  fmt.Sprintf("%s/%s/BeiJing", info.WuJiangCn, info.Name),
				},
			}
			if info.IsReal {
				toAddInfo["gongji"] = map[string]interface{}{
					"x": config.Gongji.X,
					"y": config.Gongji.Y,
				}
			}

			if playerInfo, ok := dynamicSkin[info.WuJiangEn]; ok {
				if _, exist := playerInfo[info.Name]; !exist {
					if toOutputMap[info.WuJiangEn] == nil {
						toOutputMap[info.WuJiangEn] = make(map[string]interface{})
					}
					toOutputMap[info.WuJiangEn][info.Name] = toAddInfo
				}
			} else {
				if toOutputMap[info.WuJiangEn] == nil {
					toOutputMap[info.WuJiangEn] = make(map[string]interface{})
				}
				toOutputMap[info.WuJiangEn][info.Name] = toAddInfo
			}
			for _, pre := range info.Prefix {
				if pre == "" {
					continue
				}
				newEnName := pre + "_" + info.WuJiangEn
				if _, exist := extend[newEnName]; !exist {
					toOutputExtendMap[newEnName] = fmt.Sprintf("decadeUI.dynamicSkin.%s", info.WuJiangEn)
				}
			}
		}
	}

	// 生成js文件, 准备导入
	filePath := "dynamicSkin_new.js"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
		return
	}
	// 及时关闭file句柄
	defer file.Close()
	// 写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	write.WriteString("var dynamicSkin = {\n")
	for enName, skins := range toOutputMap {
		write.WriteString("    ")
		write.WriteString(enName)
		write.WriteString(": {\n")
		for skinName, value := range skins {
			write.WriteString("        ")
			write.WriteString(skinName)
			write.WriteString(": {\n")
			newVal, _ := value.(map[string]interface{})
			for k, v := range newVal {
				write.WriteString("            ")
				write.WriteString(k)
				write.WriteString(": ")

				switch t := v.(type) {
				case []float64:
					write.WriteString(fmt.Sprintf("[%v, %v],\n", t[0], t[1]))
				case string:
					write.WriteString(fmt.Sprintf("\"%s\",\n", t))
				case map[string]interface{}:
					write.WriteString("{\n")
					for kk, vv := range t {
						write.WriteString("                ")
						write.WriteString(kk)
						write.WriteString(": ")
						switch tt := vv.(type) {
						case []float64:
							write.WriteString(fmt.Sprintf("[%v, %v],\n", tt[0], tt[1]))
						case string:
							write.WriteString(fmt.Sprintf("\"%s\",\n", tt))
						default:
							write.WriteString(fmt.Sprintf("%v,\n", vv))
						}
					}
					write.WriteString("            },\n")
				default:
					write.WriteString(fmt.Sprintf("%v,\n", v))
				}
			}
			write.WriteString("        },\n")
		}
		write.WriteString("    },\n")
	}
	write.WriteString("};\n\n")

	write.WriteString("var extend = {\n")
	for enName, val := range toOutputExtendMap {
		write.WriteString(fmt.Sprintf("    %s: %s,\n", enName, val))
	}
	write.WriteString("};\n\n")
	write.Flush()
}

func parseSkinJs() (map[string]map[string]interface{}, map[string]string) {
	dsPath := "dynamicSkin.js"
	// 执行js, 获取已经填写的参数, 防止重复生成
	data, err := ioutil.ReadFile(dsPath)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	// 参考 https://www.ohyee.cc/post/note_go_js
	vm := goja.New()
	_, err = vm.RunString(string(data))
	if err != nil {
		fmt.Println("解析动皮参数js出错", err)
		return nil, nil
	}
	dynamicSkin := make(map[string]map[string]interface{})
	if err = vm.ExportTo(vm.Get("dynamicSkin"), &dynamicSkin); err != nil {
		fmt.Println("解析dynamicSkin出错", err)
	}
	extend := make(map[string]string)
	if err = vm.ExportTo(vm.Get("extend"), &extend); err != nil {
		fmt.Println("解析extend出错", err)
	}
	fmt.Println(dynamicSkin, extend)
	return dynamicSkin, extend
}

// 生成所有的动皮参数存放入excel
func createExcel(skinIds []string) {

	// 首先读取已经存在的excel数据, 防止重复填写
	excelName := config.ExcelName
	toReadFileName := excelName + ".xlsx"

	skinInfo := make(map[string][]string)
	if gulu.File.IsExist(toReadFileName) {
		output, _ := xlsx.FileToSlice(toReadFileName)
		// 读取第一个sheet的数据
		for rowIndex, row := range output[0] {
			if rowIndex == 0 {
				continue
			}
			if len(row) == 0 {
				continue
			}
			skinInfo[row[0]] = row
		}
	}

	// 创建excel表格, 读取已经填写的数据, 增加新的数据
	sort.SliceStable(skinIds, func(i, j int) bool {
		x, _ := strconv.ParseInt(skinIds[i], 10, 64)
		y, _ := strconv.ParseInt(skinIds[j], 10, 64)
		return x < y
	})
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("映射表")
	if err != nil {
		fmt.Println("创建sheet错误", err)
		return
	}
	row := sheet.AddRow()
	for _, h := range excelHeader {
		cell := row.AddCell()
		cell.Value = h
	}

	// 填写数据
	for _, skinId := range skinIds {
		row := sheet.AddRow()
		if r, ok := skinInfo[skinId]; ok {
			for _, v := range r {
				cell := row.AddCell()
				cell.Value = v
			}
		} else {
			cell := row.AddCell()
			cell.Value = skinId
		}
	}

	err = file.Save(fmt.Sprintf("%s_新.xlsx", excelName))
	if err != nil {
		fmt.Println("生成excel错误", err)
	}
}
