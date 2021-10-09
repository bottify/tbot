module tbot

go 1.16

require (
	github.com/sirupsen/logrus v1.8.1
	github.com/wdvxdr1123/ZeroBot v1.2.4
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/driver/sqlite v1.1.6
	gorm.io/gorm v1.21.16
)

replace github.com/wdvxdr1123/ZeroBot => ./lib3rd/ZeroBot
