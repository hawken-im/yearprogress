---
title: 用 Go 语言写焦虑发生器并发布到 Rum 上·第三篇
date: 2022-02-20 15:47:53
tags:
---

初学 Go 语言是在去年 11 月 20 日。到今年 2022 年的 2 月 3 日，我写了一个小 bot 连续运行成功，发布到了 github 上开源。花了两个多月时间。我感觉效率还是不错的。
于是我这就来记录这段学习经历：
目的一是让同样正在跨编程入门这道薛定谔之忽高忽低门槛的小白同学一些参考，更顺利的入门编程；
目的二是回顾并巩固自己过去的学习，为下一步继续学习打好基础；
目的三是让更多的人能够对 Rum 这个新的东西感兴趣。
希望能成功达成目的：

## 变得更酷之定时任务
到现在，我们已经可以给 Rum 发送进度条了。当时我本人，在写到这一步的时候已经偷偷自调闹钟，在整点向 Rum 的“去中心微博”那个种子网络发了两三次了。
但是自己调闹钟来发内容，这根本不 bot！根本都不酷！
这一步我们要加入定时任务这个 new feature。

经过多番研究
>（踩了不少坑，关于 cron job 的繁琐本人在 Rum 上的朋友圈里吐槽了一番）

还是确定了引用这个包：
``` Go
import{
	cron "github.com/robfig/cron/v3"//前面的cron是自己取的包的别名
}
```

选定了这个包，就通过阅读官方文档，找到新建计划任务的方法：
``` Go
c := cron.New(cron.WithLocation(time.UTC))
```

cron.New() 可以新建一个实例，里面的参数是指定时区，我们这里就用 UTC 时间。
接着应用实例 c 的方法 AddFunc 就好，如下：
``` Go
c.AddFunc([计划任务时间], func() {[要执行的函数]})
```
\[计划任务时间\]采用的是 linux 著名的 crontab 计划任务常用的格式，这个具体怎么弄除了自己去查询，还有个神奇的网站帮我们去做计划任务的时间：
https://crontab.guru/

AddFunc 方法可以使用多次，也就是添加多个计划，之后再用：
``` Go
c.Start()
```
才可以开始计划好的任务。

一开始我是每 24 小时发送一次，感觉有些打扰别人的时间线，经过一番折腾，最终把发送进度条的频率设置为每 1% 发送一次。

这个的算法是这样思考的：
在接近一个整数百分比的时间前，每分钟一个百分比，算出 15 分钟 15 个百分比，每个百分比作一次被减数，去减下一个整数百分比的时间。直到整数百分比减去算出来的百分比小于 0.00001。那么就设定一个计划任务，在算出来的那个时间发布进度条。然后，休眠 85 个小时，因为一年的 1% 差不多是 87个小时。
如果这 15 分钟内没有一个百分比达成这个条件，则休眠 15 分钟，在下一个 15 分钟唤醒程序，再算一次。
打太多字读者也许会有点难读，我们放代码吧，代码配合注释可能更容易理解：
``` Go
func main() {
	c := cron.New(cron.WithLocation(time.UTC))
	url := "https://127.0.0.1:8002/api/v1/group/content" //Rum 定义的 api

	for {
		startTime := time.Date(2022, time.Now().UTC().Month(), time.Now().UTC().Day(), time.Now().UTC().Hour(), time.Now().Minute(), 0, 0, time.UTC) //开始时间
		for x := 0; x <= 14; x++ {                                                                                                                   //循环15次，下一个15分钟每分钟一次
			addMinutes, _ := time.ParseDuration(fmt.Sprintf("%dm", x))//每次循环，在开始时间前加x分钟
			realTimePerc := timePerc(startTime.Add(addMinutes))
			roundPerc := math.Ceil(realTimePerc*100) / 100 //计算下一个整数百分比
			differVal := roundPerc - realTimePerc          //计算差值，差值接近于零代表时间接近整数百分比了
			if differVal < 0.00001 {                       //每分钟计算一次，每分钟是一年的0.000002，因此精确到小数点后5位
				realTime := startTime.Add(addMinutes)
				nextPostTime := fmt.Sprintf("%d %d %d %d *", realTime.Minute(), realTime.Hour(), realTime.Day(), realTime.Month())
				progressBar := printBar(roundPerc)
				c.AddFunc(nextPostTime, func() { postToRum("2022 进度条", progressBar, "fe2842cb-db6b-4e8a-b007-e83e5603131c", url) }) //设置定时任务
				c.Start()                                                                                                           //开始定时任务
				fmt.Println("######## went to sleep for 85 hours ########")                                                         //休眠85个小时，因为一个百分比大概接近87个小时
				time.Sleep(85 * time.Hour)
				break
			}
		}
		fmt.Println("######## went to sleep ########") //休眠15分钟
		time.Sleep(15 * time.Minute)
		c.Stop()
		fmt.Println("############ awaken ###########") //唤醒
	}
}
```

