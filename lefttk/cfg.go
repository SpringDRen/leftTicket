package lefttk

import (
	"github.com/SpringDRen/easylog"
	"gopkg.in/ini.v1"
)

//读取ini文件中的某个section，返回map键值对，忽略行内注释
func GetSectionCfg(file string, sectionName string) map[string]string {
	result := make(map[string]string)
	//忽略行内注释
	cfg, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, file)
	if err != nil {
		easylog.Error("can't load cfg", file, err)
		return nil
	}
	section, err := cfg.GetSection(sectionName)
	if err != nil {
		easylog.Error("can't read section", sectionName, err)
		return nil
	}
	for _, k := range section.Keys() {
		result[k.Name()] = k.String()
	}
	return result
}