上面的 for 循环是放在 main 里，是一个死循环，我的理想情况是这个代码能够一年都不停地运行不会出错。写这篇文章的时候已经稳定运行了 3 个百分点。
这里的变量 startTime 是每一个 15 分钟判断开始的时间。设定的是整秒数。然后嵌套一个运行 15 次的 for 循环，每次循环都会在 startTime 的基础上增加一分钟并判断这个时间和整百分点时间差多少百分比。直到相差小于 0.00001。
再嵌套的一个 for 循环会遍历我的 config 文件里的 Rum 种子网络 ID。

## 变得更酷之记录日志
最终实现自己的目标前，会出不少问题，要定位这些问题，最好的方法就是每一步都输出一个结果。
特别是我们上一步那种循环套循环，还要长期运行的代码，没有日志很难定位问题出在哪里。
于是我引入了日志。
也是老办法，查了下有哪些好用的包。试水了两三个，最后选到了这个：
``` Go
import (
	log "github.com/sirupsen/logrus"
)
```

非常不好意思的告诉读者们，我作为小白一开始连前面的 log 是别名都不懂，有了别名，以后要引用这个包就不需要写全名，而只需要写 log 这个小短词了。

阅读文档得知，将 log 输出到文件的语句如下：
``` Go
f, err := os.OpenFile("YP.log", os.O_WRONLY|os.O_CREATE, 0755) //log file
if err != nil {
	panic(err)
}
log.SetOutput(f)
```

先用 os.OpenFile 新建或打开一个文档，我这里取名叫 "YP.log"，将文档赋值给变量 f，然后调用 log.SetOutput(f) 就可以将日志统统输出到文档 YP.log 里了。
在需要记录日志的地方，调用 log.Info(\[日志内容\]) 就可以了。
于是我在主循环中加入了一些日志，好让我可以观察程序的运行：
``` Go
func main() {
	f, err := os.OpenFile("YP.log", os.O_WRONLY|os.O_CREATE, 0755) //log file
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)

	c := cron.New(cron.WithLocation(time.UTC))
	url := "https://127.0.0.1:8002/api/v1/group/content" //Rum 定义的 api

	for {
		startTime := time.Date(2022, time.Now().UTC().Month(), time.Now().UTC().Day(), time.Now().UTC().Hour(), time.Now().Minute(), 0, 0, time.UTC) //开始时间
		log.Info("startTime:", startTime)                                                                                                            //记录一下循环开始时间
		for x := 0; x <= 14; x++ {                                                                                                                   //循环15次，下一个15分钟每分钟一次
			addMinutes, _ := time.ParseDuration(fmt.Sprintf("%dm", x)) //每次循环，在开始时间前加x分钟
			log.Info("addMinutes:", addMinutes)                        //记录一下每次加的时间对不对
			realTimePerc := timePerc(startTime.Add(addMinutes))
			log.Info("realTimePerc:", realTimePerc)        //加了时间之后的百分比，记录一下这个增长过程
			roundPerc := math.Ceil(realTimePerc*100) / 100 //计算下一个整数百分比
			log.Info("roundPerc:", roundPerc)              //虽然每次都是一样的值，但还是想看看
			differVal := roundPerc - realTimePerc          //计算差值，差值接近于零代表时间接近整数百分比了
			log.Info("differVal:", differVal)              //看看差值的变化过程，越来越接近于零
			if differVal < 0.00001 {                       //每分钟计算一次，每分钟是一年的0.000002，因此精确到小数点后5位
				realTime := startTime.Add(addMinutes)
				log.Info("differVal less than 0:", differVal) //终于到整百分点了，记录一个
				nextPostTime := fmt.Sprintf("%d %d %d %d *", realTime.Minute(), realTime.Hour(), realTime.Day(), realTime.Month())
				log.Info("nextPostTime:", nextPostTime) //报告具体的整百分点发布时间
				progressBar := printBar(roundPerc)
				c.AddFunc(nextPostTime, func() { postToRum("2022 进度条", progressBar, "fe2842cb-db6b-4e8a-b007-e83e5603131c", url) }) //设置定时任务
				c.Start()
				log.Info("######## went to sleep for 85 hours ########")    //日志里也记录一下                                                                         //开始定时任务
				fmt.Println("######## went to sleep for 85 hours ########") //休眠85个小时，因为一个百分比大概接近87个小时
				time.Sleep(85 * time.Hour)
				break
			}
		}
		log.Info("######## went to sleep ########")
		fmt.Println("######## went to sleep ########") //休眠15分钟
		time.Sleep(15 * time.Minute)
		c.Stop()
		log.Info("############ awaken ###########")
		fmt.Println("############ awaken ###########") //唤醒
	}
}
```

按照惯例，我把代码的全貌展示在 github 上。作为小白本白，我经常复制粘贴了别人的代码再修修补补一下就没法成功运行了，所以我这里提供一个完整的可以执行的代码，修修补补出了问题可以回滚：

https://github.com/hawken-im/yearprogress/tree/main/Step2